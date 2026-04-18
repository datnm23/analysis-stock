---
title: "Kết nối CrawlAgent → ForecastService qua Redis"
description: "Thay placeholder text bằng tin tức thực từ crawl-agent khi ForecastService gọi sentiment analysis"
status: completed
priority: P1
effort: 6h
branch: main
tags: [crawl-agent, forecast, redis, sentiment, integration]
blockedBy: []
blocks: []
created: 2026-04-18
completed: 2026-04-18
---

# Kết nối CrawlAgent → ForecastService qua Redis

## Overview

**Vấn đề:** `ForecastService.Forecast()` (dòng 103-104 `forecast_service.go`) gọi sentiment analysis với text giả:
```go
{ID: symbol, Content: fmt.Sprintf("Phân tích cổ phiếu %s", symbol)}
```
Kết quả: mọi sentiment score đều vô nghĩa vì không dựa trên tin tức thực.

**Giải pháp:** Dùng Redis List làm intermediary:
- CrawlAgent push news vào `news:{symbol}:recent` sau mỗi lần scrape
- ForecastService đọc từ Redis trước khi gọi sentiment
- Fallback graceful khi không có news trong Redis

## Architecture

```
CrawlAgent (Python)          Redis              ForecastService (Go)
     │                         │                        │
     │  scrape + dedup          │                        │
     │──────────────────────────▶  LPUSH news:VNM:recent │
     │                         │  LPUSH news:HPG:recent  │
     │                         │  (TTL 24h, max 20 items)│
     │                         │                        │
     │                         │◀──── LRANGE news:VNM:recent 0 9
     │                         │                        │
     │                         │  [{title, content,     │
     │                         │   source, published_at}]│──▶ SentimentClient.Analyze()
```

## Phases

| Phase | Name | Status |
|-------|------|--------|
| 1 | [NewsPublisher — crawl-agent](./phase-01-news-publisher.md) | Completed |
| 2 | [Tích hợp vào Scheduler & Router](./phase-02-integrate-publisher.md) | Completed |
| 3 | [ForecastService đọc news từ Redis](./phase-03-forecast-fetch-news.md) | Completed |
| 4 | [Tests & Validation](./phase-04-tests.md) | Completed |

## Dependencies

- Redis đang chạy và được cả 2 service (crawl-agent Python + go-services) kết nối
- crawl-agent đã có `app.state.redis_client` (initialized trong `main.py`)
- go-services đã có `redis.Client` trong `ForecastService`

## Redis Schema

```
Key:   news:{SYMBOL}:recent          (e.g., news:VNM:recent)
Type:  List (JSON strings)
Max:   20 items (LTRIM 0 19)
TTL:   86400s (24h)

Item JSON:
{
  "id": "unique-hash",
  "title": "Vinamilk báo lãi Q1...",
  "content": "Nội dung bài viết (max 500 chars)...",
  "source": "cafef",
  "published_at": "2026-04-18T08:30:00+07:00",
  "symbols": ["VNM"]
}
```
