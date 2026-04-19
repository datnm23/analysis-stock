package services

import (
	"strconv"
	"time"
)

// TrendingItem represents one symbol in the trending list.
type TrendingItem struct {
	Symbol         string    `json:"symbol"`
	ArticleCount   int       `json:"count"`
	Recommendation string    `json:"recommendation"`
	LastPublished  time.Time `json:"last_published"`
}

// ScreenerItem holds the latest AI analysis snapshot per symbol.
type ScreenerItem struct {
	Symbol         string    `json:"symbol"`
	Recommendation string    `json:"recommendation"`
	Confidence     float64   `json:"confidence"`
	TechnicalScore float64   `json:"technical_score"`
	SentimentScore float64   `json:"sentiment_score"`
	LastAnalyzed   time.Time `json:"last_analyzed"`
	ArticleSlug    string    `json:"article_slug"`
}

// Trending returns top-N symbols with the most articles in the last `days` days.
func (s *ArticleService) Trending(days, limit int) ([]TrendingItem, error) {
	type row struct {
		Symbol        string    `gorm:"column:symbol"`
		ArticleCount  int       `gorm:"column:article_count"`
		Recommendation string   `gorm:"column:recommendation"`
		LastPublished time.Time `gorm:"column:last_published"`
	}

	var rows []row
	err := s.db.Raw(`
		SELECT
			a.symbol,
			COUNT(*) AS article_count,
			(
				SELECT forecast_data::jsonb->>'recommendation'
				FROM articles a2
				WHERE a2.symbol = a.symbol
				  AND a2.status = 'published'
				  AND (a2.forecast_data IS NOT NULL AND a2.forecast_data != '')
				ORDER BY COALESCE(a2.published_at, a2.created_at) DESC
				LIMIT 1
			) AS recommendation,
			MAX(COALESCE(a.published_at, a.created_at)) AS last_published
		FROM articles a
		WHERE a.status = 'published'
		  AND COALESCE(a.published_at, a.created_at) >= NOW() - ($1 || ' days')::interval
		GROUP BY a.symbol
		ORDER BY article_count DESC
		LIMIT $2
	`, days, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]TrendingItem, len(rows))
	for i, r := range rows {
		items[i] = TrendingItem{
			Symbol:         r.Symbol,
			ArticleCount:   r.ArticleCount,
			Recommendation: r.Recommendation,
			LastPublished:  r.LastPublished,
		}
	}
	return items, nil
}

// Screener returns the latest AI analysis per symbol, with optional recommendation filter.
// minConfidence: 0.0 = no filter; recommendation: "" = all.
func (s *ArticleService) Screener(recommendation string, minConfidence float64, limit int) ([]ScreenerItem, error) {
	type row struct {
		Symbol         string    `gorm:"column:symbol"`
		Recommendation string    `gorm:"column:recommendation"`
		Confidence     float64   `gorm:"column:confidence"`
		TechnicalScore float64   `gorm:"column:technical_score"`
		SentimentScore float64   `gorm:"column:sentiment_score"`
		LastAnalyzed   time.Time `gorm:"column:last_analyzed"`
		ArticleSlug    string    `gorm:"column:article_slug"`
	}

	baseQ := `
		SELECT DISTINCT ON (symbol)
			symbol,
			forecast_data::jsonb->>'recommendation'            AS recommendation,
			COALESCE((forecast_data::jsonb->>'confidence')::float, 0)       AS confidence,
			COALESCE((forecast_data::jsonb->>'technical_score')::float, 0)  AS technical_score,
			COALESCE((forecast_data::jsonb->>'sentiment_score')::float, 0)  AS sentiment_score,
			COALESCE(published_at, created_at)                 AS last_analyzed,
			slug                                               AS article_slug
		FROM articles
		WHERE status = 'published'
		  AND forecast_data IS NOT NULL
		  AND forecast_data != ''
		  AND forecast_data::jsonb ? 'recommendation'
	`

	args := []interface{}{}
	if recommendation != "" {
		baseQ += " AND forecast_data::jsonb->>'recommendation' = $1"
		args = append(args, recommendation)
	}
	if minConfidence > 0 {
		ph := len(args) + 1
		baseQ += " AND (forecast_data::jsonb->>'confidence')::float >= $" + itoa(ph)
		args = append(args, minConfidence)
	}
	ph := len(args) + 1
	baseQ += " ORDER BY symbol, COALESCE(published_at, created_at) DESC LIMIT $" + itoa(ph)
	args = append(args, limit)

	var rows []row
	if err := s.db.Raw(baseQ, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]ScreenerItem, len(rows))
	for i, r := range rows {
		items[i] = ScreenerItem{
			Symbol:         r.Symbol,
			Recommendation: r.Recommendation,
			Confidence:     r.Confidence,
			TechnicalScore: r.TechnicalScore,
			SentimentScore: r.SentimentScore,
			LastAnalyzed:   r.LastAnalyzed,
			ArticleSlug:    r.ArticleSlug,
		}
	}
	return items, nil
}

func itoa(n int) string { return strconv.Itoa(n) }
