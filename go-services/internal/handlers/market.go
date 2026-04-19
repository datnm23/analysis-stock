package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// IndexSnapshot holds one index value + daily change.
type IndexSnapshot struct {
	Symbol    string  `json:"symbol"`
	Close     float64 `json:"close"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"change_pct"`
	Volume    int64   `json:"volume"`
}

// indexList is the 4 symbols shown in the index bar.
var indexList = []string{"VNINDEX", "VN30", "HNXINDEX", "UPCOMINDEX"}

// in-memory cache for indices (avoids Redis round-trip for 4 small values)
var (
	indicesCacheMu sync.RWMutex
	indicesCache   []IndexSnapshot
	indicesExpiry  time.Time
	indicesTTL     = 2 * time.Minute
)

// MarketIndices returns last-close + daily change for the 4 main indices.
// GET /api/v1/market/indices
func MarketIndices() gin.HandlerFunc {
	return func(c *gin.Context) {
		indicesCacheMu.RLock()
		if time.Now().Before(indicesExpiry) && indicesCache != nil {
			cached := indicesCache
			indicesCacheMu.RUnlock()
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, gin.H{"indices": cached})
			return
		}
		indicesCacheMu.RUnlock()

		ctx, cancel := context.WithTimeout(c.Request.Context(), 12*time.Second)
		defer cancel()

		type result struct {
			snap IndexSnapshot
			err  error
		}
		results := make([]result, len(indexList))
		var wg sync.WaitGroup

		for i, sym := range indexList {
			wg.Add(1)
			go func(idx int, symbol string) {
				defer wg.Done()
				bars, err := fetchVCIOHLCV(ctx, symbol, 5)
				if err != nil || len(bars) < 2 {
					results[idx] = result{err: err}
					return
				}
				cur := bars[len(bars)-1]
				prev := bars[len(bars)-2]
				chg := cur.Close - prev.Close
				chgPct := 0.0
				if prev.Close != 0 {
					chgPct = chg / prev.Close * 100
				}
				results[idx] = result{snap: IndexSnapshot{
					Symbol:    symbol,
					Close:     cur.Close,
					Change:    chg,
					ChangePct: chgPct,
					Volume:    cur.Volume,
				}}
			}(i, sym)
		}
		wg.Wait()

		snapshots := make([]IndexSnapshot, 0, len(indexList))
		for _, r := range results {
			if r.err == nil {
				snapshots = append(snapshots, r.snap)
			}
		}

		indicesCacheMu.Lock()
		indicesCache = snapshots
		indicesExpiry = time.Now().Add(indicesTTL)
		indicesCacheMu.Unlock()

		c.Header("X-Cache", "MISS")
		c.JSON(http.StatusOK, gin.H{"indices": snapshots})
	}
}

// MarketBoard proxies to crawl-agent price board with Redis caching.
// GET /api/v1/market/board?exchange=ALL
func MarketBoard(marketURL string, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		exchange := strings.ToUpper(c.DefaultQuery("exchange", "ALL"))
		cacheKey := fmt.Sprintf("market:board:%s", exchange)

		if rdb != nil {
			if cached, err := rdb.Get(c.Request.Context(), cacheKey).Bytes(); err == nil {
				c.Header("X-Cache", "HIT")
				c.Data(http.StatusOK, "application/json", cached)
				return
			}
		}

		upstream := fmt.Sprintf("%s/market/board?exchange=%s", marketURL, exchange)
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, upstream, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
			return
		}

		client := &http.Client{Timeout: 60 * time.Second} // vnstock can be slow
		resp, err := client.Do(req)
		if err != nil {
			// Graceful degradation: return empty board instead of 502
			c.Header("X-Cache", "MISS")
			c.JSON(http.StatusOK, gin.H{"symbols": []struct{}{}, "total": 0, "error": "price board service unavailable"})
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 MB cap
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read upstream response"})
			return
		}

		if rdb != nil && resp.StatusCode == http.StatusOK {
			rdb.Set(c.Request.Context(), cacheKey, body, 5*time.Minute)
		}

		c.Header("X-Cache", "MISS")
		c.Data(resp.StatusCode, "application/json", body)
	}
}
