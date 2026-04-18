---
phase: 3
title: "ForecastService đọc news từ Redis"
status: completed
effort: 2h
---

# Phase 3: ForecastService đọc news từ Redis (Go)

## Context Links
- Phase trước: [phase-02-integrate-publisher.md](./phase-02-integrate-publisher.md)
- File sửa: `go-services/internal/services/forecast_service.go`
- Tham khảo: `go-services/internal/services/sentiment_client.go` (TextItem struct)

## Overview

Thêm hàm `fetchRecentNews()` vào `ForecastService` để đọc news từ Redis key `news:{symbol}:recent`,
rồi thay placeholder text trong `Forecast()` bằng news thực.

## Key Insights

- `ForecastService` đã có `redis *redis.Client` field → không cần thêm dependency
- `TextItem` struct đã có `ID`, `Content`, `Source`, `PublishedAt` — đủ để map từ news JSON
- Cần giới hạn số lượng items gửi sentiment (max 10) để tránh PhoBERT timeout
- **Fallback quan trọng:** nếu Redis không có news → dùng placeholder cũ (không break existing behavior)
- Combine title + content làm `Content` để tối đa hóa thông tin cho PhoBERT

## Related Code Files

**Sửa:**
- `go-services/internal/services/forecast_service.go`

## Implementation Steps

### 1. Thêm struct `newsItem` (internal, unexported)

Thêm vào đầu file sau các imports:

```go
// newsItem mirrors the JSON stored by crawl-agent in Redis.
type newsItem struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    Content     string `json:"content"`
    Source      string `json:"source"`
    PublishedAt string `json:"published_at"`
    URL         string `json:"url"`
}
```

### 2. Thêm hàm `fetchRecentNews()`

```go
const (
    newsRedisKeyPrefix = "news"
    newsMaxItems       = 10  // max items to send to sentiment service
)

// fetchRecentNews reads recent news for a symbol from Redis.
// Returns nil slice (not error) when no news is available — callers degrade gracefully.
func (s *ForecastService) fetchRecentNews(ctx context.Context, symbol string) []TextItem {
    if s.redis == nil {
        return nil
    }

    key := fmt.Sprintf("%s:%s:recent", newsRedisKeyPrefix, symbol)
    raw, err := s.redis.LRange(ctx, key, 0, int64(newsMaxItems-1)).Result()
    if err != nil || len(raw) == 0 {
        return nil
    }

    items := make([]TextItem, 0, len(raw))
    for _, r := range raw {
        var n newsItem
        if err := json.Unmarshal([]byte(r), &n); err != nil {
            continue
        }
        content := strings.TrimSpace(n.Title + " " + n.Content)
        if content == "" {
            continue
        }
        var publishedAt time.Time
        if n.PublishedAt != "" {
            publishedAt, _ = time.Parse(time.RFC3339, n.PublishedAt)
        }
        items = append(items, TextItem{
            ID:          n.ID,
            Content:     content,
            Source:      n.Source,
            PublishedAt: publishedAt,
        })
    }
    return items
}
```

### 3. Sửa `Forecast()` — thay placeholder bằng news thực

**Trước (dòng 103-105):**
```go
sentResult, err := s.sentimentClient.Analyze(ctx, []TextItem{
    {ID: symbol, Content: fmt.Sprintf("Phân tích cổ phiếu %s", symbol)},
})
```

**Sau:**
```go
// Fetch recent news for this symbol from Redis (populated by crawl-agent)
newsItems := s.fetchRecentNews(ctx, symbol)
if len(newsItems) == 0 {
    // Fallback: neutral placeholder when no news available
    newsItems = []TextItem{
        {ID: symbol, Content: fmt.Sprintf("Phân tích cổ phiếu %s", symbol)},
    }
    slog.Debug("No recent news in Redis, using placeholder", "symbol", symbol)
}

sentResult, err := s.sentimentClient.Analyze(ctx, newsItems)
```

### 4. Thêm reasoning khi dùng real news

Trong block build `reasoning` (sau dòng 155), thêm:

```go
if len(newsItems) > 0 && newsItems[0].ID != symbol {
    reasoning = append(reasoning,
        fmt.Sprintf("Sentiment based on %d recent news articles", len(newsItems)),
    )
}
```

## Todo List

- [ ] Thêm `newsItem` struct vào `forecast_service.go`
- [ ] Thêm constants `newsRedisKeyPrefix`, `newsMaxItems`
- [ ] Implement `fetchRecentNews()` method
- [ ] Sửa `Forecast()` — gọi `fetchRecentNews()` trước `sentimentClient.Analyze()`
- [ ] Thêm fallback khi `newsItems` rỗng
- [ ] Thêm reasoning line khi dùng real news
- [ ] Kiểm tra compile: `cd go-services && go build ./...`

## Success Criteria

- `Forecast("VNM")` khi Redis có news → `SentimentResult` dựa trên tin tức thực
- `Forecast("VNM")` khi Redis rỗng → fallback placeholder, không error
- `ForecastResult.Reasoning` có dòng "Sentiment based on N recent news articles"
- `go build ./...` không lỗi

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| JSON parse lỗi với 1 item | Low | `continue` on error, skip bad item |
| Redis LRange timeout | Low | Context đã có timeout từ handler |
| newsItems rỗng sau filter | Medium | Fallback placeholder |
| PhoBERT quá tải với 10 items | Low | `newsMaxItems = 10`, tăng timeout sentiment client nếu cần |

## Security Considerations

- Content từ Redis là do crawl-agent push — đã qua dedup filter, không cần sanitize thêm
- `LRange` key `news:{symbol}:recent` — symbol đến từ input user → **KHÔNG** dùng symbol trực tiếp trong Lua/eval, chỉ dùng trong `fmt.Sprintf` cho key name (safe với go-redis)
