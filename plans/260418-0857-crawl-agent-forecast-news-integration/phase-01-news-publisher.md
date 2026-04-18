---
phase: 1
title: "NewsPublisher — crawl-agent"
status: completed
effort: 1.5h
---

# Phase 1: NewsPublisher Service (crawl-agent)

## Context Links
- Main plan: [plan.md](./plan.md)
- Config: `crawl-agent/app/config.py`
- Main app: `crawl-agent/app/main.py`

## Overview

Tạo `NewsPublisher` service trong crawl-agent để push news items vào Redis theo từng symbol.

## Key Insights

- `app.state.redis_client` đã available (initialized trong `main.py` lifespan)
- Mỗi news item đã có `symbols: List[str]` từ `symbol_detector.extract_symbols()`
- Cần push 1 item vào nhiều symbol keys (1 bài viết nhắc đến VNM và HPG → push vào cả 2)
- Content cần truncate về 500 chars để tránh bloat Redis và giảm token khi gửi PhoBERT

## Requirements

- `LPUSH news:{symbol}:recent {json}` cho mỗi symbol trong item
- `LTRIM news:{symbol}:recent 0 19` — giữ tối đa 20 items mới nhất
- `EXPIRE news:{symbol}:recent 86400` — TTL 24h (reset mỗi lần có news mới)
- Không block scraping nếu Redis lỗi (fire-and-forget với try/except)
- Thêm 2 settings vào config: `news_redis_max_items` (default 20), `news_redis_ttl_hours` (default 24)

## Related Code Files

**Tạo mới:**
- `crawl-agent/app/services/news_publisher.py`

**Sửa:**
- `crawl-agent/app/config.py` — thêm 2 settings

## Implementation Steps

### 1. Thêm settings vào `config.py`

```python
# Redis news cache settings
news_redis_max_items: int = 20      # max items per symbol
news_redis_ttl_hours: int = 24      # TTL in hours
```

### 2. Tạo `crawl-agent/app/services/news_publisher.py`

```python
"""
Publishes scraped news items to Redis per stock symbol.

Each symbol gets a Redis List key: news:{symbol}:recent
Items are JSON-encoded dicts with title, content, source, published_at.
List is capped at max_items and given a TTL to prevent stale data.
"""
import json
import logging
from typing import Any, Dict, List, Optional

import redis.asyncio as redis

logger = logging.getLogger(__name__)

_REDIS_KEY_PREFIX = "news"


class NewsPublisher:
    def __init__(
        self,
        redis_client: redis.Redis,
        max_items: int = 20,
        ttl_hours: int = 24,
    ):
        self._redis = redis_client
        self._max_items = max_items
        self._ttl_seconds = ttl_hours * 3600

    async def publish_item(self, item: Dict[str, Any]) -> int:
        """Push one news item to Redis for each detected symbol.

        Returns count of symbols published to.
        """
        symbols: List[str] = item.get("symbols", [])
        if not symbols:
            return 0

        payload = json.dumps(
            {
                "id": item.get("id", ""),
                "title": item.get("title", "")[:200],
                "content": (item.get("content") or item.get("summary", ""))[:500],
                "source": item.get("source", "unknown"),
                "published_at": item.get("published_at", ""),
                "url": item.get("link", item.get("url", "")),
                "symbols": symbols,
            },
            ensure_ascii=False,
        )

        count = 0
        for symbol in symbols:
            key = f"{_REDIS_KEY_PREFIX}:{symbol}:recent"
            try:
                pipe = self._redis.pipeline()
                pipe.lpush(key, payload)
                pipe.ltrim(key, 0, self._max_items - 1)
                pipe.expire(key, self._ttl_seconds)
                await pipe.execute()
                count += 1
            except Exception as exc:
                logger.warning("Failed to publish news for %s: %s", symbol, exc)

        return count

    async def publish_batch(self, items: List[Dict[str, Any]]) -> Dict[str, int]:
        """Publish multiple items. Returns {symbols_count, items_published}."""
        total_symbols = 0
        items_published = 0
        for item in items:
            n = await self.publish_item(item)
            if n > 0:
                total_symbols += n
                items_published += 1
        return {"items_published": items_published, "symbol_keys_updated": total_symbols}
```

## Todo List

- [ ] Thêm `news_redis_max_items` và `news_redis_ttl_hours` vào `config.py`
- [ ] Tạo `crawl-agent/app/services/news_publisher.py`
- [ ] Tạo `crawl-agent/app/services/__init__.py` nếu chưa có (export `NewsPublisher`)

## Success Criteria

- `NewsPublisher` class có thể instantiate với redis client
- `publish_item()` push đúng JSON format vào Redis
- `publish_batch()` xử lý nhiều items
- Không raise exception khi Redis unavailable (chỉ log warning)
- Pipeline LPUSH + LTRIM + EXPIRE chạy atomic

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Redis unavailable | Low | try/except per symbol, log warning |
| Symbol list rỗng | Medium | Early return if not symbols |
| Content quá dài → bloat Redis | Low | Truncate title 200, content 500 chars |
