---
phase: 4
title: "Tests & Validation"
status: completed
effort: 1h
---

# Phase 4: Tests & Validation

## Context Links
- Phase trước: [phase-03-forecast-fetch-news.md](./phase-03-forecast-fetch-news.md)
- Go tests: `go-services/internal/services/forecast_service_test.go`
- Python tests: `crawl-agent/` (nếu có test dir)

## Overview

Viết tests bao phủ 2 path chính: real news từ Redis và fallback placeholder.

## Related Code Files

**Sửa:**
- `go-services/internal/services/forecast_service_test.go`

**Tạo mới:**
- `crawl-agent/tests/test_news_publisher.py`

## Implementation Steps

### 1. Go — test `fetchRecentNews`

Thêm vào `forecast_service_test.go`:

```go
func TestFetchRecentNews_WithData(t *testing.T) {
    // Setup miniredis
    mr := miniredis.RunT(t)
    rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

    // Push fake news items
    item := `{"id":"n1","title":"Vinamilk tăng trưởng","content":"Doanh thu Q1 tăng 15%","source":"cafef","published_at":"2026-04-18T08:00:00+07:00"}`
    mr.LPush("news:VNM:recent", item)

    svc := &ForecastService{redis: rdb}
    items := svc.fetchRecentNews(context.Background(), "VNM")

    assert.Len(t, items, 1)
    assert.Contains(t, items[0].Content, "Vinamilk tăng trưởng")
    assert.Equal(t, "cafef", items[0].Source)
}

func TestFetchRecentNews_EmptyRedis(t *testing.T) {
    mr := miniredis.RunT(t)
    rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

    svc := &ForecastService{redis: rdb}
    items := svc.fetchRecentNews(context.Background(), "VNM")

    assert.Nil(t, items)
}

func TestFetchRecentNews_NilRedis(t *testing.T) {
    svc := &ForecastService{redis: nil}
    items := svc.fetchRecentNews(context.Background(), "VNM")
    assert.Nil(t, items)
}
```

### 2. Python — test `NewsPublisher`

```python
# crawl-agent/tests/test_news_publisher.py
import json
import pytest
import fakeredis.aioredis as fakeredis
from app.services.news_publisher import NewsPublisher


@pytest.mark.asyncio
async def test_publish_item_single_symbol():
    redis = fakeredis.FakeRedis()
    publisher = NewsPublisher(redis, max_items=20, ttl_hours=24)

    item = {
        "id": "abc123",
        "title": "VNM tăng trưởng",
        "content": "Doanh thu Q1...",
        "source": "cafef",
        "published_at": "2026-04-18T08:00:00+07:00",
        "symbols": ["VNM"],
    }
    count = await publisher.publish_item(item)
    assert count == 1

    raw = await redis.lrange("news:VNM:recent", 0, -1)
    assert len(raw) == 1
    data = json.loads(raw[0])
    assert data["title"] == "VNM tăng trưởng"
    assert data["source"] == "cafef"


@pytest.mark.asyncio
async def test_publish_item_multiple_symbols():
    redis = fakeredis.FakeRedis()
    publisher = NewsPublisher(redis)

    item = {"id": "x1", "title": "HPG và VNM", "content": "...",
            "source": "vnexpress", "symbols": ["HPG", "VNM"]}
    count = await publisher.publish_item(item)
    assert count == 2

    assert await redis.llen("news:HPG:recent") == 1
    assert await redis.llen("news:VNM:recent") == 1


@pytest.mark.asyncio
async def test_publish_item_no_symbols():
    redis = fakeredis.FakeRedis()
    publisher = NewsPublisher(redis)
    count = await publisher.publish_item({"id": "x", "title": "...", "symbols": []})
    assert count == 0


@pytest.mark.asyncio
async def test_ltrim_enforced():
    redis = fakeredis.FakeRedis()
    publisher = NewsPublisher(redis, max_items=3)

    for i in range(5):
        await publisher.publish_item(
            {"id": f"n{i}", "title": f"News {i}", "content": "", "source": "s",
             "symbols": ["VNM"]}
        )

    assert await redis.llen("news:VNM:recent") == 3
```

### 3. Chạy tests

```bash
# Go
cd /media/datnm/Data/Java/analysis-stock/go-services
go test ./internal/services/... -run TestFetchRecentNews -v

# Python
cd /media/datnm/Data/Java/analysis-stock/crawl-agent
pip install fakeredis pytest-asyncio
pytest tests/test_news_publisher.py -v
```

## Todo List

- [ ] Thêm 3 test cases Go vào `forecast_service_test.go`
- [ ] Tạo `crawl-agent/tests/test_news_publisher.py` với 4 test cases
- [ ] Chạy Go tests: `go test ./internal/services/... -v`
- [ ] Chạy Python tests: `pytest tests/test_news_publisher.py -v`
- [ ] Sửa lỗi nếu có, chạy lại cho đến khi pass

## Success Criteria

- Tất cả Go tests pass: `PASS`
- Tất cả Python tests pass
- `go build ./...` không lỗi
- Không có test nào dùng mock thay cho real Redis (dùng miniredis/fakeredis)
