package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// OHLCVBar is one candlestick day for the chart API.
type OHLCVBar struct {
	Time   string  `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

type kbsBar struct {
	T string  `json:"t"`
	O float64 `json:"o"`
	H float64 `json:"h"`
	L float64 `json:"l"`
	C float64 `json:"c"`
	V int64   `json:"v"`
}

type kbsResp struct {
	DataDay []kbsBar `json:"data_day"`
}

// VCI (Vietcap) types — used for index symbols only.
type vciRequest struct {
	TimeFrame string   `json:"timeFrame"`
	Symbols   []string `json:"symbols"`
	To        int64    `json:"to"`
	CountBack int      `json:"countBack"`
}

type vciColumn struct {
	Symbol string    `json:"symbol"`
	O      []float64 `json:"o"`
	H      []float64 `json:"h"`
	L      []float64 `json:"l"`
	C      []float64 `json:"c"`
	V      []float64 `json:"v"`
	T      []string  `json:"t"`
}

const (
	kbsBase        = "https://kbbuddywts.kbsec.com.vn/iis-server/investment/stocks"
	vciBase        = "https://trading.vietcap.com.vn/api/chart/OHLCChart/gap-chart"
	kbsBodyLimitMB = 5 << 20 // 5 MB guard
	chartCacheTTL  = 5 * time.Minute
)

// indexSymbolMap maps route param → VCI symbol name.
var indexSymbolMap = map[string]string{
	"VNINDEX":    "VNINDEX",
	"VN30":       "VN30",
	"HNXINDEX":   "HNXIndex",
	"UPCOMINDEX": "HNXUpcomIndex",
	"UPCOM":      "HNXUpcomIndex",
}

func isIndex(symbol string) bool {
	_, ok := indexSymbolMap[symbol]
	return ok
}

// symbolRe guards against path-traversal and injection via the :symbol route param.
var symbolRe = regexp.MustCompile(`^[A-Z0-9]{2,10}$`)

// ── In-memory cache (no Redis dependency needed for public chart data) ──────

type cachedChart struct {
	bars      []OHLCVBar
	expiresAt time.Time
}

var (
	chartCache   = make(map[string]*cachedChart)
	chartCacheMu sync.RWMutex
)

func getCachedChart(key string) ([]OHLCVBar, bool) {
	chartCacheMu.RLock()
	defer chartCacheMu.RUnlock()
	v, ok := chartCache[key]
	if ok && time.Now().Before(v.expiresAt) {
		return v.bars, true
	}
	return nil, false
}

func setCachedChart(key string, bars []OHLCVBar) {
	chartCacheMu.Lock()
	defer chartCacheMu.Unlock()
	chartCache[key] = &cachedChart{bars: bars, expiresAt: time.Now().Add(chartCacheTTL)}
}

// ── KBS fetch ─────────────────────────────────────────────────────────────────

func fetchKBSOHLCV(ctx context.Context, symbol string, days int) ([]OHLCVBar, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)
	url := fmt.Sprintf(
		"%s/%s/data_day?sdate=%s&edate=%s",
		kbsBase, symbol,
		start.Format("02-01-2006"),
		end.Format("02-01-2006"),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; VNStock-Hybrid/1.0)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("KBS request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("KBS returned %d", resp.StatusCode)
	}

	var raw kbsResp
	if err := json.NewDecoder(io.LimitReader(resp.Body, kbsBodyLimitMB)).Decode(&raw); err != nil {
		return nil, fmt.Errorf("KBS decode: %w", err)
	}

	// Deduplicate and parse dates.
	seen := make(map[string]bool, len(raw.DataDay))
	bars := make([]OHLCVBar, 0, len(raw.DataDay))
	for _, b := range raw.DataDay {
		dateStr := b.T
		if len(dateStr) >= 10 {
			dateStr = dateStr[:10] // keep YYYY-MM-DD
		}
		if seen[dateStr] {
			continue // skip duplicate dates
		}
		seen[dateStr] = true
		bars = append(bars, OHLCVBar{
			Time:   dateStr,
			Open:   b.O,
			High:   b.H,
			Low:    b.L,
			Close:  b.C,
			Volume: b.V,
		})
	}

	// KBS normally returns newest-first; reverse only when that's actually the case.
	if len(bars) > 1 && bars[0].Time > bars[1].Time {
		for i, j := 0, len(bars)-1; i < j; i, j = i+1, j-1 {
			bars[i], bars[j] = bars[j], bars[i]
		}
	}
	return bars, nil
}

// fetchVCIOHLCV fetches OHLCV data from VCI (Vietcap) for index symbols.
func fetchVCIOHLCV(ctx context.Context, symbol string, days int) ([]OHLCVBar, error) {
	vciSymbol := indexSymbolMap[symbol]
	body, _ := json.Marshal(vciRequest{
		TimeFrame: "ONE_DAY",
		Symbols:   []string{vciSymbol},
		To:        time.Now().Unix(),
		CountBack: days + 60, // extra bars for SMA warmup
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, vciBase, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; VNStock-Hybrid/1.0)")
	req.Header.Set("Referer", "https://trading.vietcap.com.vn/")
	req.Header.Set("Origin", "https://trading.vietcap.com.vn")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("VCI request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VCI returned %d", resp.StatusCode)
	}

	var cols []vciColumn
	if err := json.NewDecoder(io.LimitReader(resp.Body, kbsBodyLimitMB)).Decode(&cols); err != nil {
		return nil, fmt.Errorf("VCI decode: %w", err)
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("VCI returned empty data for %s", symbol)
	}

	col := cols[0]
	n := len(col.T)
	seen := make(map[string]bool, n)
	bars := make([]OHLCVBar, 0, n)

	for i := 0; i < n; i++ {
		if i >= len(col.O) || i >= len(col.H) || i >= len(col.L) || i >= len(col.C) {
			break
		}
		// Skip non-trading day placeholders
		if col.O[i] == 0 && col.H[i] == 0 {
			continue
		}
		ts, err := strconv.ParseInt(col.T[i], 10, 64)
		if err != nil {
			continue
		}
		dateStr := time.Unix(ts, 0).UTC().Format("2006-01-02")
		if seen[dateStr] {
			continue
		}
		seen[dateStr] = true

		var vol int64
		if i < len(col.V) {
			vol = int64(col.V[i])
		}
		bars = append(bars, OHLCVBar{
			Time:   dateStr,
			Open:   col.O[i],
			High:   col.H[i],
			Low:    col.L[i],
			Close:  col.C[i],
			Volume: vol,
		})
	}

	// VCI returns oldest-first; ensure ascending order
	if len(bars) > 1 && bars[0].Time > bars[1].Time {
		for i, j := 0, len(bars)-1; i < j; i, j = i+1, j-1 {
			bars[i], bars[j] = bars[j], bars[i]
		}
	}
	return bars, nil
}

// ChartData returns OHLCV bars for the given symbol.
// GET /api/v1/chart/:symbol?days=90
func ChartData() gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Param("symbol")
		if !symbolRe.MatchString(symbol) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid symbol"})
			return
		}

		days := 120
		if d := c.Query("days"); d != "" {
			if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 365 {
				days = v
			}
		}

		cacheKey := fmt.Sprintf("%s:%d", symbol, days)
		if cached, ok := getCachedChart(cacheKey); ok {
			c.Header("Cache-Control", "public, max-age=300")
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, gin.H{"symbol": symbol, "bars": cached})
			return
		}

		var bars []OHLCVBar
		var err error
		if isIndex(symbol) {
			bars, err = fetchVCIOHLCV(c.Request.Context(), symbol, days)
		} else {
			bars, err = fetchKBSOHLCV(c.Request.Context(), symbol, days)
		}
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		setCachedChart(cacheKey, bars)

		c.Header("Cache-Control", "public, max-age=300")
		c.Header("X-Cache", "MISS")
		c.JSON(http.StatusOK, gin.H{"symbol": symbol, "bars": bars})
	}
}
