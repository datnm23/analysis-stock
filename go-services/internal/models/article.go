package models

import (
	"time"

	"gorm.io/gorm"
)

type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusRejected  ArticleStatus = "rejected"
)

type Article struct {
	ID           uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol       string        `gorm:"size:10;index" json:"symbol"`
	Title        string        `gorm:"not null" json:"title"`
	Slug         string        `gorm:"uniqueIndex;size:200" json:"slug"`
	Content      string        `gorm:"type:text" json:"content"`
	Summary      string        `gorm:"type:text" json:"summary"`
	SourceURLs   string        `gorm:"type:text" json:"source_urls"`    // JSON array string
	ForecastData string        `gorm:"type:jsonb" json:"forecast_data"` // snapshot at generation time
	ImageURL     *string       `gorm:"size:500" json:"image_url,omitempty"`
	Status       ArticleStatus `gorm:"size:20;default:draft;index" json:"status"`
	PublishedAt  *time.Time    `json:"published_at,omitempty"`
	CreatedAt    time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
