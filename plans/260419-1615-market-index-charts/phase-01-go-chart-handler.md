---
title: Phase 01 – Go chart.go VCI routing
status: pending
file: go-services/internal/handlers/chart.go
---

# Phase 01 – Go chart.go: thêm VCI fetch + index routing

## Overview

Thêm VCI API fetch cho 4 index symbols, giữ nguyên KBS cho tất cả cổ phiếu thường.
ChartData handler chỉ đổi routing logic, cache/response format không đổi.

## VCI API

```
POST https://trading.vietcap.com.vn/api/chart/OHLCChart/gap-chart
Headers:
  Content-Type: application/json
  User-Agent: Mozilla/5.0 (compatible; VNStock-Hybrid/1.0)
  Referer: https://trading.vietcap.com.vn/
  Origin: https://trading.vietcap.com.vn

Body:
  {"timeFrame":"ONE_DAY","symbols":["VNINDEX"],"to":<unix_ts>,"countBack":<N>}

Response (columnar, array of 1 element):
  [{"symbol":"VNINDEX","o":[...],"h":[...],"l":[...],"c":[...],"v":[...],"t":["1744588800",...],...}]
```

## Symbol Mapping

```go
var indexSymbolMap = map[string]string{
    "VNINDEX":    "VNINDEX",
    "VN30":       "VN30",
    "HNXINDEX":   "HNXIndex",
    "UPCOMINDEX": "HNXUpcomIndex",
    "UPCOM":      "HNXUpcomIndex",
}
```

## Implementation Steps

1. Thêm constants + types mới sau block `kbsResp`:

```go
const vciBase = "https://trading.vietcap.com.vn/api/chart/OHLCChart/gap-chart"

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
```

2. Thêm `indexSymbolMap` và helper `isIndex()`:

```go
var indexSymbolMap = map[string]string{
    "VNINDEX": "VNINDEX", "VN30": "VN30",
    "HNXINDEX": "HNXIndex", "UPCOMINDEX": "HNXUpcomIndex", "UPCOM": "HNXUpcomIndex",
}

func isIndex(symbol string) bool {
    _, ok := indexSymbolMap[symbol]
    return ok
}
```

3. Thêm `fetchVCIOHLCV()` sau `fetchKBSOHLCV()`:
   - POST body: `timeFrame=ONE_DAY`, `symbols=[vciSymbol]`, `to=time.Now().Unix()`, `countBack=days+60`
   - Parse `t[i]` (string unix seconds) → `time.Unix(ts, 0).Format("2006-01-02")`
   - Skip bars nếu `O[i]==0 && H[i]==0` (VCI returns zeros for non-trading days)
   - Deduplicate by date (same logic as KBS)
   - VCI trả về oldest-first → không cần reverse

4. Sửa `ChartData()` handler — thêm routing:
```go
var bars []OHLCVBar
var err error
if isIndex(symbol) {
    bars, err = fetchVCIOHLCV(c.Request.Context(), symbol, days)
} else {
    bars, err = fetchKBSOHLCV(c.Request.Context(), symbol, days)
}
```

5. Xóa `c.Header("Access-Control-Allow-Origin", "*")` khỏi ChartData (đã có CORS middleware toàn cục).

## Success Criteria

- `GET /api/v1/chart/VNINDEX?days=90` → 200, bars với time YYYY-MM-DD
- `GET /api/v1/chart/VNM?days=90` → vẫn dùng KBS, không đổi
- Go compiles: `cd go-services && go build ./...`
