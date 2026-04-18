"""
Publishes scraped news items to Redis per stock symbol.

Each symbol gets a Redis List key: news:{symbol}:recent
Items are JSON-encoded dicts with title, content, source, published_at.
List is capped at max_items (LTRIM) and given a TTL to prevent stale data.
"""

import json
import logging
from typing import Any, Dict, List

import redis.asyncio as aioredis

logger = logging.getLogger(__name__)

_KEY_PREFIX = "news"


class NewsPublisher:
    def __init__(
        self,
        redis_client: aioredis.Redis,
        max_items: int = 20,
        ttl_hours: int = 24,
    ) -> None:
        self._redis = redis_client
        self._max_items = max_items
        self._ttl_seconds = ttl_hours * 3600

    async def publish_item(self, item: Dict[str, Any]) -> int:
        """Push one news item to Redis for each detected symbol.

        Returns count of symbol keys updated.
        Failures are logged and skipped — never raises.
        """
        symbols: List[str] = item.get("symbols", [])
        if not symbols:
            return 0

        payload = json.dumps(
            {
                "id": item.get("id", ""),
                "title": (item.get("title") or "")[:200],
                "content": (item.get("content") or item.get("summary") or "")[:500],
                "source": item.get("source", "unknown"),
                "published_at": item.get("published_at", ""),
                "url": item.get("link") or item.get("url", ""),
                "symbols": symbols,
            },
            ensure_ascii=False,
        )

        count = 0
        for symbol in symbols:
            key = f"{_KEY_PREFIX}:{symbol}:recent"
            try:
                pipe = self._redis.pipeline()
                pipe.lpush(key, payload)
                pipe.ltrim(key, 0, self._max_items - 1)
                pipe.expire(key, self._ttl_seconds)
                await pipe.execute()
                count += 1
            except Exception as exc:
                logger.warning("Redis publish failed for %s: %s", symbol, exc)

        return count

    async def publish_batch(self, items: List[Dict[str, Any]]) -> Dict[str, int]:
        """Publish multiple items. Returns stats dict."""
        items_published = 0
        symbol_keys_updated = 0
        for item in items:
            n = await self.publish_item(item)
            if n > 0:
                items_published += 1
                symbol_keys_updated += n
        return {
            "items_published": items_published,
            "symbol_keys_updated": symbol_keys_updated,
        }
