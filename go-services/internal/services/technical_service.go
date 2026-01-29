package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"vnstock-hybrid/internal/indicators"
	"vnstock-hybrid/internal/models"
	"vnstock-hybrid/pkg/vnstock"
)

// TechnicalService handles technical analysis
type TechnicalService struct {
	db           *gorm.DB
	redis        *redis.Client
	marketClient *vnstock.Client
}

// TechnicalResult represents the result of technical analysis
type TechnicalResult struct {
	Symbol      string                      `json:"symbol"`
	Timestamp   time.Time                   `json:"timestamp"`
	Price       PriceData                   `json:"price"`
	RSI         float64                     `json:"rsi"`
	MACD        *indicators.MACD            `json:"macd"`
	Bollinger   *indicators.BollingerBands  `json:"bollinger"`
	Stochastic  *indicators.Stochastic      `json:"stochastic"`
	ADX         *indicators.ADX             `json:"adx"`
	SMA20       float64                     `json:"sma_20"`
	SMA50       float64                     `json:"sma_50"`
	EMA12       float64                     `json:"ema_12"`
	EMA26       float64                     `json:"ema_26"`
	ATR         float64                     `json:"atr"`
	VWAP        float64                     `json:"vwap"`
	Signal      string                      `json:"signal"`
	Confidence  float64                     `json:"confidence"`
	Score       float64                     `json:"score"`
	Reasons     []string                    `json:"reasons"`
}

// PriceData represents current price information
type PriceData struct {
	Open         float64 `json:"open"`
	High         float64 `json:"high"`
	Low          float64 `json:"low"`
	Close        float64 `json:"close"`
	Volume       int64   `json:"volume"`
	ChangePercent float64 `json:"change_percent"`
}

// NewTechnicalService creates a new technical analysis service
func NewTechnicalService(db *gorm.DB, redis *redis.Client, client *vnstock.Client) *TechnicalService {
	return &TechnicalService{
		db:           db,
		redis:        redis,
		marketClient: client,
	}
}

// Analyze performs technical analysis for a single symbol
func (s *TechnicalService) Analyze(ctx context.Context, symbol string) (*TechnicalResult, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("technical:%s:latest", symbol)
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var result TechnicalResult
			if json.Unmarshal([]byte(cached), &result) == nil {
				return &result, nil
			}
		}
	}

	// Fetch historical data (use mock data for now)
	history := s.marketClient.GetMockData(symbol, 100)
	if len(history) < 26 {
		return nil, fmt.Errorf("insufficient data for analysis")
	}

	// Extract price arrays
	closes := make([]float64, len(history))
	highs := make([]float64, len(history))
	lows := make([]float64, len(history))
	volumes := make([]int64, len(history))

	for i, h := range history {
		closes[i] = h.Close
		highs[i] = h.High
		lows[i] = h.Low
		volumes[i] = h.Volume
	}

	// Calculate indicators concurrently
	var wg sync.WaitGroup
	var rsiVal float64
	var macdVal *indicators.MACD
	var bbVal *indicators.BollingerBands
	var stochVal *indicators.Stochastic
	var adxVal *indicators.ADX
	var sma20Val, sma50Val, ema12Val, ema26Val, atrVal, vwapVal float64

	wg.Add(8)

	go func() {
		defer wg.Done()
		rsiVal = indicators.RSILatest(closes, 14)
	}()

	go func() {
		defer wg.Done()
		macdVal = indicators.CalculateMACD(closes, 12, 26, 9)
	}()

	go func() {
		defer wg.Done()
		bbVal = indicators.CalculateBollingerBands(closes, 20, 2.0)
	}()

	go func() {
		defer wg.Done()
		stochVal = indicators.CalculateStochastic(highs, lows, closes, 14, 3)
	}()

	go func() {
		defer wg.Done()
		adxVal = indicators.CalculateADX(highs, lows, closes, 14)
	}()

	go func() {
		defer wg.Done()
		sma20Val = indicators.SMALatest(closes, 20)
		sma50Val = indicators.SMALatest(closes, 50)
	}()

	go func() {
		defer wg.Done()
		ema12Val = indicators.EMALatest(closes, 12)
		ema26Val = indicators.EMALatest(closes, 26)
	}()

	go func() {
		defer wg.Done()
		atrVal = indicators.ATRLatest(highs, lows, closes, 14)
		vwapVal = indicators.VWAPLatest(highs, lows, closes, volumes)
	}()

	wg.Wait()

	// Current price data
	latest := history[len(history)-1]
	previous := history[len(history)-2]
	changePercent := ((latest.Close - previous.Close) / previous.Close) * 100

	// Generate signals
	signal, confidence, score, reasons := s.generateSignals(
		closes[len(closes)-1],
		rsiVal, macdVal, bbVal, stochVal, adxVal,
		sma20Val, sma50Val,
		float64(latest.Volume), float64(previous.Volume),
	)

	result := &TechnicalResult{
		Symbol:    symbol,
		Timestamp: time.Now(),
		Price: PriceData{
			Open:          latest.Open,
			High:          latest.High,
			Low:           latest.Low,
			Close:         latest.Close,
			Volume:        latest.Volume,
			ChangePercent: changePercent,
		},
		RSI:        rsiVal,
		MACD:       macdVal,
		Bollinger:  bbVal,
		Stochastic: stochVal,
		ADX:        adxVal,
		SMA20:      sma20Val,
		SMA50:      sma50Val,
		EMA12:      ema12Val,
		EMA26:      ema26Val,
		ATR:        atrVal,
		VWAP:       vwapVal,
		Signal:     signal,
		Confidence: confidence,
		Score:      score,
		Reasons:    reasons,
	}

	// Cache result
	if s.redis != nil {
		if data, err := json.Marshal(result); err == nil {
			s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
		}
	}

	// Store in database
	s.storeResult(ctx, result)

	return result, nil
}

// AnalyzeBatch performs analysis for multiple symbols concurrently
func (s *TechnicalService) AnalyzeBatch(ctx context.Context, symbols []string) (map[string]*TechnicalResult, error) {
	results := make(map[string]*TechnicalResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limit concurrency
	semaphore := make(chan struct{}, 10)

	for _, sym := range symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result, err := s.Analyze(ctx, symbol)
			if err != nil {
				return // Skip failed symbols
			}

			mu.Lock()
			results[symbol] = result
			mu.Unlock()
		}(sym)
	}

	wg.Wait()
	return results, nil
}

// generateSignals generates trading signals based on indicators
func (s *TechnicalService) generateSignals(
	price, rsi float64,
	macd *indicators.MACD,
	bb *indicators.BollingerBands,
	stoch *indicators.Stochastic,
	adx *indicators.ADX,
	sma20, sma50 float64,
	currentVolume, avgVolume float64,
) (signal string, confidence, score float64, reasons []string) {
	score = 0
	reasons = []string{}

	// RSI Analysis
	if rsi > 0 {
		if rsi < 30 {
			score += 2
			reasons = append(reasons, fmt.Sprintf("RSI quá bán (%.1f < 30) - Tín hiệu mua mạnh", rsi))
		} else if rsi < 40 {
			score += 1
			reasons = append(reasons, fmt.Sprintf("RSI thấp (%.1f) - Xu hướng tăng có thể", rsi))
		} else if rsi > 70 {
			score -= 2
			reasons = append(reasons, fmt.Sprintf("RSI quá mua (%.1f > 70) - Nguy cơ điều chỉnh", rsi))
		} else if rsi > 60 {
			score -= 1
			reasons = append(reasons, fmt.Sprintf("RSI cao (%.1f) - Cần thận trọng", rsi))
		}
	}

	// MACD Analysis
	if macd != nil {
		if macd.Histogram > 0 && macd.MACDLine > macd.SignalLine {
			score += 2
			reasons = append(reasons, "MACD cắt lên Signal - Tín hiệu tăng")
		} else if macd.Histogram < 0 && macd.MACDLine < macd.SignalLine {
			score -= 2
			reasons = append(reasons, "MACD cắt xuống Signal - Tín hiệu giảm")
		}

		if macd.MACDLine > 0 {
			score += 0.5
		} else {
			score -= 0.5
		}
	}

	// Moving Average Analysis
	if sma20 > 0 {
		if price > sma20 {
			score += 1
			reasons = append(reasons, fmt.Sprintf("Giá trên SMA20 (%.0f) - Xu hướng tăng ngắn hạn", sma20))
		} else {
			score -= 1
			reasons = append(reasons, fmt.Sprintf("Giá dưới SMA20 (%.0f) - Xu hướng giảm ngắn hạn", sma20))
		}
	}

	if sma20 > 0 && sma50 > 0 {
		if sma20 > sma50 {
			score += 1
			reasons = append(reasons, "SMA20 > SMA50 - Golden Cross, xu hướng tăng")
		} else {
			score -= 1
			reasons = append(reasons, "SMA20 < SMA50 - Death Cross, xu hướng giảm")
		}
	}

	// Bollinger Bands Analysis
	if bb != nil {
		if price < bb.Lower {
			score += 1.5
			reasons = append(reasons, fmt.Sprintf("Giá chạm dải BB dưới (%.0f) - Oversold", bb.Lower))
		} else if price > bb.Upper {
			score -= 1.5
			reasons = append(reasons, fmt.Sprintf("Giá chạm dải BB trên (%.0f) - Overbought", bb.Upper))
		}
	}

	// Stochastic Analysis
	if stoch != nil {
		if stoch.K < 20 && stoch.D < 20 {
			score += 1
			reasons = append(reasons, fmt.Sprintf("Stochastic oversold (%.1f) - Tín hiệu mua", stoch.K))
		} else if stoch.K > 80 && stoch.D > 80 {
			score -= 1
			reasons = append(reasons, fmt.Sprintf("Stochastic overbought (%.1f) - Tín hiệu bán", stoch.K))
		}
	}

	// ADX (Trend Strength)
	if adx != nil {
		if adx.ADX > 25 {
			reasons = append(reasons, fmt.Sprintf("ADX = %.1f - Xu hướng mạnh", adx.ADX))
		} else {
			reasons = append(reasons, fmt.Sprintf("ADX = %.1f - Thị trường sideway", adx.ADX))
		}
	}

	// Volume Analysis
	if avgVolume > 0 {
		volumeRatio := currentVolume / avgVolume
		if volumeRatio > 1.5 {
			reasons = append(reasons, fmt.Sprintf("Khối lượng tăng %.1fx - Dòng tiền mạnh", volumeRatio))
			score += 0.5
		} else if volumeRatio < 0.5 {
			reasons = append(reasons, fmt.Sprintf("Khối lượng thấp %.1fx - Dòng tiền yếu", volumeRatio))
			score -= 0.5
		}
	}

	// Generate final recommendation
	if score >= 4 {
		signal = "STRONG_BUY"
		confidence = min(95, 70+(score-4)*5)
	} else if score >= 2 {
		signal = "BUY"
		confidence = min(85, 60+(score-2)*5)
	} else if score >= -2 {
		signal = "HOLD"
		confidence = 50 + abs(score)*5
	} else if score >= -4 {
		signal = "SELL"
		confidence = min(85, 60+abs(score+2)*5)
	} else {
		signal = "STRONG_SELL"
		confidence = min(95, 70+abs(score+4)*5)
	}

	return signal, confidence, score, reasons
}

func (s *TechnicalService) storeResult(ctx context.Context, result *TechnicalResult) {
	if s.db == nil {
		return
	}

	analysis := &models.TechnicalAnalysis{
		Symbol:     result.Symbol,
		Timestamp:  result.Timestamp,
		OpenPrice:  result.Price.Open,
		HighPrice:  result.Price.High,
		LowPrice:   result.Price.Low,
		ClosePrice: result.Price.Close,
		Volume:     result.Price.Volume,
		RSI14:      &result.RSI,
		SMA20:      &result.SMA20,
		SMA50:      &result.SMA50,
		EMA12:      &result.EMA12,
		EMA26:      &result.EMA26,
		ATR:        &result.ATR,
		Signal:     result.Signal,
		Confidence: &result.Confidence,
		Score:      &result.Score,
	}

	if result.MACD != nil {
		analysis.MACDLine = &result.MACD.MACDLine
		analysis.MACDSignal = &result.MACD.SignalLine
		analysis.MACDHistogram = &result.MACD.Histogram
	}

	if result.Bollinger != nil {
		analysis.BBUpper = &result.Bollinger.Upper
		analysis.BBMiddle = &result.Bollinger.Middle
		analysis.BBLower = &result.Bollinger.Lower
	}

	if result.Stochastic != nil {
		analysis.StochK = &result.Stochastic.K
		analysis.StochD = &result.Stochastic.D
	}

	if result.ADX != nil {
		analysis.ADX = &result.ADX.ADX
	}

	s.db.Create(analysis)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
