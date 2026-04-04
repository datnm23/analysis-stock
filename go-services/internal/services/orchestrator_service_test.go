package services

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestDefaultWatchlist_ReturnsDefaults(t *testing.T) {
	// Ensure env is not set
	os.Unsetenv("WATCHLIST_SYMBOLS")

	list := defaultWatchlist()
	if len(list) != 15 {
		t.Errorf("expected 15 default symbols, got %d", len(list))
	}
	if list[0] != "VNM" {
		t.Errorf("expected first symbol to be VNM, got %s", list[0])
	}
}

func TestDefaultWatchlist_ReadsFromEnv(t *testing.T) {
	os.Setenv("WATCHLIST_SYMBOLS", "AAA,BBB,CCC")
	defer os.Unsetenv("WATCHLIST_SYMBOLS")

	list := defaultWatchlist()
	if len(list) != 3 {
		t.Errorf("expected 3 symbols from env, got %d", len(list))
	}
	if list[0] != "AAA" || list[1] != "BBB" || list[2] != "CCC" {
		t.Errorf("unexpected symbols: %v", list)
	}
}

func TestDefaultWatchlist_TrimsWhitespace(t *testing.T) {
	os.Setenv("WATCHLIST_SYMBOLS", " VNM , FPT , HPG ")
	defer os.Unsetenv("WATCHLIST_SYMBOLS")

	list := defaultWatchlist()
	if len(list) != 3 {
		t.Errorf("expected 3 symbols, got %d", len(list))
	}
	if list[0] != "VNM" || list[1] != "FPT" || list[2] != "HPG" {
		t.Errorf("expected trimmed symbols, got %v", list)
	}
}

func TestDefaultWatchlist_EmptyEnvFallsBack(t *testing.T) {
	os.Setenv("WATCHLIST_SYMBOLS", "")
	defer os.Unsetenv("WATCHLIST_SYMBOLS")

	list := defaultWatchlist()
	if len(list) != 15 {
		t.Errorf("expected 15 default symbols on empty env, got %d", len(list))
	}
}

func TestDailyReportResult_MarketSummary(t *testing.T) {
	svc := NewOrchestratorService(nil, nil, nil)

	tests := []struct {
		name     string
		report   *DailyReportResult
		contains string
	}{
		{
			"no data",
			&DailyReportResult{TotalSymbolsAnalyzed: 0},
			"Không có dữ liệu",
		},
		{
			"bullish market",
			&DailyReportResult{
				TotalSymbolsAnalyzed: 10,
				BuySignals:           7,
				SellSignals:          2,
				HoldSignals:          1,
			},
			"tích cực",
		},
		{
			"bearish market",
			&DailyReportResult{
				TotalSymbolsAnalyzed: 10,
				BuySignals:           1,
				SellSignals:          7,
				HoldSignals:          2,
			},
			"tiêu cực",
		},
		{
			"neutral market",
			&DailyReportResult{
				TotalSymbolsAnalyzed: 10,
				BuySignals:           3,
				SellSignals:          3,
				HoldSignals:          4,
			},
			"trung lập",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := svc.generateMarketSummary(tt.report)
			if !strings.Contains(summary, tt.contains) {
				t.Errorf("summary %q does not contain %q", summary, tt.contains)
			}
		})
	}
}

func TestGenerateDailyReport_EmptySymbolsList(t *testing.T) {
	mock := &MockMarketDataProvider{
		HistoricalData: generateTestOHLCV(100),
	}
	techSvc := NewTechnicalService(nil, nil, mock)
	forecastSvc := NewForecastService(nil, nil, techSvc, nil)
	orchSvc := NewOrchestratorService(nil, nil, forecastSvc)

	// Empty symbols → use default watchlist
	report, err := orchSvc.GenerateDailyReport(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalSymbolsAnalyzed == 0 {
		t.Error("expected at least 1 analyzed symbol")
	}
	if report.Date == "" {
		t.Error("expected non-empty date")
	}
}

func TestGenerateDailyReport_CustomSymbols(t *testing.T) {
	mock := &MockMarketDataProvider{
		HistoricalData: generateTestOHLCV(100),
	}
	techSvc := NewTechnicalService(nil, nil, mock)
	forecastSvc := NewForecastService(nil, nil, techSvc, nil)
	orchSvc := NewOrchestratorService(nil, nil, forecastSvc)

	symbols := []string{"VNM", "FPT"}
	report, err := orchSvc.GenerateDailyReport(context.Background(), symbols)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalSymbolsAnalyzed > 2 {
		t.Errorf("expected at most 2 analyzed symbols, got %d", report.TotalSymbolsAnalyzed)
	}
}

