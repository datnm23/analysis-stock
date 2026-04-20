# Software Requirements Specification (SRS)
# VN Stock Analysis System — Hybrid Architecture
## Version 2.0

**Document ID:** SRS-VNSTOCK-HYBRID-002
**Date:** February 2026
**Status:** Active
**Architecture:** Hybrid (Go + Python) on Docker Compose
**Supersedes:** SRS-VNStock-Analysis-System-v1.0, SRS-VNStock-Go-GCP-v1.0, SRS-VNStock-Hybrid-v1.0

---

## Document Control

| Field | Value |
|-------|-------|
| **Version** | 2.0 |
| **Date** | 2026-02-24 |
| **Status** | Active |
| **Architecture** | Hybrid Microservices (Go 1.22 + Python 3.9) |
| **Deployment** | Docker Compose with Profiles |
| **Reference Standard** | IEEE 830-1998 |

### Change History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-28 | Team | Initial draft (Python/AWS) |
| 1.0-Go | 2026-01-28 | Team | Go/GCP variant |
| 1.0-Hybrid | 2026-01-28 | Team | Hybrid variant (GCP Cloud Run) |
| **2.0** | **2026-02-24** | **Team** | **Consolidated SRS matching actual codebase: Docker Compose deployment, n8n orchestration, actual project structure** |

---

## Table of Contents

1. [Introduction](#1-introduction)
2. [Overall Description](#2-overall-description)
3. [External Interface Requirements](#3-external-interface-requirements)
4. [Functional Requirements](#4-functional-requirements)
5. [Non-Functional Requirements](#5-non-functional-requirements)
6. [Data Requirements](#6-data-requirements)
7. [System Constraints](#7-system-constraints)
8. [Deployment Architecture](#8-deployment-architecture)
9. [Appendices](#9-appendices)

---

## 1. Introduction

### 1.1 Purpose

This SRS defines the requirements for the **VN Stock Analysis System** — an AI-powered automated stock analysis platform for Vietnamese securities markets (HSX/HNX/UPCOM). This v2.0 consolidates and supersedes all previous SRS documents to accurately reflect the **actual implemented architecture**.

**Intended Audience:**
- Development Team
- QA Engineers
- Product Managers
- System Architects
- DevOps Engineers

### 1.2 Scope

**System Name:** VN Stock Analysis System (Hệ Thống Phân Tích Chứng Khoán Việt Nam)

**Capabilities:**
- Automated data collection from Vietnamese financial news sources (RSS)
- Technical analysis with 10+ indicators (RSI, MACD, BB, SMA, EMA, Stochastic, ADX, ATR, VWAP)
- Vietnamese NLP sentiment analysis using PhoBERT
- Multi-agent synthesis for investment recommendations
- Automated daily workflow via n8n
- Real-time alerts via Telegram
- Web dashboard for visualization
- RESTful API services

**Out of Scope:**
- Automated trade execution
- Licensed financial advice
- Real-time tick data (uses daily OHLCV)
- International markets
- Mobile native applications (Phase 3)

### 1.3 Architecture Rationale

| Component | Language | Rationale |
|-----------|----------|-----------|
| API Gateway | Go 1.22 / Gin | High throughput, low latency, statically compiled |
| Technical Agent | Go 1.22 / Gin | CPU-intensive concurrent indicator calculations |
| Forecast Agent | Go 1.22 / Gin | Lightweight aggregation and weighted scoring |
| Master Orchestrator | Go 1.22 / Gin | Workflow coordination, report generation |
| **Sentiment Agent** | **Python 3.9 / FastAPI** | PhoBERT model requires transformers + PyTorch ecosystem |
| Workflow Engine | n8n | Visual workflow builder, scheduling, error handling |
| Infrastructure | Docker Compose | Single-command deployment, profile-based environments |

### 1.4 Definitions, Acronyms, and Abbreviations

| Term | Definition |
|------|------------|
| HSX | Ho Chi Minh Stock Exchange (HOSE) |
| HNX | Hanoi Stock Exchange |
| UPCOM | Unlisted Public Company Market |
| VN-Index | Vietnam Stock Market Benchmark Index |
| OHLCV | Open, High, Low, Close, Volume |
| RSI | Relative Strength Index |
| MACD | Moving Average Convergence Divergence |
| BB | Bollinger Bands |
| SMA/EMA | Simple/Exponential Moving Average |
| ADX | Average Directional Index |
| ATR | Average True Range |
| VWAP | Volume Weighted Average Price |
| PhoBERT | Vietnamese BERT Language Model (vinai/phobert-base) |
| n8n | Workflow Automation Platform |
| "Lùa gà" | Vietnamese slang: pump and dump scheme |
| "Cá mập" | Vietnamese slang: large institutional investor |
| "Bắt đáy" | Vietnamese slang: bottom fishing |
| "Cắt lỗ" | Vietnamese slang: stop loss |

### 1.5 References

1. Go Programming Language (https://go.dev) — v1.22
2. Gin Web Framework (https://gin-gonic.com)
3. GORM ORM (https://gorm.io)
4. FastAPI Framework (https://fastapi.tiangolo.com)
5. PhoBERT Model (vinai/phobert-base)
6. vnstock Python Library (https://github.com/thinh-vu/vnstock)
7. n8n Documentation (https://docs.n8n.io)
8. Docker Compose Specification (https://docs.docker.com/compose)
9. Vietnamese Securities Law and Circular 96/2020/TT-BTC
10. Project Architecture Overview: `docs/introduce/ARCHITECTURE_OVERVIEW.md`
11. Project Introduction: `docs/introduce/PROJECT_INTRODUCTION.md`

---

## 2. Overall Description

### 2.1 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     DATA SOURCES                             │
│  RSS Feeds (VnEconomy, CafeF, VietStock, Đầu tư)           │
│  Market APIs (vnstock) │ Social Media │ Internal DB          │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│             n8n WORKFLOW ORCHESTRATION (:5678)               │
│  Scheduling │ RSS Scraping │ Data Pipeline │ Error Handler   │
└──────────────────────────┬──────────────────────────────────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
┌──────────────────────────────────────────────────────────────┐
│                  GO API GATEWAY (:8080)                       │
│  Gin Router │ Rate Limiting │ Logging │ Request Validation   │
└─────┬──────────────┬──────────────┬──────────────┬──────────┘
      │              │              │              │
      ▼              ▼              ▼              ▼
┌───────────┐ ┌───────────┐ ┌────────────┐ ┌──────────────┐
│ Technical │ │ Sentiment │ │  Forecast  │ │   Master     │
│   Agent   │ │   Agent   │ │   Agent    │ │ Orchestrator │
│ Go :8081  │ │ Py :8000  │ │ Go :8082   │ │  Go :8083    │
└─────┬─────┘ └─────┬─────┘ └─────┬──────┘ └──────┬───────┘
      │              │              │              │
      └──────────────┴──────────┬───┴──────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              ▼                 ▼                  ▼
┌──────────────────┐ ┌────────────────┐ ┌──────────────────┐
│   PostgreSQL 15  │ │   Redis 7      │ │   AWS S3 /       │
│   (:5432)        │ │   (:6379)      │ │   MinIO          │
└──────────────────┘ └────────────────┘ └──────────────────┘
```

### 2.2 Service Specifications

| Service | Language | Framework | Port | Memory | CPU | Container |
|---------|----------|-----------|------|--------|-----|-----------|
| API Gateway | Go 1.22 | Gin | 8080 | 512MB | 1 | vnstock-api-gateway |
| Technical Agent | Go 1.22 | Gin | 8081 | 1GB | 2 | vnstock-technical |
| Sentiment Agent | Python 3.9 | FastAPI | 8000 | **4GB** | 2 | vnstock-sentiment |
| Forecast Agent | Go 1.22 | Gin | 8082 | 1GB | 1 | vnstock-forecast |
| Master Orchestrator | Go 1.22 | Gin | 8083 | 1GB | 1 | vnstock-orchestrator |
| n8n Workflow | Node.js | n8n | 5678 | 1GB | 1 | vnstock-n8n |

**Note:** Sentiment Agent requires **4GB minimum** for PhoBERT model loading (~2.5GB model + inference overhead).

### 2.3 Product Functions

| ID | Function | Service | Description |
|----|----------|---------|-------------|
| F1 | Data Collection | n8n + RSS | Automated scraping from VnEconomy, CafeF, VietStock |
| F2 | Technical Analysis | Go Technical Agent | Calculate 10+ indicators, generate buy/sell/hold signals |
| F3 | Sentiment Analysis | Python Sentiment Agent | Vietnamese NLP with PhoBERT + slang dictionary |
| F4 | Forecast Synthesis | Go Forecast Agent | Weighted scoring (40/30/30) for final recommendations |
| F5 | Report Generation | Go Master Orchestrator | Daily Markdown/HTML/JSON reports |
| F6 | Real-time Alerting | Go Master + Telegram | Telegram bot notifications and custom alerts |
| F7 | Web Dashboard | Next.js | Interactive charts and live updates |
| F8 | API Services | Go API Gateway | RESTful endpoints for third-party integration |
| F9 | Workflow Automation | n8n | Daily 8 AM pipeline with error handling |

### 2.4 User Classes

| ID | User Class | Description | Primary Functions |
|----|-----------|-------------|-------------------|
| UC1 | Retail Investors | Individual Vietnamese stock market investors | F2, F3, F4, F5, F6, F7 |
| UC2 | Active Traders | Day traders needing real-time signals | F2, F6, F8 |
| UC3 | System Administrators | System operators and DevOps | All |
| UC4 | API Consumers | Third-party applications and bots | F8 |
| UC5 | Financial Analysts | Professional researchers | F2, F3, F5, F7 |

### 2.5 Operating Environment

| Component | Technology | Specification |
|-----------|------------|---------------|
| Container Runtime | Docker Compose | Version 2.0+ |
| OS | Linux (Docker) | Alpine-based images |
| Database | PostgreSQL 15 | Alpine image, persistent volume |
| Cache | Redis 7 | Alpine image, AOF persistence |
| Object Storage | MinIO (dev) / AWS S3 (prod) | S3-compatible API |
| Workflow | n8n | Latest image, PostgreSQL backend |
| Network | Docker bridge | 172.28.0.0/16 subnet |
| Region | ap-southeast-1 | For AWS S3 (Singapore, closest to Vietnam) |

---

## 3. External Interface Requirements

### 3.1 User Interfaces

#### UI-1: Web Dashboard
- **Framework:** Next.js (port 3000)
- **Features:** Interactive charts, live stock analysis, daily reports viewer
- **Deployment:** Docker container or standalone

#### UI-2: Telegram Bot
- **Integration:** Telegram Bot API via Master Orchestrator
- **Commands:** `/analyze VNM`, `/subscribe VNM HPG FPT`, `/alert_settings`
- **Output:** Formatted analysis reports with signals and confidence scores

#### UI-3: n8n Workflow UI
- **URL:** http://localhost:5678
- **Auth:** Basic auth (N8N_USER / N8N_PASSWORD)
- **Purpose:** Visual workflow editor, monitoring, manual triggers

### 3.2 Software Interfaces

#### SI-1: Go API Gateway (Gin)

```
GET  /health                        → Health check
GET  /api/v1/technical/{symbol}     → Technical analysis (single)
POST /api/v1/technical/batch        → Technical analysis (batch, max 50)
POST /api/v1/sentiment              → Sentiment analysis (proxied to Python)
POST /api/v1/synthesize             → Full analysis synthesis
GET  /api/v1/reports/daily          → Daily report
```

#### SI-2: Python Sentiment Service (FastAPI)

```
POST /analyze                       → Analyze text sentiment (batch)
GET  /health                        → Health check + model status
```

#### SI-3: Market Data Client (vnstock)

```go
// pkg/vnstock/client.go
type Client struct { ... }
func (c *Client) GetMockData(symbol string, days int) []HistoricalData
func (c *Client) GetHistoricalData(ctx context.Context, symbol string, days int) ([]HistoricalData, error)
```

### 3.3 Inter-Service Communication

| From | To | Protocol | Purpose |
|------|----|----------|---------|
| n8n | API Gateway | HTTP/JSON | Trigger analysis workflow |
| API Gateway | Technical Agent | HTTP/JSON | Route technical analysis requests |
| API Gateway | Sentiment Agent | HTTP/JSON | Proxy sentiment requests to Python |
| API Gateway | Forecast Agent | HTTP/JSON | Route forecast requests |
| Master Orchestrator | All Agents | HTTP/JSON | Coordinate analysis pipeline |
| Master Orchestrator | Telegram API | HTTPS | Send reports and alerts |
| All Go Services | PostgreSQL | TCP :5432 | Data persistence |
| All Go Services | Redis | TCP :6379 | Caching and rate limiting |
| n8n | PostgreSQL | TCP :5432 | n8n workflow storage |

### 3.4 External Data Sources

| Source | Type | Data | Frequency |
|--------|------|------|-----------|
| VnEconomy | RSS Feed | Financial news | Daily scrape |
| CafeF | RSS Feed | Market analysis, company news | Daily scrape |
| VietStock | RSS Feed | Stock-specific articles | Daily scrape |
| Đầu tư | RSS Feed | Investment news | Daily scrape |
| vnstock API | REST API | OHLCV historical data | On-demand |
| Social Media | Various | Telegram groups, Facebook | Phase 2 |

---

## 4. Functional Requirements

### 4.1 API Gateway (FR-API)

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-API-001 | Accept REST requests for stock analysis | Must | ✅ Implemented |
| FR-API-002 | Route requests to appropriate agents via HTTP | Must | ✅ Implemented |
| FR-API-003 | Validate stock symbol format (`^[A-Z]{3}$`) | Must | ✅ Implemented |
| FR-API-004 | Implement rate limiting via Redis middleware | Must | ✅ Implemented |
| FR-API-005 | Request/response structured logging | Should | ✅ Implemented |
| FR-API-006 | API versioning (v1) | Should | ✅ Implemented |
| FR-API-007 | Health check endpoint | Must | ✅ Implemented |
| FR-API-008 | JWT authentication for protected endpoints | Should | Planned |
| FR-API-009 | CORS configuration | Should | Planned |

### 4.2 Technical Analysis Agent (FR-TA)

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-TA-001 | Calculate RSI (14-period default) | Must | ✅ `indicators/rsi.go` |
| FR-TA-002 | Calculate MACD (12, 26, 9 default) | Must | ✅ `indicators/macd.go` |
| FR-TA-003 | Calculate Bollinger Bands (20, 2 default) | Must | ✅ `indicators/bollinger.go` |
| FR-TA-004 | Calculate SMA (multiple periods) | Must | ✅ `indicators/sma.go` |
| FR-TA-005 | Calculate EMA (multiple periods) | Must | ✅ `indicators/ema.go` |
| FR-TA-006 | Calculate Stochastic Oscillator | Must | ✅ `indicators/stochastic.go` |
| FR-TA-007 | Calculate ADX | Should | ✅ `indicators/adx.go` |
| FR-TA-008 | Calculate ATR | Should | ✅ `indicators/atr.go` |
| FR-TA-009 | Calculate VWAP | Should | ✅ `indicators/vwap.go` |
| FR-TA-010 | Generate buy/sell/hold signals | Must | ✅ `services/technical_service.go` |
| FR-TA-011 | Calculate signal confidence score (0-100) | Must | ✅ Implemented |
| FR-TA-012 | Cache results in Redis (TTL: 5min) | Must | ✅ Implemented |
| FR-TA-013 | Support batch processing (up to 50 symbols) | Should | ✅ `handlers/analysis.go` |
| FR-TA-014 | Concurrent indicator calculation (goroutines) | Must | ✅ `sync.WaitGroup` |
| FR-TA-015 | Unit tests for all indicators | Must | ✅ `indicators/indicators_test.go` |

### 4.3 Sentiment Analysis Agent (FR-SA) — Python

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-SA-001 | Load PhoBERT model on startup | Must | ✅ `models/phobert.py` |
| FR-SA-002 | Classify sentiment (positive/negative/neutral) | Must | ✅ Implemented |
| FR-SA-003 | Calculate confidence score (0-100) | Must | ✅ Implemented |
| FR-SA-004 | Extract stock symbols from text | Must | ✅ `sentiment_analyzer.py` |
| FR-SA-005 | Apply Vietnamese stock slang dictionary | Must | ✅ `slang_dictionary.json` |
| FR-SA-006 | Handle batch text analysis (up to 100 items) | Must | ✅ Implemented |
| FR-SA-007 | Preprocess Vietnamese text | Must | ✅ Implemented |
| FR-SA-008 | Adjust confidence based on slang modifiers | Should | ✅ Implemented |
| FR-SA-009 | Expose REST API (FastAPI) | Must | ✅ `routers/analyze.py` |
| FR-SA-010 | Health check with model readiness status | Must | ✅ `routers/health.py` |
| FR-SA-011 | Cache sentiment results (Redis) | Should | Planned |
| FR-SA-012 | Pub/Sub async processing (GCP) | Could | Optional |

**Vietnamese Slang Dictionary:**

| Slang | Meaning | Sentiment Modifier |
|-------|---------|-------------------|
| lùa gà | pump and dump | Strongly Negative |
| cá mập | big investor/whale | Neutral |
| gom hàng | accumulating | Positive |
| xả hàng | dumping/selling off | Negative |
| tay to | major player | Neutral |
| đánh úp | surprise attack | Negative |
| bắt đáy | catching the bottom | Positive |
| cắt lỗ | stop loss | Neutral |
| fomo | fear of missing out | Negative |
| sideway | sideways trend | Neutral |

### 4.4 Forecast Agent (FR-FC)

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-FC-001 | Aggregate technical analysis (weight: 40%) | Must | ✅ Implemented |
| FR-FC-002 | Aggregate sentiment analysis (weight: 30%) | Must | ✅ Implemented |
| FR-FC-003 | Aggregate market context (weight: 30%) | Must | ✅ Implemented |
| FR-FC-004 | Generate recommendation (STRONG_BUY/BUY/HOLD/SELL/STRONG_SELL) | Must | ✅ Implemented |
| FR-FC-005 | Calculate combined confidence score | Must | ✅ Implemented |
| FR-FC-006 | Provide price targets (support/resistance) | Should | ✅ Implemented |
| FR-FC-007 | Risk level assessment (HIGH/MEDIUM/LOW) | Should | ✅ Implemented |

### 4.5 Master Orchestrator (FR-MO)

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-MO-001 | Coordinate multi-agent analysis workflow | Must | ✅ Implemented |
| FR-MO-002 | Generate daily analysis reports (JSON/Markdown) | Must | ✅ Implemented |
| FR-MO-003 | Send Telegram alerts | Must | ✅ Implemented |
| FR-MO-004 | Track analysis job status | Must | ✅ Implemented |
| FR-MO-005 | Retry failed agent calls (max 3) | Must | ✅ Implemented |
| FR-MO-006 | Detect hot stocks from news mentions | Should | ✅ Implemented |
| FR-MO-007 | Send email notifications | Should | Planned |

### 4.6 Workflow Automation (FR-WF) — n8n

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-WF-001 | Daily trigger at 8:00 AM (weekdays, VN time) | Must | ✅ `n8n-workflow-daily-analysis.json` |
| FR-WF-002 | RSS feed scraping from 4+ sources | Must | ✅ Implemented |
| FR-WF-003 | Spam filtering with keyword blacklist | Must | ✅ Implemented |
| FR-WF-004 | Hot stock detection (most-mentioned symbols) | Must | ✅ Implemented |
| FR-WF-005 | Trigger analysis pipeline via API | Must | ✅ Implemented |
| FR-WF-006 | Error handling and retry logic | Must | ✅ `n8n-workflow-error-handler.json` |
| FR-WF-007 | Report distribution (Telegram, Web) | Must | ✅ Implemented |

---

## 5. Non-Functional Requirements

### 5.1 Performance

| ID | Requirement | Target | Current |
|----|-------------|--------|---------|
| NFR-P-001 | Technical analysis (single stock) | < 3s | ~2s |
| NFR-P-002 | Sentiment analysis (100 articles) | < 60s | ~45s |
| NFR-P-003 | Full daily report (10 stocks) | < 10 min | ~5 min |
| NFR-P-004 | RSS scraping (100 articles) | < 60s | ~30s |
| NFR-P-005 | API response (cached) | < 100ms | ~50ms |
| NFR-P-006 | Batch analysis (50 symbols) | < 30s | ~25s |

### 5.2 Scalability

| ID | Requirement | Implementation |
|----|-------------|----------------|
| NFR-S-001 | Horizontal scaling of Go agents | Stateless Docker containers, `docker-compose scale` |
| NFR-S-002 | Concurrent indicator calculation | Go goroutines + `sync.WaitGroup` |
| NFR-S-003 | Batch processing up to 50 symbols | Parallel goroutine execution |
| NFR-S-004 | Connection pooling | GORM pool for PostgreSQL, go-redis pool for Redis |
| NFR-S-005 | Sentiment service memory limit | Docker `deploy.resources.limits: 4G` |

### 5.3 Availability

| Requirement | Target |
|-------------|--------|
| System Uptime | 99.5% |
| Recovery Time Objective (RTO) | < 10 minutes |
| Recovery Point Objective (RPO) | < 1 hour |
| Health Checks | Every 30s per service |
| Auto-restart | `restart: unless-stopped` policy |

### 5.4 Security

| ID | Requirement | Implementation |
|----|-------------|----------------|
| NFR-SEC-001 | Network isolation | Docker bridge network (172.28.0.0/16) |
| NFR-SEC-002 | n8n authentication | Basic auth (N8N_USER/N8N_PASSWORD) |
| NFR-SEC-003 | Input validation | Gin binding + regex (`^[A-Z]{3}$`), max batch 50 |
| NFR-SEC-004 | Rate limiting | Redis-backed: 100/1000/10000 req/h by tier |
| NFR-SEC-005 | Secrets management | `.env` file, never committed to VCS |
| NFR-SEC-006 | SSL/TLS (production) | Nginx reverse proxy with SSL termination |
| NFR-SEC-007 | Data privacy | GDPR/PDPA compliant, no personal data stored |
| NFR-SEC-008 | API authentication | JWT tokens (planned) |

### 5.5 Monitoring & Observability

| ID | Requirement | Implementation |
|----|-------------|----------------|
| NFR-O-001 | Metrics collection | Prometheus (profile: monitoring) |
| NFR-O-002 | Dashboard visualization | Grafana with pre-configured dashboards |
| NFR-O-003 | Structured logging | JSON format with timestamp, service, symbol, duration |
| NFR-O-004 | Health endpoints | `/health` on all services |
| NFR-O-005 | Container health checks | Docker HEALTHCHECK with curl |

---

## 6. Data Requirements

### 6.1 PostgreSQL Schema

```sql
-- Database: vnstock (PostgreSQL 15)

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

-- Sentiment analysis results
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

-- Forecast results
CREATE TABLE forecasts (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    technical_score DECIMAL(5, 2),
    sentiment_score DECIMAL(5, 2),
    market_score DECIMAL(5, 2),
    recommendation VARCHAR(20) CHECK (recommendation IN
        ('STRONG_BUY', 'BUY', 'HOLD', 'SELL', 'STRONG_SELL')),
    confidence DECIMAL(5, 2),
    support_price DECIMAL(12, 2),
    resistance_price DECIMAL(12, 2),
    reasoning TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Daily reports
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

### 6.2 Redis Cache Keys

```
# Technical analysis cache
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

### 6.3 S3 / MinIO Storage Structure

```
vnstock-data/
├── raw-data/
│   ├── news/{YYYY-MM-DD}/         # Raw RSS articles
│   ├── market-data/{YYYY-MM-DD}/  # OHLCV data
│   └── social/{YYYY-MM-DD}/       # Social media data
├── processed/
│   ├── sentiment/{YYYY-MM-DD}/    # Processed sentiment results
│   ├── technical/{YYYY-MM-DD}/    # Processed technical results
│   └── combined/{YYYY-MM-DD}/     # Combined analysis
└── reports/
    ├── daily/{YYYY-MM-DD}/        # Daily reports (JSON, Markdown, HTML)
    ├── weekly/{YYYY-WW}/          # Weekly summaries
    └── alerts/{YYYY-MM-DD}/       # Alert logs
```

---

## 7. System Constraints

### 7.1 Vietnamese Market Constraints

| Constraint | Details |
|------------|---------|
| Trading Hours | 9:00 AM – 3:00 PM Vietnam time (UTC+7) |
| Pre-market | 8:30 – 9:00 AM (order entry only) |
| Lunch Break | 11:30 AM – 1:00 PM |
| ATC Session | 2:30 – 2:45 PM (closing auction) |
| Exchanges | HSX, HNX, UPCOM |
| Symbol Format | 3 uppercase letters (e.g., VNM, FPT, VIC) |
| Price Limits | ±7% (HSX), ±10% (HNX), ±15% (UPCOM) |
| Settlement | T+2 |

### 7.2 Technical Constraints

| Constraint | Limit | Mitigation |
|------------|-------|------------|
| PhoBERT model memory | ~2.5GB | Sentiment service capped at 4GB RAM |
| Docker network | Bridge mode | Subnet 172.28.0.0/16 |
| Batch analysis | Max 50 symbols | Validation in handler |
| Sentiment batch | Max 100 texts | FastAPI validation |
| n8n workflow timeout | Configurable | Error handler workflow |
| PostgreSQL connections | Pool-limited | GORM connection pooling |
| Redis memory | Host-dependent | TTL-based eviction |

### 7.3 Deployment Constraints

| Constraint | Details |
|------------|---------|
| Docker required | All services run as containers |
| Single-host default | Docker Compose (no Kubernetes) |
| Volume persistence | Named volumes for DB, Redis, model cache |
| Environment variables | All config via `.env` file |
| Model download | PhoBERT auto-downloaded on first start (~120s) |

---

## 8. Deployment Architecture

### 8.1 Docker Compose Services

| Service | Image | Port | Profile | Health Check |
|---------|-------|------|---------|-------------|
| api-gateway | go.Dockerfile (api-gateway) | 8080 | default | curl /health |
| technical-agent | go.Dockerfile (technical-agent) | 8081 | default | — |
| forecast-agent | go.Dockerfile (forecast-agent) | 8082 | default | — |
| master-orchestrator | go.Dockerfile (master-orchestrator) | 8083 | default | — |
| sentiment | python-sentiment/Dockerfile | 8000 | default | curl /health (start_period: 120s) |
| n8n | n8nio/n8n:latest | 5678 | default | — |
| postgres | postgres:15-alpine | 5432 | default | pg_isready |
| redis | redis:7-alpine | 6379 | default | redis-cli ping |
| nginx | nginx:alpine | 80, 443 | production | — |
| prometheus | prom/prometheus:latest | 9090 | monitoring | — |
| grafana | grafana/grafana:latest | 3001 | monitoring | — |
| minio | minio/minio:latest | 9000, 9001 | development | — |

### 8.2 Deployment Profiles

| Profile | Command | Services |
|---------|---------|----------|
| Default | `docker-compose up -d` | Core: API Gateway, 4 Agents, n8n, PostgreSQL, Redis |
| Development | `--profile development` | Core + MinIO (local S3) |
| Monitoring | `--profile monitoring` | Core + Prometheus + Grafana |
| Production | `--profile production` | Core + Nginx (SSL, reverse proxy) |
| Full | `--profile production --profile monitoring` | All services |

### 8.3 Cost Estimate (Self-Hosted)

| Component | Specification | Est. Cost/Month |
|-----------|--------------|-----------------|
| VPS/Cloud VM | 4 vCPU, 16GB RAM | $40-80 |
| AWS S3 (optional) | 50GB Standard, ap-southeast-1 | $1-2 |
| Domain + SSL | Let's Encrypt | Free |
| **Total** | | **~$40-80/month** |

---

## 9. Appendices

### Appendix A: Actual Project Structure

```
analysis-stock/
├── go-services/                    # All Go microservices
│   ├── cmd/
│   │   ├── api-gateway/main.go
│   │   └── technical-agent/main.go
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── database/{postgres,redis}.go
│   │   ├── handlers/{analysis,health}.go
│   │   ├── indicators/{rsi,macd,bollinger,sma,ema,stochastic,adx,atr,vwap}.go
│   │   ├── indicators/indicators_test.go
│   │   ├── middleware/{logging,ratelimit}.go
│   │   ├── models/stock.go
│   │   └── services/{technical_service,sentiment_client}.go
│   ├── pkg/vnstock/client.go
│   ├── go.mod
│   └── Makefile
├── python-sentiment/               # Python sentiment service
│   ├── app/
│   │   ├── main.py, config.py, __init__.py
│   │   ├── models/phobert.py
│   │   ├── services/sentiment_analyzer.py
│   │   ├── routers/{analyze,health}.py
│   │   └── data/slang_dictionary.json
│   ├── requirements.txt
│   └── Dockerfile
├── scripts/init.sql
├── docker/go.Dockerfile
├── docker-compose.yml
├── n8n-workflow-daily-analysis.json
├── n8n-workflow-error-handler.json
├── technical_agent.py              # Standalone Python agent (legacy)
├── quickstart.sh
├── CLAUDE.md
└── README.md
```

### Appendix B: API Response Schemas

```json
// TechnicalResult (GET /api/v1/technical/{symbol})
{
  "symbol": "VNM",
  "timestamp": "2026-02-24T08:00:00Z",
  "price": { "open": 75.0, "high": 76.5, "low": 74.2, "close": 76.0, "volume": 1234567 },
  "rsi": 62.5,
  "macd": { "macd_line": 0.45, "signal_line": 0.32, "histogram": 0.13 },
  "bollinger": { "upper": 78.5, "middle": 75.0, "lower": 71.5 },
  "stochastic": { "k": 75.2, "d": 68.4 },
  "adx": { "adx": 25.3, "plus_di": 28.1, "minus_di": 18.5 },
  "sma_20": 74.8,
  "sma_50": 73.2,
  "ema_12": 75.5,
  "ema_26": 74.1,
  "atr": 1.85,
  "vwap": 75.3,
  "signal": "BUY",
  "confidence": 72.5,
  "score": 68.0,
  "reasons": ["RSI in neutral zone trending up", "MACD bullish crossover", "Price above SMA 20/50"]
}

// SentimentResult (POST /analyze)
{
  "results": [
    {
      "text": "VNM công bố lợi nhuận tăng 20%",
      "sentiment": "positive",
      "confidence": 92.5,
      "symbols": ["VNM"],
      "keywords": ["lợi nhuận", "tăng"]
    }
  ]
}
```

### Appendix C: Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| POSTGRES_PASSWORD | Yes | — | PostgreSQL password |
| REDIS_PASSWORD | No | — | Redis auth password |
| N8N_USER | Yes | admin | n8n basic auth user |
| N8N_PASSWORD | Yes | — | n8n basic auth password |
| N8N_HOST | No | localhost | n8n hostname |
| TELEGRAM_BOT_TOKEN | Yes | — | Telegram bot token |
| ANTHROPIC_API_KEY | No | — | Claude API key |
| OPENAI_API_KEY | No | — | OpenAI API key |
| GCP_PROJECT_ID | No | — | GCP project (optional Pub/Sub) |
| GRAFANA_USER | No | admin | Grafana admin user |
| GRAFANA_PASSWORD | No | — | Grafana admin password |
| MINIO_ROOT_USER | No | minioadmin | MinIO root user |
| MINIO_ROOT_PASSWORD | No | — | MinIO root password |
| ENABLE_RUMOR_DETECTION | No | false | Feature flag |
| ENABLE_ML_FORECAST | No | false | Feature flag |

### Appendix D: Vietnamese Market Hours Schedule

| Session | Time (ICT/UTC+7) | Activity |
|---------|-------------------|----------|
| Pre-market | 08:30 – 09:00 | Order entry |
| Morning Session | 09:00 – 11:30 | Continuous trading |
| Lunch Break | 11:30 – 13:00 | No trading |
| Afternoon Session | 13:00 – 14:30 | Continuous trading |
| ATC | 14:30 – 14:45 | Closing auction |
| Post-market | 14:45 – 15:00 | Settlement |

### Appendix E: Document Cross-References

| Document | Location | Description |
|----------|----------|-------------|
| Implementation Guide v2.0 | `docs/Implementation-Guide-v2.0.md` | Step-by-step implementation guide |
| Project Introduction | `docs/introduce/PROJECT_INTRODUCTION.md` | Project overview (bilingual) |
| Architecture Overview | `docs/introduce/ARCHITECTURE_OVERVIEW.md` | Reference architecture analysis |
| CLAUDE.md | `CLAUDE.md` | AI coding assistant context |
| README.md | `README.md` | Quick start and overview |

---

**End of SRS Document v2.0**
