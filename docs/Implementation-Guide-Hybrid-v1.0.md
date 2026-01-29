# Implementation Guide
# VN Stock Analysis System - Hybrid Architecture
## Version 1.0

---

## Table of Contents

1. [Project Structure](#1-project-structure)
2. [Development Environment](#2-development-environment)
3. [Go Services Implementation](#3-go-services-implementation)
4. [Python Sentiment Service](#4-python-sentiment-service)
5. [Inter-Service Communication](#5-inter-service-communication)
6. [Infrastructure (Terraform)](#6-infrastructure-terraform)
7. [Deployment](#7-deployment)
8. [Monitoring & Operations](#8-monitoring--operations)

---

## 1. Project Structure

```
vnstock-hybrid/
├── go-services/                    # All Go microservices
│   ├── cmd/
│   │   ├── api-gateway/
│   │   │   └── main.go
│   │   ├── technical-agent/
│   │   │   └── main.go
│   │   ├── forecast-agent/
│   │   │   └── main.go
│   │   └── master-orchestrator/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── database/
│   │   │   ├── postgres.go
│   │   │   └── redis.go
│   │   ├── handlers/
│   │   │   ├── analysis.go
│   │   │   ├── technical.go
│   │   │   └── health.go
│   │   ├── indicators/
│   │   │   ├── rsi.go
│   │   │   ├── macd.go
│   │   │   ├── bollinger.go
│   │   │   ├── sma.go
│   │   │   ├── ema.go
│   │   │   └── indicators_test.go
│   │   ├── models/
│   │   │   ├── stock.go
│   │   │   ├── analysis.go
│   │   │   └── forecast.go
│   │   ├── pubsub/
│   │   │   ├── publisher.go
│   │   │   └── subscriber.go
│   │   ├── services/
│   │   │   ├── technical_service.go
│   │   │   ├── sentiment_client.go
│   │   │   ├── forecast_service.go
│   │   │   └── market_data.go
│   │   └── middleware/
│   │       ├── auth.go
│   │       ├── ratelimit.go
│   │       └── logging.go
│   ├── pkg/
│   │   └── vnstock/
│   │       └── client.go           # Vietnamese stock data client
│   ├── go.mod
│   ├── go.sum
│   └── Makefile
│
├── python-sentiment/               # Python sentiment service
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py
│   │   ├── config.py
│   │   ├── models/
│   │   │   ├── __init__.py
│   │   │   └── phobert.py
│   │   ├── services/
│   │   │   ├── __init__.py
│   │   │   ├── sentiment_analyzer.py
│   │   │   └── symbol_extractor.py
│   │   ├── routers/
│   │   │   ├── __init__.py
│   │   │   ├── analyze.py
│   │   │   └── health.py
│   │   ├── pubsub/
│   │   │   ├── __init__.py
│   │   │   └── subscriber.py
│   │   └── data/
│   │       └── slang_dictionary.json
│   ├── requirements.txt
│   ├── Dockerfile
│   └── pyproject.toml
│
├── terraform/                      # GCP infrastructure
│   ├── main.tf
│   ├── variables.tf
│   ├── outputs.tf
│   ├── modules/
│   │   ├── cloud-run/
│   │   ├── cloud-sql/
│   │   ├── memorystore/
│   │   ├── pubsub/
│   │   └── storage/
│   └── environments/
│       ├── dev.tfvars
│       └── prod.tfvars
│
├── workflows/                      # Cloud Workflows
│   ├── daily-analysis.yaml
│   └── batch-analysis.yaml
│
├── docker/
│   ├── go.Dockerfile
│   └── python.Dockerfile
│
├── scripts/
│   ├── deploy.sh
│   ├── local-dev.sh
│   └── migrate.sh
│
├── docker-compose.yaml             # Local development
└── README.md
```

---

## 2. Development Environment

### 2.1 Prerequisites

```bash
# Go 1.22+
go version  # go1.22.0 or higher

# Python 3.11+
python3 --version  # 3.11.x

# Docker & Docker Compose
docker --version
docker-compose --version

# Google Cloud SDK
gcloud --version

# Terraform
terraform --version  # 1.5+
```

### 2.2 Local Development Setup

```bash
# Clone repository
git clone https://github.com/your-org/vnstock-hybrid.git
cd vnstock-hybrid

# Start local infrastructure
docker-compose up -d postgres redis

# Go services
cd go-services
go mod download
make run-api-gateway    # Terminal 1
make run-technical      # Terminal 2
make run-forecast       # Terminal 3
make run-orchestrator   # Terminal 4

# Python sentiment service
cd python-sentiment
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python -m app.main  # Terminal 5
```

### 2.3 Docker Compose (Local)

```yaml
# docker-compose.yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: vnstock
      POSTGRES_USER: vnstock
      POSTGRES_PASSWORD: vnstock123
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  # Go API Gateway
  api-gateway:
    build:
      context: ./go-services
      dockerfile: ../docker/go.Dockerfile
      args:
        SERVICE: api-gateway
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - SENTIMENT_SERVICE_URL=http://sentiment:8000
    depends_on:
      - postgres
      - redis

  # Go Technical Agent
  technical-agent:
    build:
      context: ./go-services
      dockerfile: ../docker/go.Dockerfile
      args:
        SERVICE: technical-agent
    ports:
      - "8081:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  # Go Forecast Agent
  forecast-agent:
    build:
      context: ./go-services
      dockerfile: ../docker/go.Dockerfile
      args:
        SERVICE: forecast-agent
    ports:
      - "8082:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  # Go Master Orchestrator
  master-orchestrator:
    build:
      context: ./go-services
      dockerfile: ../docker/go.Dockerfile
      args:
        SERVICE: master-orchestrator
    ports:
      - "8083:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - TECHNICAL_SERVICE_URL=http://technical-agent:8080
      - SENTIMENT_SERVICE_URL=http://sentiment:8000
      - FORECAST_SERVICE_URL=http://forecast-agent:8080
    depends_on:
      - postgres
      - redis
      - technical-agent
      - sentiment
      - forecast-agent

  # Python Sentiment Service
  sentiment:
    build:
      context: ./python-sentiment
      dockerfile: Dockerfile
    ports:
      - "8000:8000"
    environment:
      - REDIS_HOST=redis
      - MODEL_NAME=vinai/phobert-base
    volumes:
      - model_cache:/app/.cache
    deploy:
      resources:
        limits:
          memory: 4G

volumes:
  postgres_data:
  redis_data:
  model_cache:
```

---

## 3. Go Services Implementation

### 3.1 Project Configuration

```go
// go-services/internal/config/config.go
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
	ProjectID string
}

type ServicesConfig struct {
	TechnicalURL  string
	SentimentURL  string
	ForecastURL   string
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
			TechnicalURL:  getEnv("TECHNICAL_SERVICE_URL", "http://localhost:8081"),
			SentimentURL:  getEnv("SENTIMENT_SERVICE_URL", "http://localhost:8000"),
			ForecastURL:   getEnv("FORECAST_SERVICE_URL", "http://localhost:8082"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
```

### 3.2 Data Models

```go
// go-services/internal/models/stock.go
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
	EMA12         *float64 `gorm:"type:decimal(12,2)" json:"ema_12"`
	EMA26         *float64 `gorm:"type:decimal(12,2)" json:"ema_26"`
	ADX           *float64 `gorm:"type:decimal(5,2)" json:"adx"`
	ATR           *float64 `gorm:"type:decimal(12,2)" json:"atr"`

	// Signal
	Signal     string   `gorm:"size:10" json:"signal"`
	Confidence *float64 `gorm:"type:decimal(5,2)" json:"confidence"`

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
	)
}
```

### 3.3 Technical Indicators

```go
// go-services/internal/indicators/rsi.go
package indicators

import "math"

// RSI calculates the Relative Strength Index
func RSI(closes []float64, period int) float64 {
	if len(closes) < period+1 {
		return 0
	}

	var gains, losses float64

	// Calculate initial average gain/loss
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

	// Apply smoothing for remaining periods
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

// RSISeries calculates RSI for entire price series
func RSISeries(closes []float64, period int) []float64 {
	if len(closes) < period+1 {
		return nil
	}

	result := make([]float64, len(closes))
	for i := 0; i < period; i++ {
		result[i] = math.NaN()
	}

	for i := period; i < len(closes); i++ {
		result[i] = RSI(closes[:i+1], period)
	}

	return result
}
```

```go
// go-services/internal/indicators/macd.go
package indicators

// MACD represents MACD indicator values
type MACD struct {
	MACDLine   float64
	SignalLine float64
	Histogram  float64
}

// CalculateMACD calculates MACD indicator
func CalculateMACD(closes []float64, fastPeriod, slowPeriod, signalPeriod int) *MACD {
	if len(closes) < slowPeriod+signalPeriod {
		return nil
	}

	emaFast := EMA(closes, fastPeriod)
	emaSlow := EMA(closes, slowPeriod)

	if len(emaFast) == 0 || len(emaSlow) == 0 {
		return nil
	}

	// Calculate MACD line (Fast EMA - Slow EMA)
	macdLine := make([]float64, len(closes))
	startIdx := slowPeriod - 1 // Start from where slow EMA is valid

	for i := startIdx; i < len(closes); i++ {
		macdLine[i] = emaFast[i] - emaSlow[i]
	}

	// Calculate Signal line (EMA of MACD line)
	validMACD := macdLine[startIdx:]
	signalEMA := EMA(validMACD, signalPeriod)

	if len(signalEMA) == 0 {
		return nil
	}

	lastMACD := macdLine[len(macdLine)-1]
	lastSignal := signalEMA[len(signalEMA)-1]

	return &MACD{
		MACDLine:   lastMACD,
		SignalLine: lastSignal,
		Histogram:  lastMACD - lastSignal,
	}
}
```

```go
// go-services/internal/indicators/bollinger.go
package indicators

import "math"

// BollingerBands represents Bollinger Bands values
type BollingerBands struct {
	Upper  float64
	Middle float64
	Lower  float64
}

// CalculateBollingerBands calculates Bollinger Bands
func CalculateBollingerBands(closes []float64, period int, stdDevMultiplier float64) *BollingerBands {
	if len(closes) < period {
		return nil
	}

	// Calculate SMA (middle band)
	sma := SMA(closes, period)
	if len(sma) == 0 {
		return nil
	}
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

```go
// go-services/internal/indicators/sma.go
package indicators

// SMA calculates Simple Moving Average
func SMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}

	result := make([]float64, len(prices))

	// Calculate first SMA
	var sum float64
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	// Calculate remaining SMAs using sliding window
	for i := period; i < len(prices); i++ {
		sum = sum - prices[i-period] + prices[i]
		result[i] = sum / float64(period)
	}

	return result
}
```

```go
// go-services/internal/indicators/ema.go
package indicators

// EMA calculates Exponential Moving Average
func EMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}

	result := make([]float64, len(prices))
	multiplier := 2.0 / float64(period+1)

	// First EMA is SMA
	var sum float64
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	// Calculate EMA for remaining values
	for i := period; i < len(prices); i++ {
		result[i] = (prices[i]-result[i-1])*multiplier + result[i-1]
	}

	return result
}
```

### 3.4 Technical Analysis Service

```go
// go-services/internal/services/technical_service.go
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

type TechnicalService struct {
	db           *gorm.DB
	redis        *redis.Client
	marketClient *vnstock.Client
}

type TechnicalResult struct {
	Symbol     string                 `json:"symbol"`
	Timestamp  time.Time              `json:"timestamp"`
	RSI        float64                `json:"rsi"`
	MACD       *indicators.MACD       `json:"macd"`
	Bollinger  *indicators.BollingerBands `json:"bollinger"`
	SMA20      float64                `json:"sma_20"`
	EMA12      float64                `json:"ema_12"`
	EMA26      float64                `json:"ema_26"`
	Signal     string                 `json:"signal"`
	Confidence float64                `json:"confidence"`
}

func NewTechnicalService(db *gorm.DB, redis *redis.Client, client *vnstock.Client) *TechnicalService {
	return &TechnicalService{
		db:           db,
		redis:        redis,
		marketClient: client,
	}
}

func (s *TechnicalService) Analyze(ctx context.Context, symbol string) (*TechnicalResult, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("technical:%s:latest", symbol)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var result TechnicalResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			return &result, nil
		}
	}

	// Fetch historical data
	history, err := s.marketClient.GetHistoricalData(ctx, symbol, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market data: %w", err)
	}

	closes := extractCloses(history)
	if len(closes) < 26 {
		return nil, fmt.Errorf("insufficient data for analysis")
	}

	// Calculate indicators concurrently
	var wg sync.WaitGroup
	var rsiVal float64
	var macdVal *indicators.MACD
	var bbVal *indicators.BollingerBands
	var sma20Val, ema12Val, ema26Val float64

	wg.Add(5)

	go func() {
		defer wg.Done()
		rsiVal = indicators.RSI(closes, 14)
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
		sma := indicators.SMA(closes, 20)
		if len(sma) > 0 {
			sma20Val = sma[len(sma)-1]
		}
	}()

	go func() {
		defer wg.Done()
		ema12 := indicators.EMA(closes, 12)
		ema26 := indicators.EMA(closes, 26)
		if len(ema12) > 0 {
			ema12Val = ema12[len(ema12)-1]
		}
		if len(ema26) > 0 {
			ema26Val = ema26[len(ema26)-1]
		}
	}()

	wg.Wait()

	// Generate signal
	signal, confidence := s.generateSignal(rsiVal, macdVal, bbVal, closes[len(closes)-1])

	result := &TechnicalResult{
		Symbol:     symbol,
		Timestamp:  time.Now(),
		RSI:        rsiVal,
		MACD:       macdVal,
		Bollinger:  bbVal,
		SMA20:      sma20Val,
		EMA12:      ema12Val,
		EMA26:      ema26Val,
		Signal:     signal,
		Confidence: confidence,
	}

	// Cache result
	if data, err := json.Marshal(result); err == nil {
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	// Store in database
	s.storeResult(ctx, result, history)

	return result, nil
}

func (s *TechnicalService) AnalyzeBatch(ctx context.Context, symbols []string) (map[string]*TechnicalResult, error) {
	results := make(map[string]*TechnicalResult)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(symbols))

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
				errChan <- fmt.Errorf("%s: %w", symbol, err)
				return
			}

			mu.Lock()
			results[symbol] = result
			mu.Unlock()
		}(sym)
	}

	wg.Wait()
	close(errChan)

	return results, nil
}

func (s *TechnicalService) generateSignal(rsi float64, macd *indicators.MACD, bb *indicators.BollingerBands, price float64) (string, float64) {
	var buySignals, sellSignals int
	var totalConfidence float64

	// RSI signals
	if rsi < 30 {
		buySignals++
		totalConfidence += (30 - rsi) / 30 * 100
	} else if rsi > 70 {
		sellSignals++
		totalConfidence += (rsi - 70) / 30 * 100
	}

	// MACD signals
	if macd != nil {
		if macd.Histogram > 0 && macd.MACDLine > macd.SignalLine {
			buySignals++
			totalConfidence += 70
		} else if macd.Histogram < 0 && macd.MACDLine < macd.SignalLine {
			sellSignals++
			totalConfidence += 70
		}
	}

	// Bollinger Bands signals
	if bb != nil {
		if price < bb.Lower {
			buySignals++
			totalConfidence += 60
		} else if price > bb.Upper {
			sellSignals++
			totalConfidence += 60
		}
	}

	// Determine final signal
	signalCount := buySignals + sellSignals
	if signalCount == 0 {
		return "HOLD", 50.0
	}

	avgConfidence := totalConfidence / float64(signalCount)
	if buySignals > sellSignals {
		return "BUY", avgConfidence
	} else if sellSignals > buySignals {
		return "SELL", avgConfidence
	}

	return "HOLD", 50.0
}

func (s *TechnicalService) storeResult(ctx context.Context, result *TechnicalResult, history []vnstock.OHLCV) {
	if len(history) == 0 {
		return
	}

	latest := history[len(history)-1]
	analysis := &models.TechnicalAnalysis{
		Symbol:        result.Symbol,
		Timestamp:     result.Timestamp,
		OpenPrice:     latest.Open,
		HighPrice:     latest.High,
		LowPrice:      latest.Low,
		ClosePrice:    latest.Close,
		Volume:        latest.Volume,
		RSI14:         &result.RSI,
		SMA20:         &result.SMA20,
		EMA12:         &result.EMA12,
		EMA26:         &result.EMA26,
		Signal:        result.Signal,
		Confidence:    &result.Confidence,
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

	s.db.Create(analysis)
}

func extractCloses(history []vnstock.OHLCV) []float64 {
	closes := make([]float64, len(history))
	for i, h := range history {
		closes[i] = h.Close
	}
	return closes
}
```

### 3.5 Sentiment Client (Go → Python)

```go
// go-services/internal/services/sentiment_client.go
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SentimentClient struct {
	baseURL    string
	httpClient *http.Client
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
	Sentiment  string   `json:"sentiment"`
	Confidence float64  `json:"confidence"`
	Symbols    []string `json:"symbols"`
	Keywords   []string `json:"keywords"`
}

func NewSentimentClient(baseURL string) *SentimentClient {
	return &SentimentClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *SentimentClient) Analyze(ctx context.Context, texts []TextItem) (*SentimentResponse, error) {
	reqBody := SentimentRequest{Texts: texts}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/analyze", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sentiment service returned status %d", resp.StatusCode)
	}

	var result SentimentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *SentimentClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sentiment service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
```

### 3.6 API Gateway

```go
// go-services/cmd/api-gateway/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"vnstock-hybrid/internal/config"
	"vnstock-hybrid/internal/handlers"
	"vnstock-hybrid/internal/middleware"
	"vnstock-hybrid/internal/models"
	"vnstock-hybrid/internal/services"
	"vnstock-hybrid/pkg/vnstock"
)

func main() {
	cfg := config.Load()

	// Database connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := models.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize services
	marketClient := vnstock.NewClient()
	technicalSvc := services.NewTechnicalService(db, rdb, marketClient)
	sentimentClient := services.NewSentimentClient(cfg.Services.SentimentURL)

	// Setup Gin
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RateLimiter(rdb, 100, time.Minute))

	// Health endpoints
	r.GET("/health", handlers.HealthCheck(db, rdb))
	r.GET("/ready", handlers.ReadinessCheck(db, rdb, sentimentClient))

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Technical analysis
		v1.GET("/technical/:symbol", handlers.TechnicalAnalysis(technicalSvc))
		v1.POST("/technical/batch", handlers.TechnicalBatch(technicalSvc))

		// Sentiment (proxy to Python)
		v1.POST("/sentiment", handlers.SentimentProxy(sentimentClient))

		// Combined analysis
		v1.POST("/analyze", handlers.FullAnalysis(technicalSvc, sentimentClient))

		// Reports
		v1.GET("/reports/daily", handlers.DailyReport(db))
	}

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("API Gateway starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
```

### 3.7 Handlers

```go
// go-services/internal/handlers/analysis.go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"vnstock-hybrid/internal/services"
)

type AnalyzeRequest struct {
	Symbols          []string `json:"symbols" binding:"required,min=1,max=50"`
	IncludeSentiment bool     `json:"include_sentiment"`
	IncludeForecast  bool     `json:"include_forecast"`
}

type AnalyzeResponse struct {
	RequestID string                       `json:"request_id"`
	Timestamp string                       `json:"timestamp"`
	Results   map[string]*AnalysisResult   `json:"results"`
}

type AnalysisResult struct {
	Symbol    string                       `json:"symbol"`
	Technical *services.TechnicalResult    `json:"technical,omitempty"`
	Sentiment *services.SentimentResult    `json:"sentiment,omitempty"`
}

func TechnicalAnalysis(svc *services.TechnicalService) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Param("symbol")
		if len(symbol) != 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid symbol format"})
			return
		}

		result, err := svc.Analyze(c.Request.Context(), symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func TechnicalBatch(svc *services.TechnicalService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Symbols []string `json:"symbols" binding:"required,min=1,max=50"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		results, err := svc.AnalyzeBatch(c.Request.Context(), req.Symbols)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"results": results})
	}
}

func SentimentProxy(client *services.SentimentClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Texts []services.TextItem `json:"texts" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := client.Analyze(c.Request.Context(), req.Texts)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func FullAnalysis(techSvc *services.TechnicalService, sentClient *services.SentimentClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AnalyzeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		// Get technical analysis
		techResults, err := techSvc.AnalyzeBatch(ctx, req.Symbols)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		results := make(map[string]*AnalysisResult)
		for symbol, tech := range techResults {
			results[symbol] = &AnalysisResult{
				Symbol:    symbol,
				Technical: tech,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"request_id": c.GetString("request_id"),
			"results":    results,
		})
	}
}
```

---

## 4. Python Sentiment Service

### 4.1 Project Structure

```
python-sentiment/
├── app/
│   ├── __init__.py
│   ├── main.py
│   ├── config.py
│   ├── models/
│   │   ├── __init__.py
│   │   └── phobert.py
│   ├── services/
│   │   ├── __init__.py
│   │   ├── sentiment_analyzer.py
│   │   └── symbol_extractor.py
│   ├── routers/
│   │   ├── __init__.py
│   │   ├── analyze.py
│   │   └── health.py
│   └── data/
│       └── slang_dictionary.json
├── requirements.txt
├── Dockerfile
└── pyproject.toml
```

### 4.2 Requirements

```txt
# requirements.txt
fastapi==0.109.0
uvicorn[standard]==0.27.0
transformers==4.37.0
torch==2.1.2
numpy==1.26.3
pydantic==2.5.3
python-dotenv==1.0.0
redis==5.0.1
google-cloud-pubsub==2.19.0
```

### 4.3 Configuration

```python
# app/config.py
from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    # Server
    host: str = "0.0.0.0"
    port: int = 8000
    debug: bool = False

    # Model
    model_name: str = "vinai/phobert-base"
    model_cache_dir: str = "./.cache"
    max_sequence_length: int = 256

    # Redis
    redis_host: str = "localhost"
    redis_port: int = 6379
    redis_password: str = ""
    redis_db: int = 0

    # GCP (for Pub/Sub)
    gcp_project_id: str = ""
    pubsub_subscription: str = "sentiment-requests-sub"
    pubsub_result_topic: str = "sentiment-results"

    class Config:
        env_file = ".env"


@lru_cache()
def get_settings() -> Settings:
    return Settings()
```

### 4.4 PhoBERT Model

```python
# app/models/phobert.py
import logging
from typing import List, Tuple

import torch
from transformers import AutoTokenizer, AutoModelForSequenceClassification

logger = logging.getLogger(__name__)


class PhoBERTSentiment:
    """Vietnamese sentiment analysis using PhoBERT."""

    def __init__(self, model_name: str = "vinai/phobert-base", cache_dir: str = "./.cache"):
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        logger.info(f"Using device: {self.device}")

        # Load tokenizer and model
        logger.info(f"Loading model: {model_name}")
        self.tokenizer = AutoTokenizer.from_pretrained(
            model_name,
            cache_dir=cache_dir
        )

        # For sentiment, we use a fine-tuned version or base model with classification head
        # In production, use a fine-tuned sentiment model
        self.model = AutoModelForSequenceClassification.from_pretrained(
            model_name,
            num_labels=3,  # positive, negative, neutral
            cache_dir=cache_dir
        )
        self.model.to(self.device)
        self.model.eval()

        self.labels = ["negative", "neutral", "positive"]
        self._model_loaded = True
        logger.info("Model loaded successfully")

    @property
    def is_loaded(self) -> bool:
        return self._model_loaded

    def predict(self, text: str, max_length: int = 256) -> Tuple[str, float]:
        """Predict sentiment for a single text."""
        with torch.no_grad():
            inputs = self.tokenizer(
                text,
                return_tensors="pt",
                truncation=True,
                max_length=max_length,
                padding=True
            ).to(self.device)

            outputs = self.model(**inputs)
            probabilities = torch.softmax(outputs.logits, dim=1)

            predicted_idx = torch.argmax(probabilities, dim=1).item()
            confidence = probabilities[0][predicted_idx].item() * 100

            return self.labels[predicted_idx], confidence

    def predict_batch(self, texts: List[str], max_length: int = 256, batch_size: int = 16) -> List[Tuple[str, float]]:
        """Predict sentiment for multiple texts."""
        results = []

        for i in range(0, len(texts), batch_size):
            batch_texts = texts[i:i + batch_size]

            with torch.no_grad():
                inputs = self.tokenizer(
                    batch_texts,
                    return_tensors="pt",
                    truncation=True,
                    max_length=max_length,
                    padding=True
                ).to(self.device)

                outputs = self.model(**inputs)
                probabilities = torch.softmax(outputs.logits, dim=1)

                for j in range(len(batch_texts)):
                    predicted_idx = torch.argmax(probabilities[j]).item()
                    confidence = probabilities[j][predicted_idx].item() * 100
                    results.append((self.labels[predicted_idx], confidence))

        return results

    def get_memory_usage(self) -> float:
        """Get current GPU memory usage in MB."""
        if torch.cuda.is_available():
            return torch.cuda.memory_allocated() / 1024 / 1024
        return 0.0
```

### 4.5 Slang Dictionary

```json
{
  "slang_mappings": {
    "lùa gà": {"meaning": "pump and dump scam", "sentiment_modifier": -0.3},
    "cá mập": {"meaning": "whale/big investor", "sentiment_modifier": 0.0},
    "gom hàng": {"meaning": "accumulating shares", "sentiment_modifier": 0.2},
    "xả hàng": {"meaning": "dumping shares", "sentiment_modifier": -0.2},
    "tay to": {"meaning": "major player", "sentiment_modifier": 0.0},
    "đánh úp": {"meaning": "surprise dump", "sentiment_modifier": -0.3},
    "bắt đáy": {"meaning": "catching the bottom", "sentiment_modifier": 0.1},
    "cắt lỗ": {"meaning": "stop loss", "sentiment_modifier": -0.1},
    "trần": {"meaning": "ceiling price", "sentiment_modifier": 0.2},
    "sàn": {"meaning": "floor price", "sentiment_modifier": -0.2},
    "xanh": {"meaning": "green/up", "sentiment_modifier": 0.1},
    "đỏ": {"meaning": "red/down", "sentiment_modifier": -0.1},
    "tím": {"meaning": "purple/limit up", "sentiment_modifier": 0.3},
    "xanh lơ": {"meaning": "ceiling limit", "sentiment_modifier": 0.2},
    "la liệt": {"meaning": "widespread red", "sentiment_modifier": -0.2},
    "bơm thổi": {"meaning": "pump artificially", "sentiment_modifier": -0.2},
    "mua đuổi": {"meaning": "chasing buy", "sentiment_modifier": 0.1},
    "bán tháo": {"meaning": "panic sell", "sentiment_modifier": -0.3},
    "đội lái": {"meaning": "market manipulator team", "sentiment_modifier": -0.2},
    "giải chấp": {"meaning": "margin call", "sentiment_modifier": -0.3}
  },
  "positive_keywords": [
    "tăng", "lợi nhuận", "kỷ lục", "đột phá", "thắng lớn",
    "doanh thu tăng", "tích cực", "triển vọng tốt", "mua mạnh",
    "vượt kỳ vọng", "tăng trưởng", "cổ tức cao"
  ],
  "negative_keywords": [
    "giảm", "thua lỗ", "sụt giảm", "bán tháo", "sập sàn",
    "nợ xấu", "phá sản", "cảnh báo", "điều tra", "gian lận",
    "thao túng", "vi phạm", "đình chỉ"
  ]
}
```

### 4.6 Sentiment Analyzer Service

```python
# app/services/sentiment_analyzer.py
import json
import logging
import re
from pathlib import Path
from typing import List, Dict, Any, Optional

from app.models.phobert import PhoBERTSentiment

logger = logging.getLogger(__name__)


class SentimentAnalyzer:
    def __init__(self, model: PhoBERTSentiment, slang_dict_path: Optional[str] = None):
        self.model = model
        self.slang_mappings = {}
        self.positive_keywords = []
        self.negative_keywords = []

        # Load slang dictionary
        dict_path = slang_dict_path or Path(__file__).parent.parent / "data" / "slang_dictionary.json"
        self._load_slang_dictionary(dict_path)

        # Symbol pattern for Vietnamese stocks (3 uppercase letters)
        self.symbol_pattern = re.compile(r'\b([A-Z]{3})\b')

    def _load_slang_dictionary(self, path: str):
        try:
            with open(path, 'r', encoding='utf-8') as f:
                data = json.load(f)
                self.slang_mappings = data.get("slang_mappings", {})
                self.positive_keywords = data.get("positive_keywords", [])
                self.negative_keywords = data.get("negative_keywords", [])
                logger.info(f"Loaded {len(self.slang_mappings)} slang mappings")
        except Exception as e:
            logger.warning(f"Failed to load slang dictionary: {e}")

    def preprocess_text(self, text: str) -> str:
        """Preprocess Vietnamese text for sentiment analysis."""
        # Normalize whitespace
        text = " ".join(text.split())

        # Replace slang with standard terms (optional, for logging)
        for slang, info in self.slang_mappings.items():
            if slang in text.lower():
                logger.debug(f"Found slang: {slang} -> {info['meaning']}")

        return text

    def extract_symbols(self, text: str) -> List[str]:
        """Extract stock symbols from text."""
        symbols = self.symbol_pattern.findall(text)
        # Filter to only valid Vietnamese stock symbols (simple validation)
        valid_symbols = [s for s in symbols if len(s) == 3]
        return list(set(valid_symbols))

    def extract_keywords(self, text: str) -> List[str]:
        """Extract relevant keywords from text."""
        keywords = []
        text_lower = text.lower()

        for kw in self.positive_keywords + self.negative_keywords:
            if kw in text_lower:
                keywords.append(kw)

        for slang in self.slang_mappings:
            if slang in text_lower:
                keywords.append(slang)

        return keywords

    def adjust_confidence(self, base_sentiment: str, base_confidence: float, text: str) -> tuple:
        """Adjust sentiment based on slang and keywords."""
        text_lower = text.lower()
        adjustment = 0.0

        # Apply slang modifiers
        for slang, info in self.slang_mappings.items():
            if slang in text_lower:
                adjustment += info.get("sentiment_modifier", 0) * 100

        # Apply keyword modifiers
        pos_count = sum(1 for kw in self.positive_keywords if kw in text_lower)
        neg_count = sum(1 for kw in self.negative_keywords if kw in text_lower)
        adjustment += (pos_count - neg_count) * 5

        # Adjust confidence
        new_confidence = min(100, max(0, base_confidence + adjustment))

        # Potentially flip sentiment if adjustment is strong
        if adjustment < -20 and base_sentiment == "positive":
            return "neutral", new_confidence
        elif adjustment > 20 and base_sentiment == "negative":
            return "neutral", new_confidence

        return base_sentiment, new_confidence

    def analyze(self, text: str) -> Dict[str, Any]:
        """Analyze sentiment of a single text."""
        processed_text = self.preprocess_text(text)
        base_sentiment, base_confidence = self.model.predict(processed_text)

        # Adjust based on domain knowledge
        sentiment, confidence = self.adjust_confidence(
            base_sentiment, base_confidence, text
        )

        return {
            "sentiment": sentiment,
            "confidence": round(confidence, 2),
            "symbols": self.extract_symbols(text),
            "keywords": self.extract_keywords(text)
        }

    def analyze_batch(self, texts: List[str]) -> List[Dict[str, Any]]:
        """Analyze sentiment of multiple texts."""
        processed_texts = [self.preprocess_text(t) for t in texts]
        predictions = self.model.predict_batch(processed_texts)

        results = []
        for i, (sentiment, confidence) in enumerate(predictions):
            adj_sentiment, adj_confidence = self.adjust_confidence(
                sentiment, confidence, texts[i]
            )
            results.append({
                "sentiment": adj_sentiment,
                "confidence": round(adj_confidence, 2),
                "symbols": self.extract_symbols(texts[i]),
                "keywords": self.extract_keywords(texts[i])
            })

        return results
```

### 4.7 FastAPI Application

```python
# app/main.py
import logging
import sys
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import get_settings
from app.models.phobert import PhoBERTSentiment
from app.services.sentiment_analyzer import SentimentAnalyzer
from app.routers import analyze, health

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)]
)
logger = logging.getLogger(__name__)

# Global model instance
model: PhoBERTSentiment = None
analyzer: SentimentAnalyzer = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Startup and shutdown events."""
    global model, analyzer
    settings = get_settings()

    # Load model on startup
    logger.info("Loading PhoBERT model...")
    model = PhoBERTSentiment(
        model_name=settings.model_name,
        cache_dir=settings.model_cache_dir
    )
    analyzer = SentimentAnalyzer(model)
    logger.info("Model loaded, service ready")

    yield

    # Cleanup on shutdown
    logger.info("Shutting down sentiment service")


def create_app() -> FastAPI:
    settings = get_settings()

    app = FastAPI(
        title="VN Stock Sentiment Service",
        description="Vietnamese stock market sentiment analysis using PhoBERT",
        version="1.0.0",
        lifespan=lifespan
    )

    # CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # Include routers
    app.include_router(analyze.router)
    app.include_router(health.router)

    return app


def get_analyzer() -> SentimentAnalyzer:
    return analyzer


def get_model() -> PhoBERTSentiment:
    return model


app = create_app()

if __name__ == "__main__":
    import uvicorn
    settings = get_settings()
    uvicorn.run(
        "app.main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug
    )
```

### 4.8 API Routers

```python
# app/routers/analyze.py
from typing import List, Optional
from datetime import datetime

from fastapi import APIRouter, HTTPException, Depends
from pydantic import BaseModel, Field

from app.main import get_analyzer
from app.services.sentiment_analyzer import SentimentAnalyzer

router = APIRouter(tags=["analyze"])


class TextItem(BaseModel):
    id: str
    content: str
    source: Optional[str] = None
    published_at: Optional[datetime] = None


class AnalyzeRequest(BaseModel):
    texts: List[TextItem] = Field(..., min_length=1, max_length=100)


class SentimentResultItem(BaseModel):
    id: str
    sentiment: str
    confidence: float
    symbols: List[str]
    keywords: List[str]


class AnalyzeResponse(BaseModel):
    results: List[SentimentResultItem]
    processing_time_ms: float
    model_version: str = "phobert-base-v1"


@router.post("/analyze", response_model=AnalyzeResponse)
async def analyze_sentiment(request: AnalyzeRequest):
    """Analyze sentiment of Vietnamese texts."""
    analyzer = get_analyzer()
    if analyzer is None:
        raise HTTPException(status_code=503, detail="Model not loaded")

    start_time = datetime.now()

    texts = [item.content for item in request.texts]
    analyses = analyzer.analyze_batch(texts)

    results = []
    for i, analysis in enumerate(analyses):
        results.append(SentimentResultItem(
            id=request.texts[i].id,
            sentiment=analysis["sentiment"],
            confidence=analysis["confidence"],
            symbols=analysis["symbols"],
            keywords=analysis["keywords"]
        ))

    processing_time = (datetime.now() - start_time).total_seconds() * 1000

    return AnalyzeResponse(
        results=results,
        processing_time_ms=round(processing_time, 2)
    )


@router.post("/analyze/single")
async def analyze_single(text: str):
    """Quick endpoint for single text analysis."""
    analyzer = get_analyzer()
    if analyzer is None:
        raise HTTPException(status_code=503, detail="Model not loaded")

    result = analyzer.analyze(text)
    return result
```

```python
# app/routers/health.py
from fastapi import APIRouter

from app.main import get_model

router = APIRouter(tags=["health"])


@router.get("/health")
async def health_check():
    """Health check endpoint."""
    model = get_model()

    return {
        "status": "healthy" if model and model.is_loaded else "unhealthy",
        "model_loaded": model.is_loaded if model else False,
        "model_name": "vinai/phobert-base",
        "memory_usage_mb": model.get_memory_usage() if model else 0
    }


@router.get("/ready")
async def readiness_check():
    """Readiness check for Kubernetes."""
    model = get_model()

    if not model or not model.is_loaded:
        return {"ready": False}, 503

    return {"ready": True}
```

### 4.9 Dockerfile

```dockerfile
# python-sentiment/Dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first for caching
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Pre-download model during build
RUN python -c "from transformers import AutoTokenizer, AutoModel; \
    AutoTokenizer.from_pretrained('vinai/phobert-base'); \
    AutoModel.from_pretrained('vinai/phobert-base')"

# Copy application code
COPY app/ ./app/

# Create cache directory
RUN mkdir -p .cache

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV PORT=8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8000/health')" || exit 1

EXPOSE 8000

CMD ["python", "-m", "app.main"]
```

---

## 5. Inter-Service Communication

### 5.1 Synchronous (HTTP)

Go API Gateway calls Python Sentiment Service via HTTP:

```go
// Example from Go
result, err := sentimentClient.Analyze(ctx, []TextItem{
    {ID: "1", Content: "VNM công bố lợi nhuận tăng 20%"},
})
```

### 5.2 Asynchronous (Pub/Sub)

For batch processing, use Pub/Sub:

```go
// go-services/internal/pubsub/publisher.go
package pubsub

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"
)

type Publisher struct {
	client *pubsub.Client
	topics map[string]*pubsub.Topic
}

func NewPublisher(ctx context.Context, projectID string) (*Publisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		client: client,
		topics: make(map[string]*pubsub.Topic),
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, topicID string, data interface{}) (string, error) {
	topic, ok := p.topics[topicID]
	if !ok {
		topic = p.client.Topic(topicID)
		p.topics[topicID] = topic
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data: jsonData,
	})

	return result.Get(ctx)
}
```

```python
# app/pubsub/subscriber.py
import asyncio
import json
import logging
from concurrent.futures import TimeoutError

from google.cloud import pubsub_v1

from app.config import get_settings
from app.main import get_analyzer

logger = logging.getLogger(__name__)


class SentimentSubscriber:
    def __init__(self):
        settings = get_settings()
        self.subscriber = pubsub_v1.SubscriberClient()
        self.subscription_path = self.subscriber.subscription_path(
            settings.gcp_project_id,
            settings.pubsub_subscription
        )
        self.publisher = pubsub_v1.PublisherClient()
        self.result_topic = self.publisher.topic_path(
            settings.gcp_project_id,
            settings.pubsub_result_topic
        )

    def process_message(self, message):
        """Process incoming sentiment request."""
        try:
            data = json.loads(message.data.decode("utf-8"))
            analyzer = get_analyzer()

            texts = [item["content"] for item in data.get("texts", [])]
            results = analyzer.analyze_batch(texts)

            # Publish results
            response = {
                "request_id": data.get("request_id"),
                "correlation_id": data.get("correlation_id"),
                "results": [
                    {
                        "id": data["texts"][i]["id"],
                        **results[i]
                    }
                    for i in range(len(results))
                ]
            }

            self.publisher.publish(
                self.result_topic,
                json.dumps(response).encode("utf-8")
            )

            message.ack()
            logger.info(f"Processed request {data.get('request_id')}")

        except Exception as e:
            logger.error(f"Error processing message: {e}")
            message.nack()

    def start(self):
        """Start listening for messages."""
        streaming_pull = self.subscriber.subscribe(
            self.subscription_path,
            callback=self.process_message
        )
        logger.info(f"Listening on {self.subscription_path}")

        with self.subscriber:
            try:
                streaming_pull.result()
            except TimeoutError:
                streaming_pull.cancel()
                streaming_pull.result()
```

---

## 6. Infrastructure (Terraform)

### 6.1 Main Configuration

```hcl
# terraform/main.tf
terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  backend "gcs" {
    bucket = "vnstock-terraform-state"
    prefix = "hybrid"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Enable required APIs
resource "google_project_service" "services" {
  for_each = toset([
    "run.googleapis.com",
    "sqladmin.googleapis.com",
    "redis.googleapis.com",
    "pubsub.googleapis.com",
    "secretmanager.googleapis.com",
    "cloudbuild.googleapis.com",
    "artifactregistry.googleapis.com",
  ])
  service            = each.key
  disable_on_destroy = false
}

# Artifact Registry for container images
resource "google_artifact_registry_repository" "vnstock" {
  location      = var.region
  repository_id = "vnstock-hybrid"
  format        = "DOCKER"
}

# VPC for private services
resource "google_compute_network" "vpc" {
  name                    = "vnstock-vpc"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "subnet" {
  name          = "vnstock-subnet"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
  network       = google_compute_network.vpc.id
}

# Cloud SQL
module "cloud_sql" {
  source     = "./modules/cloud-sql"
  project_id = var.project_id
  region     = var.region
  network_id = google_compute_network.vpc.id
}

# Memorystore (Redis)
module "memorystore" {
  source     = "./modules/memorystore"
  project_id = var.project_id
  region     = var.region
  network_id = google_compute_network.vpc.id
}

# Pub/Sub Topics
module "pubsub" {
  source     = "./modules/pubsub"
  project_id = var.project_id
}

# Cloud Storage
module "storage" {
  source     = "./modules/storage"
  project_id = var.project_id
  region     = var.region
}

# Cloud Run Services
module "cloud_run" {
  source              = "./modules/cloud-run"
  project_id          = var.project_id
  region              = var.region
  vpc_connector       = google_vpc_access_connector.connector.id
  db_connection_name  = module.cloud_sql.connection_name
  redis_host          = module.memorystore.host
  artifact_registry   = google_artifact_registry_repository.vnstock.name
}

# VPC Connector for Cloud Run -> VPC
resource "google_vpc_access_connector" "connector" {
  name          = "vnstock-connector"
  region        = var.region
  ip_cidr_range = "10.8.0.0/28"
  network       = google_compute_network.vpc.name
}
```

### 6.2 Cloud Run Module

```hcl
# terraform/modules/cloud-run/main.tf
variable "project_id" {}
variable "region" {}
variable "vpc_connector" {}
variable "db_connection_name" {}
variable "redis_host" {}
variable "artifact_registry" {}

# Go API Gateway
resource "google_cloud_run_v2_service" "api_gateway" {
  name     = "api-gateway"
  location = var.region

  template {
    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${var.artifact_registry}/api-gateway:latest"

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      env {
        name  = "SENTIMENT_SERVICE_URL"
        value = google_cloud_run_v2_service.sentiment.uri
      }

      env {
        name  = "REDIS_HOST"
        value = var.redis_host
      }

      env {
        name = "DB_PASSWORD"
        value_source {
          secret_key_ref {
            secret  = "db-password"
            version = "latest"
          }
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    vpc_access {
      connector = var.vpc_connector
      egress    = "PRIVATE_RANGES_ONLY"
    }
  }
}

# Go Technical Agent
resource "google_cloud_run_v2_service" "technical" {
  name     = "technical-agent"
  location = var.region

  template {
    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${var.artifact_registry}/technical-agent:latest"

      resources {
        limits = {
          cpu    = "2"
          memory = "1Gi"
        }
      }

      env {
        name  = "REDIS_HOST"
        value = var.redis_host
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 5
    }

    vpc_access {
      connector = var.vpc_connector
      egress    = "PRIVATE_RANGES_ONLY"
    }
  }
}

# Python Sentiment Agent
resource "google_cloud_run_v2_service" "sentiment" {
  name     = "sentiment-agent"
  location = var.region

  template {
    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${var.artifact_registry}/sentiment-agent:latest"

      resources {
        limits = {
          cpu    = "2"
          memory = "4Gi"
        }
      }

      startup_probe {
        http_get {
          path = "/health"
        }
        initial_delay_seconds = 30
        period_seconds        = 10
        failure_threshold     = 10
      }

      env {
        name  = "MODEL_NAME"
        value = "vinai/phobert-base"
      }
    }

    # Keep warm for faster response (model loading takes time)
    scaling {
      min_instance_count = 1
      max_instance_count = 3
    }

    timeout = "300s"  # 5 minutes for model loading

    vpc_access {
      connector = var.vpc_connector
      egress    = "PRIVATE_RANGES_ONLY"
    }
  }
}

# Allow API Gateway to call other services
resource "google_cloud_run_v2_service_iam_member" "sentiment_invoker" {
  location = var.region
  name     = google_cloud_run_v2_service.sentiment.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_cloud_run_v2_service.api_gateway.template[0].service_account}"
}

output "api_gateway_url" {
  value = google_cloud_run_v2_service.api_gateway.uri
}

output "sentiment_url" {
  value = google_cloud_run_v2_service.sentiment.uri
}
```

---

## 7. Deployment

### 7.1 Build and Deploy Script

```bash
#!/bin/bash
# scripts/deploy.sh

set -e

PROJECT_ID=${GCP_PROJECT_ID:-"your-project-id"}
REGION=${GCP_REGION:-"asia-southeast1"}
REGISTRY="${REGION}-docker.pkg.dev/${PROJECT_ID}/vnstock-hybrid"

echo "Building and deploying VN Stock Hybrid..."

# Build Go services
echo "Building Go services..."
cd go-services

for service in api-gateway technical-agent forecast-agent master-orchestrator; do
    echo "Building $service..."
    docker build \
        --build-arg SERVICE=$service \
        -t ${REGISTRY}/${service}:latest \
        -f ../docker/go.Dockerfile \
        .
    docker push ${REGISTRY}/${service}:latest
done

cd ..

# Build Python sentiment service
echo "Building Python sentiment service..."
cd python-sentiment
docker build -t ${REGISTRY}/sentiment-agent:latest .
docker push ${REGISTRY}/sentiment-agent:latest
cd ..

# Deploy with Terraform
echo "Deploying infrastructure..."
cd terraform
terraform init
terraform apply -var="project_id=${PROJECT_ID}" -var="region=${REGION}" -auto-approve

# Output URLs
echo ""
echo "Deployment complete!"
terraform output
```

### 7.2 Go Dockerfile

```dockerfile
# docker/go.Dockerfile
FROM golang:1.22-alpine AS builder

ARG SERVICE

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/service ./cmd/${SERVICE}

# Final image
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/service /service

EXPOSE 8080

ENTRYPOINT ["/service"]
```

---

## 8. Monitoring & Operations

### 8.1 Health Checks

All services expose:
- `/health` - Liveness check
- `/ready` - Readiness check (dependencies verified)

### 8.2 Logging

```go
// Structured logging in Go
log.Info("Analysis completed",
    zap.String("symbol", symbol),
    zap.Duration("duration", duration),
    zap.String("signal", result.Signal),
)
```

```python
# Structured logging in Python
logger.info("Sentiment analysis completed",
    extra={
        "request_id": request_id,
        "texts_count": len(texts),
        "processing_time_ms": duration
    }
)
```

### 8.3 Metrics

Key metrics to monitor:
- Request latency (p50, p95, p99)
- Error rate
- Model inference time (Python)
- Cache hit rate (Redis)
- Pub/Sub message backlog

### 8.4 Alerts

```yaml
# Cloud Monitoring alert policies
- displayName: "High Error Rate - API Gateway"
  conditions:
    - conditionThreshold:
        filter: 'resource.type="cloud_run_revision" AND metric.type="run.googleapis.com/request_count"'
        comparison: COMPARISON_GT
        thresholdValue: 0.05  # 5% error rate
        duration: 300s

- displayName: "Sentiment Service Down"
  conditions:
    - conditionThreshold:
        filter: 'resource.type="cloud_run_revision" AND resource.labels.service_name="sentiment-agent"'
        comparison: COMPARISON_LT
        thresholdValue: 1  # No healthy instances
        duration: 60s
```

---

## Summary

This hybrid architecture provides:

1. **Performance**: Go handles high-throughput API and technical analysis
2. **ML Capability**: Python with PhoBERT for Vietnamese NLP
3. **Scalability**: Independent scaling of each service
4. **Cost Efficiency**: Go services scale to 0, Python stays warm for fast inference
5. **Maintainability**: Clear separation of concerns

**Estimated Monthly Cost**: $160-210/month (lower than full GCP due to Go efficiency)

**Next Steps**:
1. Set up GCP project and enable APIs
2. Deploy infrastructure with Terraform
3. Build and push container images
4. Configure Cloud Scheduler for daily analysis
5. Set up monitoring dashboards and alerts
