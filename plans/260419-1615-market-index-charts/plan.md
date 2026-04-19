---
title: Market Index Charts + /market Page
status: pending
priority: high
created: 2026-04-19
---

# Market Index Charts + /market Page

Thêm chart VNIndex/VN30/HNXINDEX/UPCOM vào trang /market mới.
Option B: giữ KBS cho cổ phiếu thường, dùng VCI API chỉ cho index symbols.

## Scope

3 files thay đổi:
1. `go-services/internal/handlers/chart.go` — route index → VCI, stock → KBS
2. `blog-site/app/market/page.tsx` — trang mới (NEW FILE)
3. `blog-site/app/layout.tsx` — thêm nav link "Thị trường"

## Phases

| Phase | File | Status |
|-------|------|--------|
| [01](phase-01-go-chart-handler.md) | chart.go VCI routing | pending |
| [02](phase-02-market-page.md) | /market page | pending |
| [03](phase-03-nav-link.md) | nav link layout.tsx | pending |

## Key Constraints

- VCI POST: `https://trading.vietcap.com.vn/api/chart/OHLCChart/gap-chart`
- Index symbol map: VNINDEX→VNINDEX, VN30→VN30, HNXINDEX→HNXIndex, UPCOMINDEX→HNXUpcomIndex
- VCI `"t"` field: string unix seconds → convert to `YYYY-MM-DD`
- countBack = days + 60 (SMA warmup)
- Cache key/TTL unchanged: `symbol:days`, 5 min
- Tab switcher: useSearchParams → wrap in Suspense
- Top rec: `/api/v1/screener?recommendation=BUY&limit=3` + SELL&limit=3
