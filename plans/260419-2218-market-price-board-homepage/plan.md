---
title: Market Price Board + Homepage Redesign
status: completed
priority: high
created: 2026-04-19
completed: 2026-04-20
blockedBy: []
blocks: []
progress: 100%
---

# Market Price Board + Homepage Redesign

Biến trang chủ thành market overview như VietStock Finance:
- Bảng giá tất cả mã (700+) via vnstock/TCBS
- Nhóm ngành tĩnh (~200 mã phổ biến)
- Index bar realtime (VCI chart API)
- AI layer: highlight mã có khuyến nghị từ screener

## Scope

**4 services thay đổi:**

| Service | Files |
|---------|-------|
| crawl-agent | `requirements.txt`, `app/routers/market.py` (NEW), `app/main.py` |
| go-services | `internal/handlers/market.go` (NEW), `internal/config/config.go`, `cmd/api-gateway/main.go` |
| blog-site | `lib/sector-mapping.ts` (NEW), `components/market-index-bar.tsx` (NEW), `components/market-board-table.tsx` (NEW), `app/page.tsx` |

## Phases

| Phase | Mô tả | Status |
|-------|-------|--------|
| [01](phase-01-crawl-agent-price-board.md) | vnstock price board endpoint | completed |
| [02](phase-02-go-market-handlers.md) | Go market indices + board proxy | completed |
| [03](phase-03-sector-mapping.md) | Sector mapping static data | completed |
| [04](phase-04-homepage-redesign.md) | Homepage redesign | completed |

## Data Flow

```
Next.js → GET /api/v1/market/indices → Go → VCI chart (last 2 bars per index)
Next.js → GET /api/v1/market/board  → Go → Redis HIT? → return
                                           → Redis MISS → crawl-agent:8085/market/board
                                                         → vnstock TCBS price_board()
                                                         → cache Redis 5min
```

## Key Constraints

- vnstock v3.4.0 source='TCBS' cho price board (miễn phí, no auth)
- crawl-agent port: 8085
- Go: thêm `MARKET_SERVICE_URL` env (default `http://localhost:8085`)
- Redis key: `market:board:{exchange}` TTL 5min; `market:indices` TTL 2min
- Sector mapping: static TS file, ~200 mã → 12 ngành
- Homepage: chỉ sửa `/` route, không đụng `/articles`, `/market`, `/screener`
