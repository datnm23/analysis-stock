---
phase: 2
title: "Tích hợp NewsPublisher vào Scheduler & Router"
status: completed
effort: 1.5h
---

# Phase 2: Tích hợp NewsPublisher vào Scheduler & Router

## Context Links
- Phase trước: [phase-01-news-publisher.md](./phase-01-news-publisher.md)
- Scheduler: `crawl-agent/app/scheduler.py`
- Router: `crawl-agent/app/routers/scrape.py`
- Main app: `crawl-agent/app/main.py`

## Overview

Tích hợp `NewsPublisher` vào 2 điểm chạy scraping:
1. `ScrapeScheduler._run_scrape()` — chạy tự động mỗi 30 phút
2. `POST /scrape/pipeline` — endpoint n8n gọi

## Key Insights

- `NewsPublisher` cần `redis_client` từ `app.state.redis_client` (chỉ có trong router context)
- Trong `scheduler.py`, không có access đến `app.state` → cần inject redis_client khi khởi tạo scheduler
- Nên publish SAU dedup để tránh push duplicate items vào Redis
- `scheduler.py` hiện tại lưu items vào JSON file → THÊM Redis publish, không thay thế (backward compat)

## Related Code Files

**Sửa:**
- `crawl-agent/app/scheduler.py` — inject redis_client, gọi `NewsPublisher.publish_batch()`
- `crawl-agent/app/routers/scrape.py` — gọi `NewsPublisher.publish_batch()` trong `/scrape/pipeline`
- `crawl-agent/app/main.py` — truyền `redis_client` vào `ScrapeScheduler`

## Implementation Steps

### 1. Sửa `ScrapeScheduler.__init__` nhận optional redis_client

```python
# scheduler.py
from app.services.news_publisher import NewsPublisher

class ScrapeScheduler:
    def __init__(
        self,
        interval_minutes: int = 30,
        data_dir: str = "data/scheduled",
        market_hours_only: bool = True,
        redis_client=None,          # NEW
    ):
        ...
        self._news_publisher = (
            NewsPublisher(
                redis_client,
                max_items=settings.news_redis_max_items,
                ttl_hours=settings.news_redis_ttl_hours,
            )
            if redis_client else None
        )
```

### 2. Gọi publish_batch sau khi collect all_items trong `_run_scrape()`

Thêm vào cuối `_run_scrape()`, sau block "Extract symbols", trước "Save to JSON":

```python
# Publish to Redis per symbol
if self._news_publisher and all_items:
    pub_stats = await self._news_publisher.publish_batch(all_items)
    stats["redis_published"] = pub_stats
    logger.info(
        "Published %d items → %d symbol keys",
        pub_stats["items_published"],
        pub_stats["symbol_keys_updated"],
    )
```

### 3. Sửa `main.py` — truyền redis_client vào scheduler

```python
# main.py lifespan, sau khi khởi tạo redis_client
if not settings.debug:
    from app.scheduler import get_scheduler
    scheduler = get_scheduler(redis_client=redis_client)   # truyền redis
    scheduler.start()
```

Sửa `get_scheduler()` để nhận optional redis_client:

```python
def get_scheduler(redis_client=None) -> ScrapeScheduler:
    global _scheduler
    if _scheduler is None:
        interval = int(os.environ.get("SCRAPE_INTERVAL_MINUTES", "30"))
        market_only = os.environ.get("SCRAPE_MARKET_HOURS_ONLY", "true").lower() == "true"
        _scheduler = ScrapeScheduler(
            interval_minutes=interval,
            market_hours_only=market_only,
            redis_client=redis_client,
        )
    return _scheduler
```

### 4. Sửa `/scrape/pipeline` trong `scrape.py`

Sau block "Merge results", thêm publish:

```python
# Publish enriched items to Redis per symbol
if request.app.state.redis_client:
    from app.services.news_publisher import NewsPublisher
    from app.config import get_settings
    settings = get_settings()
    publisher = NewsPublisher(
        request.app.state.redis_client,
        max_items=settings.news_redis_max_items,
        ttl_hours=settings.news_redis_ttl_hours,
    )
    pub_stats = await publisher.publish_batch(items)
    stats["redis_published"] = pub_stats
```

## Todo List

- [ ] Sửa `ScrapeScheduler.__init__()` nhận `redis_client` parameter
- [ ] Thêm `_news_publisher` attribute vào `ScrapeScheduler`
- [ ] Thêm publish call trong `_run_scrape()` sau symbol extraction
- [ ] Sửa `get_scheduler()` nhận optional `redis_client`
- [ ] Sửa `main.py` lifespan — truyền `redis_client` vào `get_scheduler()`
- [ ] Sửa `POST /scrape/pipeline` — thêm publish step sau merge

## Success Criteria

- Sau mỗi scheduler run, Redis có keys `news:{symbol}:recent`
- `POST /scrape/pipeline` cũng publish vào Redis
- Nếu Redis down → scraping tiếp tục bình thường (chỉ skip publish)
- `stats` response có field `redis_published` với count

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Scheduler không có redis context | Medium | inject via `get_scheduler(redis_client)` |
| Double publish (scheduler + pipeline) | Low | LPUSH + LTRIM — duplicates sẽ bị trim tự nhiên |
| Import circular | Low | Import `NewsPublisher` local trong function nếu cần |
