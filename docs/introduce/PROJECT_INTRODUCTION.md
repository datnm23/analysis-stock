# VNStock Analysis System — Giới Thiệu Dự Án / Project Introduction

> **Purpose / Mục đích**: Tài liệu giới thiệu toàn diện về Hệ Thống Phân Tích Thị Trường Chứng Khoán Việt Nam — bao gồm giá trị kinh doanh, kiến trúc, công nghệ, và hướng dẫn triển khai.
>
> **Based on**: Universal Project Introduction Template — áp dụng từ phân tích kiến trúc của dự án World Monitor.

---

<!-- ============================================================ -->
<!-- ENGLISH VERSION -->
<!-- ============================================================ -->

# 🇬🇧 ENGLISH VERSION

---

# 🇻🇳 VNStock Analysis System

> **AI-powered automated stock analysis platform for the Vietnamese securities market (HSX/HNX/UPCOM)**

| Metadata | Value |
|---|---|
| **Version** | 1.0.0 |
| **License** | MIT |
| **Status** | Beta |
| **Last Updated** | February 2026 |

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Business Value & Objectives](#business-value--objectives)
3. [Key Features](#key-features)
4. [System Architecture](#system-architecture)
5. [Technology Stack](#technology-stack)
6. [Core Components](#core-components)
7. [Data Flow](#data-flow)
8. [Security Model](#security-model)
9. [Scalability & Performance](#scalability--performance)
10. [Deployment Architecture](#deployment-architecture)
11. [Getting Started](#getting-started)
12. [Configuration](#configuration)
13. [Contributing](#contributing)
14. [Roadmap](#roadmap)
15. [License](#license)

---

## Executive Summary

VNStock Analysis System is an AI-powered automated stock analysis platform specifically designed for the Vietnamese securities market (HSX, HNX, UPCOM). It combines multi-agent AI architecture, Vietnamese natural language processing (PhoBERT), real-time technical indicators, and automated workflow orchestration to deliver actionable investment insights.

The system aggregates data from multiple Vietnamese financial news sources (VnEconomy, CafeF, VietStock), social media channels, and market data APIs. A pipeline of specialized AI agents — Technical, Sentiment, Forecast, and Master Orchestrator — processes this data to generate daily analysis reports with buy/sell/hold signals, confidence scores, and risk assessments.

Unlike commercial alternatives that cost thousands of dollars and primarily serve English-speaking markets, VNStock is 100% open-source, purpose-built for Vietnamese market specifics including Vietnamese NLP with stock-market slang understanding, local market hours (9:00 AM – 3:00 PM VN time), and domestic exchange data integration.

**Target Users**: Individual investors, retail traders, financial analysts, fintech developers, and stock research communities in Vietnam.

**Core Problem**: Vietnamese retail investors lack affordable, automated, AI-driven analysis tools that understand local market dynamics, Vietnamese-language sentiment, and domestic news sources.

**Solution Approach**: A hybrid Go/Python microservices architecture with specialized AI agents, n8n workflow automation, and real-time Telegram/Web delivery — all deployable via Docker Compose.

---

## Business Value & Objectives

| Problem | Solution |
|---|---|
| Commercial analysis tools cost $10K+/year and don't support Vietnamese markets well | 100% open-source platform, purpose-built for HSX/HNX/UPCOM |
| Manual analysis of 15+ indicators across multiple stocks takes hours | Automated multi-agent system analyzes 10+ stocks with 15+ indicators in ~5 minutes |
| No Vietnamese NLP sentiment tools for stock market context | PhoBERT-based sentiment analysis with custom Vietnamese stock slang dictionary |
| Scattered data across VnEconomy, CafeF, VietStock, social media | Unified data pipeline with automated scraping, filtering, and aggregation via n8n |
| No real-time alerts for market events | Telegram bot + Web dashboard with automated daily reports and custom alert triggers |

### Key Metrics / KPIs

| Metric | Target | Current |
|---|---|---|
| Technical Analysis (per stock) | < 3s | ~2s |
| Sentiment Analysis (100 articles) | < 60s | ~45s |
| Full Daily Report (10 stocks) | < 10 min | ~5 min |
| System Uptime | 99.5% | Beta |
| Supported Technical Indicators | 15+ | 10 (RSI, MACD, BB, SMA, EMA, Stochastic, ADX, ATR, VWAP) |

---

## Key Features

### 📊 Multi-Dimensional Analysis

- **Technical Analysis Engine** — RSI, MACD, Bollinger Bands, SMA/EMA, Stochastic, ADX, ATR, VWAP with automated buy/sell/hold signal generation and confidence scoring
- **Batch Analysis** — Analyze up to 50 symbols concurrently with parallel indicator calculation

### 🤖 AI/ML Capabilities

- **Vietnamese NLP Sentiment** — PhoBERT model with custom stock-market slang dictionary (e.g., "lùa gà" = pump and dump)
- **Multi-Agent Synthesis** — Technical (40%) + Sentiment (30%) + Market Context (30%) weighted scoring for final recommendations
- **Rumor Detection** — Anti-spam filtering and rumor classification from social media channels

### 📰 Automated Data Pipeline

- **RSS Aggregation** — Automated scraping from VnEconomy, CafeF, VietStock, Đầu tư via n8n workflows
- **Hot Stock Detection** — Real-time identification of most-mentioned symbols across all sources
- **Daily Workflow** — Fully automated pipeline triggered at 8:00 AM weekdays: Scrape → Filter → Analyze → Report → Distribute

### 🔔 Real-Time Distribution

- **Telegram Bot** — Daily reports, custom alerts, and on-demand stock analysis via `/analyze VNM`
- **Web Dashboard** — Next.js frontend with live updates and interactive charts
- **Email Alerts** — Configurable email notifications for high-risk events and strong signals

---

## System Architecture

### Architectural Pattern

**Primary Pattern**: Hybrid Microservices (Go + Python) with Workflow Orchestration

| Aspect | Pattern | Rationale |
|---|---|---|
| **Overall Structure** | Hybrid Microservices (Go + Python) | Go for high-throughput API/analysis, Python for ML/NLP — best tool for each job |
| **API Design** | RESTful with Gin router | Simple, fast, well-suited for JSON API consumption by bots and dashboards |
| **Data Processing** | Multi-Agent Pipeline | Specialized agents enable independent scaling and domain expertise isolation |
| **Orchestration** | n8n Workflow Engine | Visual workflow builder, easy scheduling, built-in error handling without code |
| **Deployment** | Docker Compose with Profiles | Single-command deployment for dev/staging/production with optional monitoring |

### Architecture Diagram

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
│ RSI,MACD  │ │ PhoBERT   │ │ Synthesis  │ │ Coordination │
│ BB,SMA... │ │ Slang NLP │ │ Weighting  │ │ Reports      │
└─────┬─────┘ └─────┬─────┘ └─────┬──────┘ └──────┬───────┘
      │              │              │              │
      └──────────────┴──────────┬───┴──────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              ▼                 ▼                  ▼
┌──────────────────┐ ┌────────────────┐ ┌──────────────────┐
│   PostgreSQL     │ │     Redis      │ │   AWS S3 /       │
│   (:5432)        │ │    (:6379)     │ │   MinIO          │
│ Stocks, Analysis │ │ Cache, Broker  │ │ Raw, Processed,  │
│ Forecasts, Rpts  │ │ Rate Limiting  │ │ Reports, Backups │
└──────────────────┘ └────────────────┘ └──────────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              ▼                 ▼                  ▼
┌──────────────────┐ ┌────────────────┐ ┌──────────────────┐
│  Telegram Bot    │ │ Web Dashboard  │ │  Email Alerts    │
│  Daily Reports   │ │ Next.js :3000  │ │  Custom Triggers │
│  /analyze CMD    │ │ Live Charts    │ │  Risk Warnings   │
└──────────────────┘ └────────────────┘ └──────────────────┘
```

---

## Technology Stack

### Languages

| Language | Role | Percentage |
|---|---|---|
| Go 1.22 | API Gateway, Technical Analysis, Forecast Agent, Master Orchestrator | ~60% |
| Python 3.9+ | Sentiment Analysis (PhoBERT NLP), ML Models | ~30% |
| TypeScript | Web Dashboard (Next.js), n8n custom nodes | ~10% |

### Frameworks & Libraries

| Category | Technology | Purpose |
|---|---|---|
| **API Gateway** | Gin (Go) | High-performance HTTP router, middleware, request validation |
| **Sentiment ML** | FastAPI (Python) | Async Python API for PhoBERT model serving |
| **NLP Model** | Transformers + PhoBERT (vinai/phobert-base) | Vietnamese language understanding for stock sentiment |
| **ML Runtime** | PyTorch 2.1 | Deep learning inference engine |
| **ORM** | GORM (Go) | PostgreSQL ORM with migration support |
| **Cache** | go-redis v9 | Redis client for caching and rate limiting |
| **Frontend** | Next.js | Web dashboard with server-side rendering |
| **Workflow** | n8n | Visual workflow automation, scheduling, and data pipeline |
| **Market Data** | vnstock | Vietnamese stock market data fetching library |
| **Testing** | Go testing + Pytest | Unit tests, integration tests, benchmark tests |
| **Cloud** | GCP Pub/Sub (optional) | Event-driven messaging for cloud deployment |

### Data Stores

| Store | Type | Role |
|---|---|---|
| PostgreSQL 15 | Relational | Stocks, technical analysis, sentiment results, forecasts, daily reports |
| Redis 7 | Key-Value | Analysis caching (TTL-based), message broker, rate limiting |
| AWS S3 / MinIO | Object Storage | Raw data, processed data, reports, backups (structured by date) |

### Infrastructure

| Component | Provider | Role |
|---|---|---|
| **Containerization** | Docker Compose | Multi-service orchestration with profiles (dev/prod/monitoring) |
| **Reverse Proxy** | Nginx | Production load balancing, SSL termination |
| **Monitoring** | Prometheus + Grafana | Metrics collection, dashboards (scraping rate, analysis duration, errors) |
| **Async Tasks** | Celery + Redis | Background task processing, scheduled jobs |
| **Object Storage** | MinIO (dev) / AWS S3 (prod) | S3-compatible storage for data pipeline |
| **CI/CD** | GitHub Actions | Automated testing and deployment |

---

## Core Components

### 1. Go Services (`go-services/`)

| Module | Responsibility |
|---|---|
| `cmd/api-gateway/` | Central API entrypoint — routes requests, applies middleware, proxies to agents |
| `cmd/technical-agent/` | Calculates 10+ technical indicators (RSI, MACD, BB, SMA, EMA, Stochastic, ADX, ATR, VWAP), generates signals |
| `internal/indicators/` | Pure-function indicator library — reusable, testable, concurrent calculation |
| `internal/services/` | Business logic layer — caching, orchestration, external service clients |
| `internal/handlers/` | HTTP handlers — request validation, symbol pattern matching, batch analysis |
| `internal/middleware/` | Cross-cutting concerns — logging, rate limiting |
| `pkg/vnstock/` | Market data client — fetches historical price data from vnstock |

### 2. Python Sentiment Service (`python-sentiment/`)

| Module | Responsibility |
|---|---|
| `app/models/phobert.py` | PhoBERT model loading, inference, prediction (positive/negative/neutral) |
| `app/services/sentiment_analyzer.py` | Text preprocessing, Vietnamese slang normalization, keyword extraction, confidence adjustment |
| `app/data/slang_dictionary.json` | Vietnamese stock-market slang mappings with sentiment modifiers |
| `app/routers/analyze.py` | FastAPI endpoints for single and batch sentiment analysis |
| `app/routers/health.py` | Health check endpoint with model readiness status |

### 3. Workflow & Orchestration

| Module | Responsibility |
|---|---|
| `n8n-workflow-daily-analysis.json` | Daily 8 AM pipeline: RSS scraping → spam filtering → hot stock detection → analysis → report → distribution |
| `n8n-workflow-error-handler.json` | Error capture, retry logic, alert notification on pipeline failures |
| `docker-compose.yml` | Full system orchestration with 10+ services, 3 profiles (dev/prod/monitoring) |

---

## Data Flow

### Primary Data Flow (Daily Pipeline)

```
[RSS Feeds: VnEconomy, CafeF, VietStock, Đầu tư]
    │
    ▼
[n8n Workflow — 8:00 AM trigger]
    │
    ├── Scraping Phase (8:00 - 8:15)
    │   └── Fetch, Parse, Spam Filter → S3 Raw Storage
    │
    ├── Hot Stock Detection (8:15 - 8:20)
    │   └── Symbol mention analysis → Top 10 stocks
    │
    ├── Analysis Phase (8:20 - 8:35)
    │   ├── Technical Agent → 15+ indicators per stock
    │   ├── Sentiment Agent → PhoBERT on 100+ articles
    │   └── Forecast Agent → Weighted synthesis
    │
    ├── Report Generation (8:35 - 8:40)
    │   └── Master Agent → Markdown/HTML/JSON reports
    │
    └── Distribution (8:40 - 8:45)
        ├── Telegram → Daily report messages
        ├── Web Dashboard → Live update push
        └── Email → Alert notifications
```

### Caching Strategy

| Tier | Scope | TTL | Purpose |
|---|---|---|---|
| Redis L1 | Per-symbol analysis | 5–15 min | Avoid redundant indicator recalculation during batch requests |
| Redis L2 | Cross-request | 1–24h | Cache daily reports, sentiment scores between workflow runs |
| S3 Archive | Persistent | Indefinite | Raw data, processed results, report history for backtesting |

### Key Data Pipelines

| Pipeline | Input | Processing | Output |
|---|---|---|---|
| Technical Analysis | Historical OHLCV data | Concurrent calculation of RSI, MACD, BB, SMA, EMA, Stochastic, ADX, ATR, VWAP → Signal generation | `TechnicalResult` with signal, confidence, score, reasons |
| Sentiment Analysis | Vietnamese news articles | Preprocess → PhoBERT inference → Slang adjustment → Keyword extraction | Sentiment (pos/neg/neutral), confidence, keywords |
| Forecast Synthesis | Technical + Sentiment + Market | Weighted scoring (40/30/30) → Support/Resistance levels | Final recommendation with price targets |
| Daily Report | All agent outputs | Master Agent aggregation → Top picks → Market summary | JSON/Markdown report → Telegram/Web/Email |

---

## Security Model

### Defense Layers

| Layer | Mechanism | Details |
|---|---|---|
| **Network** | Docker bridge network (172.28.0.0/16), Nginx SSL | Services communicate internally; external access via Nginx reverse proxy only |
| **Authentication** | n8n basic auth, Telegram bot token, API keys | n8n protected with username/password; API keys for external service access |
| **Input Validation** | Gin binding + regex validation | Stock symbols validated against `^[A-Z]{3}$` pattern; batch requests capped at 50 symbols |
| **Rate Limiting** | Redis-backed middleware | Anonymous: 100 req/h, Registered: 1,000 req/h, Premium: 10,000 req/h |
| **Secrets Management** | Environment variables (`.env` file) | DB passwords, API keys, tokens stored in `.env`, never committed to VCS |
| **Data Privacy** | Encrypted user data, GDPR/PDPA compliant | No sensitive personal data stored; market data is public |

### Privacy Architecture

| Level | Mode | Data Locality |
|---|---|---|
| Development | Docker Compose + MinIO | 100% local — all data stays on machine |
| Production | Docker Compose + AWS S3 | Hybrid — analysis local, storage on S3 (ap-southeast-1) |
| Cloud-Optional | GCP Pub/Sub integration | Event-driven messaging to cloud services when enabled |

---

## Scalability & Performance

### Backend Optimization

| Technique | Description |
|---|---|
| Concurrent Indicator Calculation | Go goroutines + `sync.WaitGroup` calculate RSI, MACD, BB, etc. in parallel per symbol |
| Redis Caching | TTL-based cache prevents redundant analysis; content-hash keys for deduplication |
| Batch Analysis | Single API call processes up to 50 symbols with concurrent goroutine execution |
| Connection Pooling | GORM connection pool for PostgreSQL; go-redis connection pool for Redis |
| Stateless Services | Each agent is independently deployable and horizontally scalable |
| Resource Limits | Docker `deploy.resources.limits` — Sentiment service capped at 4GB RAM for model loading |

### Performance Benchmarks (4 CPU, 8GB RAM)

| Operation | Duration |
|---|---|
| Single stock technical analysis | ~2 seconds |
| 100 article sentiment analysis | ~45 seconds |
| Full daily report (10 stocks) | ~5 minutes |
| RSS scraping (100 articles) | ~30 seconds |

---

## Deployment Architecture

### Environments

| Environment | URL | Purpose |
|---|---|---|
| **Production** | http://your-server:8080 (API), :5678 (n8n), :3000 (Web) | Live system with Nginx, monitoring, and S3 storage |
| **Staging** | Same ports, `--profile monitoring` | Pre-production with Prometheus + Grafana enabled |
| **Development** | http://localhost:8080/8081/8082/8083 | Local Docker Compose with MinIO instead of S3 |

### Service Matrix

| Service | Port | Build Command | Container |
|---|---|---|---|
| API Gateway (Go) | 8080 | `docker-compose up api-gateway` | vnstock-api-gateway |
| Technical Agent (Go) | 8081 | `docker-compose up technical-agent` | vnstock-technical |
| Forecast Agent (Go) | 8082 | `docker-compose up forecast-agent` | vnstock-forecast |
| Master Orchestrator (Go) | 8083 | `docker-compose up master-orchestrator` | vnstock-orchestrator |
| Sentiment Agent (Python) | 8000 | `docker-compose up sentiment` | vnstock-sentiment |
| n8n Workflow | 5678 | `docker-compose up n8n` | vnstock-n8n |
| PostgreSQL | 5432 | `docker-compose up postgres` | vnstock-db |
| Redis | 6379 | `docker-compose up redis` | vnstock-cache |
| Grafana (optional) | 3001 | `docker-compose --profile monitoring up grafana` | vnstock-grafana |
| Prometheus (optional) | 9090 | `docker-compose --profile monitoring up prometheus` | vnstock-prometheus |
| MinIO (dev) | 9000/9001 | `docker-compose --profile development up minio` | vnstock-minio |
| Nginx (prod) | 80/443 | `docker-compose --profile production up nginx` | vnstock-nginx |

---

## Getting Started

### Prerequisites

| Requirement | Version | Notes |
|---|---|---|
| Docker | ≥ 24.0 | Required for all services |
| Docker Compose | ≥ 2.0 | Multi-service orchestration |
| Python | ≥ 3.9 | Only for standalone agent development |
| Go | ≥ 1.22 | Only for Go service development |
| Node.js | ≥ 18.0 | Only for web dashboard development |
| AWS CLI | ≥ 2.0 | Optional — for S3 storage setup |

### Quick Start

```bash
# Clone the repository
git clone https://github.com/your-username/vnstock-analysis.git
cd vnstock-analysis

# Setup environment variables
cp .env.example .env
# Edit .env with your configuration (DB passwords, Telegram token, etc.)

# Start all core services
docker-compose up -d

# Or use the quickstart script
./quickstart.sh

# Access services:
# n8n Workflow UI:    http://localhost:5678
# API Gateway:        http://localhost:8080
# Sentiment Service:  http://localhost:8000
# Web Dashboard:      http://localhost:3000

# Test the API
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/technical/VNM
```

### Environment Variables

```bash
cp .env.example .env
```

| Group | Variables | Required | Notes |
|---|---|---|---|
| **Database** | `POSTGRES_PASSWORD` | Yes | PostgreSQL password for vnstock DB |
| **Cache** | `REDIS_PASSWORD` | Optional | Redis authentication password |
| **n8n** | `N8N_USER`, `N8N_PASSWORD`, `N8N_HOST` | Yes | n8n basic auth and host config |
| **Telegram** | `TELEGRAM_BOT_TOKEN` | Yes | Bot token from @BotFather |
| **AI/LLM** | `ANTHROPIC_API_KEY`, `OPENAI_API_KEY` | Optional | For Claude/GPT-based synthesis |
| **Cloud** | `GCP_PROJECT_ID` | Optional | GCP Pub/Sub integration |
| **Storage** | AWS S3 credentials | Optional | For production S3 storage |
| **Monitoring** | `GRAFANA_USER`, `GRAFANA_PASSWORD` | Optional | Grafana dashboard access |
| **Dev Storage** | `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD` | Optional | Local S3-compatible storage |
| **Feature Flags** | `ENABLE_RUMOR_DETECTION`, `ENABLE_ML_FORECAST` | Optional | Toggle advanced features |

---

## Configuration

### Docker Compose Profiles

| Profile | Command | Description |
|---|---|---|
| Default (no profile) | `docker-compose up -d` | Core services: API Gateway, agents, n8n, PostgreSQL, Redis |
| Development | `docker-compose --profile development up -d` | Core + MinIO (local S3) |
| Monitoring | `docker-compose --profile monitoring up -d` | Core + Prometheus + Grafana |
| Production | `docker-compose --profile production up -d` | Core + Nginx reverse proxy |
| Full Stack | `docker-compose --profile production --profile monitoring up -d` | All services enabled |

### Technical Analysis Tuning

| Parameter | Default | Description |
|---|---|---|
| `RSI_PERIOD` | 14 | RSI calculation period |
| `MACD_FAST` | 12 | MACD fast EMA period |
| `MACD_SLOW` | 26 | MACD slow EMA period |
| `BOLLINGER_PERIOD` | 20 | Bollinger Bands period |
| `SMA_PERIODS` | 20, 50 | SMA calculation periods |
| `STOCHASTIC_K` | 14 | Stochastic %K period |

### Sentiment Analysis Tuning

| Parameter | Default | Description |
|---|---|---|
| `MODEL_NAME` | `vinai/phobert-base` | HuggingFace model name |
| `CONFIDENCE_THRESHOLD` | 0.7 | Minimum confidence for signal generation |
| `BATCH_SIZE` | 32 | Inference batch size |
| `MODEL_CACHE_DIR` | `/app/.cache` | Model weight cache directory |

---

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

```bash
# Development workflow
docker-compose up -d                 # Start all services
go test ./go-services/...            # Run Go tests
pytest python-sentiment/             # Run Python tests
docker-compose logs -f agent-service # Monitor logs

# Standalone development
python technical_agent.py            # Run technical agent directly
cd go-services && make build         # Build Go services
```

### Code Style

- **Go**: Follow standard Go conventions, table-driven tests, error wrapping with `%w`
- **Python**: PEP 8, type hints, structured logging
- **Commit Messages**: Conventional Commits format

---

## Roadmap

### Phase 1: MVP (Month 1-2) ✅
- [x] Core infrastructure setup (Docker Compose, PostgreSQL, Redis)
- [x] Technical analysis agent (Go) — RSI, MACD, BB, SMA, EMA, Stochastic, ADX, ATR, VWAP
- [x] Sentiment analysis agent (Python) — PhoBERT with Vietnamese slang dictionary
- [x] n8n workflow automation — daily pipeline
- [x] Telegram bot integration

### Phase 2: Enhancement (Month 3-4)
- [ ] Social media scraping (Facebook Groups, Telegram channels)
- [ ] Advanced ML models (LSTM, Transformer for price prediction)
- [ ] Rumor detection system
- [ ] Web dashboard with real-time updates (Next.js)
- [ ] Email notification system

### Phase 3: Scale (Month 5-6)
- [ ] Portfolio tracking & optimization
- [ ] Mobile app (React Native)
- [ ] Custom alerts engine
- [ ] Backtesting framework
- [ ] Multi-user support with authentication

### Phase 4: Monetization (Month 7+)
- [ ] Premium subscription tiers
- [ ] API marketplace
- [ ] White-label solution
- [ ] Institutional-grade analytics
- [ ] Partner integrations (brokerages, fintech)

---

## License

MIT License — see [LICENSE](../../LICENSE) for details.

---

## Author

**datnm** — [GitHub](https://github.com/your-username)

---

<!-- ============================================================ -->
<!-- VIETNAMESE VERSION -->
<!-- ============================================================ -->

# 🇻🇳 PHIÊN BẢN TIẾNG VIỆT

---

# 🇻🇳 Hệ Thống Phân Tích Chứng Khoán VNStock

> **Nền tảng phân tích chứng khoán tự động bằng AI cho thị trường Việt Nam (HSX/HNX/UPCOM)**

| Thông tin | Giá trị |
|---|---|
| **Phiên bản** | 1.0.0 |
| **Giấy phép** | MIT |
| **Trạng thái** | Beta |
| **Cập nhật lần cuối** | Tháng 2, 2026 |

---

## Mục Lục

1. [Tóm Tắt Tổng Quan](#tóm-tắt-tổng-quan)
2. [Giá Trị Kinh Doanh & Mục Tiêu](#giá-trị-kinh-doanh--mục-tiêu)
3. [Tính Năng Chính](#tính-năng-chính)
4. [Kiến Trúc Hệ Thống](#kiến-trúc-hệ-thống)
5. [Công Nghệ Sử Dụng](#công-nghệ-sử-dụng)
6. [Các Thành Phần Cốt Lõi](#các-thành-phần-cốt-lõi)
7. [Luồng Dữ Liệu](#luồng-dữ-liệu)
8. [Mô Hình Bảo Mật](#mô-hình-bảo-mật)
9. [Khả Năng Mở Rộng & Hiệu Năng](#khả-năng-mở-rộng--hiệu-năng)
10. [Kiến Trúc Triển Khai](#kiến-trúc-triển-khai)
11. [Bắt Đầu Nhanh](#bắt-đầu-nhanh)
12. [Cấu Hình](#cấu-hình)
13. [Đóng Góp](#đóng-góp)
14. [Lộ Trình Phát Triển](#lộ-trình-phát-triển)
15. [Giấy Phép](#giấy-phép)

---

## Tóm Tắt Tổng Quan

VNStock Analysis System là nền tảng phân tích chứng khoán tự động bằng AI, được thiết kế đặc biệt cho thị trường chứng khoán Việt Nam (HSX, HNX, UPCOM). Hệ thống kết hợp kiến trúc đa tác nhân AI, xử lý ngôn ngữ tự nhiên tiếng Việt (PhoBERT), chỉ báo kỹ thuật thời gian thực, và tự động hóa quy trình làm việc để cung cấp phân tích đầu tư chất lượng.

Hệ thống tổng hợp dữ liệu từ nhiều nguồn tin tài chính Việt Nam (VnEconomy, CafeF, VietStock), kênh mạng xã hội, và API dữ liệu thị trường. Một pipeline gồm các tác nhân AI chuyên biệt — Technical, Sentiment, Forecast, và Master Orchestrator — xử lý dữ liệu này để tạo báo cáo phân tích hàng ngày với tín hiệu mua/bán/giữ, điểm tin cậy, và đánh giá rủi ro.

Khác với các giải pháp thương mại đắt đỏ và chủ yếu phục vụ thị trường nói tiếng Anh, VNStock là mã nguồn mở 100%, được xây dựng riêng cho đặc thù thị trường Việt Nam bao gồm NLP tiếng Việt với hiểu biết về tiếng lóng chứng khoán, giờ giao dịch nội địa (9:00 AM – 3:00 PM giờ VN), và tích hợp dữ liệu sàn trong nước.

**Đối tượng sử dụng**: Nhà đầu tư cá nhân, trader, phân tích viên tài chính, lập trình viên fintech, và cộng đồng nghiên cứu chứng khoán tại Việt Nam.

**Vấn đề cốt lõi**: Nhà đầu tư nhỏ lẻ Việt Nam thiếu công cụ phân tích tự động, AI-driven, giá cả phải chăng, hiểu được động lực thị trường nội địa, tâm lý tiếng Việt, và nguồn tin trong nước.

**Cách tiếp cận giải pháp**: Kiến trúc microservices hybrid Go/Python với các AI agent chuyên biệt, n8n workflow automation, và phân phối real-time qua Telegram/Web — tất cả triển khai bằng Docker Compose.

---

## Giá Trị Kinh Doanh & Mục Tiêu

| Vấn đề | Giải pháp |
|---|---|
| Công cụ phân tích thương mại tốn $10K+/năm, không hỗ trợ tốt thị trường VN | Nền tảng mã nguồn mở 100%, xây dựng riêng cho HSX/HNX/UPCOM |
| Phân tích thủ công 15+ chỉ báo cho nhiều mã mất hàng giờ | Hệ thống đa tác nhân tự động phân tích 10+ mã với 15+ chỉ báo trong ~5 phút |
| Không có công cụ phân tích tâm lý tiếng Việt cho chứng khoán | Phân tích tâm lý dựa trên PhoBERT với từ điển tiếng lóng chứng khoán Việt |
| Dữ liệu phân tán qua VnEconomy, CafeF, VietStock, mạng xã hội | Pipeline dữ liệu thống nhất với tự động thu thập, lọc, tổng hợp qua n8n |
| Không có cảnh báo real-time cho sự kiện thị trường | Telegram bot + Web dashboard với báo cáo hàng ngày tự động và cảnh báo tùy chỉnh |

### Chỉ Số Đo Lường / KPIs

| Chỉ số | Mục tiêu | Hiện tại |
|---|---|---|
| Phân tích kỹ thuật (mỗi mã) | < 3s | ~2s |
| Phân tích tâm lý (100 bài) | < 60s | ~45s |
| Báo cáo hàng ngày đầy đủ (10 mã) | < 10 phút | ~5 phút |
| Uptime hệ thống | 99.5% | Beta |
| Số chỉ báo kỹ thuật | 15+ | 10 (RSI, MACD, BB, SMA, EMA, Stochastic, ADX, ATR, VWAP) |

---

## Tính Năng Chính

### 📊 Phân Tích Đa Chiều

- **Phân tích kỹ thuật** — RSI, MACD, Bollinger Bands, SMA/EMA, Stochastic, ADX, ATR, VWAP với tự động tạo tín hiệu mua/bán/giữ và điểm tin cậy
- **Phân tích hàng loạt** — Phân tích đồng thời tối đa 50 mã với tính toán chỉ báo song song

### 🤖 Khả Năng AI/ML

- **NLP Tiếng Việt** — Mô hình PhoBERT với từ điển tiếng lóng chứng khoán (ví dụ: "lùa gà" = pump and dump)
- **Tổng hợp Đa Tác Nhân** — Trọng số Kỹ thuật (40%) + Tâm lý (30%) + Bối cảnh thị trường (30%) cho khuyến nghị cuối cùng
- **Phát hiện tin đồn** — Lọc spam và phân loại tin đồn từ kênh mạng xã hội

### 📰 Pipeline Dữ Liệu Tự Động

- **Tổng hợp RSS** — Tự động thu thập từ VnEconomy, CafeF, VietStock, Đầu tư qua n8n workflow
- **Phát hiện Mã Nóng** — Nhận diện real-time các mã được nhắc đến nhiều nhất trên tất cả nguồn
- **Quy trình hàng ngày** — Pipeline tự động hoàn toàn vào 8:00 AM ngày trong tuần: Thu thập → Lọc → Phân tích → Báo cáo → Phân phối

### 🔔 Phân Phối Thời Gian Thực

- **Telegram Bot** — Báo cáo hàng ngày, cảnh báo tùy chỉnh, phân tích theo yêu cầu qua `/analyze VNM`
- **Web Dashboard** — Giao diện Next.js với cập nhật trực tiếp và biểu đồ tương tác
- **Email Alerts** — Thông báo email cho sự kiện rủi ro cao và tín hiệu mạnh

---

## Kiến Trúc Hệ Thống

### Mô Hình Kiến Trúc

**Mô hình chính**: Hybrid Microservices (Go + Python) với Điều phối Workflow

| Khía cạnh | Mô hình | Lý do |
|---|---|---|
| **Cấu trúc tổng thể** | Hybrid Microservices (Go + Python) | Go cho API/phân tích hiệu năng cao, Python cho ML/NLP — công cụ tốt nhất cho từng việc |
| **Thiết kế API** | RESTful với Gin router | Đơn giản, nhanh, phù hợp cho JSON API consumption bởi bot và dashboard |
| **Xử lý dữ liệu** | Pipeline Đa Tác Nhân | Các agent chuyên biệt cho phép scale độc lập và cô lập chuyên môn từng domain |
| **Điều phối** | n8n Workflow Engine | Xây dựng workflow trực quan, lên lịch dễ dàng, xử lý lỗi tích hợp không cần code |
| **Triển khai** | Docker Compose với Profiles | Triển khai một lệnh cho dev/staging/production với monitoring tùy chọn |

### Sơ Đồ Kiến Trúc

```
┌─────────────────────────────────────────────────────────────┐
│                    NGUỒN DỮ LIỆU                             │
│  RSS Feeds (VnEconomy, CafeF, VietStock, Đầu tư)           │
│  Market APIs (vnstock) │ Mạng xã hội │ Database nội bộ      │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│         ĐIỀU PHỐI WORKFLOW n8n (:5678)                       │
│  Lên lịch │ Thu thập RSS │ Pipeline │ Xử lý lỗi            │
└──────────────────────────┬──────────────────────────────────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
┌──────────────────────────────────────────────────────────────┐
│               GO API GATEWAY (:8080)                         │
│  Gin Router │ Rate Limiting │ Logging │ Xác thực request     │
└─────┬──────────────┬──────────────┬──────────────┬──────────┘
      │              │              │              │
      ▼              ▼              ▼              ▼
┌───────────┐ ┌───────────┐ ┌────────────┐ ┌──────────────┐
│ Technical │ │ Sentiment │ │  Forecast  │ │   Master     │
│   Agent   │ │   Agent   │ │   Agent    │ │ Orchestrator │
│ Go :8081  │ │ Py :8000  │ │ Go :8082   │ │  Go :8083    │
│ RSI,MACD  │ │ PhoBERT   │ │ Tổng hợp   │ │ Điều phối    │
│ BB,SMA... │ │ NLP Việt  │ │ Trọng số   │ │ Báo cáo      │
└─────┬─────┘ └─────┬─────┘ └─────┬──────┘ └──────┬───────┘
      │              │              │              │
      └──────────────┴──────────┬───┴──────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              ▼                 ▼                  ▼
┌──────────────────┐ ┌────────────────┐ ┌──────────────────┐
│   PostgreSQL     │ │     Redis      │ │   AWS S3 /       │
│   (:5432)        │ │    (:6379)     │ │   MinIO          │
│ Mã CK, Phân tích│ │ Cache, Broker  │ │ Raw, Processed,  │
│ Dự báo, Báo cáo │ │ Rate Limiting  │ │ Báo cáo, Backup  │
└──────────────────┘ └────────────────┘ └──────────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              ▼                 ▼                  ▼
┌──────────────────┐ ┌────────────────┐ ┌──────────────────┐
│  Telegram Bot    │ │ Web Dashboard  │ │  Email Alerts    │
│  Báo cáo hàng   │ │ Next.js :3000  │ │  Cảnh báo tùy   │
│  ngày, /analyze  │ │ Biểu đồ live  │ │  chỉnh rủi ro   │
└──────────────────┘ └────────────────┘ └──────────────────┘
```

---

## Công Nghệ Sử Dụng

### Ngôn Ngữ Lập Trình

| Ngôn ngữ | Vai trò | Tỷ lệ |
|---|---|---|
| Go 1.22 | API Gateway, Phân tích kỹ thuật, Forecast Agent, Master Orchestrator | ~60% |
| Python 3.9+ | Phân tích tâm lý (PhoBERT NLP), Mô hình ML | ~30% |
| TypeScript | Web Dashboard (Next.js), n8n custom nodes | ~10% |

### Framework & Thư Viện

| Danh mục | Công nghệ | Mục đích |
|---|---|---|
| **API Gateway** | Gin (Go) | HTTP router hiệu năng cao, middleware, xác thực request |
| **Sentiment ML** | FastAPI (Python) | API Python bất đồng bộ cho serving mô hình PhoBERT |
| **Mô hình NLP** | Transformers + PhoBERT (vinai/phobert-base) | Hiểu ngôn ngữ tiếng Việt cho phân tích tâm lý chứng khoán |
| **ML Runtime** | PyTorch 2.1 | Engine suy luận deep learning |
| **ORM** | GORM (Go) | ORM PostgreSQL với hỗ trợ migration |
| **Cache** | go-redis v9 | Redis client cho caching và rate limiting |
| **Frontend** | Next.js | Web dashboard với server-side rendering |
| **Workflow** | n8n | Tự động hóa workflow trực quan, lên lịch, và pipeline dữ liệu |
| **Dữ liệu thị trường** | vnstock | Thư viện lấy dữ liệu chứng khoán Việt Nam |
| **Testing** | Go testing + Pytest | Unit test, integration test, benchmark test |
| **Cloud** | GCP Pub/Sub (tùy chọn) | Messaging hướng sự kiện cho triển khai cloud |

### Kho Dữ Liệu

| Kho | Loại | Vai trò |
|---|---|---|
| PostgreSQL 15 | Quan hệ | Mã CK, phân tích kỹ thuật, kết quả tâm lý, dự báo, báo cáo hàng ngày |
| Redis 7 | Key-Value | Cache phân tích (TTL-based), message broker, rate limiting |
| AWS S3 / MinIO | Object Storage | Dữ liệu thô, dữ liệu đã xử lý, báo cáo, backup (phân theo ngày) |

### Hạ Tầng

| Thành phần | Nhà cung cấp | Vai trò |
|---|---|---|
| **Container hóa** | Docker Compose | Điều phối đa dịch vụ với profiles (dev/prod/monitoring) |
| **Reverse Proxy** | Nginx | Cân bằng tải production, kết thúc SSL |
| **Giám sát** | Prometheus + Grafana | Thu thập metrics, dashboard (tỷ lệ thu thập, thời gian phân tích, lỗi) |
| **Tác vụ bất đồng bộ** | Celery + Redis | Xử lý tác vụ nền, công việc theo lịch |
| **Lưu trữ đối tượng** | MinIO (dev) / AWS S3 (prod) | Lưu trữ tương thích S3 cho pipeline dữ liệu |
| **CI/CD** | GitHub Actions | Testing và deployment tự động |

---

## Các Thành Phần Cốt Lõi

### 1. Dịch Vụ Go (`go-services/`)

| Module | Trách nhiệm |
|---|---|
| `cmd/api-gateway/` | Điểm vào API trung tâm — routing request, áp dụng middleware, proxy đến các agent |
| `cmd/technical-agent/` | Tính toán 10+ chỉ báo kỹ thuật (RSI, MACD, BB, SMA, EMA, Stochastic, ADX, ATR, VWAP), tạo tín hiệu |
| `internal/indicators/` | Thư viện chỉ báo thuần hàm — tái sử dụng, kiểm thử được, tính toán song song |
| `internal/services/` | Tầng logic nghiệp vụ — caching, điều phối, client dịch vụ ngoài |
| `internal/handlers/` | HTTP handlers — xác thực request, khớp pattern mã CK, phân tích hàng loạt |
| `internal/middleware/` | Cross-cutting concerns — logging, rate limiting |
| `pkg/vnstock/` | Client dữ liệu thị trường — lấy dữ liệu giá lịch sử từ vnstock |

### 2. Dịch Vụ Sentiment Python (`python-sentiment/`)

| Module | Trách nhiệm |
|---|---|
| `app/models/phobert.py` | Tải mô hình PhoBERT, suy luận, dự đoán (positive/negative/neutral) |
| `app/services/sentiment_analyzer.py` | Tiền xử lý văn bản, chuẩn hóa tiếng lóng Việt, trích xuất từ khóa, điều chỉnh độ tin cậy |
| `app/data/slang_dictionary.json` | Ánh xạ tiếng lóng chứng khoán Việt Nam với hệ số tâm lý |
| `app/routers/analyze.py` | FastAPI endpoints cho phân tích tâm lý đơn lẻ và hàng loạt |
| `app/routers/health.py` | Endpoint kiểm tra sức khỏe với trạng thái sẵn sàng mô hình |

### 3. Workflow & Điều Phối

| Module | Trách nhiệm |
|---|---|
| `n8n-workflow-daily-analysis.json` | Pipeline hàng ngày 8 AM: Thu thập RSS → Lọc spam → Phát hiện mã nóng → Phân tích → Báo cáo → Phân phối |
| `n8n-workflow-error-handler.json` | Bắt lỗi, logic thử lại, thông báo cảnh báo khi pipeline thất bại |
| `docker-compose.yml` | Điều phối toàn hệ thống với 10+ dịch vụ, 3 profiles (dev/prod/monitoring) |

---

## Luồng Dữ Liệu

### Luồng Dữ Liệu Chính (Pipeline Hàng Ngày)

```
[RSS Feeds: VnEconomy, CafeF, VietStock, Đầu tư]
    │
    ▼
[n8n Workflow — Kích hoạt 8:00 AM]
    │
    ├── Giai đoạn Thu thập (8:00 - 8:15)
    │   └── Fetch, Parse, Lọc Spam → Lưu S3 Raw
    │
    ├── Phát hiện Mã Nóng (8:15 - 8:20)
    │   └── Phân tích lượt nhắc → Top 10 mã
    │
    ├── Giai đoạn Phân tích (8:20 - 8:35)
    │   ├── Technical Agent → 15+ chỉ báo/mã
    │   ├── Sentiment Agent → PhoBERT trên 100+ bài
    │   └── Forecast Agent → Tổng hợp có trọng số
    │
    ├── Tạo Báo cáo (8:35 - 8:40)
    │   └── Master Agent → Markdown/HTML/JSON
    │
    └── Phân phối (8:40 - 8:45)
        ├── Telegram → Tin nhắn báo cáo hàng ngày
        ├── Web Dashboard → Cập nhật trực tiếp
        └── Email → Thông báo cảnh báo
```

### Chiến Lược Cache

| Tầng | Phạm vi | TTL | Mục đích |
|---|---|---|---|
| Redis L1 | Phân tích từng mã | 5–15 phút | Tránh tính toán lại chỉ báo khi batch request |
| Redis L2 | Xuyên request | 1–24h | Cache báo cáo hàng ngày, điểm tâm lý giữa các lần workflow |
| S3 Archive | Lưu trữ lâu dài | Vĩnh viễn | Dữ liệu thô, kết quả đã xử lý, lịch sử báo cáo cho backtesting |

### Pipeline Dữ Liệu Chính

| Pipeline | Đầu vào | Xử lý | Đầu ra |
|---|---|---|---|
| Phân tích Kỹ thuật | Dữ liệu OHLCV lịch sử | Tính song song RSI, MACD, BB, SMA, EMA, Stochastic, ADX, ATR, VWAP → Tạo tín hiệu | `TechnicalResult` với signal, confidence, score, reasons |
| Phân tích Tâm lý | Bài báo tiếng Việt | Tiền xử lý → Suy luận PhoBERT → Điều chỉnh tiếng lóng → Trích xuất từ khóa | Tâm lý (pos/neg/neutral), confidence, keywords |
| Tổng hợp Dự báo | Kỹ thuật + Tâm lý + Thị trường | Chấm điểm trọng số (40/30/30) → Mức hỗ trợ/kháng cự | Khuyến nghị cuối với mục tiêu giá |
| Báo cáo Hàng ngày | Tất cả đầu ra agent | Master Agent tổng hợp → Top picks → Tóm tắt thị trường | JSON/Markdown → Telegram/Web/Email |

---

## Mô Hình Bảo Mật

### Các Tầng Phòng Thủ

| Tầng | Cơ chế | Chi tiết |
|---|---|---|
| **Mạng** | Docker bridge network (172.28.0.0/16), Nginx SSL | Dịch vụ giao tiếp nội bộ; truy cập ngoài chỉ qua Nginx reverse proxy |
| **Xác thực** | n8n basic auth, Telegram bot token, API keys | n8n bảo vệ bằng username/password; API key cho truy cập dịch vụ ngoài |
| **Xác thực đầu vào** | Gin binding + regex validation | Mã CK xác thực theo pattern `^[A-Z]{3}$`; batch request giới hạn 50 mã |
| **Giới hạn tốc độ** | Redis-backed middleware | Anonymous: 100 req/h, Registered: 1,000 req/h, Premium: 10,000 req/h |
| **Quản lý bí mật** | Biến môi trường (file `.env`) | Mật khẩu DB, API key, token lưu trong `.env`, không bao giờ commit vào VCS |
| **Quyền riêng tư** | Dữ liệu mã hóa, tuân thủ GDPR/PDPA | Không lưu trữ dữ liệu cá nhân nhạy cảm; dữ liệu thị trường là công khai |

### Kiến Trúc Quyền Riêng Tư

| Cấp độ | Chế độ | Vị trí dữ liệu |
|---|---|---|
| Development | Docker Compose + MinIO | 100% local — toàn bộ dữ liệu trên máy |
| Production | Docker Compose + AWS S3 | Hybrid — phân tích local, lưu trữ trên S3 (ap-southeast-1) |
| Cloud-Optional | Tích hợp GCP Pub/Sub | Messaging hướng sự kiện đến cloud khi bật |

---

## Khả Năng Mở Rộng & Hiệu Năng

### Tối Ưu Backend

| Kỹ thuật | Mô tả |
|---|---|
| Tính toán Chỉ báo Song song | Go goroutines + `sync.WaitGroup` tính RSI, MACD, BB, v.v. song song cho mỗi mã |
| Redis Caching | Cache TTL-based ngăn phân tích dư thừa; content-hash key cho loại bỏ trùng lặp |
| Phân tích Hàng loạt | Một API call xử lý tối đa 50 mã với goroutine execution song song |
| Connection Pooling | GORM connection pool cho PostgreSQL; go-redis connection pool cho Redis |
| Dịch vụ Stateless | Mỗi agent triển khai và scale ngang độc lập |
| Giới hạn Tài nguyên | Docker `deploy.resources.limits` — Dịch vụ Sentiment giới hạn 4GB RAM cho model loading |

### Hiệu Năng Benchmark (4 CPU, 8GB RAM)

| Thao tác | Thời gian |
|---|---|
| Phân tích kỹ thuật 1 mã | ~2 giây |
| Phân tích tâm lý 100 bài | ~45 giây |
| Báo cáo hàng ngày đầy đủ (10 mã) | ~5 phút |
| Thu thập RSS (100 bài) | ~30 giây |

---

## Kiến Trúc Triển Khai

### Môi Trường

| Môi trường | URL | Mục đích |
|---|---|---|
| **Production** | http://your-server:8080 (API), :5678 (n8n), :3000 (Web) | Hệ thống live với Nginx, monitoring, S3 |
| **Staging** | Cùng cổng, `--profile monitoring` | Tiền production với Prometheus + Grafana |
| **Development** | http://localhost:8080/8081/8082/8083 | Docker Compose local với MinIO thay S3 |

### Ma Trận Dịch Vụ

| Dịch vụ | Cổng | Lệnh khởi chạy | Container |
|---|---|---|---|
| API Gateway (Go) | 8080 | `docker-compose up api-gateway` | vnstock-api-gateway |
| Technical Agent (Go) | 8081 | `docker-compose up technical-agent` | vnstock-technical |
| Forecast Agent (Go) | 8082 | `docker-compose up forecast-agent` | vnstock-forecast |
| Master Orchestrator (Go) | 8083 | `docker-compose up master-orchestrator` | vnstock-orchestrator |
| Sentiment Agent (Python) | 8000 | `docker-compose up sentiment` | vnstock-sentiment |
| n8n Workflow | 5678 | `docker-compose up n8n` | vnstock-n8n |
| PostgreSQL | 5432 | `docker-compose up postgres` | vnstock-db |
| Redis | 6379 | `docker-compose up redis` | vnstock-cache |
| Grafana (tùy chọn) | 3001 | `docker-compose --profile monitoring up grafana` | vnstock-grafana |
| Prometheus (tùy chọn) | 9090 | `docker-compose --profile monitoring up prometheus` | vnstock-prometheus |
| MinIO (dev) | 9000/9001 | `docker-compose --profile development up minio` | vnstock-minio |
| Nginx (prod) | 80/443 | `docker-compose --profile production up nginx` | vnstock-nginx |

---

## Bắt Đầu Nhanh

### Yêu Cầu Tiên Quyết

| Yêu cầu | Phiên bản | Ghi chú |
|---|---|---|
| Docker | ≥ 24.0 | Bắt buộc cho tất cả dịch vụ |
| Docker Compose | ≥ 2.0 | Điều phối đa dịch vụ |
| Python | ≥ 3.9 | Chỉ cho phát triển standalone agent |
| Go | ≥ 1.22 | Chỉ cho phát triển Go service |
| Node.js | ≥ 18.0 | Chỉ cho phát triển web dashboard |
| AWS CLI | ≥ 2.0 | Tùy chọn — cho thiết lập S3 storage |

### Khởi Động Nhanh

```bash
# Clone repository
git clone https://github.com/your-username/vnstock-analysis.git
cd vnstock-analysis

# Thiết lập biến môi trường
cp .env.example .env
# Chỉnh sửa .env với cấu hình (mật khẩu DB, Telegram token, v.v.)

# Khởi chạy tất cả dịch vụ
docker-compose up -d

# Hoặc dùng script khởi động nhanh
./quickstart.sh

# Truy cập dịch vụ:
# n8n Workflow UI:    http://localhost:5678
# API Gateway:        http://localhost:8080
# Sentiment Service:  http://localhost:8000
# Web Dashboard:      http://localhost:3000

# Kiểm tra API
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/technical/VNM
```

### Biến Môi Trường

```bash
cp .env.example .env
```

| Nhóm | Biến | Bắt buộc | Ghi chú |
|---|---|---|---|
| **Database** | `POSTGRES_PASSWORD` | Có | Mật khẩu PostgreSQL cho vnstock DB |
| **Cache** | `REDIS_PASSWORD` | Tùy chọn | Mật khẩu xác thực Redis |
| **n8n** | `N8N_USER`, `N8N_PASSWORD`, `N8N_HOST` | Có | n8n basic auth và cấu hình host |
| **Telegram** | `TELEGRAM_BOT_TOKEN` | Có | Bot token từ @BotFather |
| **AI/LLM** | `ANTHROPIC_API_KEY`, `OPENAI_API_KEY` | Tùy chọn | Cho tổng hợp dựa trên Claude/GPT |
| **Cloud** | `GCP_PROJECT_ID` | Tùy chọn | Tích hợp GCP Pub/Sub |
| **Lưu trữ** | AWS S3 credentials | Tùy chọn | Cho storage S3 production |
| **Giám sát** | `GRAFANA_USER`, `GRAFANA_PASSWORD` | Tùy chọn | Truy cập dashboard Grafana |
| **Feature Flags** | `ENABLE_RUMOR_DETECTION`, `ENABLE_ML_FORECAST` | Tùy chọn | Bật/tắt tính năng nâng cao |

---

## Cấu Hình

### Docker Compose Profiles

| Profile | Lệnh | Mô tả |
|---|---|---|
| Mặc định | `docker-compose up -d` | Dịch vụ cốt lõi: API Gateway, agents, n8n, PostgreSQL, Redis |
| Development | `docker-compose --profile development up -d` | Cốt lõi + MinIO (S3 local) |
| Monitoring | `docker-compose --profile monitoring up -d` | Cốt lõi + Prometheus + Grafana |
| Production | `docker-compose --profile production up -d` | Cốt lõi + Nginx reverse proxy |
| Full Stack | `docker-compose --profile production --profile monitoring up -d` | Tất cả dịch vụ |

### Tinh Chỉnh Phân Tích Kỹ Thuật

| Tham số | Mặc định | Mô tả |
|---|---|---|
| `RSI_PERIOD` | 14 | Chu kỳ tính RSI |
| `MACD_FAST` | 12 | Chu kỳ EMA nhanh MACD |
| `MACD_SLOW` | 26 | Chu kỳ EMA chậm MACD |
| `BOLLINGER_PERIOD` | 20 | Chu kỳ Bollinger Bands |
| `SMA_PERIODS` | 20, 50 | Chu kỳ tính SMA |
| `STOCHASTIC_K` | 14 | Chu kỳ Stochastic %K |

### Tinh Chỉnh Phân Tích Tâm Lý

| Tham số | Mặc định | Mô tả |
|---|---|---|
| `MODEL_NAME` | `vinai/phobert-base` | Tên model HuggingFace |
| `CONFIDENCE_THRESHOLD` | 0.7 | Ngưỡng tin cậy tối thiểu cho tạo tín hiệu |
| `BATCH_SIZE` | 32 | Kích thước batch suy luận |
| `MODEL_CACHE_DIR` | `/app/.cache` | Thư mục cache trọng số model |

---

## Đóng Góp

Xem [CONTRIBUTING.md](../../CONTRIBUTING.md) để biết hướng dẫn chi tiết.

```bash
# Quy trình phát triển
docker-compose up -d                 # Khởi chạy tất cả dịch vụ
go test ./go-services/...            # Chạy Go test
pytest python-sentiment/             # Chạy Python test
docker-compose logs -f agent-service # Theo dõi log

# Phát triển standalone
python technical_agent.py            # Chạy technical agent trực tiếp
cd go-services && make build         # Build Go services
```

### Quy Tắc Code

- **Go**: Tuân thủ Go conventions, table-driven tests, error wrapping với `%w`
- **Python**: PEP 8, type hints, structured logging
- **Commit Messages**: Conventional Commits format

---

## Lộ Trình Phát Triển

### Giai đoạn 1: MVP (Tháng 1-2) ✅
- [x] Thiết lập hạ tầng cốt lõi (Docker Compose, PostgreSQL, Redis)
- [x] Technical analysis agent (Go) — RSI, MACD, BB, SMA, EMA, Stochastic, ADX, ATR, VWAP
- [x] Sentiment analysis agent (Python) — PhoBERT với từ điển tiếng lóng Việt
- [x] n8n workflow automation — pipeline hàng ngày
- [x] Tích hợp Telegram bot

### Giai đoạn 2: Nâng Cấp (Tháng 3-4)
- [ ] Thu thập mạng xã hội (Facebook Groups, kênh Telegram)
- [ ] Mô hình ML nâng cao (LSTM, Transformer cho dự đoán giá)
- [ ] Hệ thống phát hiện tin đồn
- [ ] Web dashboard với cập nhật real-time (Next.js)
- [ ] Hệ thống thông báo email

### Giai đoạn 3: Mở Rộng (Tháng 5-6)
- [ ] Theo dõi & tối ưu danh mục
- [ ] Ứng dụng di động (React Native)
- [ ] Engine cảnh báo tùy chỉnh
- [ ] Framework backtesting
- [ ] Hỗ trợ đa người dùng với xác thực

### Giai đoạn 4: Thương Mại Hóa (Tháng 7+)
- [ ] Gói đăng ký premium
- [ ] API marketplace
- [ ] Giải pháp white-label
- [ ] Phân tích cấp tổ chức
- [ ] Tích hợp đối tác (công ty chứng khoán, fintech)

---

## Giấy Phép

MIT License — xem [LICENSE](../../LICENSE) để biết chi tiết.

---

## Tác Giả

**datnm** — [GitHub](https://github.com/your-username)

---

## ⚖️ Tuyên Bố Miễn Trừ Trách Nhiệm

Hệ thống này được phát triển cho mục đích giáo dục và nghiên cứu. Tác giả không chịu trách nhiệm cho bất kỳ tổn thất tài chính nào phát sinh từ việc sử dụng hệ thống. Các phân tích và khuyến nghị chỉ mang tính chất tham khảo — người dùng phải tự nghiên cứu kỹ lưỡng và chịu trách nhiệm cho các quyết định đầu tư của mình.

---

**Made with ❤️ in Vietnam 🇻🇳**
