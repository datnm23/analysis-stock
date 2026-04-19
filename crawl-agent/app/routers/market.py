"""Market data router — price board via KBS ISS API (~realtime)."""
import asyncio
import json
import logging
from datetime import datetime, timezone
from typing import Any

import httpx
import redis.asyncio as aioredis
from fastapi import APIRouter, Depends, HTTPException, Query, Request

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/market", tags=["market"])

CACHE_TTL = 300  # 5 minutes
VALID_EXCHANGES = {"HSX", "HNX", "UPCOM", "ALL"}
BATCH_SIZE = 100

# KBS ISS endpoint (real-time price board)
_KBS_ISS_URL = "https://kbbuddywts.kbsec.com.vn/iis-server/investment/stock/iss"
_KBS_HEADERS = {
    "Content-Type": "application/json",
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "Accept-Language": "en-US,en;q=0.9,vi;q=0.8",
    "x-lang": "vi",
    "Referer": "https://banggia.kb.com.vn/",
}

# VCI listing endpoint for symbol list
_VCI_SYMBOLS_URL = "https://trading.vietcap.com.vn/api/price/symbols/getAll"


async def _get_redis(request: Request) -> aioredis.Redis | None:
    return getattr(request.app.state, "redis_client", None)


def _cache_key(exchange: str) -> str:
    return f"market:board:{exchange}"


def _kbs_row_to_dict(row: dict) -> dict | None:
    """Convert a KBS ISS row to our normalized format."""
    symbol = row.get("SB") or row.get("IN", "")
    if not symbol:
        return None

    def _f(key: str, default: float = 0.0) -> float:
        v = row.get(key)
        if v is None:
            return default
        try:
            return float(v)
        except (ValueError, TypeError):
            return default

    def _i(key: str, default: int = 0) -> int:
        v = row.get(key)
        if v is None:
            return default
        try:
            return int(float(v))
        except (ValueError, TypeError):
            return default

    close = _f("CP") or _f("AP")  # close price, fallback to average
    ref = _f("RE")
    change = close - ref if close and ref else _f("CH")
    change_pct = _f("CHP")

    return {
        "symbol":     symbol,
        "price":      close,
        "change":     change,
        "change_pct": change_pct,
        "volume":     _i("TT"),      # total accumulated volume
        "ceiling":    _f("CL"),
        "floor":      _f("FL"),
        "ref":        ref,
    }


async def _fetch_symbols(exchange: str) -> list[str]:
    """Get all stock symbols from VCI listing."""
    board_filter = {"HSX": "HSX", "HNX": "HNX", "UPCOM": "UPCOM"}.get(exchange)
    async with httpx.AsyncClient(timeout=15) as client:
        r = await client.get(_VCI_SYMBOLS_URL, headers={"User-Agent": "Mozilla/5.0"})
        r.raise_for_status()
        data = r.json()

    symbols = []
    for item in data:
        sym_type = item.get("type", "")
        board = item.get("board", "")
        if sym_type != "STOCK":
            continue
        if board_filter and board != board_filter:
            continue
        symbols.append(item["symbol"])
    return symbols


async def _fetch_kbs_batch(client: httpx.AsyncClient, symbols: list[str]) -> list[dict]:
    """Fetch one batch of symbols from KBS ISS."""
    payload = json.dumps({"code": ",".join(symbols)})
    try:
        r = await client.post(_KBS_ISS_URL, headers=_KBS_HEADERS, content=payload, timeout=30)
        if r.status_code not in (200, 201):
            logger.warning("KBS ISS batch returned %d", r.status_code)
            return []
        data = r.json()
        if not isinstance(data, list):
            return []
        rows = []
        for row in data:
            norm = _kbs_row_to_dict(row)
            if norm:
                rows.append(norm)
        return rows
    except Exception as e:
        logger.warning("KBS ISS batch error: %s", e)
        return []


async def _fetch_price_board(exchange: str) -> dict:
    """Async price board fetch: VCI symbol list → KBS ISS batched."""
    try:
        symbols = await _fetch_symbols(exchange)
    except Exception as e:
        raise RuntimeError(f"Could not fetch symbol list: {e}") from e

    if not symbols:
        return {"symbols": [], "total": 0, "updated_at": datetime.now(timezone.utc).isoformat()}

    logger.info("Fetching price board for %d symbols (exchange=%s)", len(symbols), exchange)

    async with httpx.AsyncClient() as client:
        tasks = [
            _fetch_kbs_batch(client, symbols[i: i + BATCH_SIZE])
            for i in range(0, len(symbols), BATCH_SIZE)
        ]
        results = await asyncio.gather(*tasks, return_exceptions=True)

    all_rows: list[dict] = []
    for res in results:
        if isinstance(res, list):
            all_rows.extend(res)

    return {
        "symbols":    all_rows,
        "total":      len(all_rows),
        "updated_at": datetime.now(timezone.utc).isoformat(),
    }


@router.get("/board")
async def price_board(
    exchange: str = Query("ALL", description="HSX | HNX | UPCOM | ALL"),
    redis_client: aioredis.Redis | None = Depends(_get_redis),
):
    """Return near-realtime price board via KBS ISS API."""
    exchange = exchange.upper()
    if exchange not in VALID_EXCHANGES:
        raise HTTPException(400, f"exchange must be one of {VALID_EXCHANGES}")

    if redis_client:
        try:
            cached = await redis_client.get(_cache_key(exchange))
            if cached:
                return json.loads(cached)
        except Exception:
            pass

    try:
        data = await _fetch_price_board(exchange)
    except RuntimeError as e:
        logger.error("Price board error: %s", e)
        raise HTTPException(502, str(e))

    if redis_client:
        try:
            await redis_client.setex(_cache_key(exchange), CACHE_TTL, json.dumps(data))
        except Exception:
            pass

    return data
