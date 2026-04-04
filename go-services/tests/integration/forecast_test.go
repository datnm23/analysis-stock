//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"vnstock-hybrid/internal/services"
	"vnstock-hybrid/internal/testutil"
)

func TestForecast_EndToEnd(t *testing.T) {
	ctx := context.Background()

	db, cleanupDB := testutil.SetupPostgres(ctx)
	defer cleanupDB()

	rdb, cleanupRedis := testutil.SetupRedis(ctx)
	defer cleanupRedis()

	// Mock sentiment server
	sentimentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sentiment":  "positive",
			"confidence": 75.0,
			"summary":    "Test sentiment result",
		})
	}))
	defer sentimentServer.Close()

	// Build service stack
	provider := &mockMarketDataProvider{data: generateTestOHLCV(100)}
	technicalSvc := services.NewTechnicalService(db, rdb, provider)
	sentimentClient := services.NewSentimentClient(sentimentServer.URL)
	forecastSvc := services.NewForecastService(db, rdb, technicalSvc, sentimentClient)

	// Run forecast
	result, err := forecastSvc.Forecast(ctx, "HPG")
	if err != nil {
		t.Fatalf("Forecast failed: %v", err)
	}

	if result.Symbol != "HPG" {
		t.Errorf("expected symbol HPG, got %s", result.Symbol)
	}
	if result.Recommendation == "" {
		t.Error("expected non-empty recommendation")
	}
	if result.CombinedScore <= 0 || result.CombinedScore > 100 {
		t.Errorf("combined score out of range: %f", result.CombinedScore)
	}

	t.Logf("Recommendation: %s (confidence: %.1f%%)", result.Recommendation, result.Confidence)
	t.Logf("Scores — Tech: %.1f, Sentiment: %.1f, Market: %.1f, Combined: %.1f",
		result.TechnicalScore, result.SentimentScore, result.MarketScore, result.CombinedScore)

	// Verify DB persistence
	var count int64
	db.Table("forecasts").Where("symbol = ?", "HPG").Count(&count)
	if count == 0 {
		t.Error("expected forecast to be stored in database")
	}
}

func TestDailyReport_EndToEnd(t *testing.T) {
	ctx := context.Background()

	db, cleanupDB := testutil.SetupPostgres(ctx)
	defer cleanupDB()

	rdb, cleanupRedis := testutil.SetupRedis(ctx)
	defer cleanupRedis()

	// Mock sentiment
	sentimentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sentiment":  "neutral",
			"confidence": 60.0,
		})
	}))
	defer sentimentServer.Close()

	provider := &mockMarketDataProvider{data: generateTestOHLCV(100)}
	technicalSvc := services.NewTechnicalService(db, rdb, provider)
	sentimentClient := services.NewSentimentClient(sentimentServer.URL)
	forecastSvc := services.NewForecastService(db, rdb, technicalSvc, sentimentClient)
	orchestratorSvc := services.NewOrchestratorService(db, rdb, forecastSvc)

	// Generate report for 3 symbols
	symbols := []string{"VNM", "FPT", "HPG"}
	report, err := orchestratorSvc.GenerateDailyReport(ctx, symbols)
	if err != nil {
		t.Fatalf("GenerateDailyReport failed: %v", err)
	}

	if report.TotalSymbolsAnalyzed != len(symbols) {
		t.Errorf("expected %d analyzed, got %d", len(symbols), report.TotalSymbolsAnalyzed)
	}
	total := report.BuySignals + report.SellSignals + report.HoldSignals
	if total != len(symbols) {
		t.Errorf("signal counts don't add up: %d + %d + %d = %d, expected %d",
			report.BuySignals, report.SellSignals, report.HoldSignals, total, len(symbols))
	}
	if report.MarketSummary == "" {
		t.Error("expected non-empty market summary")
	}

	t.Logf("Report: %d analyzed, Buy:%d Sell:%d Hold:%d",
		report.TotalSymbolsAnalyzed, report.BuySignals, report.SellSignals, report.HoldSignals)

	// Verify DB persistence
	var count int64
	db.Table("daily_reports").Count(&count)
	if count == 0 {
		t.Error("expected daily report to be stored in database")
	}

	// Verify cache + retrieval
	retrieved, err := orchestratorSvc.GetLatestReport(ctx)
	if err != nil {
		t.Fatalf("GetLatestReport failed: %v", err)
	}
	if retrieved.TotalSymbolsAnalyzed != report.TotalSymbolsAnalyzed {
		t.Errorf("retrieved report mismatch: got %d, want %d", retrieved.TotalSymbolsAnalyzed, report.TotalSymbolsAnalyzed)
	}
}
