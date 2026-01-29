package models

import (
	"time"

	"gorm.io/gorm"
)

type Stock struct {
	Symbol    string    `gorm:"primaryKey;size:10" json:"symbol"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Exchange  string    `gorm:"size:10;not null" json:"exchange"`
	Industry  string    `gorm:"size:100" json:"industry"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type TechnicalAnalysis struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Symbol    string    `gorm:"size:10;not null;index:idx_tech_symbol_time" json:"symbol"`
	Timestamp time.Time `gorm:"not null;index:idx_tech_symbol_time" json:"timestamp"`

	// Price data
	OpenPrice  float64 `gorm:"type:decimal(12,2)" json:"open_price"`
	HighPrice  float64 `gorm:"type:decimal(12,2)" json:"high_price"`
	LowPrice   float64 `gorm:"type:decimal(12,2)" json:"low_price"`
	ClosePrice float64 `gorm:"type:decimal(12,2)" json:"close_price"`
	Volume     int64   `json:"volume"`

	// Indicators
	RSI14         *float64 `gorm:"type:decimal(5,2)" json:"rsi_14"`
	MACDLine      *float64 `gorm:"type:decimal(10,4)" json:"macd_line"`
	MACDSignal    *float64 `gorm:"type:decimal(10,4)" json:"macd_signal"`
	MACDHistogram *float64 `gorm:"type:decimal(10,4)" json:"macd_histogram"`
	BBUpper       *float64 `gorm:"type:decimal(12,2)" json:"bb_upper"`
	BBMiddle      *float64 `gorm:"type:decimal(12,2)" json:"bb_middle"`
	BBLower       *float64 `gorm:"type:decimal(12,2)" json:"bb_lower"`
	SMA20         *float64 `gorm:"type:decimal(12,2)" json:"sma_20"`
	SMA50         *float64 `gorm:"type:decimal(12,2)" json:"sma_50"`
	EMA12         *float64 `gorm:"type:decimal(12,2)" json:"ema_12"`
	EMA26         *float64 `gorm:"type:decimal(12,2)" json:"ema_26"`
	ADX           *float64 `gorm:"type:decimal(5,2)" json:"adx"`
	ATR           *float64 `gorm:"type:decimal(12,2)" json:"atr"`
	StochK        *float64 `gorm:"type:decimal(5,2)" json:"stoch_k"`
	StochD        *float64 `gorm:"type:decimal(5,2)" json:"stoch_d"`

	// Signal
	Signal     string   `gorm:"size:15" json:"signal"`
	Confidence *float64 `gorm:"type:decimal(5,2)" json:"confidence"`
	Score      *float64 `gorm:"type:decimal(5,2)" json:"score"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type SentimentAnalysis struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Symbol      *string   `gorm:"size:10;index" json:"symbol"`
	SourceURL   string    `gorm:"type:text" json:"source_url"`
	SourceType  string    `gorm:"size:50" json:"source_type"`
	TextContent string    `gorm:"type:text" json:"text_content"`
	Sentiment   string    `gorm:"size:20" json:"sentiment"`
	Confidence  float64   `gorm:"type:decimal(5,2)" json:"confidence"`
	Keywords    []string  `gorm:"type:text[];serializer:json" json:"keywords"`
	PublishedAt time.Time `json:"published_at"`
	AnalyzedAt  time.Time `gorm:"autoCreateTime" json:"analyzed_at"`
}

type Forecast struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Symbol          string    `gorm:"size:10;not null;index" json:"symbol"`
	Timestamp       time.Time `gorm:"not null" json:"timestamp"`
	TechnicalScore  float64   `gorm:"type:decimal(5,2)" json:"technical_score"`
	SentimentScore  float64   `gorm:"type:decimal(5,2)" json:"sentiment_score"`
	MarketScore     float64   `gorm:"type:decimal(5,2)" json:"market_score"`
	Recommendation  string    `gorm:"size:20" json:"recommendation"`
	Confidence      float64   `gorm:"type:decimal(5,2)" json:"confidence"`
	SupportPrice    *float64  `gorm:"type:decimal(12,2)" json:"support_price"`
	ResistancePrice *float64  `gorm:"type:decimal(12,2)" json:"resistance_price"`
	Reasoning       string    `gorm:"type:text" json:"reasoning"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type DailyReport struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	ReportDate           time.Time `gorm:"uniqueIndex;not null" json:"report_date"`
	TotalSymbolsAnalyzed int       `json:"total_symbols_analyzed"`
	BuySignals           int       `json:"buy_signals"`
	SellSignals          int       `json:"sell_signals"`
	HoldSignals          int       `json:"hold_signals"`
	TopPicks             string    `gorm:"type:jsonb" json:"top_picks"`
	MarketSummary        string    `gorm:"type:text" json:"market_summary"`
	ReportJSON           string    `gorm:"type:jsonb" json:"report_json"`
	ReportURL            string    `gorm:"type:text" json:"report_url"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (TechnicalAnalysis) TableName() string {
	return "technical_analysis"
}

func (SentimentAnalysis) TableName() string {
	return "sentiment_analysis"
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Stock{},
		&TechnicalAnalysis{},
		&SentimentAnalysis{},
		&Forecast{},
		&DailyReport{},
	)
}
