package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"vnstock-hybrid/internal/models"
)

// ForecastService combines technical and sentiment analysis for forecasts
type ForecastService struct {
	db              *gorm.DB
	redis           *redis.Client
	technicalSvc    *TechnicalService
	sentimentClient *SentimentClient
	anomalyDetector *AnomalyDetector
}

// ForecastResult represents a combined forecast
type ForecastResult struct {
	Symbol          string    `json:"symbol"`
	Timestamp       time.Time `json:"timestamp"`
	TechnicalScore  float64   `json:"technical_score"`
	SentimentScore  float64   `json:"sentiment_score"`
	MarketScore     float64   `json:"market_score"`
	CombinedScore   float64   `json:"combined_score"`
	Recommendation  string    `json:"recommendation"`
	Confidence      float64   `json:"confidence"`
	SupportPrice    float64   `json:"support_price,omitempty"`
	ResistancePrice float64   `json:"resistance_price,omitempty"`
	Reasoning       []string  `json:"reasoning"`
	// Phase 2+3 additions
	AnomalyDetected bool    `json:"anomaly_detected,omitempty"`
	WeightsUsed     Weights `json:"weights_used"`
}

// Weights represents the dynamic weight allocation for a single forecast.
type Weights struct {
	Technical float64 `json:"technical"`
	Sentiment float64 `json:"sentiment"`
	Market    float64 `json:"market"`
}

// Base weights — these are the starting point before adaptive adjustments.
const (
	baseTechnicalWeight = 0.50
	baseSentimentWeight = 0.25
	baseMarketWeight    = 0.25
)

// NewForecastService creates a new forecast service
func NewForecastService(db *gorm.DB, rdb *redis.Client, techSvc *TechnicalService, sentClient *SentimentClient) *ForecastService {
	var ad *AnomalyDetector
	if rdb != nil {
		ad = NewAnomalyDetector(rdb)
	}
	return &ForecastService{
		db:              db,
		redis:           rdb,
		technicalSvc:    techSvc,
		sentimentClient: sentClient,
		anomalyDetector: ad,
	}
}

// Forecast generates a combined forecast for a symbol
func (s *ForecastService) Forecast(ctx context.Context, symbol string) (*ForecastResult, error) {
	// Check cache
	cacheKey := fmt.Sprintf("forecast:%s:latest", symbol)
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var result ForecastResult
			if json.Unmarshal([]byte(cached), &result) == nil {
				return &result, nil
			}
		}
	}

	// Get technical analysis
	techResult, err := s.technicalSvc.Analyze(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("technical analysis failed for %s: %w", symbol, err)
	}

	// Normalize technical score to 0-100 range
	techScore := normalizeTechnicalScore(techResult.Score)

	// Get sentiment score (optional — degrade gracefully)
	sentScore := 50.0 // Neutral default
	sentReasons := []string{}
	sentConfidence := 0.0 // track to feed adaptive weights

	if s.sentimentClient != nil {
		sentResult, err := s.sentimentClient.Analyze(ctx, []TextItem{
			{ID: symbol, Content: fmt.Sprintf("Phân tích cổ phiếu %s", symbol)},
		})
		if err != nil {
			slog.Warn("Sentiment service unavailable, using neutral score", "symbol", symbol, "error", err)
			sentReasons = append(sentReasons, "Sentiment service unavailable — using neutral score")
		} else if len(sentResult.Results) > 0 {
			sentScore = sentimentToScore(sentResult.Results[0].Sentiment, sentResult.Results[0].Confidence)
			sentConfidence = sentResult.Results[0].Confidence
		}
	}

	// Market context score (enhanced — volume, ADX, ATR, SMA trend)
	marketScore := calculateMarketScore(techResult)

	// --- Phase 2: Anomaly detection ---
	anomalyDetected := false
	anomalyPenalty := 1.0
	if s.anomalyDetector != nil {
		alert, alertErr := s.anomalyDetector.RecordAndCheck(ctx, symbol, sentScore)
		if alertErr != nil {
			slog.Warn("Anomaly detection error", "symbol", symbol, "error", alertErr)
		}
		if alert != nil {
			anomalyDetected = true
			anomalyPenalty = alert.WeightPenalty()
			sentReasons = append(sentReasons,
				fmt.Sprintf("⚠ Anomaly detected: %s (%.1fσ deviation, penalty=%.0f%%)",
					alert.Severity, alert.Deviation, anomalyPenalty*100),
			)
		}
	}

	// --- Phase 3: Adaptive weights ---
	w := adaptiveWeights(techResult, sentConfidence, anomalyPenalty)

	// Weighted combination
	combinedScore := techScore*w.Technical + sentScore*w.Sentiment + marketScore*w.Market

	// Generate recommendation with confidence gate
	recommendation, confidence := scoreToRecommendation(combinedScore)

	// Confidence Gate: force HOLD when confidence is too low
	if confidence < 40 {
		recommendation = "HOLD"
		sentReasons = append(sentReasons, "Low confidence — forcing HOLD recommendation")
	}

	// Build reasoning
	reasoning := make([]string, 0, len(techResult.Reasons)+len(sentReasons)+5)
	reasoning = append(reasoning, techResult.Reasons...)
	reasoning = append(reasoning, sentReasons...)
	reasoning = append(reasoning,
		fmt.Sprintf("Technical Score: %.1f (weight: %.0f%%)", techScore, w.Technical*100),
		fmt.Sprintf("Sentiment Score: %.1f (weight: %.0f%%)", sentScore, w.Sentiment*100),
		fmt.Sprintf("Market Score: %.1f (weight: %.0f%%)", marketScore, w.Market*100),
		fmt.Sprintf("Combined Score: %.1f", combinedScore),
	)

	// Safe access to Bollinger (may be nil if insufficient data)
	var supportPrice, resistancePrice float64
	if techResult.Bollinger != nil {
		supportPrice = techResult.Bollinger.Lower
		resistancePrice = techResult.Bollinger.Upper
	}

	result := &ForecastResult{
		Symbol:          symbol,
		Timestamp:       time.Now(),
		TechnicalScore:  techScore,
		SentimentScore:  sentScore,
		MarketScore:     marketScore,
		CombinedScore:   combinedScore,
		Recommendation:  recommendation,
		Confidence:      confidence,
		SupportPrice:    supportPrice,
		ResistancePrice: resistancePrice,
		Reasoning:       reasoning,
		AnomalyDetected: anomalyDetected,
		WeightsUsed:     w,
	}

	// Cache result (5 minutes)
	if s.redis != nil {
		if data, err := json.Marshal(result); err == nil {
			if err := s.redis.Set(ctx, cacheKey, data, 5*time.Minute).Err(); err != nil {
				slog.Warn("Failed to cache forecast result", "symbol", symbol, "error", err)
			}
		}
	}

	// Persist to database
	s.storeForecast(ctx, result)

	return result, nil
}

func (s *ForecastService) storeForecast(ctx context.Context, result *ForecastResult) {
	if s.db == nil {
		return
	}

	forecast := &models.Forecast{
		Symbol:         result.Symbol,
		Timestamp:      result.Timestamp,
		TechnicalScore: result.TechnicalScore,
		SentimentScore: result.SentimentScore,
		MarketScore:    result.MarketScore,
		Recommendation: result.Recommendation,
		Confidence:     result.Confidence,
		Reasoning:      strings.Join(result.Reasoning, " | "),
	}

	// Only persist support/resistance if Bollinger bands were computed
	if result.SupportPrice != 0 {
		forecast.SupportPrice = &result.SupportPrice
	}
	if result.ResistancePrice != 0 {
		forecast.ResistancePrice = &result.ResistancePrice
	}

	if err := s.db.Create(forecast).Error; err != nil {
		slog.Warn("Failed to store forecast", "symbol", result.Symbol, "error", err)
	}
}

// ---------------------------------------------------------------------------
// Scoring helpers
// ---------------------------------------------------------------------------

// normalizeTechnicalScore normalizes the raw score (-10..+10) to 0..100
func normalizeTechnicalScore(score float64) float64 {
	normalized := (score + 10) / 20 * 100
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 100 {
		normalized = 100
	}
	return normalized
}

// sentimentToScore converts sentiment label + confidence to 0-100 score
func sentimentToScore(sentiment string, confidence float64) float64 {
	switch sentiment {
	case "positive":
		return 50 + confidence/2 // 50-100
	case "negative":
		return 50 - confidence/2 // 0-50
	default:
		return 50 // neutral
	}
}

// adaptiveWeights computes dynamic weights based on market conditions,
// sentiment reliability, and anomaly detection results.
func adaptiveWeights(tech *TechnicalResult, sentConfidence float64, anomalyPenalty float64) Weights {
	tw := baseTechnicalWeight
	sw := baseSentimentWeight
	mw := baseMarketWeight

	// 1. Strong trend → increase technical influence
	if tech.ADX != nil && tech.ADX.ADX > 30 {
		tw += 0.10
		sw -= 0.05
		mw -= 0.05
	}

	// 2. Low sentiment confidence → reduce sentiment, add to technical
	if sentConfidence < 40 {
		delta := sw * 0.5 // halve sentiment weight
		sw -= delta
		tw += delta
	}

	// 3. Anomaly penalty → further reduce sentiment
	if anomalyPenalty < 1.0 {
		reduction := sw * (1.0 - anomalyPenalty)
		sw -= reduction
		tw += reduction * 0.6
		mw += reduction * 0.4
	}

	// Ensure weights sum to 1.0 and are non-negative
	sw = math.Max(0.02, sw) // never fully zero — keep minimal signal
	total := tw + sw + mw
	tw /= total
	sw /= total
	mw /= total

	return Weights{
		Technical: math.Round(tw*100) / 100,
		Sentiment: math.Round(sw*100) / 100,
		Market:    math.Round(mw*100) / 100,
	}
}

// calculateMarketScore derives a market context score from technical data,
// enriched with VN-Index correlation and foreign investor flow when available.
func calculateMarketScore(tech *TechnicalResult) float64 {
	score := 50.0 // Start neutral

	// 1. ADX trend strength: strong trend is actionable
	if tech.ADX != nil {
		if tech.ADX.ADX > 30 {
			score += 15 // Strong trend
		} else if tech.ADX.ADX > 20 {
			score += 5 // Moderate trend
		}
		// Directional bias from DI
		if tech.ADX.PlusDI > tech.ADX.MinusDI {
			score += 5 // Bullish direction
		} else if tech.ADX.MinusDI > tech.ADX.PlusDI {
			score -= 5 // Bearish direction
		}
	}

	// 2. ATR-based volatility regime
	if tech.ATR > 0 && tech.Price.Close > 0 {
		atrPct := tech.ATR / tech.Price.Close * 100
		switch {
		case atrPct < 1.5:
			score += 10 // Very low vol — stable
		case atrPct < 3.0:
			score += 5 // Normal
		case atrPct > 5.0:
			score -= 10 // High vol — risky
		case atrPct > 8.0:
			score -= 20 // Extreme vol
		}
	}

	// 3. SMA trend alignment (price vs SMA20)
	if tech.SMA20 > 0 && tech.Price.Close > 0 {
		smaDiff := (tech.Price.Close - tech.SMA20) / tech.SMA20 * 100
		switch {
		case smaDiff > 5:
			score += 10 // Well above SMA20 — bullish
		case smaDiff > 0:
			score += 5 // Above SMA20
		case smaDiff < -5:
			score -= 10 // Well below SMA20 — bearish
		case smaDiff < 0:
			score -= 5 // Below SMA20
		}
	}

	// 4. VN-Index correlation (if available via MarketContext)
	if tech.VNIndexChangePct != 0 {
		if tech.VNIndexChangePct > 1.0 {
			score += 5 // Broad market rising
		} else if tech.VNIndexChangePct < -1.0 {
			score -= 5 // Broad market falling
		}
	}

	// 5. Foreign investor flow (if available via MarketContext)
	if tech.ForeignNetBuyVol != 0 {
		if tech.ForeignNetBuyVol > 0 {
			score += 5 // Foreign net buying — bullish signal
		} else {
			score -= 5 // Foreign net selling — bearish signal
		}
	}

	// Clamp to [0, 100]
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

// scoreToRecommendation converts a 0-100 combined score to a recommendation
func scoreToRecommendation(score float64) (string, float64) {
	switch {
	case score >= 80:
		return "STRONG_BUY", min(95, 70+(score-80))
	case score >= 60:
		return "BUY", min(85, 60+(score-60)/2)
	case score >= 40:
		return "HOLD", 50 + math.Abs(score-50)
	case score >= 20:
		return "SELL", min(85, 60+(40-score)/2)
	default:
		return "STRONG_SELL", min(95, 70+(20-score))
	}
}
