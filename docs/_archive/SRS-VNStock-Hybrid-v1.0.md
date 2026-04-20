# Software Requirements Specification
# VN Stock Analysis System - Hybrid Architecture
## Version 1.0

---

## Document Control

| Field | Value |
|-------|-------|
| **Version** | 1.0 |
| **Date** | 2026-01-28 |
| **Status** | Draft |
| **Architecture** | Hybrid (Go + Python) on GCP |

---

## 1. Introduction

### 1.1 Purpose

This SRS defines requirements for the VN Stock Analysis System using a **hybrid architecture**:
- **Go/Gin**: API Gateway, Technical Analysis Agent, Forecast Agent, Master Orchestrator
- **Python/FastAPI**: Sentiment Analysis Agent (PhoBERT Vietnamese NLP)

### 1.2 Architecture Rationale

| Component | Language | Reason |
|-----------|----------|--------|
| API Gateway | Go | High throughput, low latency |
| Technical Agent | Go | CPU-intensive calculations, concurrent processing |
| Forecast Agent | Go | Lightweight aggregation logic |
| Master Orchestrator | Go | Workflow coordination, Pub/Sub handling |
| **Sentiment Agent** | **Python** | PhoBERT model, transformers library, Vietnamese NLP |

### 1.3 Scope

The system analyzes Vietnamese stock market data from HSX/HNX/UPCOM exchanges through:
- Real-time technical indicator calculation (Go)
- Vietnamese NLP sentiment analysis (Python + PhoBERT)
- ML-based price forecasting (Go + Vertex AI)
- Automated Telegram/Email alerts (Go)

---

## 2. System Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                    GOOGLE CLOUD PLATFORM                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Cloud Scheduler ──► Cloud Workflows                                 │
│                          │                                           │
│                          ▼                                           │
│              ┌─── Cloud Pub/Sub ────────────────────┐               │
│              │           │                           │               │
│              ▼           ▼                           ▼               │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐      │
│  │  Cloud Run:     │  │  Cloud Run:     │  │  Cloud Run:     │      │
│  │  API Gateway    │  │  Technical      │  │  Forecast       │      │
│  │  (Go/Gin)       │  │  Agent (Go)     │  │  Agent (Go)     │      │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘      │
│           │                     │                     │              │
│           │         ┌───────────┴───────────┐         │              │
│           │         ▼                       ▼         │              │
│           │  ┌─────────────────┐   ┌─────────────────┐│              │
│           │  │  Cloud Run:     │   │  Cloud Run:     ││              │
│           │  │  Sentiment      │   │  Master         ││              │
│           │  │  Agent (Python) │   │  Orchestrator   ││              │
│           │  │  + PhoBERT      │   │  (Go)           ││              │
│           │  └────────┬────────┘   └────────┬────────┘│              │
│           │           │                     │         │              │
│           └───────────┼─────────────────────┼─────────┘              │
│                       ▼                     ▼                        │
│           ┌─────────────────────────────────────────┐               │
│           │              Cloud SQL                   │               │
│           │            (PostgreSQL)                  │               │
│           └─────────────────────────────────────────┘               │
│                              │                                       │
│           ┌──────────────────┼──────────────────┐                   │
│           ▼                  ▼                  ▼                    │
│    ┌──────────┐      ┌──────────────┐    ┌──────────┐               │
│    │Memorystore│      │Cloud Storage │    │ Vertex   │               │
│    │ (Redis)  │      │   (GCS)      │    │   AI     │               │
│    └──────────┘      └──────────────┘    └──────────┘               │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 Service Communication

```
┌─────────────────────────────────────────────────────────────────────┐
│                    INTER-SERVICE COMMUNICATION                       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   Synchronous (gRPC/HTTP):                                          │
│   ┌──────────┐     HTTP/JSON      ┌──────────────┐                  │
│   │ Go API   │ ◄─────────────────► │ Python       │                  │
│   │ Gateway  │                     │ Sentiment    │                  │
│   └──────────┘                     └──────────────┘                  │
│                                                                      │
│   Asynchronous (Pub/Sub):                                           │
│   ┌──────────┐     Pub/Sub        ┌──────────────┐                  │
│   │ Go       │ ──────────────────► │ Python       │                  │
│   │ Technical│  topic: sentiment   │ Sentiment    │                  │
│   └──────────┘     -requests       └──────────────┘                  │
│                                          │                           │
│                        Pub/Sub           │                           │
│                  topic: sentiment-results│                           │
│                                          ▼                           │
│                                   ┌──────────────┐                  │
│                                   │ Go Master    │                  │
│                                   │ Orchestrator │                  │
│                                   └──────────────┘                  │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.3 Service Specifications

| Service | Language | Framework | Memory | CPU | Scaling |
|---------|----------|-----------|--------|-----|---------|
| API Gateway | Go 1.22 | Gin | 512MB | 1 | 0-10 |
| Technical Agent | Go 1.22 | Gin | 1GB | 2 | 0-5 |
| Sentiment Agent | Python 3.11 | FastAPI | 4GB | 2 | 1-3 |
| Forecast Agent | Go 1.22 | Gin | 1GB | 1 | 0-3 |
| Master Orchestrator | Go 1.22 | Gin | 1GB | 1 | 1-2 |

**Note:** Sentiment Agent requires 4GB minimum for PhoBERT model loading.

---

## 3. Functional Requirements

### 3.1 Go Services

#### 3.1.1 API Gateway (FR-API)

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-API-001 | Accept REST requests for stock analysis | Must |
| FR-API-002 | Route requests to appropriate agents | Must |
| FR-API-003 | Aggregate responses from multiple agents | Must |
| FR-API-004 | Implement rate limiting (100 req/min) | Must |
| FR-API-005 | JWT authentication for protected endpoints | Must |
| FR-API-006 | Request/response logging | Should |
| FR-API-007 | API versioning (v1, v2) | Should |

#### 3.1.2 Technical Analysis Agent (FR-TA)

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-TA-001 | Calculate RSI (14-period default) | Must |
| FR-TA-002 | Calculate MACD (12, 26, 9 default) | Must |
| FR-TA-003 | Calculate Bollinger Bands (20, 2 default) | Must |
| FR-TA-004 | Calculate SMA/EMA (multiple periods) | Must |
| FR-TA-005 | Calculate Stochastic Oscillator | Must |
| FR-TA-006 | Calculate ADX | Should |
| FR-TA-007 | Calculate ATR | Should |
| FR-TA-008 | Calculate VWAP | Should |
| FR-TA-009 | Generate buy/sell/hold signals | Must |
| FR-TA-010 | Calculate signal confidence score (0-100) | Must |
| FR-TA-011 | Cache indicator results in Redis (TTL: 5min) | Must |
| FR-TA-012 | Support batch processing (up to 50 symbols) | Should |
| FR-TA-013 | Concurrent indicator calculation | Must |

#### 3.1.3 Forecast Agent (FR-FC)

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-FC-001 | Aggregate technical analysis (weight: 40%) | Must |
| FR-FC-002 | Aggregate sentiment analysis (weight: 30%) | Must |
| FR-FC-003 | Aggregate market context (weight: 30%) | Must |
| FR-FC-004 | Generate final recommendation | Must |
| FR-FC-005 | Calculate combined confidence score | Must |
| FR-FC-006 | Provide price targets (support/resistance) | Should |
| FR-FC-007 | Optional Vertex AI integration | Could |

#### 3.1.4 Master Orchestrator (FR-MO)

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-MO-001 | Coordinate analysis workflow | Must |
| FR-MO-002 | Handle Pub/Sub message routing | Must |
| FR-MO-003 | Generate daily analysis reports | Must |
| FR-MO-004 | Send Telegram alerts | Must |
| FR-MO-005 | Send email notifications | Should |
| FR-MO-006 | Track analysis job status | Must |
| FR-MO-007 | Retry failed agent calls (max 3) | Must |

### 3.2 Python Services

#### 3.2.1 Sentiment Analysis Agent (FR-SA)

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-SA-001 | Load PhoBERT model on startup | Must |
| FR-SA-002 | Analyze Vietnamese news text | Must |
| FR-SA-003 | Classify sentiment (positive/negative/neutral) | Must |
| FR-SA-004 | Calculate confidence score (0-100) | Must |
| FR-SA-005 | Extract stock symbols from text | Must |
| FR-SA-006 | Apply Vietnamese stock slang dictionary | Must |
| FR-SA-007 | Handle batch text analysis (up to 100 items) | Must |
| FR-SA-008 | Cache sentiment results (TTL: 1 hour) | Should |
| FR-SA-009 | Expose REST API for synchronous calls | Must |
| FR-SA-010 | Subscribe to Pub/Sub for async processing | Must |
| FR-SA-011 | Health check endpoint with model status | Must |

**Vietnamese Slang Dictionary Examples:**
| Slang | Meaning | Sentiment |
|-------|---------|-----------|
| lùa gà | pump and dump | Negative |
| cá mập | big investor/whale | Neutral |
| gom hàng | accumulating | Positive |
| xả hàng | dumping/selling off | Negative |
| tay to | major player | Neutral |
| đánh úp | surprise attack | Negative |
| bắt đáy | catching the bottom | Positive |
| cắt lỗ | stop loss | Neutral |

---

## 4. API Specifications

### 4.1 Go API Gateway Endpoints

```yaml
openapi: 3.0.3
info:
  title: VN Stock Analysis API
  version: 1.0.0

paths:
  /api/v1/analyze:
    post:
      summary: Full stock analysis
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/AnalyzeRequest'
      responses:
        200:
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AnalyzeResponse'

  /api/v1/technical/{symbol}:
    get:
      summary: Technical analysis only
      parameters:
        - name: symbol
          in: path
          required: true
          schema:
            type: string
            pattern: '^[A-Z]{3}$'
      responses:
        200:
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TechnicalResult'

  /api/v1/sentiment:
    post:
      summary: Sentiment analysis (proxied to Python)
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/SentimentRequest'
      responses:
        200:
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SentimentResult'

  /api/v1/reports/daily:
    get:
      summary: Get daily analysis report
      responses:
        200:
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DailyReport'

components:
  schemas:
    AnalyzeRequest:
      type: object
      required:
        - symbols
      properties:
        symbols:
          type: array
          items:
            type: string
          maxItems: 50
        include_sentiment:
          type: boolean
          default: true
        include_forecast:
          type: boolean
          default: true

    AnalyzeResponse:
      type: object
      properties:
        request_id:
          type: string
        timestamp:
          type: string
          format: date-time
        results:
          type: array
          items:
            $ref: '#/components/schemas/StockAnalysis'

    StockAnalysis:
      type: object
      properties:
        symbol:
          type: string
        technical:
          $ref: '#/components/schemas/TechnicalResult'
        sentiment:
          $ref: '#/components/schemas/SentimentResult'
        forecast:
          $ref: '#/components/schemas/ForecastResult'

    TechnicalResult:
      type: object
      properties:
        rsi:
          type: number
        macd:
          type: object
          properties:
            macd_line:
              type: number
            signal_line:
              type: number
            histogram:
              type: number
        bollinger:
          type: object
          properties:
            upper:
              type: number
            middle:
              type: number
            lower:
              type: number
        signal:
          type: string
          enum: [BUY, SELL, HOLD]
        confidence:
          type: number
          minimum: 0
          maximum: 100

    SentimentResult:
      type: object
      properties:
        sentiment:
          type: string
          enum: [positive, negative, neutral]
        confidence:
          type: number
          minimum: 0
          maximum: 100
        source_count:
          type: integer
        keywords:
          type: array
          items:
            type: string

    ForecastResult:
      type: object
      properties:
        recommendation:
          type: string
          enum: [STRONG_BUY, BUY, HOLD, SELL, STRONG_SELL]
        confidence:
          type: number
        price_target:
          type: object
          properties:
            support:
              type: number
            resistance:
              type: number
        reasoning:
          type: string
```

### 4.2 Python Sentiment Service API

```yaml
openapi: 3.0.3
info:
  title: Sentiment Analysis Service
  version: 1.0.0

paths:
  /analyze:
    post:
      summary: Analyze text sentiment
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required:
                - texts
              properties:
                texts:
                  type: array
                  items:
                    type: string
                  maxItems: 100
      responses:
        200:
          content:
            application/json:
              schema:
                type: object
                properties:
                  results:
                    type: array
                    items:
                      type: object
                      properties:
                        text:
                          type: string
                        sentiment:
                          type: string
                        confidence:
                          type: number
                        symbols:
                          type: array
                          items:
                            type: string

  /health:
    get:
      summary: Health check with model status
      responses:
        200:
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                  model_loaded:
                    type: boolean
                  model_name:
                    type: string
                  memory_usage_mb:
                    type: number
```

---

## 5. Data Models

### 5.1 PostgreSQL Schema

```sql
-- Stock symbols
CREATE TABLE stocks (
    symbol VARCHAR(10) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    exchange VARCHAR(10) NOT NULL CHECK (exchange IN ('HSX', 'HNX', 'UPCOM')),
    industry VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Technical analysis results
CREATE TABLE technical_analysis (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL REFERENCES stocks(symbol),
    timestamp TIMESTAMPTZ NOT NULL,

    -- Price data
    open_price DECIMAL(12, 2),
    high_price DECIMAL(12, 2),
    low_price DECIMAL(12, 2),
    close_price DECIMAL(12, 2),
    volume BIGINT,

    -- Indicators
    rsi_14 DECIMAL(5, 2),
    macd_line DECIMAL(10, 4),
    macd_signal DECIMAL(10, 4),
    macd_histogram DECIMAL(10, 4),
    bb_upper DECIMAL(12, 2),
    bb_middle DECIMAL(12, 2),
    bb_lower DECIMAL(12, 2),
    sma_20 DECIMAL(12, 2),
    ema_12 DECIMAL(12, 2),
    ema_26 DECIMAL(12, 2),
    adx DECIMAL(5, 2),
    atr DECIMAL(12, 2),

    -- Signal
    signal VARCHAR(10) CHECK (signal IN ('BUY', 'SELL', 'HOLD')),
    confidence DECIMAL(5, 2),

    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(symbol, timestamp)
);

-- Sentiment analysis results
CREATE TABLE sentiment_analysis (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) REFERENCES stocks(symbol),
    source_url TEXT,
    source_type VARCHAR(50),

    -- Sentiment data
    text_content TEXT,
    sentiment VARCHAR(20) CHECK (sentiment IN ('positive', 'negative', 'neutral')),
    confidence DECIMAL(5, 2),
    keywords TEXT[],

    -- Metadata
    published_at TIMESTAMPTZ,
    analyzed_at TIMESTAMPTZ DEFAULT NOW(),
    model_version VARCHAR(50) DEFAULT 'phobert-base'
);

-- Forecast results
CREATE TABLE forecasts (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL REFERENCES stocks(symbol),
    timestamp TIMESTAMPTZ NOT NULL,

    -- Component scores
    technical_score DECIMAL(5, 2),
    sentiment_score DECIMAL(5, 2),
    market_score DECIMAL(5, 2),

    -- Final recommendation
    recommendation VARCHAR(20) CHECK (recommendation IN
        ('STRONG_BUY', 'BUY', 'HOLD', 'SELL', 'STRONG_SELL')),
    confidence DECIMAL(5, 2),

    -- Price targets
    support_price DECIMAL(12, 2),
    resistance_price DECIMAL(12, 2),

    reasoning TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Daily reports
CREATE TABLE daily_reports (
    id BIGSERIAL PRIMARY KEY,
    report_date DATE UNIQUE NOT NULL,

    -- Summary
    total_symbols_analyzed INT,
    buy_signals INT,
    sell_signals INT,
    hold_signals INT,

    -- Top picks
    top_picks JSONB,
    market_summary TEXT,

    -- Report content
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

### 5.2 Redis Cache Keys

```
# Technical analysis cache
technical:{symbol}:latest          -> JSON (TTL: 5 min)
technical:{symbol}:{date}          -> JSON (TTL: 24 hours)

# Sentiment cache
sentiment:{symbol}:latest          -> JSON (TTL: 1 hour)
sentiment:batch:{hash}             -> JSON (TTL: 1 hour)

# Rate limiting
ratelimit:{client_id}              -> Counter (TTL: 1 min)

# Job tracking
job:{job_id}                       -> JSON (status, progress)
job:{job_id}:results               -> JSON (partial results)

# Model status
model:sentiment:status             -> JSON (loaded, memory)
```

---

## 6. Pub/Sub Topics

### 6.1 Topic Definitions

| Topic | Publisher | Subscriber | Message Schema |
|-------|-----------|------------|----------------|
| `analysis-requests` | API Gateway | Master Orchestrator | AnalysisRequest |
| `technical-requests` | Master Orchestrator | Technical Agent | TechnicalRequest |
| `technical-results` | Technical Agent | Master Orchestrator | TechnicalResult |
| `sentiment-requests` | Master Orchestrator | Sentiment Agent (Python) | SentimentRequest |
| `sentiment-results` | Sentiment Agent | Master Orchestrator | SentimentResult |
| `forecast-requests` | Master Orchestrator | Forecast Agent | ForecastRequest |
| `forecast-results` | Forecast Agent | Master Orchestrator | ForecastResult |
| `alerts` | Master Orchestrator | Telegram Bot, Email Service | Alert |

### 6.2 Message Schemas

```json
// AnalysisRequest
{
  "request_id": "uuid",
  "symbols": ["VNM", "FPT", "VIC"],
  "options": {
    "include_sentiment": true,
    "include_forecast": true
  },
  "callback_url": "optional",
  "timestamp": "2026-01-28T08:00:00Z"
}

// SentimentRequest (to Python service)
{
  "request_id": "uuid",
  "correlation_id": "parent-request-id",
  "texts": [
    {
      "id": "1",
      "content": "VNM công bố lợi nhuận tăng 20%",
      "source": "cafef.vn",
      "published_at": "2026-01-28T07:30:00Z"
    }
  ]
}

// SentimentResult (from Python service)
{
  "request_id": "uuid",
  "correlation_id": "parent-request-id",
  "results": [
    {
      "id": "1",
      "sentiment": "positive",
      "confidence": 92.5,
      "symbols": ["VNM"],
      "keywords": ["lợi nhuận", "tăng"]
    }
  ],
  "processing_time_ms": 150,
  "model_version": "phobert-base-v1"
}
```

---

## 7. Deployment Architecture

### 7.1 GCP Services

| Component | GCP Service | Configuration |
|-----------|-------------|---------------|
| Go Services | Cloud Run | Gen2, min 0, max 10 |
| Python Sentiment | Cloud Run | Gen2, min 1, max 3, 4GB RAM |
| Database | Cloud SQL | PostgreSQL 15, db-f1-micro |
| Cache | Memorystore | Redis 7.0, 1GB Basic |
| Storage | Cloud Storage | Standard, lifecycle rules |
| Messaging | Pub/Sub | Default config |
| Scheduler | Cloud Scheduler | Cron expressions |
| Workflows | Cloud Workflows | YAML definitions |
| Secrets | Secret Manager | Versioned secrets |
| Monitoring | Cloud Monitoring | Custom dashboards |

### 7.2 Container Images

```yaml
# Go services (multi-stage build)
# Base image: golang:1.22-alpine
# Final image: gcr.io/distroless/static-debian12
# Size: ~15-25 MB

# Python sentiment service
# Base image: python:3.11-slim
# Final image: includes PhoBERT model
# Size: ~2.5 GB (with model cached)
```

### 7.3 Scaling Configuration

```yaml
# Go API Gateway
apiVersion: serving.knative.dev/v1
kind: Service
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "0"
        autoscaling.knative.dev/maxScale: "10"
        autoscaling.knative.dev/target: "100"  # concurrent requests
    spec:
      containerConcurrency: 100
      containers:
        - resources:
            limits:
              cpu: "1"
              memory: "512Mi"

# Python Sentiment Agent
apiVersion: serving.knative.dev/v1
kind: Service
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"  # Always warm (model loaded)
        autoscaling.knative.dev/maxScale: "3"
        autoscaling.knative.dev/target: "10"   # Lower due to model inference
    spec:
      containerConcurrency: 10
      timeoutSeconds: 300  # Model loading can take time
      containers:
        - resources:
            limits:
              cpu: "2"
              memory: "4Gi"
```

---

## 8. Non-Functional Requirements

### 8.1 Performance

| Metric | Requirement | Notes |
|--------|-------------|-------|
| API Response (technical) | < 200ms p95 | Go services |
| API Response (sentiment) | < 500ms p95 | Python service |
| Full Analysis | < 2s p95 | All agents combined |
| Concurrent Requests | 1000+ | API Gateway |
| Daily Throughput | 50,000+ symbols | Batch processing |

### 8.2 Availability

| Requirement | Target |
|-------------|--------|
| Uptime SLA | 99.5% |
| Recovery Time Objective | < 5 minutes |
| Recovery Point Objective | < 1 hour |

### 8.3 Security

| Requirement | Implementation |
|-------------|----------------|
| API Authentication | JWT tokens |
| Service-to-Service | IAM + service accounts |
| Secrets | GCP Secret Manager |
| Network | VPC + private IPs |
| Data Encryption | At rest and in transit |

---

## 9. Cost Estimate

### 9.1 Monthly Cost Breakdown

| Service | Specification | Est. Cost |
|---------|--------------|-----------|
| Cloud Run (Go x4) | 1-2 vCPU, 512MB-1GB | $40-60 |
| Cloud Run (Python) | 2 vCPU, 4GB, min 1 | $60-80 |
| Cloud SQL | db-f1-micro, 20GB | $20-30 |
| Memorystore | 1GB Basic | $35 |
| Cloud Storage | 50GB Standard | $2 |
| Pub/Sub | 2M messages | $2 |
| Secret Manager | 10 secrets | $0.50 |
| Cloud Logging | 5GB | $2.50 |
| **Total** | | **~$160-210/month** |

### 9.2 Cost Optimization Notes

- Python service min=1 required for PhoBERT model loading time
- Go services can scale to 0 during off-hours
- Consider preemptible/spot for batch processing
- Use committed use discounts for sustained workloads

---

## 10. Appendix

### 10.1 Vietnamese Stock Market Hours

| Session | Time (ICT) | Activity |
|---------|------------|----------|
| Pre-market | 08:30-09:00 | Order entry |
| Morning | 09:00-11:30 | Continuous trading |
| Lunch | 11:30-13:00 | Break |
| Afternoon | 13:00-14:30 | Continuous trading |
| ATC | 14:30-14:45 | Closing auction |
| Post-market | 14:45-15:00 | Settlement |

### 10.2 Symbol Format

- HSX: 3 uppercase letters (e.g., VNM, FPT, VIC)
- HNX: 3 uppercase letters (e.g., PVS, SHS)
- UPCOM: 3 uppercase letters (e.g., VNZ, BSR)

### 10.3 Document References

| Document | Description |
|----------|-------------|
| `docs/Implementation-Guide-Hybrid-v1.0.md` | Implementation guide for hybrid architecture |
| `docs/SRS-VNStock-Go-GCP-v1.0.md` | Full Go/GCP SRS |
| `docs/SRS-VNStock-Analysis-System-v1.0.md` | Original Python/AWS SRS |
