---
phase: 1
title: "DB Model & Migration"
status: completed
effort: 1h
completed: 2026-04-18
---

# Phase 1: DB Model & Migration (go-services)

## Context Links
- Next: [phase-02-go-api.md](./phase-02-go-api.md)
- Existing models: `go-services/internal/models/stock.go`
- DB connection: `go-services/internal/database/postgres.go`
- Auto-migration: `go-services/cmd/api-gateway/main.go`

## Overview

Thêm `Article` GORM model vào go-services. GORM auto-migration sẽ tự tạo table khi service start.

## Related Code Files

**Tạo mới:**
- `go-services/internal/models/article.go`

**Sửa:**
- `go-services/cmd/api-gateway/main.go` — thêm `&models.Article{}` vào AutoMigrate list

## Implementation Steps

### 1. Tạo `go-services/internal/models/article.go`

```go
package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusRejected  ArticleStatus = "rejected"
)

type Article struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol       string         `gorm:"index;size:10" json:"symbol"`
	Title        string         `gorm:"not null" json:"title"`
	Slug         string         `gorm:"uniqueIndex;size:200" json:"slug"`
	Content      string         `gorm:"type:text" json:"content"`
	Summary      string         `gorm:"type:text" json:"summary"`
	SourceURLs   pq.StringArray `gorm:"type:text[]" json:"source_urls"`
	ForecastData datatypes.JSON `gorm:"type:jsonb" json:"forecast_data"`
	Status       ArticleStatus  `gorm:"size:20;default:draft;index" json:"status"`
	PublishedAt  *time.Time     `json:"published_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
```

> **Lưu ý:** `pq.StringArray` cần `github.com/lib/pq` đã có sẵn trong go.mod (driver PostgreSQL). `datatypes.JSON` từ `gorm.io/datatypes`.

### 2. Kiểm tra `go.mod` có `gorm.io/datatypes` chưa

```bash
cd go-services && grep "gorm.io/datatypes" go.mod
```

Nếu chưa có:
```bash
go get gorm.io/datatypes
```

### 3. Sửa `main.go` — thêm Article vào AutoMigrate

Tìm block AutoMigrate trong `go-services/cmd/api-gateway/main.go`:

```go
// Trước
db.AutoMigrate(
    &models.Stock{},
    &models.TechnicalAnalysis{},
    &models.SentimentAnalysis{},
    &models.Forecast{},
    &models.DailyReport{},
)

// Sau — thêm &models.Article{}
db.AutoMigrate(
    &models.Stock{},
    &models.TechnicalAnalysis{},
    &models.SentimentAnalysis{},
    &models.Forecast{},
    &models.DailyReport{},
    &models.Article{},
)
```

### 4. Compile check

```bash
cd go-services && go build ./...
```

## Todo List

- [ ] Tạo `go-services/internal/models/article.go`
- [ ] Kiểm tra và thêm `gorm.io/datatypes` dependency nếu cần
- [ ] Thêm `&models.Article{}` vào AutoMigrate
- [ ] Chạy `go build ./...` — pass

## Success Criteria

- `go build ./...` không lỗi
- Table `articles` tự tạo trong PostgreSQL khi service start
- Schema: id, symbol, title, slug, content, summary, source_urls, forecast_data, status, published_at, created_at, updated_at

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| `gorm.io/datatypes` chưa có | Low | `go get gorm.io/datatypes` |
| `pq.StringArray` không tương thích | Low | fallback: dùng `datatypes.JSON` cho source_urls |
