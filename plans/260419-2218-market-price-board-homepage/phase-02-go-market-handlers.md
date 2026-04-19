---
title: Phase 02 – Go market handlers (indices + board proxy)
status: completed
progress: 100%
completed: 2026-04-20
---

# Phase 02 – Go: market indices + price board proxy

## Overview

Thêm 2 endpoint vào Go API Gateway:
- `GET /api/v1/market/indices` — 4 index values từ VCI (tái dùng fetchVCIOHLCV)
- `GET /api/v1/market/board` — proxy + Redis cache → crawl-agent price board

## Files

- **Tạo mới**: `go-services/internal/handlers/market.go`
- **Sửa**: `go-services/internal/config/config.go` — thêm `MarketURL`
- **Sửa**: `go-services/cmd/api-gateway/main.go` — đăng ký routes

## Implementation Steps

### 1. `internal/config/config.go`

Thêm vào `ServicesConfig`:
```go
MarketURL string  // crawl-agent price board service
```

Thêm vào `Load()` → `Services`:
```go
MarketURL: getEnv("MARKET_SERVICE_URL", "http://localhost:8085"),
```

### 2. Tạo `internal/handlers/market.go`

```go
package handlers

import (
	"context"
	"encoding/json"
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

var (
	indicesCacheMu sync.RWMutex
	indicesCache   []IndexSnapshot
	indicesExpiry  time.Time
	indicesTTL     = 2 * time.Minute
)

// indexList is the 4 symbols shown in the index bar.
var indexList = []string{"VNINDEX", "VN30", "HNXINDEX", "UPCOMINDEX"}

// MarketIndices returns last-close + daily change for the 4 main indices.
// GET /api/v1/market/indices
func MarketIndices() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In-memory cache (no Redis needed for 4 small payloads)
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

		// Fetch all 4 indices concurrently
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

		// Update in-memory cache
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

		// Redis cache check
		if rdb != nil {
			cached, err := rdb.Get(c.Request.Context(), cacheKey).Bytes()
			if err == nil {
				c.Header("X-Cache", "HIT")
				c.Header("Content-Type", "application/json")
				c.Data(http.StatusOK, "application/json", cached)
				return
			}
		}

		// Forward to crawl-agent
		upstream := fmt.Sprintf("%s/market/board?exchange=%s", marketURL, exchange)
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, upstream, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
			return
		}

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "price board service unavailable", "symbols": []struct{}{}})
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 MB
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read response"})
			return
		}

		// Cache in Redis on success
		if rdb != nil && resp.StatusCode == http.StatusOK {
			rdb.Set(c.Request.Context(), cacheKey, body, 5*time.Minute)
		}

		c.Header("X-Cache", "MISS")
		c.Data(resp.StatusCode, "application/json", body)
	}
}
```

### 3. `cmd/api-gateway/main.go`

Trong phần "Initialize services", thêm:
```go
// Market service
marketURL := cfg.Services.MarketURL
```

Trong phần routes v1:
```go
v1.GET("/market/indices", handlers.MarketIndices())
v1.GET("/market/board",   handlers.MarketBoard(marketURL, infra.Redis))
```

## Note: import redis

`go-services` đã có redis dependency. `MarketBoard` nhận `*redis.Client` — nếu `infra.Redis == nil` thì skip cache, vẫn proxy được.

## Success Criteria

```bash
curl http://localhost:8080/api/v1/market/indices
# {"indices":[{"symbol":"VNINDEX","close":1817,"change":5.2,"change_pct":0.29,...}]}

curl "http://localhost:8080/api/v1/market/board?exchange=HSX" | python3 -c "
import json,sys; d=json.load(sys.stdin); print('total:', d.get('total'))
"
# total: 400+
```

- `go build ./...` clean
- X-Cache: HIT on second request
