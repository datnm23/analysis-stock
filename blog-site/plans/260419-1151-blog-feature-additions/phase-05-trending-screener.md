# Phase 5: Trending Widget + Screener

**Priority:** Medium  
**Effort:** ~1.5 ngày  
**Status:** pending  
**Depends on:** Phase 1 (Go handler pattern), Phase 2 (filter pattern)

## Overview

Hai widget độc lập:
1. **Trending Widget** — Top 5 mã được phân tích nhiều nhất gần đây (homepage + sidebar)
2. **Screener** — Bảng lọc nhanh: sector, recommendation, lọc theo RSI/confidence

## Related Code Files

**Tạo mới:**
- `blog-site/components/trending-stocks.tsx`
- `blog-site/components/screener-table.tsx`
- `blog-site/app/screener/page.tsx`
- `go-services/internal/handlers/trending.go`

**Sửa:**
- `blog-site/app/page.tsx` — thêm TrendingStocks widget
- `go-services/cmd/api-gateway/main.go` — register routes mới

## Part A: Trending Widget

### Backend

1. Tạo `go-services/internal/handlers/trending.go`:
```go
// GET /api/v1/trending?days=7&limit=5
// Query: SELECT symbol, COUNT(*) as article_count,
//        MAX(published_at) as last_published,
//        MAX(forecast_data) as latest_forecast  -- lấy từ bài mới nhất
//        FROM articles WHERE status='published'
//          AND published_at >= NOW() - INTERVAL '$days days'
//        GROUP BY symbol ORDER BY article_count DESC LIMIT $limit
// Cache: 10 phút (in-memory, same pattern as chart cache)
```

2. Response:
```json
{
  "trending": [
    { "symbol": "VNM", "count": 5, "recommendation": "BUY", "confidence": 0.87 },
    ...
  ]
}
```

### Frontend

3. Tạo `blog-site/components/trending-stocks.tsx` (Server component):
```tsx
// Fetch /api/v1/trending?days=7&limit=5
// Render bảng compact:
//   Rank | Symbol | Số bài | Khuyến nghị | Link →
// Style: brutalist table (border-3, alternating bg)
// Fallback: skeleton nếu loading
```

4. Thêm vào `app/page.tsx` sau hero section:
```tsx
<section className="mb-12">
  <h2 className="text-2xl font-black uppercase mb-4">Hot trong 7 ngày</h2>
  <Suspense fallback={<TrendingSkeleton />}>
    <TrendingStocks />
  </Suspense>
</section>
```

## Part B: Screener

### Backend

5. Tạo `GET /api/v1/screener`:
```go
// Params: sector, recommendation, min_confidence, sort_by (date|confidence|symbol)
// Query: SELECT DISTINCT ON (symbol) symbol, forecast_data, published_at
//        FROM articles WHERE status='published'
//        [+ filters] ORDER BY symbol, published_at DESC
// Returns: 1 row per symbol (latest analysis)
```

6. Response:
```json
{
  "results": [
    {
      "symbol": "VNM", "recommendation": "BUY", "confidence": 0.87,
      "technical_score": 7.2, "sentiment_score": 6.8,
      "last_analyzed": "2026-04-19", "article_slug": "vnm-20260419-..."
    }
  ],
  "total": 42
}
```

### Frontend

7. Tạo `blog-site/app/screener/page.tsx`:
```tsx
// Client component (interactive filters)
// Layout:
//   Filter bar: [TẤT CẢ | MUA | GIỮ | BÁN] + [Confidence > 70%] toggle
//   Table: Symbol | Khuyến nghị | Độ tin cậy | Kỹ thuật | Sentiment | Cập nhật | [Xem →]
//   Sort: click header để sort
```

8. Tạo `blog-site/components/screener-table.tsx`:
```tsx
// Props: results[], sortBy, onSort
// Sortable columns: confidence, technical_score, last_analyzed
// Row click → link đến article detail
// Recommendation: colored badge (MUA=green, BÁN=red, GIỮ=yellow)
```

9. Thêm link "Screener" vào navbar trong `layout.tsx`:
```tsx
<a href="/screener" className="...">Screener</a>
```

## UI Mockup — Screener

```
┌──────────────────────────────────────────────────────────────────┐
│  SCREENER — 42 mã                                                │
│  [TẤT CẢ] [■ MUA (18)] [GIỮ (15)] [BÁN (9)]  [Tin cậy > 70%] │
├────────┬──────────┬───────────┬──────────┬───────────┬──────────┤
│ Mã  ▲  │ Khuyến  │ Tin cậy ↓ │ KT Score │ Sentiment │ Cập nhật │
├────────┼──────────┼───────────┼──────────┼───────────┼──────────┤
│ VNM   │ 🟢 MUA  │ ██████ 87%│  7.2/10  │  6.8/10   │ hôm nay  │
│ HPG   │ 🟢 MUA  │ █████  82%│  8.1/10  │  5.9/10   │ hôm nay  │
│ VCB   │ 🟡 GIỮ  │ ████   74%│  6.3/10  │  7.1/10   │ hôm qua  │
└────────┴──────────┴───────────┴──────────┴───────────┴──────────┘
```

## Sector Filter (future)

Sector mapping cho symbols VN — hardcode trong frontend:
```ts
const SECTOR_MAP: Record<string, string> = {
  VNM: "Tiêu dùng", HPG: "Thép", VCB: "Ngân hàng",
  VIC: "Bất động sản", FPT: "Công nghệ", ...
}
```
Phase này có thể bỏ qua sector filter nếu data không có sẵn.

## Success Criteria

- [ ] Trending widget hiển thị top 5 trên homepage
- [ ] `/screener` filter BUY/SELL/HOLD hoạt động
- [ ] Sort theo confidence/score hoạt động
- [ ] Row click → article detail
- [ ] Cache trending data 10 phút (không gọi DB mỗi request)

## Risk Assessment

| Rủi ro | Mitigation |
|--------|------------|
| Go chưa có DB connection | Check Phase 1 risk; screener hoàn toàn depends on DB access |
| forecast_data JSON parsing trong SQL | Dùng `->>` operator của PostgreSQL JSONB |
| Screener chậm nếu nhiều symbols | Add index trên `(symbol, published_at DESC, status)` |
