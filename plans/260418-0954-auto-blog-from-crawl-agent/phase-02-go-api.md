---
phase: 2
title: "Go API Endpoints"
status: completed
effort: 2h
completed: 2026-04-18
---

# Phase 2: Go API Endpoints (go-services)

## Context Links
- Phase trước: [phase-01-db-model.md](./phase-01-db-model.md)
- Existing handler pattern: `go-services/internal/handlers/forecast.go`
- Router setup: `go-services/cmd/api-gateway/main.go`
- Model: `go-services/internal/models/article.go` (từ Phase 1)

## Overview

Thêm 4 endpoints cho articles:

| Method | Path | Auth | Mục đích |
|--------|------|------|---------|
| `GET` | `/api/v1/articles` | Public | List articles (filter by status, symbol) |
| `GET` | `/api/v1/articles/:slug` | Public | Single article by slug |
| `POST` | `/api/v1/articles` | Internal | Crawl-agent tạo draft |
| `PATCH` | `/api/v1/articles/:id/status` | Admin key | Admin approve/reject |

## Related Code Files

**Tạo mới:**
- `go-services/internal/services/article_service.go`
- `go-services/internal/handlers/articles.go`

**Sửa:**
- `go-services/cmd/api-gateway/main.go` — register routes

## Implementation Steps

### 1. Tạo `go-services/internal/services/article_service.go`

```go
package services

import (
	"fmt"
	"time"

	"go-services/internal/models"
	"gorm.io/gorm"
)

type ArticleService struct {
	db *gorm.DB
}

func NewArticleService(db *gorm.DB) *ArticleService {
	return &ArticleService{db: db}
}

type CreateArticleInput struct {
	Symbol       string   `json:"symbol" binding:"required"`
	Title        string   `json:"title" binding:"required"`
	Content      string   `json:"content" binding:"required"`
	Summary      string   `json:"summary"`
	SourceURLs   []string `json:"source_urls"`
	ForecastData []byte   `json:"forecast_data"`
}

type UpdateStatusInput struct {
	Status models.ArticleStatus `json:"status" binding:"required,oneof=published rejected"`
}

func (s *ArticleService) ListArticles(status string, symbol string, limit, offset int) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64
	q := s.db.Model(&models.Article{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if symbol != "" {
		q = q.Where("symbol = ?", symbol)
	}
	q.Count(&total)
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&articles).Error
	return articles, total, err
}

func (s *ArticleService) GetBySlug(slug string) (*models.Article, error) {
	var article models.Article
	err := s.db.Where("slug = ? AND status = ?", slug, models.ArticleStatusPublished).First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (s *ArticleService) Create(input CreateArticleInput) (*models.Article, error) {
	slug := generateSlug(input.Symbol, input.Title)
	article := &models.Article{
		Symbol:     input.Symbol,
		Title:      input.Title,
		Slug:       slug,
		Content:    input.Content,
		Summary:    input.Summary,
		SourceURLs: input.SourceURLs,
		Status:     models.ArticleStatusDraft,
	}
	if len(input.ForecastData) > 0 {
		article.ForecastData = input.ForecastData
	}
	if err := s.db.Create(article).Error; err != nil {
		return nil, err
	}
	return article, nil
}

func (s *ArticleService) UpdateStatus(id uint, status models.ArticleStatus) (*models.Article, error) {
	var article models.Article
	if err := s.db.First(&article, id).Error; err != nil {
		return nil, err
	}
	updates := map[string]any{"status": status}
	if status == models.ArticleStatusPublished {
		now := time.Now()
		updates["published_at"] = now
	}
	if err := s.db.Model(&article).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

func generateSlug(symbol, title string) string {
	// simple slug: symbol-YYYYMMDD-truncated-title
	date := time.Now().Format("20060102")
	short := sanitizeSlug(title)
	if len(short) > 50 {
		short = short[:50]
	}
	return fmt.Sprintf("%s-%s-%s", symbol, date, short)
}

func sanitizeSlug(s string) string {
	// lowercase, replace spaces with hyphens, strip non-alphanumeric
	result := make([]rune, 0, len(s))
	for _, r := range []rune(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			result = append(result, r)
		case r >= 'A' && r <= 'Z':
			result = append(result, r+32)
		case r == ' ' || r == '-':
			result = append(result, '-')
		}
	}
	return string(result)
}
```

### 2. Tạo `go-services/internal/handlers/articles.go`

```go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-services/internal/models"
	"go-services/internal/services"
)

type ArticlesHandler struct {
	svc         *services.ArticleService
	internalKey string // simple shared secret for internal POSTs
}

func NewArticlesHandler(svc *services.ArticleService, internalKey string) *ArticlesHandler {
	return &ArticlesHandler{svc: svc, internalKey: internalKey}
}

func (h *ArticlesHandler) List(c *gin.Context) {
	status := c.DefaultQuery("status", "published")
	symbol := c.Query("symbol")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}
	articles, total, err := h.svc.ListArticles(status, symbol, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"articles": articles, "total": total, "limit": limit, "offset": offset})
}

func (h *ArticlesHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	article, err := h.svc.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	c.JSON(http.StatusOK, article)
}

func (h *ArticlesHandler) Create(c *gin.Context) {
	// internal-only: verify shared key
	if h.internalKey != "" && c.GetHeader("X-Internal-Key") != h.internalKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var input services.CreateArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	article, err := h.svc.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, article)
}

func (h *ArticlesHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input services.UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	article, err := h.svc.UpdateStatus(uint(id64), input.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, article)
}
```

### 3. Sửa `main.go` — register routes

Trong hàm `setupRouter()` hoặc nơi khai báo routes:

```go
// Khởi tạo service và handler
articleSvc := services.NewArticleService(db)
internalKey := os.Getenv("INTERNAL_API_KEY") // thêm env var
articlesHandler := handlers.NewArticlesHandler(articleSvc, internalKey)

// Trong group v1
v1.GET("/articles", articlesHandler.List)
v1.GET("/articles/:slug", articlesHandler.GetBySlug)
v1.POST("/articles", articlesHandler.Create)
v1.PATCH("/articles/:id/status", articlesHandler.UpdateStatus)
```

### 4. Thêm env var vào `.env.example`

```env
INTERNAL_API_KEY=change-me-secret-key
```

### 5. Compile check

```bash
cd go-services && go build ./...
```

## Todo List

- [ ] Tạo `go-services/internal/services/article_service.go`
- [ ] Tạo `go-services/internal/handlers/articles.go`
- [ ] Sửa `main.go` — register 4 routes
- [ ] Thêm `INTERNAL_API_KEY` vào `.env.example`
- [ ] Chạy `go build ./...` — pass

## Success Criteria

- `GET /api/v1/articles?status=published` trả về JSON list
- `GET /api/v1/articles/:slug` trả về article hoặc 404
- `POST /api/v1/articles` với header `X-Internal-Key` tạo draft
- `PATCH /api/v1/articles/:id/status` cập nhật status và published_at

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Slug collision | Low | UniqueIndex trên slug; nếu conflict thêm random suffix |
| Vietnamese title → slug rỗng | Medium | Fallback slug = `{symbol}-{timestamp}` nếu sanitize ra empty |
