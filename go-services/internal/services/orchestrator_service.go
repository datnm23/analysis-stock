package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"vnstock-hybrid/internal/models"
)

// reportConcurrencyLimit is the max number of concurrent forecast analyses per report.
const reportConcurrencyLimit = 5

// OrchestratorService coordinates all sub-agents and generates reports
type OrchestratorService struct {
	db              *gorm.DB
	redis           *redis.Client
	forecastSvc     *ForecastService
	defaultSymbols  []string // parsed once at init
}

// DailyReportResult represents the daily analysis report
type DailyReportResult struct {
	Date                 string                       `json:"date"`
	TotalSymbolsAnalyzed int                          `json:"total_symbols_analyzed"`
	BuySignals           int                          `json:"buy_signals"`
	SellSignals          int                          `json:"sell_signals"`
	HoldSignals          int                          `json:"hold_signals"`
	TopPicks             []ForecastResult             `json:"top_picks"`
	MarketSummary        string                       `json:"market_summary"`
	Results              map[string]*ForecastResult   `json:"results"`
	GeneratedAt          time.Time                    `json:"generated_at"`
}

// NewOrchestratorService creates a new orchestrator service
func NewOrchestratorService(db *gorm.DB, rdb *redis.Client, forecastSvc *ForecastService) *OrchestratorService {
	return &OrchestratorService{
		db:             db,
		redis:          rdb,
		forecastSvc:    forecastSvc,
		defaultSymbols: defaultWatchlist(),
	}
}

// GenerateDailyReport runs full analysis for a list of symbols
func (s *OrchestratorService) GenerateDailyReport(ctx context.Context, symbols []string) (*DailyReportResult, error) {
	// Check cache
	today := time.Now().Format("2006-01-02")
	cacheKey := fmt.Sprintf("report:daily:%s", today)
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var result DailyReportResult
			if json.Unmarshal([]byte(cached), &result) == nil {
				return &result, nil
			}
		}
	}

	// If no symbols provided, use default watchlist
	if len(symbols) == 0 {
		symbols = s.defaultSymbols
	}

	// Analyze all symbols concurrently
	results := make(map[string]*ForecastResult)
	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, reportConcurrencyLimit)

	for _, sym := range symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()

			// IMP-3: Respect parent context cancellation
			select {
			case <-ctx.Done():
				slog.Warn("Context cancelled, skipping analysis", "symbol", symbol)
				return
			case semaphore <- struct{}{}:
			}
			defer func() { <-semaphore }()

			result, err := s.forecastSvc.Forecast(ctx, symbol)
			if err != nil {
				slog.Warn("Forecast failed", "symbol", symbol, "error", err)
				return
			}

			mu.Lock()
			results[symbol] = result
			mu.Unlock()
		}(sym)
	}
	wg.Wait()

	// Aggregate results
	report := &DailyReportResult{
		Date:                 today,
		TotalSymbolsAnalyzed: len(results),
		Results:              results,
		GeneratedAt:          time.Now(),
	}

	// Count signals and find top picks
	topPicks := []ForecastResult{}
	for _, r := range results {
		switch r.Recommendation {
		case "STRONG_BUY", "BUY":
			report.BuySignals++
		case "STRONG_SELL", "SELL":
			report.SellSignals++
		default:
			report.HoldSignals++
		}

		// Top picks: STRONG_BUY or BUY with high confidence
		if (r.Recommendation == "STRONG_BUY" || r.Recommendation == "BUY") && r.Confidence >= 70 {
			topPicks = append(topPicks, *r)
		}
	}
	// Sort top picks by confidence descending for deterministic ordering
	sort.Slice(topPicks, func(i, j int) bool {
		return topPicks[i].Confidence > topPicks[j].Confidence
	})
	report.TopPicks = topPicks

	// Generate market summary
	report.MarketSummary = s.generateMarketSummary(report)

	// Cache report (24 hours)
	if s.redis != nil {
		if data, err := json.Marshal(report); err == nil {
			if err := s.redis.Set(ctx, cacheKey, data, 24*time.Hour).Err(); err != nil {
				slog.Warn("Failed to cache daily report", "date", today, "error", err)
			}
		}
	}

	// Persist to database
	s.storeReport(ctx, report)

	return report, nil
}

// GetLatestReport retrieves the most recent daily report.
// It first checks Redis cache, then falls back to the database.
func (s *OrchestratorService) GetLatestReport(ctx context.Context) (*DailyReportResult, error) {
	today := time.Now().Format("2006-01-02")
	cacheKey := fmt.Sprintf("report:daily:%s", today)

	// Try Redis cache first
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var result DailyReportResult
			if json.Unmarshal([]byte(cached), &result) == nil {
				return &result, nil
			}
		}
	}

	// Fallback to database
	if s.db != nil {
		var dbReport models.DailyReport
		if err := s.db.Order("report_date DESC").First(&dbReport).Error; err == nil {
			var result DailyReportResult
			if dbReport.ReportJSON != "" {
				if json.Unmarshal([]byte(dbReport.ReportJSON), &result) == nil {
					return &result, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no daily report found for %s", today)
}

func (s *OrchestratorService) generateMarketSummary(report *DailyReportResult) string {
	total := report.TotalSymbolsAnalyzed
	if total == 0 {
		return "Không có dữ liệu phân tích."
	}

	buyPct := float64(report.BuySignals) / float64(total) * 100
	sellPct := float64(report.SellSignals) / float64(total) * 100

	var sentiment string
	switch {
	case buyPct > 60:
		sentiment = "tích cực"
	case sellPct > 60:
		sentiment = "tiêu cực"
	default:
		sentiment = "trung lập"
	}

	return fmt.Sprintf(
		"Phân tích %d mã: %d mua (%.0f%%), %d bán (%.0f%%), %d giữ. Tâm lý thị trường: %s. Top picks: %d mã.",
		total, report.BuySignals, buyPct, report.SellSignals, sellPct,
		report.HoldSignals, sentiment, len(report.TopPicks),
	)
}

func (s *OrchestratorService) storeReport(ctx context.Context, report *DailyReportResult) {
	if s.db == nil {
		return
	}

	reportDate, err := time.Parse("2006-01-02", report.Date)
	if err != nil {
		slog.Error("Failed to parse report date, skipping DB store", "date", report.Date, "error", err)
		return
	}
	topPicksJSON, _ := json.Marshal(report.TopPicks)
	reportJSON, _ := json.Marshal(report)

	dbReport := &models.DailyReport{
		ReportDate:           reportDate,
		TotalSymbolsAnalyzed: report.TotalSymbolsAnalyzed,
		BuySignals:           report.BuySignals,
		SellSignals:          report.SellSignals,
		HoldSignals:          report.HoldSignals,
		TopPicks:             string(topPicksJSON),
		MarketSummary:        report.MarketSummary,
		ReportJSON:           string(reportJSON),
	}

	// Upsert: update if report for same date already exists
	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "report_date"}},
		DoUpdates: clause.AssignmentColumns([]string{"total_symbols_analyzed", "buy_signals", "sell_signals", "hold_signals", "top_picks", "market_summary", "report_json"}),
	}).Create(dbReport).Error; err != nil {
		slog.Warn("Failed to store daily report", "error", err)
	}
}

func defaultWatchlist() []string {
	// IMP-6: Read from environment variable if set
	if env := os.Getenv("WATCHLIST_SYMBOLS"); env != "" {
		symbols := strings.Split(env, ",")
		result := make([]string, 0, len(symbols))
		for _, s := range symbols {
			s = strings.TrimSpace(s)
			if s != "" {
				result = append(result, s)
			}
		}
		if len(result) > 0 {
			return result
		}
	}

	return []string{
		"VNM", "FPT", "VIC", "HPG", "VHM",
		"VCB", "BID", "TCB", "MBB", "VPB",
		"MWG", "SSI", "VND", "PNJ", "GAS",
	}
}
