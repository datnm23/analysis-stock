package handlers

import (
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

const (
	kbsBase         = "https://kbbuddywts.kbsec.com.vn/iis-server/investment/stocks"
	kbsBodyLimitMB  = 5 << 20 // 5 MB guard against unexpectedly large responses
	chartCacheTTL   = 5 * time.Minute
)

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
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Cache-Control", "public, max-age=300")
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, gin.H{"symbol": symbol, "bars": cached})
			return
		}

		bars, err := fetchKBSOHLCV(c.Request.Context(), symbol, days)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		setCachedChart(cacheKey, bars)

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Cache-Control", "public, max-age=300")
		c.Header("X-Cache", "MISS")
		c.JSON(http.StatusOK, gin.H{"symbol": symbol, "bars": bars})
	}
}
