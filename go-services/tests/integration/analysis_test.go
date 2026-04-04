//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"vnstock-hybrid/internal/services"
	"vnstock-hybrid/internal/testutil"
	"vnstock-hybrid/pkg/vnstock"
)

func TestTechnicalAnalysis_EndToEnd(t *testing.T) {
	ctx := context.Background()

	// Start real Postgres + Redis
	db, cleanupDB := testutil.SetupPostgres(ctx)
	defer cleanupDB()

	rdb, cleanupRedis := testutil.SetupRedis(ctx)
	defer cleanupRedis()

	// Use mock market data provider
	mockData := generateTestOHLCV(100)
	provider := &mockMarketDataProvider{data: mockData}

	svc := services.NewTechnicalService(db, rdb, provider)

	// First call: should compute and cache
	result, err := svc.Analyze(ctx, "VNM")
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.Symbol != "VNM" {
		t.Errorf("expected symbol VNM, got %s", result.Symbol)
	}
	if result.Signal == "" {
		t.Error("expected non-empty signal")
	}
	if result.Confidence <= 0 {
		t.Error("expected positive confidence")
	}

	t.Logf("Signal: %s, Score: %.2f, Confidence: %.2f", result.Signal, result.Score, result.Confidence)
	t.Logf("RSI: %.2f, ATR: %.2f", result.RSI, result.ATR)

	// Verify cache hit on second call
	result2, err := svc.Analyze(ctx, "VNM")
	if err != nil {
		t.Fatalf("Second Analyze failed: %v", err)
	}
	if result2.Symbol != result.Symbol {
		t.Error("cached result should return same symbol")
	}

	// Verify stored in database
	var count int64
	db.Table("technical_analyses").Where("symbol = ?", "VNM").Count(&count)
	if count == 0 {
		t.Error("expected technical analysis to be stored in database")
	}
}

func TestTechnicalBatch_EndToEnd(t *testing.T) {
	ctx := context.Background()

	db, cleanupDB := testutil.SetupPostgres(ctx)
	defer cleanupDB()

	rdb, cleanupRedis := testutil.SetupRedis(ctx)
	defer cleanupRedis()

	provider := &mockMarketDataProvider{data: generateTestOHLCV(100)}
	svc := services.NewTechnicalService(db, rdb, provider)

	symbols := []string{"VNM", "FPT", "VIC"}
	results, err := svc.AnalyzeBatch(ctx, symbols)
	if err != nil {
		t.Fatalf("AnalyzeBatch failed: %v", err)
	}

	if len(results) != len(symbols) {
		t.Errorf("expected %d results, got %d", len(symbols), len(results))
	}

	for _, sym := range symbols {
		if r, ok := results[sym]; !ok {
			t.Errorf("missing result for %s", sym)
		} else if r.Signal == "" {
			t.Errorf("empty signal for %s", sym)
		}
	}
}

// --- Helpers ---

type mockMarketDataProvider struct {
	data []vnstock.OHLCV
}

func (m *mockMarketDataProvider) GetHistoricalData(ctx context.Context, symbol string, days int) ([]vnstock.OHLCV, error) {
	if len(m.data) > days {
		return m.data[:days], nil
	}
	return m.data, nil
}

func (m *mockMarketDataProvider) GetMockData(symbol string, days int) []vnstock.OHLCV {
	return generateTestOHLCV(days)
}

func (m *mockMarketDataProvider) GetMarketIndex(_ context.Context, index string) (*vnstock.MarketIndex, error) {
	return &vnstock.MarketIndex{Name: index, Value: 1200.0, Change: 5.0, ChangePct: 0.42, Volume: 500000000}, nil
}

func (m *mockMarketDataProvider) GetForeignFlow(_ context.Context, symbol string) (*vnstock.ForeignFlow, error) {
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
