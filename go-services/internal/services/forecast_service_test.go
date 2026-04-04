package services

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"vnstock-hybrid/internal/indicators"
	"vnstock-hybrid/pkg/vnstock"
)

// MockMarketDataProvider implements vnstock.MarketDataProvider for testing.
// CallCount uses atomic int32 so it is safe for concurrent use.
type MockMarketDataProvider struct {
	HistoricalData []vnstock.OHLCV
	HistoricalErr  error
	CallCount      int32 // accessed via atomic ops
}

func (m *MockMarketDataProvider) GetHistoricalData(ctx context.Context, symbol string, days int) ([]vnstock.OHLCV, error) {
	atomic.AddInt32(&m.CallCount, 1)
	if m.HistoricalErr != nil {
		return nil, m.HistoricalErr
	}
	return m.HistoricalData, nil
}

func (m *MockMarketDataProvider) GetMockData(symbol string, days int) []vnstock.OHLCV {
	return generateTestOHLCV(days)
}

func (m *MockMarketDataProvider) GetMarketIndex(_ context.Context, index string) (*vnstock.MarketIndex, error) {
	return &vnstock.MarketIndex{Name: index, Value: 1200.0, Change: 5.0, ChangePct: 0.42, Volume: 500000000}, nil
}

func (m *MockMarketDataProvider) GetForeignFlow(_ context.Context, symbol string) (*vnstock.ForeignFlow, error) {
	return &vnstock.ForeignFlow{Symbol: symbol, NetBuyVol: 100000, NetBuyVal: 5000000000, BuyVol: 300000, SellVol: 200000}, nil
}

func generateTestOHLCV(days int) []vnstock.OHLCV {
	data := make([]vnstock.OHLCV, days)
	base := 50000.0
	for i := 0; i < days; i++ {
		data[i] = vnstock.OHLCV{
			Date:   time.Now().AddDate(0, 0, -days+i+1),
			Open:   base + float64(i)*10,
			High:   base + float64(i)*10 + 200,
			Low:    base + float64(i)*10 - 200,
			Close:  base + float64(i)*15,
			Volume: int64(1000000 + i*10000),
		}
	}
	return data
}

// --- Forecast helper function tests ---

func TestNormalizeTechnicalScore(t *testing.T) {
	tests := []struct {
		name  string
		score float64
		want  float64
	}{
		{"max positive", 10, 100},
		{"max negative", -10, 0},
		{"neutral", 0, 50},
		{"positive 5", 5, 75},
		{"negative 5", -5, 25},
		{"above range", 15, 100}, // clamped
		{"below range", -15, 0},  // clamped
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTechnicalScore(tt.score)
			if got != tt.want {
				t.Errorf("normalizeTechnicalScore(%v) = %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}

func TestSentimentToScore(t *testing.T) {
	tests := []struct {
		name       string
		sentiment  string
		confidence float64
		want       float64
	}{
		{"positive high", "positive", 80, 90},
		{"positive low", "positive", 20, 60},
		{"negative high", "negative", 80, 10},
		{"negative low", "negative", 20, 40},
		{"neutral", "neutral", 80, 50},
		{"unknown", "mixed", 80, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sentimentToScore(tt.sentiment, tt.confidence)
			if got != tt.want {
				t.Errorf("sentimentToScore(%q, %v) = %v, want %v", tt.sentiment, tt.confidence, got, tt.want)
			}
		})
	}
}

func TestScoreToRecommendation(t *testing.T) {
	tests := []struct {
		name      string
		score     float64
		wantRec   string
	}{
		{"strong buy", 90, "STRONG_BUY"},
		{"buy", 70, "BUY"},
		{"hold neutral", 50, "HOLD"},
		{"sell", 30, "SELL"},
		{"strong sell", 10, "STRONG_SELL"},
		{"boundary 80", 80, "STRONG_BUY"},
		{"boundary 60", 60, "BUY"},
		{"boundary 40", 40, "HOLD"},
		{"boundary 20", 20, "SELL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, confidence := scoreToRecommendation(tt.score)
			if rec != tt.wantRec {
				t.Errorf("scoreToRecommendation(%v) rec = %q, want %q", tt.score, rec, tt.wantRec)
			}
			if confidence <= 0 || confidence > 100 {
				t.Errorf("scoreToRecommendation(%v) confidence = %v, want 0 < c <= 100", tt.score, confidence)
			}
		})
	}
}

func TestCalculateMarketScore(t *testing.T) {
	tests := []struct {
		name string
		tech *TechnicalResult
		want float64
	}{
		{
			"neutral no data",
			&TechnicalResult{Price: PriceData{Close: 50000}},
			50,
		},
		{
			"strong trend with ADX",
			&TechnicalResult{
				Price:  PriceData{Close: 50000, Volume: 1000000},
				ADX:    &indicators.ADX{ADX: 30},
				ATR:    500, // 1% = low volatility
			},
			65, // 50 + 15 (ADX>25) — adjusted for current scoring
		},
		{
			"high volatility",
			&TechnicalResult{
				Price:  PriceData{Close: 50000, Volume: 1000000},
				ATR:    3000, // 6% = high volatility
			},
			40, // 50 - 10 (ATR>5%)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMarketScore(tt.tech)
			if got != tt.want {
				t.Errorf("calculateMarketScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TechnicalService with mock

func TestTechnicalService_AnalyzeWithMock(t *testing.T) {
	mock := &MockMarketDataProvider{
		HistoricalData: generateTestOHLCV(100),
	}

	svc := NewTechnicalService(nil, nil, mock)
	result, err := svc.Analyze(context.Background(), "VNM")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Symbol != "VNM" {
		t.Errorf("expected symbol VNM, got %s", result.Symbol)
	}
	if result.Signal == "" {
		t.Error("expected non-empty signal")
	}
	if atomic.LoadInt32(&mock.CallCount) != 1 {
		t.Errorf("expected 1 API call, got %d", atomic.LoadInt32(&mock.CallCount))
	}
}

func TestTechnicalService_AnalyzeFailFastOnError(t *testing.T) {
	mock := &MockMarketDataProvider{
		HistoricalErr: context.DeadlineExceeded,
	}

	svc := NewTechnicalService(nil, nil, mock)
	_, err := svc.Analyze(context.Background(), "FPT")

	if err == nil {
		t.Fatal("expected error when API fails, got nil (fail-fast behavior)")
	}
}

func TestTechnicalService_AnalyzeInsufficientData(t *testing.T) {
	mock := &MockMarketDataProvider{
		HistoricalData: generateTestOHLCV(10), // Only 10 points, need 26+
	}

	svc := NewTechnicalService(nil, nil, mock)
	_, err := svc.Analyze(context.Background(), "VNM")

	if err == nil {
		t.Fatal("expected error for insufficient data")
	}
}
