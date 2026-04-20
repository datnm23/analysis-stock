# Implementation Guide
# VN Stock Analysis System — Hybrid Architecture
## Version 2.0

**Document ID:** IG-VNSTOCK-HYBRID-002
**Date:** February 2026
**Status:** Active
**Architecture:** Hybrid (Go 1.22 + Python 3.11) on Docker Compose
**Companion SRS:** SRS-VNStock-Analysis-System-v2.0.md
**Supersedes:** Implementation-Guide-Hybrid-v1.0, Implementation-Guide-Go-GCP-v1.0, Implementation-Guide-VNStock-v1.0

---

## Document Control

| Field | Value |
|-------|-------|
| **Version** | 2.0 |
| **Date** | 2026-02-24 |
| **Status** | Active |
| **Audience** | Developers, DevOps Engineers, Contributors |
| **Companion SRS** | `docs/SRS-VNStock-Analysis-System-v2.0.md` |

### Change History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-28 | Team | Initial (VNStock Python/AWS) |
| 1.0-Go | 2026-01-28 | Team | Go/GCP variant |
| 1.0-Hybrid | 2026-01-28 | Team | Hybrid variant (GCP Cloud Run) |
| **2.0** | **2026-02-24** | **Team** | **Consolidated guide matching actual codebase: Docker Compose deployment, n8n orchestration, actual project structure, verified code examples** |

---

## Table of Contents

1. [Project Structure](#1-project-structure)
2. [Development Environment](#2-development-environment)
3. [Go Services Implementation](#3-go-services-implementation)
4. [Python Sentiment Service](#4-python-sentiment-service)
5. [Inter-Service Communication](#5-inter-service-communication)
6. [Docker Compose Infrastructure](#6-docker-compose-infrastructure)
7. [n8n Workflow Automation](#7-n8n-workflow-automation)
8. [Deployment](#8-deployment)
9. [Monitoring & Operations](#9-monitoring--operations)
10. [Testing](#10-testing)
11. [Troubleshooting](#11-troubleshooting)

---

## 1. Project Structure

The following tree reflects the **actual** project structure as of v2.0:

```
analysis-stock/
├── go-services/                        # All Go microservices
│   ├── cmd/
│   │   ├── api-gateway/
│   │   │   └── main.go                 # API Gateway entry point
│   │   └── technical-agent/
│   │       └── main.go                 # Technical Agent entry point
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go               # Environment-based configuration
│   │   ├── database/
│   │   │   ├── postgres.go             # GORM PostgreSQL connection
│   │   │   └── redis.go               # go-redis client
│   │   ├── handlers/
│   │   │   ├── analysis.go            # Analysis endpoint handlers
│   │   │   └── health.go              # Health/readiness checks
│   │   ├── indicators/                 # Technical indicator calculations
│   │   │   ├── rsi.go
│   │   │   ├── macd.go
│   │   │   ├── bollinger.go
│   │   │   ├── sma.go
│   │   │   ├── ema.go
│   │   │   ├── stochastic.go
│   │   │   ├── adx.go
│   │   │   ├── atr.go
│   │   │   ├── vwap.go
│   │   │   └── indicators_test.go     # Unit tests for all indicators
│   │   ├── middleware/
│   │   │   ├── logging.go             # Request/response logging
│   │   │   └── ratelimit.go           # Redis-backed rate limiting
│   │   ├── models/
│   │   │   └── stock.go               # GORM models (Stock, TechnicalAnalysis, etc.)
│   │   └── services/
│   │       ├── technical_service.go    # Core analysis logic + signal generation
│   │       └── sentiment_client.go     # HTTP client for Python sentiment
│   ├── pkg/
│   │   └── vnstock/
│   │       └── client.go              # Vietnamese stock market data client
│   ├── go.mod                          # Go 1.22, Gin, GORM, go-redis, GCP Pub/Sub
│   ├── go.sum
│   └── Makefile                        # Build, test, and run targets
│
├── python-sentiment/                   # Python sentiment analysis service
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py                    # FastAPI application + lifespan
│   │   ├── config.py                  # Pydantic Settings configuration
│   │   ├── models/
│   │   │   ├── __init__.py
│   │   │   └── phobert.py            # PhoBERT model wrapper (vinai/phobert-base)
│   │   ├── services/
│   │   │   ├── __init__.py
│   │   │   └── sentiment_analyzer.py  # Analysis + slang dictionary + symbol extraction
│   │   ├── routers/
│   │   │   ├── __init__.py
│   │   │   ├── analyze.py            # POST /analyze endpoint
│   │   │   └── health.py             # GET /health endpoint
│   │   └── data/
│   │       └── slang_dictionary.json  # Vietnamese stock market slang → sentiment modifiers
│   ├── requirements.txt               # FastAPI, transformers, torch, etc.
│   └── Dockerfile                     # Python 3.11-slim, model caching
│
├── scripts/
│   └── init.sql                       # PostgreSQL schema initialization
│
├── docker/
│   └── go.Dockerfile                  # Multi-stage Go build (distroless)
│
├── docker-compose.yml                 # Full Docker Compose with profiles
├── n8n-workflow-daily-analysis.json    # n8n daily workflow definition
├── n8n-workflow-error-handler.json     # n8n error handler workflow
├── quickstart.sh                      # One-command setup script
├── technical_agent.py                 # Standalone Python agent (legacy/reference)
├── CLAUDE.md                          # AI coding assistant context
└── README.md                          # Project overview and quick start
```

### Key Differences from v1.0 Guide

| Aspect | v1.0 (Hybrid) | v2.0 (Actual) |
|--------|---------------|---------------|
| Go entry points | 4 (api-gateway, technical, forecast, orchestrator) | 2 (`cmd/api-gateway`, `cmd/technical-agent`) |
| Python files | `symbol_extractor.py`, `pubsub/subscriber.py` | Not present; logic in `sentiment_analyzer.py` |
| Infrastructure | Terraform/GCP Cloud Run | Docker Compose only |
| Async messaging | GCP Pub/Sub (primary) | HTTP only (Pub/Sub optional/planned) |
| Deployment | `scripts/deploy.sh` → GCR | `docker-compose up -d` |
| Workflow engine | GCP Cloud Workflows | n8n (self-hosted in Docker) |

---

## 2. Development Environment

### 2.1 Prerequisites

```bash
# Go 1.22+
go version                      # go1.22.x or higher

# Python 3.11+ (for local development of sentiment service)
python3 --version               # 3.11.x

# Docker & Docker Compose
docker --version                # 24.x+
docker compose version          # v2.x+

# Make (for Go service targets)
make --version
```

> **Note:** GCP SDK and Terraform are **not required** for the default Docker Compose deployment. They are only needed if you plan to deploy to GCP Cloud Run in the future.

### 2.2 Quick Start (Docker Compose)

The fastest way to start the entire stack:

```bash
# Clone repository
git clone https://github.com/your-org/analysis-stock.git
cd analysis-stock

# Run the quick start script
chmod +x quickstart.sh
./quickstart.sh
```

Or manually:

```bash
# 1. Create environment file
cp .env.example .env
# Edit .env with your POSTGRES_PASSWORD, N8N_PASSWORD, TELEGRAM_BOT_TOKEN, etc.

# 2. Start core services
docker compose up -d

# 3. Verify services are running
docker compose ps
curl http://localhost:8080/health        # API Gateway
curl http://localhost:8000/health        # Sentiment Service
curl http://localhost:5678               # n8n UI
```

### 2.3 Local Development (Without Docker)

```bash
# Terminal 1: Start infrastructure
docker compose up -d postgres redis

# Terminal 2: Go API Gateway
cd go-services
go mod download
make run-api-gateway

# Terminal 3: Go Technical Agent
cd go-services
make run-technical

# Terminal 4: Python Sentiment Service
cd python-sentiment
python -m venv venv
source venv/bin/activate        # Linux/Mac
# .\venv\Scripts\Activate       # Windows
pip install -r requirements.txt
python -m app.main
```

### 2.4 Makefile Targets

```bash
cd go-services

make build              # Build all Go services
make test               # Run all Go tests
make test-indicators    # Run indicator tests only
make run-api-gateway    # Run API Gateway locally
make run-technical      # Run Technical Agent locally
make run-forecast       # Run Forecast Agent locally
make run-orchestrator   # Run Master Orchestrator locally
make deps               # Download and tidy Go dependencies
make clean              # Remove build artifacts
```

---

## 3. Go Services Implementation

### 3.1 Configuration (`internal/config/config.go`)

All configuration is via environment variables with sensible defaults. This is the **actual** code:

```go
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	PubSub   PubSubConfig
	Services ServicesConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type PubSubConfig struct {
	ProjectID string        // Optional: GCP Pub/Sub project ID
}

type ServicesConfig struct {
	TechnicalURL string     // http://localhost:8081 or http://technical-agent:8080
	SentimentURL string     // http://localhost:8000 or http://sentiment:8000
	ForecastURL  string     // http://localhost:8082 or http://forecast-agent:8080
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "vnstock"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "vnstock"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		PubSub: PubSubConfig{
			ProjectID: getEnv("GCP_PROJECT_ID", ""),
		},
		Services: ServicesConfig{
			TechnicalURL: getEnv("TECHNICAL_SERVICE_URL", "http://localhost:8081"),
			SentimentURL: getEnv("SENTIMENT_SERVICE_URL", "http://localhost:8000"),
			ForecastURL:  getEnv("FORECAST_SERVICE_URL", "http://localhost:8082"),
		},
	}
}

// Helper functions: getEnv, getIntEnv, getDurationEnv
```

**Key environment variables** (see also Appendix A):

| Variable | Default | Docker Compose |
|----------|---------|----------------|
| `PORT` | `8080` | Same |
| `DB_HOST` | `localhost` | `postgres` |
| `DB_PORT` | `5432` | `5432` |
| `DB_USER` | `vnstock` | `vnstock` |
| `DB_PASSWORD` | (empty) | `${POSTGRES_PASSWORD}` |
| `REDIS_HOST` | `localhost` | `redis` |
| `SENTIMENT_SERVICE_URL` | `http://localhost:8000` | `http://sentiment:8000` |
| `TECHNICAL_SERVICE_URL` | `http://localhost:8081` | `http://technical-agent:8080` |

### 3.2 Database Connection (`internal/database/`)

```go
// internal/database/postgres.go
func NewPostgresDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// internal/database/redis.go
func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	// Verify connection...
	return rdb, nil
}
```

### 3.3 Data Models (`internal/models/stock.go`)

GORM models that map to the PostgreSQL schema defined in `scripts/init.sql`:

```go
package models

import (
	"time"
	"gorm.io/gorm"
)

type Stock struct {
	Symbol    string    `gorm:"primaryKey;size:10" json:"symbol"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Exchange  string    `gorm:"size:10;not null" json:"exchange"`    // HSX, HNX, UPCOM
	Industry  string    `gorm:"size:100" json:"industry"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type TechnicalAnalysis struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Symbol    string    `gorm:"size:10;not null;index:idx_tech_symbol_time" json:"symbol"`
	Timestamp time.Time `gorm:"not null;index:idx_tech_symbol_time" json:"timestamp"`

	// OHLCV Price Data
	OpenPrice  float64 `gorm:"type:decimal(12,2)" json:"open_price"`
	HighPrice  float64 `gorm:"type:decimal(12,2)" json:"high_price"`
	LowPrice   float64 `gorm:"type:decimal(12,2)" json:"low_price"`
	ClosePrice float64 `gorm:"type:decimal(12,2)" json:"close_price"`
	Volume     int64   `json:"volume"`

	// Technical Indicators
	RSI14         *float64 `gorm:"type:decimal(5,2)" json:"rsi_14"`
	MACDLine      *float64 `gorm:"type:decimal(10,4)" json:"macd_line"`
	MACDSignal    *float64 `gorm:"type:decimal(10,4)" json:"macd_signal"`
	MACDHistogram *float64 `gorm:"type:decimal(10,4)" json:"macd_histogram"`
	BBUpper       *float64 `gorm:"type:decimal(12,2)" json:"bb_upper"`
	BBMiddle      *float64 `gorm:"type:decimal(12,2)" json:"bb_middle"`
	BBLower       *float64 `gorm:"type:decimal(12,2)" json:"bb_lower"`
	SMA20         *float64 `gorm:"type:decimal(12,2)" json:"sma_20"`
	EMA12         *float64 `gorm:"type:decimal(12,2)" json:"ema_12"`
	EMA26         *float64 `gorm:"type:decimal(12,2)" json:"ema_26"`
	ADX           *float64 `gorm:"type:decimal(5,2)" json:"adx"`
	ATR           *float64 `gorm:"type:decimal(12,2)" json:"atr"`

	// Signal
	Signal     string   `gorm:"size:10" json:"signal"`
	Confidence *float64 `gorm:"type:decimal(5,2)" json:"confidence"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
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

// AutoMigrate creates/updates all tables
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Stock{}, &TechnicalAnalysis{}, &SentimentAnalysis{}, &Forecast{})
}
```

### 3.4 Technical Indicators (`internal/indicators/`)

All indicators are implemented as pure functions with no external dependencies. Each file contains one indicator:

#### RSI (Relative Strength Index)

```go
// internal/indicators/rsi.go
package indicators

// RSI calculates the Relative Strength Index using smoothed averages
func RSI(closes []float64, period int) float64 {
	if len(closes) < period+1 {
		return 0
	}

	var gains, losses float64
	for i := 1; i <= period; i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	// Apply Wilder's smoothing for remaining periods
	for i := period + 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			avgGain = (avgGain*float64(period-1) + change) / float64(period)
			avgLoss = (avgLoss * float64(period-1)) / float64(period)
		} else {
			avgGain = (avgGain * float64(period-1)) / float64(period)
			avgLoss = (avgLoss*float64(period-1) - change) / float64(period)
		}
	}

	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100.0 - (100.0 / (1.0 + rs))
}
```

#### MACD (Moving Average Convergence Divergence)

```go
// internal/indicators/macd.go
package indicators

type MACD struct {
	MACDLine   float64
	SignalLine float64
	Histogram  float64
}

func CalculateMACD(closes []float64, fastPeriod, slowPeriod, signalPeriod int) *MACD {
	if len(closes) < slowPeriod+signalPeriod {
		return nil
	}

	emaFast := EMA(closes, fastPeriod)
	emaSlow := EMA(closes, slowPeriod)

	// MACD Line = Fast EMA - Slow EMA
	macdLine := make([]float64, len(closes))
	startIdx := slowPeriod - 1
	for i := startIdx; i < len(closes); i++ {
		macdLine[i] = emaFast[i] - emaSlow[i]
	}

	// Signal Line = EMA of MACD Line
	validMACD := macdLine[startIdx:]
	signalEMA := EMA(validMACD, signalPeriod)

	lastMACD := macdLine[len(macdLine)-1]
	lastSignal := signalEMA[len(signalEMA)-1]

	return &MACD{
		MACDLine:   lastMACD,
		SignalLine: lastSignal,
		Histogram:  lastMACD - lastSignal,
	}
}
```

#### Bollinger Bands

```go
// internal/indicators/bollinger.go
package indicators

import "math"

type BollingerBands struct {
	Upper  float64
	Middle float64
	Lower  float64
}

func CalculateBollingerBands(closes []float64, period int, stdDevMultiplier float64) *BollingerBands {
	if len(closes) < period {
		return nil
	}

	sma := SMA(closes, period)
	middle := sma[len(sma)-1]

	// Calculate standard deviation
	recentPrices := closes[len(closes)-period:]
	var sum float64
	for _, price := range recentPrices {
		diff := price - middle
		sum += diff * diff
	}
	stdDev := math.Sqrt(sum / float64(period))

	return &BollingerBands{
		Upper:  middle + (stdDevMultiplier * stdDev),
		Middle: middle,
		Lower:  middle - (stdDevMultiplier * stdDev),
	}
}
```

#### SMA & EMA (Moving Averages)

```go
// internal/indicators/sma.go
func SMA(prices []float64, period int) []float64 {
	if len(prices) < period { return nil }
	result := make([]float64, len(prices))
	var sum float64
	for i := 0; i < period; i++ { sum += prices[i] }
	result[period-1] = sum / float64(period)
	for i := period; i < len(prices); i++ {
		sum = sum - prices[i-period] + prices[i]
		result[i] = sum / float64(period)
	}
	return result
}

// internal/indicators/ema.go
func EMA(prices []float64, period int) []float64 {
	if len(prices) < period { return nil }
	result := make([]float64, len(prices))
	multiplier := 2.0 / float64(period+1)
	var sum float64
	for i := 0; i < period; i++ { sum += prices[i] }
	result[period-1] = sum / float64(period)
	for i := period; i < len(prices); i++ {
		result[i] = (prices[i]-result[i-1])*multiplier + result[i-1]
	}
	return result
}
```

#### Additional Indicators

| File | Indicator | Parameters |
|------|-----------|------------|
| `stochastic.go` | Stochastic Oscillator | %K period (14), %D period (3) |
| `adx.go` | Average Directional Index | Period (14) |
| `atr.go` | Average True Range | Period (14) |
| `vwap.go` | Volume Weighted Average Price | Intraday data |

### 3.5 Technical Analysis Service (`internal/services/technical_service.go`)

The core analysis service orchestrates indicator calculation with concurrent goroutines:

```go
package services

type TechnicalService struct {
	db           *gorm.DB
	redis        *redis.Client
	marketClient *vnstock.Client
}

type TechnicalResult struct {
	Symbol     string                      `json:"symbol"`
	Timestamp  time.Time                   `json:"timestamp"`
	RSI        float64                     `json:"rsi"`
	MACD       *indicators.MACD            `json:"macd"`
	Bollinger  *indicators.BollingerBands  `json:"bollinger"`
	SMA20      float64                     `json:"sma_20"`
	EMA12      float64                     `json:"ema_12"`
	EMA26      float64                     `json:"ema_26"`
	Signal     string                      `json:"signal"`
	Confidence float64                     `json:"confidence"`
}

func (s *TechnicalService) Analyze(ctx context.Context, symbol string) (*TechnicalResult, error) {
	// 1. Check Redis cache (TTL: 5 min)
	cacheKey := fmt.Sprintf("technical:%s:latest", symbol)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Return cached result
	}

	// 2. Fetch historical data (100 days)
	history, err := s.marketClient.GetHistoricalData(ctx, symbol, 100)
	closes := extractCloses(history)

	// 3. Calculate indicators concurrently
	var wg sync.WaitGroup
	wg.Add(5)
	go func() { defer wg.Done(); rsiVal = indicators.RSI(closes, 14) }()
	go func() { defer wg.Done(); macdVal = indicators.CalculateMACD(closes, 12, 26, 9) }()
	go func() { defer wg.Done(); bbVal = indicators.CalculateBollingerBands(closes, 20, 2.0) }()
	go func() { defer wg.Done(); /* SMA */ }()
	go func() { defer wg.Done(); /* EMA */ }()
	wg.Wait()

	// 4. Generate signal (BUY/SELL/HOLD) with confidence score
	signal, confidence := s.generateSignal(rsiVal, macdVal, bbVal, closes[len(closes)-1])

	// 5. Cache result in Redis
	// 6. Store in PostgreSQL

	return result, nil
}

func (s *TechnicalService) AnalyzeBatch(ctx context.Context, symbols []string) (map[string]*TechnicalResult, error) {
	// Concurrent analysis with semaphore (max 10 goroutines)
}

func (s *TechnicalService) generateSignal(rsi, macd, bb, price) (string, float64) {
	// Multi-indicator signal synthesis:
	// - RSI < 30 → Buy signal
	// - RSI > 70 → Sell signal
	// - MACD histogram > 0, line > signal → Buy
	// - Price < BB lower → Buy
	// - Price > BB upper → Sell
	// Returns weighted signal + confidence (0-100)
}
```

### 3.6 Sentiment Client (`internal/services/sentiment_client.go`)

The Go service communicates with the Python sentiment service over HTTP:

```go
type SentimentClient struct {
	baseURL    string
	httpClient *http.Client     // 30s timeout
}

type SentimentRequest struct {
	Texts []TextItem `json:"texts"`
}

type TextItem struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	Source      string    `json:"source,omitempty"`
	PublishedAt time.Time `json:"published_at,omitempty"`
}

type SentimentResponse struct {
	Results []SentimentResult `json:"results"`
}

type SentimentResult struct {
	ID         string   `json:"id"`
	Sentiment  string   `json:"sentiment"`     // positive, negative, neutral
	Confidence float64  `json:"confidence"`     // 0-100
	Symbols    []string `json:"symbols"`        // Extracted stock symbols
	Keywords   []string `json:"keywords"`
}

func (c *SentimentClient) Analyze(ctx context.Context, texts []TextItem) (*SentimentResponse, error) {
	// POST http://sentiment:8000/analyze
	// Content-Type: application/json
}

func (c *SentimentClient) Health(ctx context.Context) error {
	// GET http://sentiment:8000/health
}
```

### 3.7 API Gateway (`cmd/api-gateway/main.go`)

The API Gateway is the main entry point for all HTTP requests:

```go
func main() {
	cfg := config.Load()

	// Database connection (optional — graceful fallback if not configured)
	var db *gorm.DB
	if cfg.Database.Password != "" {
		db, err = database.NewPostgresDB(cfg.Database)
	}

	// Redis connection (optional — rate limiting disabled without Redis)
	var rdb *redis.Client
	if cfg.Redis.Host != "" {
		rdb, err = database.NewRedisClient(cfg.Redis)
	}

	// Initialize services
	marketClient := vnstock.NewClient()
	technicalSvc := services.NewTechnicalService(db, rdb, marketClient)
	sentimentClient := services.NewSentimentClient(cfg.Services.SentimentURL)

	// Gin router setup
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	if rdb != nil {
		r.Use(middleware.RateLimiter(rdb, 100, time.Minute))
	}

	// Routes
	r.GET("/health", handlers.HealthCheck(db, rdb))
	r.GET("/ready", handlers.ReadinessCheck(db, rdb, sentimentClient))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/technical/:symbol", handlers.TechnicalAnalysis(technicalSvc))
		v1.POST("/technical/batch",  handlers.TechnicalBatch(technicalSvc))
		v1.POST("/sentiment",        handlers.SentimentProxy(sentimentClient))
		v1.POST("/analyze",          handlers.FullAnalysis(technicalSvc, sentimentClient))
	}

	// Graceful shutdown with signal handling (SIGINT, SIGTERM)
	srv := &http.Server{Addr: ":" + cfg.Server.Port, Handler: r}
	// ...
}
```

**Important:** The API Gateway connects to PostgreSQL, Redis, and the Python Sentiment Service optionally — it starts successfully even if some dependencies are unavailable. This makes local development easier.

### 3.8 Handlers (`internal/handlers/analysis.go`)

```go
// GET /api/v1/technical/:symbol
func TechnicalAnalysis(svc *services.TechnicalService) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Param("symbol")
		if len(symbol) != 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid symbol format"})
			return
		}
		result, err := svc.Analyze(c.Request.Context(), symbol)
		// ...
		c.JSON(http.StatusOK, result)
	}
}

// POST /api/v1/technical/batch  — max 50 symbols
func TechnicalBatch(svc *services.TechnicalService) gin.HandlerFunc { ... }

// POST /api/v1/sentiment  — proxy to Python sentiment service
func SentimentProxy(client *services.SentimentClient) gin.HandlerFunc { ... }

// POST /api/v1/analyze  — combined technical + sentiment
func FullAnalysis(techSvc *services.TechnicalService, sentClient *services.SentimentClient) gin.HandlerFunc { ... }
```

### 3.9 Middleware

#### Logging (`internal/middleware/logging.go`)

- Logs: method, path, status, latency, client IP
- JSON-structured output for production

#### Rate Limiting (`internal/middleware/ratelimit.go`)

- Redis-backed sliding window rate limiter
- Default: 100 requests per minute per IP
- Only active when Redis is connected

### 3.10 Market Data Client (`pkg/vnstock/client.go`)

```go
// pkg/vnstock/client.go
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type HistoricalData struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

func (c *Client) GetMockData(symbol string, days int) []HistoricalData { ... }
func (c *Client) GetHistoricalData(ctx context.Context, symbol string, days int) ([]HistoricalData, error) { ... }
```

> **Note:** Currently uses mock data for development. Production integration with the `vnstock` Python library or direct API calls is planned.

---

## 4. Python Sentiment Service

### 4.1 Project Structure

```
python-sentiment/
├── app/
│   ├── __init__.py
│   ├── main.py                    # FastAPI app + lifespan (model loading)
│   ├── config.py                  # Pydantic Settings
│   ├── models/
│   │   └── phobert.py             # PhoBERT model wrapper
│   ├── services/
│   │   └── sentiment_analyzer.py  # Analysis pipeline + slang dictionary
│   ├── routers/
│   │   ├── analyze.py             # POST /analyze
│   │   └── health.py              # GET /health
│   └── data/
│       └── slang_dictionary.json  # Vietnamese stock slang → sentiment modifiers
├── requirements.txt
└── Dockerfile
```

### 4.2 Dependencies (`requirements.txt`)

```txt
fastapi==0.109.0
uvicorn[standard]==0.27.0
transformers==4.37.0
torch==2.1.2
numpy==1.26.3
pydantic==2.5.3
pydantic-settings==2.1.0
python-dotenv==1.0.0
redis==5.0.1
google-cloud-pubsub==2.19.0
```

> **Note:** `google-cloud-pubsub` is included for future async messaging but not actively used in the current Docker Compose deployment.

### 4.3 Configuration (`app/config.py`)

```python
from pydantic_settings import BaseSettings
from functools import lru_cache

class Settings(BaseSettings):
    host: str = "0.0.0.0"
    port: int = 8000
    debug: bool = False
    model_name: str = "vinai/phobert-base"
    model_cache_dir: str = "./.cache"
    max_sequence_length: int = 256
    redis_host: str = "localhost"
    redis_port: int = 6379
    redis_password: str = ""
    redis_db: int = 0
    gcp_project_id: str = ""

    class Config:
        env_file = ".env"

@lru_cache()
def get_settings() -> Settings:
    return Settings()
```

### 4.4 PhoBERT Model (`app/models/phobert.py`)

The PhoBERT wrapper handles model loading, single and batch prediction:

```python
class PhoBERTSentiment:
    """Vietnamese sentiment analysis using vinai/phobert-base."""

    def __init__(self, model_name="vinai/phobert-base", cache_dir="./.cache"):
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        self.tokenizer = AutoTokenizer.from_pretrained(model_name, cache_dir=cache_dir)
        self.model = AutoModelForSequenceClassification.from_pretrained(
            model_name, num_labels=3, cache_dir=cache_dir  # positive, negative, neutral
        )
        self.model.to(self.device).eval()
        self.labels = ["negative", "neutral", "positive"]

    def predict(self, text: str) -> Tuple[str, float]:
        """Single text → (sentiment, confidence%)"""

    def predict_batch(self, texts: List[str], batch_size=16) -> List[Tuple[str, float]]:
        """Batch prediction for efficiency"""

    @property
    def is_loaded(self) -> bool: ...
    def get_memory_usage(self) -> float: ...  # GPU memory in MB
```

**Memory Note:** PhoBERT requires ~2.5GB for model weights. The Docker container is capped at 4GB to include inference overhead. First start requires model download (~120s).

### 4.5 Sentiment Analyzer (`app/services/sentiment_analyzer.py`)

The main analysis pipeline applies PhoBERT + domain-specific knowledge:

```python
class SentimentAnalyzer:
    def __init__(self, model: PhoBERTSentiment, slang_dict_path=None):
        self.model = model
        self.slang_mappings = {}       # Loaded from slang_dictionary.json
        self.positive_keywords = []
        self.negative_keywords = []
        self.symbol_pattern = re.compile(r'\b([A-Z]{3})\b')  # VNM, FPT, etc.

    def preprocess_text(self, text: str) -> str:
        """Normalize whitespace, detect slang occurrences."""

    def extract_symbols(self, text: str) -> List[str]:
        """Extract 3-letter uppercase stock symbols from text."""

    def extract_keywords(self, text: str) -> List[str]:
        """Find positive/negative keywords and slang in text."""

    def adjust_confidence(self, base_sentiment, base_confidence, text) -> Tuple[str, float]:
        """Apply slang modifiers and keyword adjustments.
        - Slang modifiers: e.g., 'lùa gà' → -0.3, 'gom hàng' → +0.2
        - Keyword counts: positive vs negative keyword ratio
        - Can flip sentiment if adjustment is strong (> ±20)
        """

    def analyze(self, text: str) -> Dict[str, Any]:
        """Single text → {sentiment, confidence, symbols, keywords}"""

    def analyze_batch(self, texts: List[str]) -> List[Dict[str, Any]]:
        """Batch analysis with preprocessing + adjustment pipeline."""
```

### 4.6 Vietnamese Stock Slang Dictionary (`app/data/slang_dictionary.json`)

| Slang | Meaning | Sentiment Modifier |
|-------|---------|-------------------|
| lùa gà | pump and dump | -0.3 (Strongly Negative) |
| cá mập | whale/big investor | 0.0 (Neutral) |
| gom hàng | accumulating | +0.2 (Positive) |
| xả hàng | dumping shares | -0.2 (Negative) |
| bắt đáy | catching the bottom | +0.1 (Positive) |
| bán tháo | panic sell | -0.3 (Strongly Negative) |
| trần / tím | ceiling price / limit up | +0.2 / +0.3 |
| sàn / đỏ | floor price / down | -0.2 / -0.1 |
| giải chấp | margin call | -0.3 (Strongly Negative) |
| đội lái | market manipulator | -0.2 (Negative) |

### 4.7 FastAPI Application (`app/main.py`)

```python
@asynccontextmanager
async def lifespan(app: FastAPI):
    global model, analyzer
    settings = get_settings()
    logger.info("Loading PhoBERT model...")
    model = PhoBERTSentiment(model_name=settings.model_name, cache_dir=settings.model_cache_dir)
    analyzer = SentimentAnalyzer(model)
    logger.info("Model loaded, service ready")
    yield
    logger.info("Shutting down sentiment service")

app = FastAPI(
    title="VN Stock Sentiment Service",
    description="Vietnamese stock market sentiment analysis using PhoBERT",
    version="1.0.0",
    lifespan=lifespan,
)
app.add_middleware(CORSMiddleware, allow_origins=["*"], ...)
app.include_router(analyze.router)
app.include_router(health.router)
```

### 4.8 API Endpoints

#### POST `/analyze`

```json
// Request
{
  "texts": [
    {"id": "1", "content": "VNM công bố lợi nhuận tăng 20%", "source": "vneconomy"},
    {"id": "2", "content": "FPT bán tháo cổ phiếu", "source": "cafef"}
  ]
}

// Response
{
  "results": [
    {"id": "1", "sentiment": "positive", "confidence": 92.5, "symbols": ["VNM"], "keywords": ["lợi nhuận", "tăng"]},
    {"id": "2", "sentiment": "negative", "confidence": 85.0, "symbols": ["FPT"], "keywords": ["bán tháo"]}
  ],
  "processing_time_ms": 145.2,
  "model_version": "phobert-base-v1"
}
```

#### GET `/health`

```json
{
  "status": "healthy",
  "model_loaded": true,
  "model_name": "vinai/phobert-base",
  "memory_usage_mb": 2512.5
}
```

### 4.9 Dockerfile (`python-sentiment/Dockerfile`)

```dockerfile
FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y build-essential curl && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Model download is done at runtime (first start ~120s)
# Uncomment below to pre-download during build:
# RUN python -c "from transformers import AutoTokenizer, AutoModel; \
#     AutoTokenizer.from_pretrained('vinai/phobert-base'); \
#     AutoModel.from_pretrained('vinai/phobert-base')"

COPY app/ ./app/
RUN mkdir -p .cache

ENV PYTHONUNBUFFERED=1
ENV PORT=8000

HEALTHCHECK --interval=30s --timeout=10s --start-period=120s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

EXPOSE 8000
CMD ["python", "-m", "app.main"]
```

---

## 5. Inter-Service Communication

### 5.1 Communication Patterns

All inter-service communication uses **synchronous HTTP/JSON**:

```
┌──────────────┐  HTTP  ┌──────────────┐  HTTP  ┌──────────────┐
│   n8n        │──────→│  API Gateway  │──────→│  Technical   │
│   (:5678)    │       │   (:8080)     │       │  Agent       │
└──────────────┘       └──────┬────────┘       │  (Go :8081)  │
                              │                 └──────────────┘
                              │ HTTP
                              ▼
                       ┌──────────────┐
                       │  Sentiment   │
                       │  Agent       │
                       │  (Py :8000)  │
                       └──────────────┘
```

### 5.2 Service Discovery (Docker Compose)

Docker Compose provides DNS-based service discovery:

| Service Name (DNS) | Internal Port | External Port | URL from other containers |
|---------------------|---------------|---------------|--------------------------|
| `api-gateway` | 8080 | 8080 | `http://api-gateway:8080` |
| `technical-agent` | 8080 | 8081 | `http://technical-agent:8080` |
| `sentiment` | 8000 | 8000 | `http://sentiment:8000` |
| `forecast-agent` | 8080 | 8082 | `http://forecast-agent:8080` |
| `master-orchestrator` | 8080 | 8083 | `http://master-orchestrator:8080` |
| `postgres` | 5432 | 5432 | `postgres:5432` |
| `redis` | 6379 | 6379 | `redis:6379` |
| `n8n` | 5678 | 5678 | `http://n8n:5678` |

### 5.3 Request Flow: Technical Analysis

```
Client → API Gateway (GET /api/v1/technical/VNM)
         │
         ├─ Check Redis cache (technical:VNM:latest)
         │   └─ Cache HIT → Return cached result
         │
         ├─ Cache MISS → Fetch 100 days OHLCV from vnstock client
         │
         ├─ Concurrent goroutines:
         │   ├─ RSI(closes, 14)
         │   ├─ CalculateMACD(closes, 12, 26, 9)
         │   ├─ CalculateBollingerBands(closes, 20, 2.0)
         │   ├─ SMA(closes, 20)
         │   └─ EMA(closes, 12), EMA(closes, 26)
         │
         ├─ Generate signal + confidence
         ├─ Cache in Redis (TTL: 5 min)
         ├─ Store in PostgreSQL
         └─ Return TechnicalResult JSON
```

### 5.4 Request Flow: Sentiment Analysis

```
Client → API Gateway (POST /api/v1/sentiment)
         │
         └─ Proxy to Python Sentiment Service (POST http://sentiment:8000/analyze)
             │
             ├─ Preprocess Vietnamese text
             ├─ PhoBERT model.predict_batch(texts)
             ├─ Extract stock symbols (regex: [A-Z]{3})
             ├─ Extract keywords (positive + negative)
             ├─ Apply slang dictionary adjustments
             └─ Return SentimentResponse JSON
```

### 5.5 Async Communication (Future/Optional)

GCP Pub/Sub support is included in dependencies (`cloud.google.com/go/pubsub`, `google-cloud-pubsub`) but not actively used. When `GCP_PROJECT_ID` is configured, services can optionally publish/subscribe to topics for batch processing.

---

## 6. Docker Compose Infrastructure

### 6.1 Full Service Map

The `docker-compose.yml` defines all services with Docker Compose profiles:

| Service | Image | Port | Profile | Health Check |
|---------|-------|------|---------|-------------|
| `api-gateway` | `go.Dockerfile (api-gateway)` | 8080 | default | `curl /health` |
| `technical-agent` | `go.Dockerfile (technical-agent)` | 8081 | default | — |
| `forecast-agent` | `go.Dockerfile (forecast-agent)` | 8082 | default | — |
| `master-orchestrator` | `go.Dockerfile (master-orchestrator)` | 8083 | default | — |
| `sentiment` | `python-sentiment/Dockerfile` | 8000 | default | `curl /health` (start_period: 120s) |
| `n8n` | `n8nio/n8n:latest` | 5678 | default | — |
| `postgres` | `postgres:15-alpine` | 5432 | default | `pg_isready` |
| `redis` | `redis:7-alpine` | 6379 | default | `redis-cli ping` |
| `nginx` | `nginx:alpine` | 80, 443 | **production** | — |
| `prometheus` | `prom/prometheus:latest` | 9090 | **monitoring** | — |
| `grafana` | `grafana/grafana:latest` | 3001 | **monitoring** | — |
| `minio` | `minio/minio:latest` | 9000, 9001 | **development** | — |

### 6.2 Deployment Profiles

```bash
# Core services (default)
docker compose up -d

# Development (adds MinIO for local S3)
docker compose --profile development up -d

# Monitoring (adds Prometheus + Grafana)
docker compose --profile monitoring up -d

# Production (adds Nginx reverse proxy with SSL)
docker compose --profile production up -d

# Full stack
docker compose --profile production --profile monitoring up -d
```

### 6.3 Network Configuration

All services share a single Docker bridge network:

```yaml
networks:
  vnstock-net:
    driver: bridge
    ipam:
      config:
        - subnet: 172.28.0.0/16
```

### 6.4 Volume Persistence

```yaml
volumes:
  postgres_data:       # PostgreSQL data (persistent)
  redis_data:          # Redis AOF (persistent)
  n8n_data:            # n8n workflows and credentials
  sentiment_cache:     # PhoBERT model cache (~2.5GB)
  grafana_data:        # Grafana dashboards and data
  prometheus_data:     # Prometheus metrics
  minio_data:          # MinIO object storage
  nginx_logs:          # Nginx access/error logs
```

### 6.5 Go Service Dockerfile (`docker/go.Dockerfile`)

Multi-stage build producing a minimal distroless image:

```dockerfile
FROM golang:1.22-alpine AS builder
ARG SERVICE
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/service ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/service /service
EXPOSE 8080
ENTRYPOINT ["/service"]
```

**Build the same image for different services** by changing the `SERVICE` build arg:

```bash
docker build --build-arg SERVICE=api-gateway -t vnstock-api-gateway -f docker/go.Dockerfile go-services/
docker build --build-arg SERVICE=technical-agent -t vnstock-technical -f docker/go.Dockerfile go-services/
```

---

## 7. n8n Workflow Automation

### 7.1 Overview

n8n runs as a Docker container and orchestrates the daily analysis pipeline:

- **Container:** `vnstock-n8n` on port 5678
- **Database:** PostgreSQL (shared instance, separate `n8n` database)
- **Timezone:** `Asia/Ho_Chi_Minh` (UTC+7)
- **Auth:** Basic auth (`N8N_USER` / `N8N_PASSWORD`)

### 7.2 Daily Analysis Workflow (`n8n-workflow-daily-analysis.json`)

**Trigger:** Cron → 8:00 AM weekdays (Vietnam time)

**Pipeline:**

```
08:00 AM Trigger
    │
    ├─ Scrape RSS feeds (VnEconomy, CafeF, VietStock, Đầu tư)
    │
    ├─ Filter spam / irrelevant articles
    │   └─ Keyword blacklist: "quảng cáo", "tuyển dụng", etc.
    │
    ├─ Detect hot stocks (most-mentioned symbols)
    │
    ├─ Call API Gateway → POST /api/v1/sentiment (batch articles)
    │
    ├─ Call API Gateway → POST /api/v1/analyze (hot stocks)
    │
    ├─ Generate daily report (Markdown)
    │
    └─ Send report via Telegram Bot
```

### 7.3 Error Handler Workflow (`n8n-workflow-error-handler.json`)

Handles failures in the main workflow:
- Retry logic (max 3 attempts)
- Error notification via Telegram
- Error logging to PostgreSQL

### 7.4 Importing Workflows

```bash
# Workflows are auto-mounted via volume:
volumes:
  - ./workflows:/home/node/.n8n/workflows

# Or import manually via n8n UI:
# 1. Open http://localhost:5678
# 2. Go to Workflows → Import
# 3. Select n8n-workflow-daily-analysis.json
```

### 7.5 n8n Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `N8N_BASIC_AUTH_ACTIVE` | `true` | Enable authentication |
| `N8N_BASIC_AUTH_USER` | `admin` | n8n login username |
| `N8N_BASIC_AUTH_PASSWORD` | (required) | n8n login password |
| `GENERIC_TIMEZONE` | `Asia/Ho_Chi_Minh` | Workflow timezone |
| `DB_TYPE` | `postgresdb` | n8n storage backend |
| `DB_POSTGRESDB_HOST` | `postgres` | PostgreSQL host (Docker DNS) |
| `DB_POSTGRESDB_DATABASE` | `n8n` | Separate database for n8n |

---

## 8. Deployment

### 8.1 Production Deployment (Docker Compose)

```bash
# 1. Prepare environment
cp .env.example .env
# Set production values: strong passwords, real API keys, SSL domain

# 2. Start with production profile
docker compose --profile production up -d

# 3. Verify
docker compose ps
curl https://yourdomain.com/health
```

### 8.2 Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `POSTGRES_PASSWORD` | **Yes** | — | PostgreSQL password |
| `N8N_PASSWORD` | **Yes** | — | n8n authentication password |
| `TELEGRAM_BOT_TOKEN` | **Yes** | — | Telegram Bot API token |
| `REDIS_PASSWORD` | No | — | Redis auth (optional) |
| `N8N_USER` | No | `admin` | n8n authentication user |
| `N8N_HOST` | No | `localhost` | n8n hostname |
| `GCP_PROJECT_ID` | No | — | GCP project (optional Pub/Sub) |
| `ANTHROPIC_API_KEY` | No | — | Claude API key (AI synthesis) |
| `OPENAI_API_KEY` | No | — | OpenAI API key (alternative) |
| `GRAFANA_USER` | No | `admin` | Grafana admin user |
| `GRAFANA_PASSWORD` | No | — | Grafana admin password |
| `MINIO_ROOT_USER` | No | `minioadmin` | MinIO root user |
| `MINIO_ROOT_PASSWORD` | No | — | MinIO root password |
| `ENABLE_RUMOR_DETECTION` | No | `false` | Feature flag |
| `ENABLE_ML_FORECAST` | No | `false` | Feature flag |

### 8.3 SSL/TLS with Nginx (Production Profile)

The `nginx` service (production profile) provides:
- SSL termination with Let's Encrypt certificates
- Reverse proxy to API Gateway (port 8080) and n8n (port 5678)
- Static file serving for web dashboard

Configuration file: `nginx/nginx.conf` (mount as read-only volume)

### 8.4 Scaling Services

```bash
# Scale Go services horizontally (stateless)
docker compose up -d --scale technical-agent=3

# Note: Sentiment service is memory-heavy (4GB per instance)
# Scale cautiously:
docker compose up -d --scale sentiment=2  # Requires 8GB+ RAM
```

### 8.5 Backup Strategy

```bash
# PostgreSQL backup
docker exec vnstock-db pg_dump -U vnstock vnstock > backup_$(date +%F).sql

# Redis backup (AOF)
docker exec vnstock-cache redis-cli BGSAVE

# Restore PostgreSQL
cat backup_2026-02-24.sql | docker exec -i vnstock-db psql -U vnstock vnstock
```

### 8.6 Cost Estimate (Self-Hosted)

| Component | Specification | Est. Cost/Month |
|-----------|---------------|-----------------|
| VPS/Cloud VM | 4 vCPU, 16GB RAM | $40–80 |
| AWS S3 (optional) | 50GB Standard | $1–2 |
| Domain + SSL | Let's Encrypt | Free |
| **Total** | | **~$40–80/month** |

---

## 9. Monitoring & Operations

### 9.1 Health Checks

All services expose health endpoints:

| Service | Endpoint | Checks |
|---------|----------|--------|
| API Gateway | `GET :8080/health` | Service alive |
| API Gateway | `GET :8080/ready` | DB + Redis + Sentiment reachable |
| Sentiment | `GET :8000/health` | Model loaded, memory usage |
| PostgreSQL | `pg_isready -U vnstock` | Connection accepting |
| Redis | `redis-cli ping` | PONG response |

### 9.2 Docker Health Check Configuration

```yaml
# API Gateway
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 10s

# Sentiment Service (needs longer start_period for model loading)
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 120s

# PostgreSQL
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U vnstock -d vnstock"]
  interval: 10s
  timeout: 5s
  retries: 5

# Redis
healthcheck:
  test: ["CMD", "redis-cli", "ping"]
  interval: 10s
  timeout: 5s
  retries: 5
```

### 9.3 Logging

#### Go Services (Structured JSON)
```json
{
  "timestamp": "2026-02-24T08:15:00Z",
  "level": "info",
  "service": "api-gateway",
  "method": "GET",
  "path": "/api/v1/technical/VNM",
  "status": 200,
  "latency_ms": 245,
  "client_ip": "172.28.0.1"
}
```

#### Python Sentiment Service
```
2026-02-24 08:15:00 - sentiment_analyzer - INFO - Sentiment analysis completed
  request_id=abc123 texts_count=15 processing_time_ms=1250
```

#### View Logs
```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f api-gateway
docker compose logs -f sentiment

# Last 100 lines
docker compose logs --tail=100 api-gateway
```

### 9.4 Monitoring Stack (Profile: monitoring)

```bash
docker compose --profile monitoring up -d
```

- **Prometheus** (`localhost:9090`): Metrics collection from all services
- **Grafana** (`localhost:3001`): Pre-configured dashboards
  - API Gateway latency (p50, p95, p99)
  - Sentiment model inference time
  - Cache hit rate
  - Error rate by service
  - Container resource usage

### 9.5 Key Metrics to Monitor

| Metric | Target | Alert Threshold |
|--------|--------|-----------------|
| API response time (p95) | < 3s | > 5s |
| Sentiment inference time | < 1s/text | > 3s |
| Cache hit rate | > 80% | < 50% |
| Error rate | < 1% | > 5% |
| Sentiment service memory | < 3.5GB | > 3.8GB |
| PostgreSQL connections | < 80 | > 90 |
| Redis memory | < 256MB | > 512MB |
| Daily workflow completion | < 10 min | > 15 min |

---

## 10. Testing

### 10.1 Go Unit Tests

```bash
cd go-services

# Run all tests
make test

# Run indicator tests only
make test-indicators

# With verbose output
go test -v ./internal/indicators/...

# With coverage
go test -cover ./...
```

**Test file:** `internal/indicators/indicators_test.go`

Tests cover:
- RSI calculation accuracy
- MACD signal line convergence
- Bollinger Bands standard deviation
- SMA/EMA moving average calculations
- Edge cases: insufficient data, zero prices, single data point

### 10.2 Python Tests

```bash
cd python-sentiment

# Run tests
python -m pytest tests/ -v

# With coverage
python -m pytest --cov=app tests/
```

### 10.3 Integration Tests

```bash
# Start full stack
docker compose up -d

# Wait for services to be healthy
sleep 30

# Test API Gateway
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/technical/VNM

# Test Sentiment Service directly
curl -X POST http://localhost:8000/analyze \
  -H "Content-Type: application/json" \
  -d '{"texts": [{"id": "1", "content": "VNM tăng mạnh hôm nay"}]}'

# Test combined analysis
curl -X POST http://localhost:8080/api/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{"symbols": ["VNM", "FPT"], "include_sentiment": true}'
```

### 10.4 Load Testing

```bash
# Install k6 or use wrk
wrk -t4 -c100 -d30s http://localhost:8080/api/v1/technical/VNM
```

---

## 11. Troubleshooting

### 11.1 Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Sentiment service takes 2+ minutes to start | PhoBERT model download (~500MB) | Pre-download in Dockerfile; use `sentiment_cache` volume |
| API Gateway can't reach sentiment | Service not ready yet | Check `start_period: 120s` in health check |
| Redis connection refused | Redis not started or wrong password | Verify `REDIS_PASSWORD` in `.env` |
| PostgreSQL init.sql not running | Volume already has data | `docker volume rm analysis-stock_postgres_data` and restart |
| n8n workflow not triggering | Wrong timezone | Verify `GENERIC_TIMEZONE=Asia/Ho_Chi_Minh` |
| Go build fails in Docker | Missing `go.sum` | Run `go mod tidy` locally first |
| Rate limiting not working | Redis not connected | API Gateway starts without Redis (graceful degradation) |

### 11.2 Service Recovery

```bash
# Restart a single service
docker compose restart sentiment

# Rebuild and restart
docker compose up -d --build api-gateway

# Full reset (WARNING: deletes all data)
docker compose down -v
docker compose up -d
```

### 11.3 Debugging

```bash
# Enter a running container
docker exec -it vnstock-api-gateway sh

# Check PostgreSQL
docker exec -it vnstock-db psql -U vnstock -d vnstock -c "SELECT count(*) FROM technical_analysis;"

# Check Redis
docker exec -it vnstock-cache redis-cli keys "technical:*"

# Check network connectivity
docker exec vnstock-api-gateway curl http://sentiment:8000/health
```

---

## Appendix A: Database Schema (`scripts/init.sql`)

```sql
-- Database: vnstock (PostgreSQL 15)

CREATE TABLE stocks (
    symbol VARCHAR(10) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    exchange VARCHAR(10) NOT NULL CHECK (exchange IN ('HSX', 'HNX', 'UPCOM')),
    industry VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE technical_analysis (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    open_price DECIMAL(12, 2),
    high_price DECIMAL(12, 2),
    low_price DECIMAL(12, 2),
    close_price DECIMAL(12, 2),
    volume BIGINT,
    rsi_14 DECIMAL(5, 2),
    macd_line DECIMAL(10, 4),
    macd_signal DECIMAL(10, 4),
    macd_histogram DECIMAL(10, 4),
    bb_upper DECIMAL(12, 2),
    bb_middle DECIMAL(12, 2),
    bb_lower DECIMAL(12, 2),
    sma_20 DECIMAL(12, 2),
    sma_50 DECIMAL(12, 2),
    ema_12 DECIMAL(12, 2),
    ema_26 DECIMAL(12, 2),
    adx DECIMAL(5, 2),
    atr DECIMAL(12, 2),
    stoch_k DECIMAL(5, 2),
    stoch_d DECIMAL(5, 2),
    signal VARCHAR(15) CHECK (signal IN ('STRONG_BUY', 'BUY', 'HOLD', 'SELL', 'STRONG_SELL')),
    confidence DECIMAL(5, 2),
    score DECIMAL(5, 2),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(symbol, timestamp)
);

CREATE TABLE sentiment_analysis (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10),
    source_url TEXT,
    source_type VARCHAR(50),
    text_content TEXT,
    sentiment VARCHAR(20) CHECK (sentiment IN ('positive', 'negative', 'neutral')),
    confidence DECIMAL(5, 2),
    keywords TEXT[],
    published_at TIMESTAMPTZ,
    analyzed_at TIMESTAMPTZ DEFAULT NOW(),
    model_version VARCHAR(50) DEFAULT 'phobert-base'
);

CREATE TABLE forecasts (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    technical_score DECIMAL(5, 2),
    sentiment_score DECIMAL(5, 2),
    market_score DECIMAL(5, 2),
    recommendation VARCHAR(20),
    confidence DECIMAL(5, 2),
    support_price DECIMAL(12, 2),
    resistance_price DECIMAL(12, 2),
    reasoning TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE daily_reports (
    id BIGSERIAL PRIMARY KEY,
    report_date DATE UNIQUE NOT NULL,
    total_symbols_analyzed INT,
    buy_signals INT,
    sell_signals INT,
    hold_signals INT,
    top_picks JSONB,
    market_summary TEXT,
    report_json JSONB,
    report_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_technical_symbol_time ON technical_analysis(symbol, timestamp DESC);
CREATE INDEX idx_sentiment_symbol ON sentiment_analysis(symbol);
CREATE INDEX idx_sentiment_analyzed ON sentiment_analysis(analyzed_at DESC);
CREATE INDEX idx_forecast_symbol_time ON forecasts(symbol, timestamp DESC);
```

## Appendix B: Redis Cache Key Schema

```
# Technical analysis
technical:{symbol}:latest          → JSON TechnicalResult (TTL: 5 min)
technical:{symbol}:{date}          → JSON (TTL: 24 hours)

# Sentiment cache
sentiment:{symbol}:latest          → JSON SentimentResult (TTL: 1 hour)
sentiment:batch:{hash}             → JSON batch results (TTL: 1 hour)

# Rate limiting
ratelimit:{client_ip}:{window}     → Counter (TTL: 1 min)

# Job tracking
job:{job_id}                       → JSON status (TTL: 1 hour)

# Daily reports
report:daily:{date}                → JSON DailyReport (TTL: 24 hours)
```

## Appendix C: API Quick Reference

```
# Health
GET  :8080/health                        → API Gateway liveness
GET  :8080/ready                         → API Gateway readiness (all deps)
GET  :8000/health                        → Sentiment service + model status

# Technical Analysis
GET  :8080/api/v1/technical/{symbol}     → Single stock analysis
POST :8080/api/v1/technical/batch        → Batch analysis (max 50)
     Body: {"symbols": ["VNM", "FPT", "VIC"]}

# Sentiment Analysis
POST :8080/api/v1/sentiment              → Proxy to Python service
     Body: {"texts": [{"id": "1", "content": "..."}]}

# Combined Analysis
POST :8080/api/v1/analyze                → Technical + Sentiment
     Body: {"symbols": ["VNM"], "include_sentiment": true}

# Direct Sentiment Service
POST :8000/analyze                       → Direct sentiment analysis
GET  :8000/health                        → Model health + memory usage
```

## Appendix D: Document Cross-References

| Document | Location | Description |
|----------|----------|-------------|
| **SRS v2.0** | `docs/SRS-VNStock-Analysis-System-v2.0.md` | Full requirements specification |
| Project Introduction | `docs/introduce/PROJECT_INTRODUCTION.md` | Project overview (bilingual EN/VI) |
| Architecture Overview | `docs/introduce/ARCHITECTURE_OVERVIEW.md` | Reference architecture analysis |
| CLAUDE.md | `CLAUDE.md` | AI coding assistant context |
| README.md | `README.md` | Quick start and overview |
| n8n Best Practices | `docs/n8n-crawling-best-practices.md` | n8n workflow guidelines |

### Superseded Documents

The following v1.0 documents are **superseded** by this guide and the SRS v2.0:

| Document | Status |
|----------|--------|
| `docs/Implementation-Guide-VNStock-v1.0.md` | Superseded (Python/AWS) |
| `docs/Implementation-Guide-Go-GCP-v1.0.md` | Superseded (Go/GCP) |
| `docs/Implementation-Guide-Hybrid-v1.0.md` | Superseded (Hybrid/GCP Cloud Run) |
| `docs/SRS-VNStock-Analysis-System-v1.0.md` | Superseded |
| `docs/SRS-VNStock-Go-GCP-v1.0.md` | Superseded |
| `docs/SRS-VNStock-Hybrid-v1.0.md` | Superseded |

---

**End of Implementation Guide v2.0**
