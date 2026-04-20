# Software Requirements Specification (SRS)
# VN Stock Analysis System - Go/GCP Edition
## Version 1.0

**Document ID:** SRS-VNSTOCK-GO-GCP-001
**Date:** January 2026
**Status:** Draft
**Technology Stack:** Go 1.22+, Google Cloud Platform

---

## Table of Contents

1. [Introduction](#1-introduction)
2. [Overall Description](#2-overall-description)
3. [External Interface Requirements](#3-external-interface-requirements)
4. [Functional Requirements](#4-functional-requirements)
5. [Non-Functional Requirements](#5-non-functional-requirements)
6. [Data Requirements](#6-data-requirements)
7. [System Constraints](#7-system-constraints)
8. [GCP Service Mapping](#8-gcp-service-mapping)
9. [Appendices](#9-appendices)

---

## 1. Introduction

### 1.1 Purpose

This SRS describes the Vietnamese Stock Market Analysis System refactored to use **Go** as the primary backend language and **Google Cloud Platform (GCP)** as the cloud infrastructure provider.

### 1.2 Scope

**System Name:** VN Stock Analysis System (Go/GCP Edition)

**Technology Stack:**
| Layer | Technology |
|-------|------------|
| Backend API | Go 1.22+ with Gin/Echo framework |
| Database | Cloud SQL (PostgreSQL) |
| Cache | Memorystore (Redis) |
| Object Storage | Cloud Storage (GCS) |
| Message Queue | Cloud Pub/Sub |
| Workflow | Cloud Workflows + Cloud Scheduler |
| Serverless | Cloud Run / Cloud Functions |
| ML/AI | Vertex AI |
| Monitoring | Cloud Monitoring + Cloud Logging |
| Container Registry | Artifact Registry |
| Secret Management | Secret Manager |
| Load Balancing | Cloud Load Balancing |
| CDN | Cloud CDN |

### 1.3 Definitions and Abbreviations

| Term | Definition |
|------|------------|
| GCP | Google Cloud Platform |
| GCS | Google Cloud Storage |
| GKE | Google Kubernetes Engine |
| Cloud Run | Serverless container platform |
| Pub/Sub | Google Cloud Pub/Sub messaging |
| Vertex AI | GCP's unified ML platform |
| Cloud SQL | Managed relational database |
| Memorystore | Managed Redis/Memcached |
| HSX/HNX/UPCOM | Vietnamese stock exchanges |

### 1.4 References

1. Go Programming Language (https://go.dev)
2. Google Cloud Documentation (https://cloud.google.com/docs)
3. Gin Web Framework (https://gin-gonic.com)
4. GORM (https://gorm.io)
5. Vietnamese Securities Regulations

---

## 2. Overall Description

### 2.1 System Architecture (GCP)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         GOOGLE CLOUD PLATFORM                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐              │
│  │ Cloud        │    │ Cloud        │    │ External     │              │
│  │ Scheduler    │───▶│ Workflows    │───▶│ Data Sources │              │
│  │ (Cron)       │    │ (Orchestrate)│    │ (RSS, APIs)  │              │
│  └──────────────┘    └──────┬───────┘    └──────────────┘              │
│                             │                                           │
│                             ▼                                           │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                     CLOUD PUB/SUB                                 │  │
│  │   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐            │  │
│  │   │ news    │  │analysis │  │ alerts  │  │ reports │            │  │
│  │   │ topic   │  │ topic   │  │ topic   │  │ topic   │            │  │
│  │   └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘            │  │
│  └────────┼────────────┼────────────┼────────────┼──────────────────┘  │
│           │            │            │            │                      │
│           ▼            ▼            ▼            ▼                      │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                        CLOUD RUN SERVICES                         │  │
│  │                                                                   │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────────┐ │  │
│  │  │ API Gateway │ │ Technical   │ │ Sentiment   │ │ Forecast   │ │  │
│  │  │ Service     │ │ Agent       │ │ Agent       │ │ Agent      │ │  │
│  │  │ (Go/Gin)    │ │ (Go)        │ │ (Go+Vertex) │ │ (Go)       │ │  │
│  │  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └─────┬──────┘ │  │
│  │         │               │               │              │         │  │
│  └─────────┼───────────────┼───────────────┼──────────────┼─────────┘  │
│            │               │               │              │            │
│            ▼               ▼               ▼              ▼            │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                      DATA LAYER                                   │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │  │
│  │  │ Cloud SQL   │  │ Memorystore │  │ Cloud       │              │  │
│  │  │ (PostgreSQL)│  │ (Redis)     │  │ Storage     │              │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                    DISTRIBUTION LAYER                             │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │  │
│  │  │ Telegram    │  │ Firebase    │  │ SendGrid    │              │  │
│  │  │ Bot (Go)    │  │ Hosting     │  │ (Email)     │              │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Product Functions

| ID | Function | GCP Service | Go Package |
|----|----------|-------------|------------|
| F1 | Data Collection | Cloud Functions + Pub/Sub | `net/http`, `encoding/xml` |
| F2 | Technical Analysis | Cloud Run | Custom Go package |
| F3 | Sentiment Analysis | Cloud Run + Vertex AI | `cloud.google.com/go/aiplatform` |
| F4 | Trend Forecasting | Cloud Run | Custom Go package |
| F5 | Report Generation | Cloud Run | `html/template`, `text/template` |
| F6 | Real-time Alerting | Pub/Sub + Cloud Functions | `cloud.google.com/go/pubsub` |
| F7 | Web Dashboard | Firebase Hosting | Next.js (unchanged) |
| F8 | Telegram Bot | Cloud Run | `go-telegram-bot-api` |
| F9 | API Services | Cloud Run + API Gateway | Gin framework |

### 2.3 User Classes

Same as original SRS:
- UC1: Retail Investors
- UC2: Active Traders
- UC3: System Administrators
- UC4: API Consumers

### 2.4 Operating Environment

| Component | GCP Service | Specification |
|-----------|-------------|---------------|
| Compute | Cloud Run | vCPU: 1-4, Memory: 512MB-8GB |
| Database | Cloud SQL | PostgreSQL 15, db-f1-micro to db-n1-standard-4 |
| Cache | Memorystore | Redis 7.0, 1-5GB |
| Storage | Cloud Storage | Standard/Nearline classes |
| Region | asia-southeast1 | Singapore (closest to Vietnam) |
| Backup Region | asia-east1 | Taiwan |

### 2.5 Design Constraints

1. **Go Idioms:** Follow Go best practices (effective Go, Go proverbs)
2. **GCP Native:** Use GCP managed services where possible
3. **12-Factor App:** Cloud-native application design
4. **Stateless Services:** All Cloud Run services must be stateless
5. **Vietnamese Market Hours:** 9:00-15:00 Vietnam time (UTC+7)

---

## 3. External Interface Requirements

### 3.1 User Interfaces

#### UI-1: Web Dashboard
- **Hosting:** Firebase Hosting with Cloud CDN
- **Framework:** Next.js 14 (SSR on Cloud Run)
- **API Calls:** Via Cloud Endpoints / API Gateway

#### UI-2: Telegram Bot
- **Runtime:** Cloud Run (Go)
- **Library:** `github.com/go-telegram-bot-api/telegram-bot-api/v5`
- **Webhook:** Cloud Run URL with HTTPS

### 3.2 Software Interfaces

#### SI-1: Market Data API (Go)

```go
// internal/market/client.go
type MarketDataClient interface {
    GetHistoricalData(ctx context.Context, symbol string, days int) ([]OHLCV, error)
    GetRealTimeQuote(ctx context.Context, symbol string) (*Quote, error)
    ListSymbols(ctx context.Context, exchange string) ([]Stock, error)
}

type OHLCV struct {
    Date   time.Time
    Open   float64
    High   float64
    Low    float64
    Close  float64
    Volume int64
}
```

#### SI-2: Cloud Storage Interface

```go
// internal/storage/gcs.go
type StorageClient interface {
    Upload(ctx context.Context, bucket, object string, data []byte) error
    Download(ctx context.Context, bucket, object string) ([]byte, error)
    List(ctx context.Context, bucket, prefix string) ([]string, error)
    Delete(ctx context.Context, bucket, object string) error
}
```

#### SI-3: Pub/Sub Interface

```go
// internal/pubsub/publisher.go
type Publisher interface {
    Publish(ctx context.Context, topic string, data []byte, attrs map[string]string) (string, error)
}

type Subscriber interface {
    Subscribe(ctx context.Context, subscription string, handler MessageHandler) error
}

type MessageHandler func(ctx context.Context, msg *pubsub.Message) error
```

### 3.3 GCP Service Interfaces

| Service | Go Package | Purpose |
|---------|------------|---------|
| Cloud SQL | `database/sql` + `github.com/lib/pq` | PostgreSQL access |
| Memorystore | `github.com/redis/go-redis/v9` | Redis caching |
| Cloud Storage | `cloud.google.com/go/storage` | Object storage |
| Pub/Sub | `cloud.google.com/go/pubsub` | Messaging |
| Secret Manager | `cloud.google.com/go/secretmanager` | Secrets |
| Vertex AI | `cloud.google.com/go/aiplatform` | ML inference |
| Cloud Logging | `cloud.google.com/go/logging` | Structured logs |

---

## 4. Functional Requirements

### 4.1 Data Collection Module (FR-DC)

#### FR-DC-001: RSS Feed Scraping (Cloud Function)

```go
// functions/scraper/main.go
type RSSScraperConfig struct {
    Feeds []struct {
        Name string `json:"name"`
        URL  string `json:"url"`
    } `json:"feeds"`
    OutputBucket string `json:"output_bucket"`
    PubSubTopic  string `json:"pubsub_topic"`
}

func ScrapeFeedsHandler(ctx context.Context, e event.Event) error {
    // Triggered by Cloud Scheduler
    // 1. Fetch RSS feeds concurrently using goroutines
    // 2. Parse XML
    // 3. Store raw data in GCS
    // 4. Publish to Pub/Sub for processing
}
```

**GCP Resources:**
- Cloud Function (Go 1.22 runtime)
- Cloud Scheduler (0 8 * * 1-5 Asia/Ho_Chi_Minh)
- Cloud Storage bucket: `vnstock-raw-data`
- Pub/Sub topic: `news-ingested`

#### FR-DC-002: Spam Filtering

```go
// internal/filter/spam.go
var spamKeywords = []string{
    "khuyến mãi", "khuyen mai",
    "đăng ký ngay", "dang ky ngay",
    "group vip", "bảo lãi",
}

func IsSpam(text string) bool {
    lower := strings.ToLower(text)
    for _, kw := range spamKeywords {
        if strings.Contains(lower, kw) {
            return true
        }
    }
    return false
}
```

### 4.2 Technical Analysis Agent (FR-TA)

#### FR-TA-001: Technical Analysis Service (Cloud Run)

```go
// cmd/technical-agent/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "vnstock/internal/analysis/technical"
)

func main() {
    r := gin.Default()

    agent := technical.NewAgent()

    r.POST("/analyze", agent.HandleAnalyze)
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy"})
    })

    r.Run(":8080")
}
```

```go
// internal/analysis/technical/agent.go
package technical

type Agent struct {
    marketClient market.Client
    cache        cache.Client
}

type AnalysisRequest struct {
    Symbol string `json:"symbol" binding:"required,len=3,alpha"`
    Days   int    `json:"days" binding:"min=30,max=365"`
}

type AnalysisResult struct {
    Symbol          string                 `json:"symbol"`
    Timestamp       time.Time              `json:"timestamp"`
    CurrentPrice    float64                `json:"current_price"`
    Recommendation  Recommendation         `json:"recommendation"`
    Confidence      float64                `json:"confidence"`
    TechnicalScore  float64                `json:"technical_score"`
    Signals         []string               `json:"signals"`
    Indicators      map[string]float64     `json:"indicators"`
    SupportResist   SupportResistance      `json:"support_resistance"`
    PriceTargets    PriceTargets           `json:"price_targets"`
}

func (a *Agent) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error) {
    // 1. Fetch historical data
    data, err := a.marketClient.GetHistoricalData(ctx, req.Symbol, req.Days)
    if err != nil {
        return nil, fmt.Errorf("fetch data: %w", err)
    }

    // 2. Calculate indicators concurrently
    var wg sync.WaitGroup
    indicators := make(map[string]float64)

    wg.Add(7)
    go a.calculateRSI(data, indicators, &wg)
    go a.calculateMACD(data, indicators, &wg)
    go a.calculateBollingerBands(data, indicators, &wg)
    go a.calculateSMA(data, indicators, &wg)
    go a.calculateStochastic(data, indicators, &wg)
    go a.calculateADX(data, indicators, &wg)
    go a.calculateVolume(data, indicators, &wg)
    wg.Wait()

    // 3. Generate signals
    signals := a.generateSignals(indicators)

    // 4. Calculate recommendation
    result := a.buildResult(req.Symbol, data, indicators, signals)

    return result, nil
}
```

#### FR-TA-002: Indicator Calculations (Go)

```go
// internal/analysis/technical/indicators.go
package technical

// RSI calculates Relative Strength Index
func RSI(closes []float64, period int) float64 {
    if len(closes) < period+1 {
        return 50.0
    }

    var gains, losses float64
    for i := 1; i <= period; i++ {
        change := closes[len(closes)-i] - closes[len(closes)-i-1]
        if change > 0 {
            gains += change
        } else {
            losses -= change
        }
    }

    avgGain := gains / float64(period)
    avgLoss := losses / float64(period)

    if avgLoss == 0 {
        return 100.0
    }

    rs := avgGain / avgLoss
    return 100.0 - (100.0 / (1.0 + rs))
}

// MACD calculates Moving Average Convergence Divergence
func MACD(closes []float64) (macd, signal, histogram float64) {
    ema12 := EMA(closes, 12)
    ema26 := EMA(closes, 26)
    macd = ema12 - ema26

    // Signal line is 9-period EMA of MACD
    // Simplified for this example
    signal = macd * 0.9
    histogram = macd - signal

    return
}

// BollingerBands calculates upper, middle, lower bands
func BollingerBands(closes []float64, period int) (upper, middle, lower float64) {
    middle = SMA(closes, period)
    stdDev := StandardDeviation(closes, period)
    upper = middle + 2*stdDev
    lower = middle - 2*stdDev
    return
}

// SMA calculates Simple Moving Average
func SMA(data []float64, period int) float64 {
    if len(data) < period {
        return 0
    }
    sum := 0.0
    for i := len(data) - period; i < len(data); i++ {
        sum += data[i]
    }
    return sum / float64(period)
}

// EMA calculates Exponential Moving Average
func EMA(data []float64, period int) float64 {
    if len(data) < period {
        return 0
    }
    multiplier := 2.0 / float64(period+1)
    ema := SMA(data[:period], period)

    for i := period; i < len(data); i++ {
        ema = (data[i]-ema)*multiplier + ema
    }
    return ema
}
```

### 4.3 Sentiment Analysis Agent (FR-SA)

#### FR-SA-001: Sentiment Service with Vertex AI

```go
// cmd/sentiment-agent/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "vnstock/internal/analysis/sentiment"
)

func main() {
    r := gin.Default()

    agent, err := sentiment.NewAgent(sentiment.Config{
        ProjectID:  os.Getenv("GCP_PROJECT"),
        Location:   "asia-southeast1",
        ModelID:    "phobert-sentiment", // Custom trained model
    })
    if err != nil {
        log.Fatal(err)
    }

    r.POST("/analyze", agent.HandleAnalyze)
    r.POST("/detect-rumors", agent.HandleDetectRumors)
    r.GET("/health", healthHandler)

    r.Run(":8080")
}
```

```go
// internal/analysis/sentiment/agent.go
package sentiment

import (
    aiplatform "cloud.google.com/go/aiplatform/apiv1"
)

type Agent struct {
    client    *aiplatform.PredictionClient
    nlp       *VietnameseNLP
    projectID string
    location  string
    endpoint  string
}

type AnalysisRequest struct {
    Texts  []string `json:"texts" binding:"required,min=1"`
    Symbol string   `json:"symbol,omitempty"`
}

type SentimentResult struct {
    Sentiment  string  `json:"sentiment"`  // positive, neutral, negative
    Confidence float64 `json:"confidence"`
    Scores     Scores  `json:"scores"`
}

type Scores struct {
    Positive float64 `json:"positive"`
    Neutral  float64 `json:"neutral"`
    Negative float64 `json:"negative"`
}

func (a *Agent) Analyze(ctx context.Context, req AnalysisRequest) (*AggregateResult, error) {
    // Filter spam
    filtered := make([]string, 0, len(req.Texts))
    for _, text := range req.Texts {
        if !a.nlp.IsSpam(text) {
            filtered = append(filtered, text)
        }
    }

    if len(filtered) == 0 {
        return a.emptyResult(req.Symbol, "All texts filtered as spam"), nil
    }

    // Process in batches for Vertex AI
    results := make([]SentimentResult, 0, len(filtered))
    batchSize := 10

    for i := 0; i < len(filtered); i += batchSize {
        end := min(i+batchSize, len(filtered))
        batch := filtered[i:end]

        batchResults, err := a.analyzeBatch(ctx, batch)
        if err != nil {
            return nil, fmt.Errorf("analyze batch: %w", err)
        }
        results = append(results, batchResults...)
    }

    return a.aggregate(req.Symbol, results), nil
}

func (a *Agent) analyzeBatch(ctx context.Context, texts []string) ([]SentimentResult, error) {
    // Preprocess texts
    processed := make([]string, len(texts))
    for i, text := range texts {
        processed[i] = a.nlp.Preprocess(text)
    }

    // Call Vertex AI endpoint
    resp, err := a.client.Predict(ctx, &aiplatformpb.PredictRequest{
        Endpoint: a.endpoint,
        Instances: textsToInstances(processed),
    })
    if err != nil {
        return nil, err
    }

    return parseVertexResponse(resp), nil
}
```

#### FR-SA-002: Vietnamese NLP Processor (Go)

```go
// internal/nlp/vietnamese.go
package nlp

import (
    "regexp"
    "strings"
)

type VietnameseNLP struct {
    slangDict    map[string]string
    validSymbols map[string]bool
    spamKeywords []string
    symbolRegex  *regexp.Regexp
}

func NewVietnameseNLP() *VietnameseNLP {
    return &VietnameseNLP{
        slangDict: map[string]string{
            "cây thông":     "bullish_pattern",
            "cay thong":     "bullish_pattern",
            "cây súng":      "bearish_pattern",
            "lùa gà":        "pump_and_dump",
            "lua ga":        "pump_and_dump",
            "múa bên trăng": "price_manipulation",
            "con tép":       "retail_investor",
            "cá mập":        "institutional_investor",
            "chốt lời":      "take_profit",
            "cắt lỗ":        "stop_loss",
            "fomo":          "fear_of_missing_out",
            "all in":        "invest_all_capital",
            "breakout":      "price_breakout",
            "sideway":       "sideways_trend",
        },
        validSymbols: map[string]bool{
            "VNM": true, "HPG": true, "VCB": true, "VHM": true, "VIC": true,
            "FPT": true, "MSN": true, "MBB": true, "TCB": true, "BID": true,
            "CTG": true, "GAS": true, "SAB": true, "VRE": true, "PLX": true,
        },
        spamKeywords: []string{
            "khuyến mãi", "khuyen mai", "đăng ký ngay", "group vip",
            "bảo lãi", "bao lai", "cam kết lời", "free signal",
        },
        symbolRegex: regexp.MustCompile(`\b([A-Z]{3})\b`),
    }
}

func (v *VietnameseNLP) Preprocess(text string) string {
    lower := strings.ToLower(text)

    for slang, meaning := range v.slangDict {
        lower = strings.ReplaceAll(lower, slang, " "+meaning+" ")
    }

    // Normalize whitespace
    return strings.Join(strings.Fields(lower), " ")
}

func (v *VietnameseNLP) ExtractSymbols(text string) []string {
    matches := v.symbolRegex.FindAllString(strings.ToUpper(text), -1)

    symbols := make([]string, 0)
    seen := make(map[string]bool)

    for _, m := range matches {
        if v.validSymbols[m] && !seen[m] {
            symbols = append(symbols, m)
            seen[m] = true
        }
    }

    return symbols
}

func (v *VietnameseNLP) IsSpam(text string) bool {
    lower := strings.ToLower(text)
    for _, kw := range v.spamKeywords {
        if strings.Contains(lower, kw) {
            return true
        }
    }
    return false
}
```

### 4.4 Forecast Agent (FR-FC)

```go
// internal/analysis/forecast/agent.go
package forecast

const (
    TechnicalWeight = 0.40
    SentimentWeight = 0.30
    MarketWeight    = 0.30
)

type Agent struct{}

type ForecastRequest struct {
    Technical     TechnicalData `json:"technical"`
    Sentiment     SentimentData `json:"sentiment"`
    MarketContext MarketData    `json:"market_context"`
}

type ForecastResult struct {
    Recommendation string   `json:"recommendation"`
    Confidence     float64  `json:"confidence"`
    RiskLevel      string   `json:"risk_level"`
    Reasoning      []string `json:"reasoning"`
    Scores         Scores   `json:"scores"`
}

func (a *Agent) Forecast(ctx context.Context, req ForecastRequest) (*ForecastResult, error) {
    // Calculate individual scores
    techScore := a.scoreTechnical(req.Technical)
    sentScore := a.scoreSentiment(req.Sentiment)
    marketScore := a.scoreMarket(req.MarketContext)

    // Weighted combination
    finalScore := techScore*TechnicalWeight +
        sentScore*SentimentWeight +
        marketScore*MarketWeight

    // Generate recommendation
    recommendation := a.getRecommendation(finalScore)
    confidence := math.Min(math.Abs(finalScore)+0.3, 1.0)
    riskLevel := a.assessRisk(req)

    return &ForecastResult{
        Recommendation: recommendation,
        Confidence:     confidence,
        RiskLevel:      riskLevel,
        Reasoning:      a.buildReasoning(req, techScore, sentScore, marketScore),
        Scores: Scores{
            Technical: techScore,
            Sentiment: sentScore,
            Market:    marketScore,
            Final:     finalScore,
        },
    }, nil
}

func (a *Agent) getRecommendation(score float64) string {
    switch {
    case score > 0.6:
        return "STRONG BUY"
    case score > 0.2:
        return "BUY"
    case score > -0.2:
        return "HOLD"
    case score > -0.6:
        return "SELL"
    default:
        return "STRONG SELL"
    }
}

func (a *Agent) assessRisk(req ForecastRequest) string {
    // Check volatility
    if req.Technical.ATR > 0.05 {
        return "HIGH"
    }

    // Check negative sentiment
    if req.Sentiment.NegativeRatio > 0.5 {
        return "HIGH"
    }

    // Check rumors
    if len(req.Sentiment.Rumors) > 0 {
        return "HIGH"
    }

    if req.Technical.ATR > 0.03 || req.Sentiment.NegativeRatio > 0.3 {
        return "MEDIUM"
    }

    return "LOW"
}
```

### 4.5 Master Agent / Orchestrator (FR-MA)

```go
// internal/orchestrator/master.go
package orchestrator

type MasterAgent struct {
    technicalClient  *http.Client
    sentimentClient  *http.Client
    forecastClient   *http.Client
    storageClient    storage.Client
    pubsubPublisher  pubsub.Publisher
}

func (m *MasterAgent) AnalyzeStock(ctx context.Context, symbol string) (*FullAnalysis, error) {
    g, ctx := errgroup.WithContext(ctx)

    var techResult *TechnicalResult
    var sentResult *SentimentResult

    // Run technical and sentiment analysis concurrently
    g.Go(func() error {
        var err error
        techResult, err = m.callTechnicalAgent(ctx, symbol)
        return err
    })

    g.Go(func() error {
        var err error
        news, _ := m.fetchNews(ctx, symbol)
        sentResult, err = m.callSentimentAgent(ctx, news)
        return err
    })

    if err := g.Wait(); err != nil {
        return nil, fmt.Errorf("analysis failed: %w", err)
    }

    // Get market context
    marketCtx, _ := m.getMarketContext(ctx)

    // Call forecast agent
    forecast, err := m.callForecastAgent(ctx, techResult, sentResult, marketCtx)
    if err != nil {
        return nil, fmt.Errorf("forecast failed: %w", err)
    }

    return &FullAnalysis{
        Symbol:     symbol,
        Technical:  techResult,
        Sentiment:  sentResult,
        Forecast:   forecast,
        Timestamp:  time.Now(),
    }, nil
}

func (m *MasterAgent) GenerateDailyReport(ctx context.Context) (*DailyReport, error) {
    // 1. Get hot stocks from news
    hotStocks, err := m.detectHotStocks(ctx)
    if err != nil {
        return nil, err
    }

    // 2. Analyze top 5 stocks concurrently
    analyses := make([]*FullAnalysis, 0, 5)
    g, ctx := errgroup.WithContext(ctx)
    mu := sync.Mutex{}

    for _, stock := range hotStocks[:min(5, len(hotStocks))] {
        stock := stock
        g.Go(func() error {
            analysis, err := m.AnalyzeStock(ctx, stock.Symbol)
            if err != nil {
                return nil // Continue on individual failure
            }
            mu.Lock()
            analyses = append(analyses, analysis)
            mu.Unlock()
            return nil
        })
    }
    g.Wait()

    // 3. Build report
    report := m.buildReport(hotStocks, analyses)

    // 4. Save to GCS
    if err := m.saveReport(ctx, report); err != nil {
        return nil, err
    }

    // 5. Publish alerts
    m.publishAlerts(ctx, analyses)

    return report, nil
}
```

### 4.6 API Gateway (FR-API)

```go
// cmd/api-gateway/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "vnstock/internal/api/handlers"
    "vnstock/internal/api/middleware"
)

func main() {
    r := gin.Default()

    // Middleware
    r.Use(middleware.Logger())
    r.Use(middleware.CORS())
    r.Use(middleware.RateLimit())

    // Health check
    r.GET("/health", handlers.HealthCheck)

    // API routes
    api := r.Group("/api")
    {
        // Analysis endpoints
        api.POST("/analyze/technical", handlers.AnalyzeTechnical)
        api.POST("/analyze/sentiment", handlers.AnalyzeSentiment)
        api.POST("/synthesize", handlers.Synthesize)

        // Reports
        api.GET("/reports/daily", handlers.GetDailyReport)
        api.GET("/stocks/:symbol", handlers.GetStockAnalysis)
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}
```

---

## 5. Non-Functional Requirements

### 5.1 Performance Requirements (GCP Optimized)

| ID | Requirement | Target | GCP Service |
|----|-------------|--------|-------------|
| NFR-P-001 | API Response (cached) | p95 < 100ms | Memorystore |
| NFR-P-002 | API Response (live) | p95 < 2s | Cloud Run |
| NFR-P-003 | Cold Start | < 3s | Cloud Run (min instances) |
| NFR-P-004 | Daily Workflow | < 5 min | Cloud Workflows |
| NFR-P-005 | Concurrent Requests | 1000 RPS | Cloud Run autoscaling |

### 5.2 Scalability (Cloud Run)

```yaml
# Cloud Run service configuration
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
        autoscaling.knative.dev/maxScale: "100"
        run.googleapis.com/cpu-throttling: "false"
    spec:
      containerConcurrency: 80
      timeoutSeconds: 300
      containers:
        - resources:
            limits:
              cpu: "2"
              memory: "2Gi"
```

### 5.3 Security Requirements

| ID | Requirement | GCP Implementation |
|----|-------------|-------------------|
| NFR-SEC-001 | Authentication | Cloud Identity / Firebase Auth |
| NFR-SEC-002 | API Keys | API Gateway + API Keys |
| NFR-SEC-003 | Secrets | Secret Manager |
| NFR-SEC-004 | Encryption | Customer-managed encryption keys (CMEK) |
| NFR-SEC-005 | Network | VPC + Cloud Armor |
| NFR-SEC-006 | IAM | Workload Identity |

### 5.4 Reliability

| ID | Requirement | GCP Implementation |
|----|-------------|-------------------|
| NFR-R-001 | Availability | 99.9% (Cloud Run SLA) |
| NFR-R-002 | Database HA | Cloud SQL HA (regional) |
| NFR-R-003 | Backup | Cloud SQL automated backups |
| NFR-R-004 | DR | Cross-region replication |

### 5.5 Observability

```go
// internal/observability/tracing.go
package observability

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func InitTracing(serviceName string) error {
    exporter, err := otlptracegrpc.New(context.Background())
    if err != nil {
        return err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String(serviceName),
        )),
    )

    otel.SetTracerProvider(tp)
    return nil
}
```

---

## 6. Data Requirements

### 6.1 Cloud SQL Schema

Same PostgreSQL schema as original, deployed on Cloud SQL:

```sql
-- Cloud SQL instance: vnstock-db
-- Region: asia-southeast1

CREATE TABLE stocks (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(255),
    exchange VARCHAR(10),
    sector VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Additional tables same as original SRS
```

### 6.2 Cloud Storage Structure

```
gs://vnstock-data/
├── raw-data/
│   ├── news/YYYY-MM-DD/
│   ├── market-data/YYYY-MM-DD/
│   └── social/YYYY-MM-DD/
├── processed/
│   ├── sentiment/YYYY-MM-DD/
│   ├── technical/YYYY-MM-DD/
│   └── combined/YYYY-MM-DD/
├── reports/
│   ├── daily/YYYY-MM-DD/
│   ├── weekly/YYYY-WW/
│   └── alerts/YYYY-MM-DD/
└── models/
    └── phobert/
```

### 6.3 Memorystore (Redis) Keys

```
# Cache key patterns
technical:{symbol}:{days}     -> Technical analysis result (TTL: 5min)
sentiment:{hash}:{symbol}     -> Sentiment result (TTL: 10min)
stock:{symbol}:latest         -> Latest full analysis (TTL: 5min)
report:daily:{date}           -> Daily report (TTL: 24h)
hot-stocks:{date}             -> Hot stocks list (TTL: 1h)
```

---

## 7. System Constraints

### 7.1 Vietnamese Market (Same as Original)

- Trading Hours: 9:00-15:00 Vietnam time
- Exchanges: HSX, HNX, UPCOM
- Symbol Format: 3 uppercase letters
- Price limits: +/- 7%

### 7.2 GCP Constraints

| Constraint | Limit | Mitigation |
|------------|-------|------------|
| Cloud Run request timeout | 60min | Use Cloud Tasks for long jobs |
| Cloud Run memory | 32GB max | Optimize memory usage |
| Pub/Sub message size | 10MB | Chunk large payloads |
| Cloud Functions timeout | 9min (Gen1), 60min (Gen2) | Use Gen2 for long tasks |
| Cloud SQL connections | Varies by tier | Use connection pooling |

### 7.3 Go Constraints

- Single binary deployment (no runtime dependencies)
- Statically compiled
- Goroutine-based concurrency
- No dynamic library loading

---

## 8. GCP Service Mapping

### 8.1 AWS to GCP Migration

| AWS Service | GCP Equivalent | Notes |
|-------------|---------------|-------|
| S3 | Cloud Storage | gsutil compatible |
| RDS PostgreSQL | Cloud SQL | Same PostgreSQL version |
| ElastiCache | Memorystore | Redis 7.0 |
| Lambda | Cloud Functions | Go 1.22 runtime |
| ECS/Fargate | Cloud Run | Serverless containers |
| SQS | Pub/Sub | Different API model |
| Step Functions | Cloud Workflows | YAML-based |
| CloudWatch | Cloud Monitoring | Prometheus compatible |
| Secrets Manager | Secret Manager | Similar API |
| API Gateway | API Gateway | OpenAPI spec |
| Cognito | Firebase Auth | Mobile-friendly |

### 8.2 Cost Estimation (Monthly)

| Service | Specification | Est. Cost |
|---------|--------------|-----------|
| Cloud Run (API) | 1 vCPU, 1GB, 100K requests | $30-50 |
| Cloud Run (Agents x3) | 2 vCPU, 2GB each | $50-80 |
| Cloud SQL | db-f1-micro, 10GB | $15-25 |
| Memorystore | 1GB Basic | $35 |
| Cloud Storage | 50GB Standard | $1-2 |
| Pub/Sub | 1M messages | $1 |
| Cloud Functions | 100K invocations | $0.40 |
| Vertex AI | Custom model hosting | $50-100 |
| **Total** | | **~$180-290/month** |

---

## 9. Appendices

### Appendix A: Go Project Structure

```
vnstock-go/
├── cmd/
│   ├── api-gateway/
│   │   └── main.go
│   ├── technical-agent/
│   │   └── main.go
│   ├── sentiment-agent/
│   │   └── main.go
│   ├── forecast-agent/
│   │   └── main.go
│   └── telegram-bot/
│       └── main.go
├── internal/
│   ├── analysis/
│   │   ├── technical/
│   │   ├── sentiment/
│   │   └── forecast/
│   ├── api/
│   │   ├── handlers/
│   │   └── middleware/
│   ├── market/
│   ├── nlp/
│   ├── storage/
│   ├── cache/
│   ├── pubsub/
│   └── config/
├── pkg/
│   └── models/
├── deployments/
│   ├── cloud-run/
│   ├── cloud-functions/
│   └── terraform/
├── scripts/
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

### Appendix B: Vietnamese Slang Dictionary (Go Map)

```go
var SlangDictionary = map[string]string{
    "cây thông":     "bullish_pattern",
    "cây súng":      "bearish_pattern",
    "múa bên trăng": "price_manipulation",
    "lùa gà":        "pump_and_dump",
    "con tép":       "retail_investor",
    "cá mập":        "institutional_investor",
    "chốt lời":      "take_profit",
    "cắt lỗ":        "stop_loss",
    "fomo":          "fear_of_missing_out",
    "all in":        "invest_all_capital",
    "sideway":       "sideways_trend",
    "breakout":      "price_breakout",
    "bắt đáy":       "bottom_fishing",
    "bắt dao rơi":   "catching_falling_knife",
}
```

### Appendix C: Terraform Resources

```hcl
# deployments/terraform/main.tf

provider "google" {
  project = var.project_id
  region  = var.region
}

# Cloud SQL
resource "google_sql_database_instance" "vnstock" {
  name             = "vnstock-db"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = "db-f1-micro"

    backup_configuration {
      enabled = true
    }
  }
}

# Memorystore
resource "google_redis_instance" "vnstock" {
  name           = "vnstock-cache"
  tier           = "BASIC"
  memory_size_gb = 1
  region         = var.region
}

# Cloud Storage
resource "google_storage_bucket" "vnstock_data" {
  name     = "${var.project_id}-vnstock-data"
  location = var.region

  lifecycle_rule {
    condition {
      age = 90
    }
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
  }
}

# Cloud Run
resource "google_cloud_run_service" "api_gateway" {
  name     = "vnstock-api"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project_id}/vnstock-api:latest"

        resources {
          limits = {
            cpu    = "2"
            memory = "1Gi"
          }
        }
      }
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/minScale" = "1"
        "autoscaling.knative.dev/maxScale" = "10"
      }
    }
  }
}
```

---

**End of SRS Document (Go/GCP Edition)**
