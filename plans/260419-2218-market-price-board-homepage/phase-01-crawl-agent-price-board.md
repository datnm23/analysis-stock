---
title: Phase 01 – crawl-agent vnstock price board endpoint
status: completed
progress: 100%
completed: 2026-04-20
---

# Phase 01 – crawl-agent: vnstock price board endpoint

## Overview

Thêm vnstock vào crawl-agent và expose endpoint `GET /market/board` trả về bảng giá tất cả mã.

## Files

- **Sửa**: `crawl-agent/requirements.txt`
- **Sửa**: `crawl-agent/app/main.py` — đăng ký router mới
- **Tạo mới**: `crawl-agent/app/routers/market.py`

## Implementation Steps

### 1. Thêm vnstock vào requirements.txt

```txt
# Market data
vnstock>=3.0.0
```

### 2. Tạo `crawl-agent/app/routers/market.py`

```python
"""Market data router — price board via vnstock."""
import json
import logging
from datetime import datetime, timezone

import redis.asyncio as redis
from fastapi import APIRouter, Depends, HTTPException, Query, Request

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/market", tags=["market"])

CACHE_TTL = 300  # 5 minutes

VALID_EXCHANGES = {"HSX", "HNX", "UPCOM", "ALL"}

async def _get_redis(request: Request):
    return request.app.state.redis_client


def _cache_key(exchange: str) -> str:
    return f"market:board:{exchange}"


async def _fetch_price_board(exchange: str) -> dict:
    """Fetch price board from TCBS via vnstock."""
    try:
        from vnstock import Vnstock
        stock = Vnstock()
        listing = stock.stock(source='VCI').listing

        # Get all symbols for the exchange
        if exchange == "ALL":
            symbols_df = listing.all_symbols()
        else:
            symbols_df = listing.symbols_by_exchange(exchange=exchange)

        if symbols_df is None or symbols_df.empty:
            return {"symbols": [], "updated_at": datetime.now(timezone.utc).isoformat()}

        symbols = symbols_df["ticker"].tolist() if "ticker" in symbols_df.columns else []
        if not symbols:
            return {"symbols": [], "updated_at": datetime.now(timezone.utc).isoformat()}

        # Fetch price board — TCBS supports bulk fetch
        quote = stock.stock(source='TCBS').quote
        board = quote.price_board(symbols_list=symbols)

        if board is None or board.empty:
            return {"symbols": [], "updated_at": datetime.now(timezone.utc).isoformat()}

        # Normalize to flat list
        result = []
        for _, row in board.iterrows():
            result.append({
                "symbol":      str(row.get("ticker", row.get("symbol", ""))),
                "price":       float(row.get("close", row.get("price", 0)) or 0),
                "change":      float(row.get("change", 0) or 0),
                "change_pct":  float(row.get("pct_change", row.get("changePct", 0)) or 0),
                "volume":      int(row.get("volume", 0) or 0),
                "ceiling":     float(row.get("ceiling", 0) or 0),
                "floor":       float(row.get("floor", 0) or 0),
                "ref":         float(row.get("ref_price", row.get("ref", 0)) or 0),
            })

        return {
            "symbols":    result,
            "total":      len(result),
            "updated_at": datetime.now(timezone.utc).isoformat(),
        }

    except ImportError:
        raise HTTPException(503, "vnstock not installed")
    except Exception as e:
        logger.error("Price board fetch error: %s", e)
        raise HTTPException(502, f"Price board unavailable: {e}")


@router.get("/board")
async def price_board(
    exchange: str = Query("ALL", description="HSX | HNX | UPCOM | ALL"),
    redis_client: redis.Redis | None = Depends(_get_redis),
):
    """Return near-realtime price board (~5min delay via TCBS)."""
    exchange = exchange.upper()
    if exchange not in VALID_EXCHANGES:
        raise HTTPException(400, f"exchange must be one of {VALID_EXCHANGES}")

    # Redis cache
    if redis_client:
        cached = await redis_client.get(_cache_key(exchange))
        if cached:
            return json.loads(cached)

    data = await _fetch_price_board(exchange)

    if redis_client:
        await redis_client.setex(_cache_key(exchange), CACHE_TTL, json.dumps(data))

    return data
```

### 3. Đăng ký router trong `app/main.py`

Thêm vào phần import routers:
```python
from app.routers import scrape, health, market
app.include_router(market.router)
```

## Fallback Strategy

Nếu vnstock fail (rate limit, API change), endpoint trả 502. Go service handle gracefully:
- Return empty board với `{"symbols": [], "error": "price board unavailable"}`
- Frontend hiển thị empty state thay vì crash

## Test

```bash
curl http://localhost:8085/market/board?exchange=HSX | python3 -c "
import json,sys; d=json.load(sys.stdin)
print('total:', d.get('total'))
print('first:', d.get('symbols', [''])[0])
"
```

## Success Criteria

- Endpoint trả về ≥100 symbols cho HSX
- Redis cache hoạt động (second call nhanh hơn)
- Fields đủ: symbol, price, change, change_pct, volume
