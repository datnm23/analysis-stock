package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"vnstock-hybrid/internal/models"

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
	ForecastData string   `json:"forecast_data"`
	ImageURL     string   `json:"image_url"` // empty string = no image
}

type UpdateStatusInput struct {
	Status string `json:"status" binding:"required,oneof=published rejected"`
}

type ArticleListResult struct {
	Articles []models.Article `json:"articles"`
	Total    int64            `json:"total"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
}

func (s *ArticleService) List(status, symbol string, limit, offset int) (*ArticleListResult, error) {
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
	if err != nil {
		return nil, err
	}
	return &ArticleListResult{Articles: articles, Total: total, Limit: limit, Offset: offset}, nil
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
	slug := buildSlug(input.Symbol, input.Title)

	// Ensure slug uniqueness by appending timestamp if needed
	var count int64
	s.db.Model(&models.Article{}).Where("slug = ?", slug).Count(&count)
	if count > 0 {
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix())
	}

	// Serialize source URLs as JSON string
	sourceURLsJSON := "[]"
	if len(input.SourceURLs) > 0 {
		if b, err := json.Marshal(input.SourceURLs); err == nil {
			sourceURLsJSON = string(b)
		}
	}

	article := &models.Article{
		Symbol:       input.Symbol,
		Title:        input.Title,
		Slug:         slug,
		Content:      input.Content,
		Summary:      input.Summary,
		SourceURLs:   sourceURLsJSON,
		ForecastData: input.ForecastData,
		ImageURL:     nullableString(input.ImageURL),
		Status:       models.ArticleStatusDraft,
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

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func buildSlug(symbol, title string) string {
	date := time.Now().Format("20060102")
	sanitized := sanitizeSlug(title)
	if sanitized == "" {
		sanitized = date
	} else if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}
	sanitized = strings.TrimRight(sanitized, "-")
	return fmt.Sprintf("%s-%s-%s", strings.ToLower(symbol), date, sanitized)
}

func sanitizeSlug(s string) string {
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevHyphen = false
		case r >= 'A' && r <= 'Z':
			b.WriteRune(unicode.ToLower(r))
			prevHyphen = false
		case r == ' ' || r == '-' || r == '_':
			if !prevHyphen {
				b.WriteRune('-')
				prevHyphen = true
			}
		}
	}
	return b.String()
}
