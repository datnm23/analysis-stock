# Software Requirements Specification (SRS)
# VN Stock Analysis System
## Version 1.0

**Document ID:** SRS-VNSTOCK-001
**Date:** January 2026
**Status:** Draft
**Reference Standard:** IEEE 830-1998

---

## Table of Contents

1. [Introduction](#1-introduction)
2. [Overall Description](#2-overall-description)
3. [External Interface Requirements](#3-external-interface-requirements)
4. [Functional Requirements](#4-functional-requirements)
5. [Non-Functional Requirements](#5-non-functional-requirements)
6. [Data Requirements](#6-data-requirements)
7. [System Constraints](#7-system-constraints)
8. [Appendices](#8-appendices)

---

## 1. Introduction

### 1.1 Purpose

This Software Requirements Specification (SRS) document describes the functional and non-functional requirements for the Vietnamese Stock Market Analysis System (VN Stock Analysis System). The system provides AI-powered automated stock analysis for the Vietnamese market (HSX/HNX/UPCOM exchanges).

**Intended Audience:**
- Development Team
- QA Engineers
- Product Managers
- System Architects
- DevOps Engineers

### 1.2 Scope

**System Name:** VN Stock Analysis System (He Thong Phan Tich Chung Khoan Viet Nam)

**Capabilities:**
- Automated data collection from Vietnamese financial news sources
- Technical analysis with 15+ indicators
- Vietnamese NLP sentiment analysis using PhoBERT
- Multi-agent synthesis for investment recommendations
- Real-time alerts via Telegram
- Web dashboard for visualization
- API services for third-party integration

**Out of Scope:**
- Automated trade execution
- Licensed financial advice
- Real-time tick data (uses daily OHLCV)
- International markets

### 1.3 Definitions, Acronyms, and Abbreviations

| Term | Definition |
|------|------------|
| HSX | Ho Chi Minh Stock Exchange (HOSE) |
| HNX | Hanoi Stock Exchange |
| UPCOM | Unlisted Public Company Market |
| VN-Index | Vietnam Stock Market Benchmark Index |
| OHLCV | Open, High, Low, Close, Volume |
| RSI | Relative Strength Index |
| MACD | Moving Average Convergence Divergence |
| SMA/EMA | Simple/Exponential Moving Average |
| ADX | Average Directional Index |
| ATR | Average True Range |
| VWAP | Volume Weighted Average Price |
| PhoBERT | Vietnamese BERT Language Model |
| n8n | Workflow Automation Platform |
| "Lua ga" | Vietnamese slang: pump and dump scheme |
| "Cay thong" | Vietnamese slang: bullish candlestick pattern |
| "Con tep" | Vietnamese slang: small retail investor |
| "Ca map" | Vietnamese slang: large institutional investor |

### 1.4 References

1. IEEE 830-1998 - Recommended Practice for Software Requirements Specifications
2. vnstock Python Library Documentation (https://github.com/thinh-vu/vnstock)
3. PhoBERT Model (vinai/phobert-base)
4. Vietnamese Securities Law and Circular 96/2020/TT-BTC
5. n8n Documentation (https://docs.n8n.io)
6. Project Architecture Document: `vnstock-analysis-architecture.md`

### 1.5 Document Overview

- **Section 2:** System context, functions, user classes, constraints
- **Section 3:** User, software, and communication interfaces
- **Section 4:** Detailed functional requirements by module
- **Section 5:** Performance, security, reliability requirements
- **Section 6:** Database schema and data storage
- **Section 7:** Vietnamese market-specific constraints
- **Section 8:** Appendices with slang dictionary and schemas

---

## 2. Overall Description

### 2.1 Product Perspective

The VN Stock Analysis System is a standalone, self-contained platform consisting of multiple microservices orchestrated via Docker Compose. It integrates with external data sources and distribution channels.

**System Architecture:**

```
┌─────────────────────────────────────────────────────────────────┐
│                     DATA INGESTION LAYER                        │
│  RSS Feeds (VnEconomy, CafeF, VietStock) │ vnstock API │ Social │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                   N8N WORKFLOW ORCHESTRATOR                      │
│            (Scheduling, Data Flow, Error Handling)               │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    AI AGENT PROCESSING LAYER                     │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐             │
│  │  Technical   │ │  Sentiment   │ │   Forecast   │             │
│  │    Agent     │ │    Agent     │ │    Agent     │             │
│  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘             │
│         └────────────────┼────────────────┘                      │
│                          ▼                                       │
│                 ┌──────────────┐                                 │
│                 │Master Agent  │                                 │
│                 │(Orchestrator)│                                 │
│                 └──────────────┘                                 │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                   OUTPUT & DISTRIBUTION LAYER                    │
│      Telegram Bot │ Web Dashboard │ Email │ S3 Archive │ API    │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Product Functions

| ID | Function | Description |
|----|----------|-------------|
| F1 | Data Collection | Automated scraping of news, market data, social media |
| F2 | Technical Analysis | Calculate 15+ technical indicators, generate signals |
| F3 | Sentiment Analysis | Vietnamese NLP analysis of news and social content |
| F4 | Trend Forecasting | Multi-factor synthesis for recommendations |
| F5 | Report Generation | Daily/weekly reports in multiple formats |
| F6 | Real-time Alerting | Push notifications for significant events |
| F7 | Web Dashboard | Interactive visualization and monitoring |
| F8 | Telegram Bot | User interaction via Telegram commands |
| F9 | API Services | RESTful API for third-party integration |

### 2.3 User Classes and Characteristics

**UC1: Retail Investors (Con Tep)**
- Primary users seeking stock analysis
- Prefer Vietnamese language interface
- Access via Telegram bot and web dashboard
- Need: Simple recommendations, alerts

**UC2: Active Traders**
- Require detailed technical indicators
- Need real-time alerts for price movements
- Want multiple watchlist management
- Need: Full analysis data, API access

**UC3: System Administrators**
- Manage n8n workflows and system health
- Configure data sources and schedules
- Monitor performance metrics
- Need: Admin interfaces, logging access

**UC4: API Consumers**
- Third-party developers integrating analysis
- Require programmatic access
- Need: REST API, documentation, rate limits

### 2.4 Operating Environment

| Component | Specification |
|-----------|---------------|
| Container Runtime | Docker 20.10+, Docker Compose 2.0+ |
| Host OS | Linux (Ubuntu 22.04+ recommended) |
| Memory | Minimum 8GB RAM (16GB for PhoBERT GPU) |
| Storage | Minimum 50GB (SSD recommended) |
| Network | Stable internet for external APIs |
| Timezone | Asia/Ho_Chi_Minh (UTC+7) |
| Cloud | AWS S3 for storage (optional MinIO for dev) |

### 2.5 Design and Implementation Constraints

1. **Vietnamese Market Hours:** System must respect trading hours (9:00-15:00 Vietnam time)
2. **vnstock Library:** Data fetching depends on vnstock library availability
3. **PhoBERT Requirements:** 4GB+ memory for model inference
4. **External API Limits:** Rate limiting on VietStock, FiinGroup APIs
5. **Legal Compliance:** All outputs must include investment disclaimer
6. **Language:** User-facing content in Vietnamese, code in English

### 2.6 Assumptions and Dependencies

**Assumptions:**
- Vietnamese stock market operates normally (no extended closures)
- RSS feeds remain accessible and maintain current format
- AWS S3 service availability > 99.9%

**Dependencies:**
- vnstock Python library for market data
- PhoBERT model from HuggingFace
- External RSS feeds (VnEconomy, CafeF, VietStock)
- Telegram Bot API
- AWS S3 for data storage

---

## 3. External Interface Requirements

### 3.1 User Interfaces

#### UI-1: Web Dashboard (Next.js)

| Screen | Description |
|--------|-------------|
| Login/Register | User authentication |
| Dashboard | Market overview, top movers, hot stocks |
| Stock Detail | Full analysis for single stock |
| Watchlist | User's monitored stocks |
| Alerts | Alert configuration and history |
| Reports | Historical report browser |
| Settings | User preferences |

**Requirements:**
- Responsive design (mobile/tablet/desktop)
- Vietnamese language interface
- Real-time updates via WebSocket
- Color-coded recommendations (green=buy, red=sell)
- Chart visualizations for technical indicators

#### UI-2: Telegram Bot Interface

| Command | Description |
|---------|-------------|
| /start | Welcome message and help |
| /analyze <SYMBOL> | Request stock analysis |
| /report | Get latest daily report |
| /subscribe <SYMBOLS> | Subscribe to alerts |
| /unsubscribe <SYMBOLS> | Unsubscribe from alerts |
| /watchlist | View/manage watchlist |
| /settings | Configure preferences |
| /help | Show available commands |

**Requirements:**
- Vietnamese language responses
- Markdown formatting for readability
- Inline keyboards for quick actions
- Rate limiting per user (10 requests/minute)

#### UI-3: n8n Admin Interface (Port 5678)

- Workflow designer and editor
- Execution history and logs
- Credential management
- Error monitoring

### 3.2 Software Interfaces

#### SI-1: vnstock Python Library

| Function | Purpose |
|----------|---------|
| `stock_historical_data()` | Fetch OHLCV data |
| `stock_intraday_data()` | Intraday prices (optional) |
| `listing_companies()` | List of tradable stocks |

**Data Format:** pandas DataFrame
**Resolution:** Daily (1D)
**History:** Up to 10 years

#### SI-2: External RSS Feeds

| Source | URL | Content |
|--------|-----|---------|
| VnEconomy | vneconomy.vn/rss/chung-khoan.rss | Financial news |
| CafeF | cafef.vn/chung-khoan.rss | Market analysis |
| VietStock | vietstock.vn/chung-khoan.rss | Stock updates |
| Dau Tu | dautucophieu.net/feed | Investment news |

**Format:** RSS 2.0 / Atom XML
**Update Frequency:** Every 15-60 minutes

#### SI-3: AWS S3 API

| Operation | Purpose |
|-----------|---------|
| PutObject | Store raw data, processed results, reports |
| GetObject | Retrieve data for analysis |
| ListObjects | Browse stored content |
| DeleteObject | Clean up old data (retention policy) |

**Authentication:** AWS Access Key + Secret Key
**Bucket:** vnstock-data (configurable)

#### SI-4: Telegram Bot API

| Method | Purpose |
|--------|---------|
| sendMessage | Send analysis results, alerts |
| getUpdates | Receive user commands |
| setWebhook | Real-time message handling |
| sendDocument | Send report files |

**Authentication:** Bot Token
**Rate Limit:** 30 messages/second

#### SI-5: Optional Paid APIs

| API | Purpose | Authentication |
|-----|---------|----------------|
| FiinGroup | Premium financial data | API Key |
| VietStock | Real-time quotes | API Key |
| Apify | Facebook scraping | API Token |

### 3.3 Communications Interfaces

| Protocol | Usage |
|----------|-------|
| HTTP/HTTPS | REST API, external services |
| WebSocket | Real-time dashboard updates |
| Redis Pub/Sub | Internal service messaging |
| SMTP | Email notifications (optional) |

**Security:**
- TLS 1.2+ for all external communications
- API endpoints require JWT authentication
- Internal services on isolated Docker network

---

## 4. Functional Requirements

### 4.1 Data Collection Module (FR-DC)

#### FR-DC-001: RSS Feed Scraping
**Priority:** High
**Input:** List of configured RSS feed URLs
**Process:**
1. Fetch RSS/Atom XML from each source
2. Parse feed entries (title, description, link, pubDate)
3. Deduplicate based on URL
4. Filter spam content

**Output:** Structured news data stored in S3
**Schedule:** Daily at 8:00 AM (weekdays)
**Error Handling:** Continue on individual feed failure, log errors

#### FR-DC-002: Spam Filtering
**Priority:** High
**Input:** Raw news articles
**Process:** Keyword-based filtering with Vietnamese spam patterns
**Blacklist Keywords:**
- "khuyen mai", "dang ky ngay" (promotional)
- "group vip", "bao lai" (scam indicators)
- "lua ga", "song danh" (manipulation warnings)

**Output:** Filtered, clean news articles
**Metric:** >95% spam removal accuracy

#### FR-DC-003: Stock Symbol Extraction
**Priority:** Medium
**Input:** News text content
**Process:**
1. Regex pattern matching: `\b[A-Z]{3}\b`
2. Validate against known HSX/HNX/UPCOM symbols
3. Count mention frequency per symbol

**Output:** Symbol mention counts per article
**Validation:** Only valid 3-letter symbols accepted

#### FR-DC-004: Market Data Fetching
**Priority:** High
**Input:** Stock symbol, date range (default 90 days)
**Process:**
1. Call vnstock library
2. Retrieve OHLCV data
3. Handle missing data gracefully
4. Cache in Redis (TTL: 1 hour)

**Output:** pandas DataFrame with price/volume data
**Error Handling:** Retry 3x with exponential backoff

#### FR-DC-005: Social Media Scraping (Phase 2)
**Priority:** Medium
**Sources:** Facebook Groups, Telegram Channels
**Process:** Use Apify/Telethon for extraction
**Output:** Social posts with metadata
**Rate Limit:** Respect platform ToS

### 4.2 Technical Analysis Agent (FR-TA)

#### FR-TA-001: RSI Calculation
**Priority:** High
**Input:** Close prices (14-day window)
**Formula:** RSI = 100 - (100 / (1 + RS)), RS = Avg Gain / Avg Loss
**Output:** RSI value (0-100)
**Signal Generation:**
- RSI < 30: Oversold (bullish signal)
- RSI > 70: Overbought (bearish signal)

#### FR-TA-002: MACD Calculation
**Priority:** High
**Input:** Close prices
**Parameters:** Fast EMA=12, Slow EMA=26, Signal=9
**Output:** MACD line, Signal line, Histogram
**Signal Generation:**
- MACD crosses above Signal: Bullish
- MACD crosses below Signal: Bearish

#### FR-TA-003: Bollinger Bands
**Priority:** High
**Input:** Close prices (20-day window)
**Formula:**
- Middle = SMA(20)
- Upper = Middle + 2*StdDev
- Lower = Middle - 2*StdDev

**Output:** Upper, Middle, Lower band values
**Signal Generation:**
- Price < Lower: Potential buy
- Price > Upper: Potential sell

#### FR-TA-004: Moving Averages
**Priority:** High
**Types:** SMA(20), SMA(50), EMA(12), EMA(26)
**Signal Generation:**
- Golden Cross (SMA20 > SMA50): Bullish
- Death Cross (SMA20 < SMA50): Bearish
- Price above SMA: Uptrend
- Price below SMA: Downtrend

#### FR-TA-005: Stochastic Oscillator
**Priority:** Medium
**Input:** High, Low, Close (14-day)
**Output:** %K, %D values (0-100)
**Signal Generation:**
- %K < 20: Oversold
- %K > 80: Overbought

#### FR-TA-006: ADX (Trend Strength)
**Priority:** Medium
**Input:** High, Low, Close (14-day)
**Output:** ADX value, +DI, -DI
**Interpretation:**
- ADX > 25: Strong trend
- ADX < 20: Weak/no trend
- +DI > -DI: Bullish
- -DI > +DI: Bearish

#### FR-TA-007: Volume Analysis
**Priority:** Medium
**Input:** Volume data
**Process:** Compare current volume to 20-day average
**Output:** Volume ratio, spike detection
**Signal:** Volume > 1.5x average = significant activity

#### FR-TA-008: Support/Resistance Calculation
**Priority:** Medium
**Algorithm:** Local minima/maxima detection (scipy.signal.argrelextrema)
**Output:** Top 3 resistance levels, Top 3 support levels
**Window:** 20-day lookback

#### FR-TA-009: Signal Generation
**Priority:** High
**Input:** All calculated indicator values
**Process:**
1. Score each indicator (-1 to +1)
2. Apply weights (RSI: 15%, MACD: 20%, BB: 15%, MA: 20%, Volume: 10%, ADX: 10%, Stoch: 10%)
3. Sum weighted scores
4. Map to recommendation

**Output:**
- Recommendation: STRONG BUY / BUY / HOLD / SELL / STRONG SELL
- Confidence: 0-100%
- Reasoning: List of contributing signals

**Thresholds:**
- Score > 0.6: STRONG BUY
- Score 0.2 to 0.6: BUY
- Score -0.2 to 0.2: HOLD
- Score -0.6 to -0.2: SELL
- Score < -0.6: STRONG SELL

### 4.3 Sentiment Analysis Agent (FR-SA)

#### FR-SA-001: Vietnamese Text Preprocessing
**Priority:** High
**Input:** Raw Vietnamese text
**Process:**
1. Lowercase conversion
2. Vietnamese slang normalization (see Appendix A)
3. Remove special characters (preserve stock symbols)
4. Tokenization

**Output:** Normalized text for model input

#### FR-SA-002: Sentiment Classification
**Priority:** High
**Model:** PhoBERT (vinai/phobert-base)
**Classes:** Positive, Neutral, Negative
**Input:** Preprocessed text (max 256 tokens)
**Output:**
- Sentiment class
- Confidence score (0-1)
- Probability distribution

**Performance Target:** >85% accuracy on Vietnamese financial text

#### FR-SA-003: Batch Sentiment Analysis
**Priority:** High
**Input:** Array of news/social texts
**Output:**
- Individual sentiment per text
- Aggregate metrics:
  - positive_ratio (0-1)
  - negative_ratio (0-1)
  - neutral_ratio (0-1)
  - overall_score (-1 to +1)

**Performance:** Process 100 texts in < 45 seconds

#### FR-SA-004: Rumor Detection
**Priority:** Medium
**Input:** Social media texts, Official news texts
**Process:**
1. Extract stock symbols from both sources
2. Identify symbols mentioned only in social media
3. Calculate mention frequency
4. Flag high-frequency social-only mentions

**Output:**
- List of potential rumor symbols
- Mention count per symbol
- Risk level (MEDIUM if < 10 mentions, HIGH if >= 10)
- Warning message in Vietnamese

#### FR-SA-005: Entity Extraction
**Priority:** Medium
**Input:** Vietnamese text
**Process:** Regex + validation against known symbols
**Output:** List of stock symbols mentioned
**Validation:** 3-letter uppercase, exists in symbol database

### 4.4 Forecast Agent (FR-FC)

#### FR-FC-001: Multi-Signal Synthesis
**Priority:** High
**Input:**
- Technical signals (from FR-TA)
- Sentiment data (from FR-SA)
- Market context

**Weights:**
- Technical: 40%
- Sentiment: 30%
- Market Context: 30%

**Output:** Combined score (0-1)

#### FR-FC-002: Market Context Analysis
**Priority:** Medium
**Inputs:**
- VN-Index daily change (%)
- Market volume vs average
- Foreign investor net flow

**Output:** Market context score (-1 to +1)
**Interpretation:**
- Positive VN-Index + high volume + foreign buying = bullish context
- Negative VN-Index + low volume + foreign selling = bearish context

#### FR-FC-003: Risk Assessment
**Priority:** High
**Factors:**
- Price volatility (ATR-based)
- Negative sentiment ratio
- Rumor presence
- Volume anomalies

**Output:**
- Risk Level: LOW / MEDIUM / HIGH
- Risk explanation in Vietnamese

**Thresholds:**
- LOW: Volatility < 3%, No rumors, Negative sentiment < 30%
- MEDIUM: Volatility 3-5%, Minor concerns
- HIGH: Volatility > 5% OR rumors detected OR negative > 50%

#### FR-FC-004: Price Target Estimation
**Priority:** Medium
**Methods:**
1. Bollinger Band mean reversion
2. Fibonacci retracement levels (23.6%, 38.2%, 50%, 61.8%)

**Output:**
- Short-term target (1-2 weeks)
- Stop loss level
- Fibonacci levels array

### 4.5 Master Agent (FR-MA)

#### FR-MA-001: Agent Orchestration
**Priority:** High
**Process:**
1. Call Technical Analysis Agent
2. Call Sentiment Analysis Agent
3. Call Forecast Agent with combined data
4. Aggregate results

**Error Handling:** Continue with partial results if one agent fails
**Logging:** Log each agent's execution time and status

#### FR-MA-002: Stock Analysis
**Priority:** High
**Input:** Stock symbol, Analysis date
**Output:** Comprehensive analysis JSON containing:
- Technical analysis results
- Sentiment analysis results
- Forecast and recommendation
- Risk assessment
- Price targets
- Generated timestamp

#### FR-MA-003: Daily Report Generation
**Priority:** High
**Trigger:** 8:00 AM weekdays (Vietnam time)
**Process:**
1. Identify hot stocks from news mentions
2. Analyze top 5 hot stocks
3. Get market overview
4. Compile report

**Output:**
- Market overview summary
- Hot stocks list with mention counts
- Top recommendations ranked by confidence
- Detailed analysis for each hot stock
- Alert summary
- Generated timestamp

**Formats:** JSON, Markdown, HTML

#### FR-MA-004: Alert Generation
**Priority:** High
**Alert Types:**
- HIGH_RISK: Risk level = HIGH
- STRONG_SIGNAL: Recommendation = STRONG BUY or STRONG SELL
- RUMOR: Rumor detected with > 5 mentions

**Action:** Push to Telegram urgent alerts channel
**Content:** Symbol, alert type, brief explanation in Vietnamese

### 4.6 Report Generation (FR-RG)

#### FR-RG-001: Markdown Report
**Priority:** High
**Template:** Jinja2 template
**Sections:**
1. Header with date
2. Market overview
3. Hot stocks summary table
4. Individual stock analysis
5. Disclaimer

**Language:** Vietnamese

#### FR-RG-002: HTML Report
**Priority:** Medium
**Features:**
- Responsive CSS styling
- Color-coded recommendations
- Gradient header design
- Stock cards layout

#### FR-RG-003: JSON Report
**Priority:** High
**Purpose:** API consumption
**Schema:** OpenAPI documented
**Content:** All analysis data in structured format

#### FR-RG-004: Disclaimer Injection
**Priority:** Critical
**Requirement:** ALL reports MUST include legal disclaimer
**Text (Vietnamese):**
```
⚠️ LƯU Ý QUAN TRỌNG:
Đây chỉ là thông tin tham khảo, không phải lời khuyên đầu tư.
Nhà đầu tư cần tự nghiên cứu và chịu trách nhiệm cho quyết định của mình.
Thông tin có thể không chính xác hoặc bị trễ. Đầu tư chứng khoán có rủi ro.
```

### 4.7 Distribution Layer (FR-DL)

#### FR-DL-001: Telegram Notification
**Priority:** High
**Channels:**
- @vnstock_daily: Daily reports
- @vnstock_alerts: High-risk/strong signal alerts

**Format:** Markdown with emoji indicators
**Rate Limit:** Max 20 messages/minute per channel

#### FR-DL-002: Web Dashboard Display
**Priority:** High
**Protocol:** WebSocket for real-time updates
**Components:**
- MarketOverview: VN-Index, market stats
- StockCard: Individual stock summaries
- AlertBanner: Urgent notifications
- AnalysisDetail: Full stock analysis

#### FR-DL-003: Email Alerts (Optional)
**Priority:** Low
**Trigger:** User-configurable
**Content:** HTML formatted daily report
**Provider:** SMTP (configurable)

#### FR-DL-004: S3 Archive
**Priority:** High
**Structure:**
```
reports/
├── daily/YYYY-MM-DD/
│   ├── summary.json
│   ├── summary.md
│   └── summary.html
├── weekly/YYYY-WW/
└── alerts/YYYY-MM-DD/
```
**Retention:** 730 days (2 years)

### 4.8 API Endpoints (FR-API)

#### FR-API-001: POST /api/analyze/technical
**Input:**
```json
{
  "symbol": "VNM",
  "days": 90
}
```
**Output:** Technical analysis result with indicators, signals, recommendation

#### FR-API-002: POST /api/analyze/sentiment
**Input:**
```json
{
  "texts": ["Article 1...", "Article 2..."],
  "symbol": "VNM"
}
```
**Output:** Sentiment scores, aggregate metrics

#### FR-API-003: POST /api/synthesize
**Input:**
```json
{
  "technical": {...},
  "sentiment": {...},
  "market_context": {...}
}
```
**Output:** Combined forecast, recommendation, risk level

#### FR-API-004: GET /api/reports/daily
**Query:** `?date=YYYY-MM-DD`
**Output:** Daily report JSON

#### FR-API-005: GET /api/stocks/{symbol}
**Output:** Latest analysis for specified symbol

#### FR-API-006: GET /health
**Output:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "services": {
    "database": "connected",
    "redis": "connected",
    "s3": "accessible"
  }
}
```

---

## 5. Non-Functional Requirements

### 5.1 Performance Requirements

| ID | Requirement | Target | Measurement |
|----|-------------|--------|-------------|
| NFR-P-001 | RSS scraping throughput | 100 articles < 30s | End-to-end time |
| NFR-P-002 | Technical analysis latency | 1 stock < 2s | API response time |
| NFR-P-003 | Sentiment analysis throughput | 100 texts < 45s | Batch processing |
| NFR-P-004 | Daily workflow completion | < 5 minutes for 10 stocks | n8n execution time |
| NFR-P-005 | API response time (cached) | p95 < 500ms | Monitoring |
| NFR-P-006 | API response time (real-time) | p95 < 3s | Monitoring |
| NFR-P-007 | Concurrent API users | 100 simultaneous | Load testing |

### 5.2 Security Requirements

| ID | Requirement | Implementation |
|----|-------------|----------------|
| NFR-SEC-001 | API Authentication | JWT tokens, 24h expiry |
| NFR-SEC-002 | Rate Limiting | Anonymous: 100/hr, Auth: 1000/hr, Premium: 10000/hr |
| NFR-SEC-003 | Credential Storage | Environment variables only |
| NFR-SEC-004 | Data in Transit | HTTPS/TLS 1.2+ |
| NFR-SEC-005 | Data at Rest | S3 server-side encryption |
| NFR-SEC-006 | Input Validation | Parameterized queries, XSS prevention |
| NFR-SEC-007 | Network Isolation | Docker bridge network for internal services |

### 5.3 Reliability Requirements

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-R-001 | System Availability | 99.5% uptime |
| NFR-R-002 | Error Recovery | Auto-retry with exponential backoff (max 5x, 300s) |
| NFR-R-003 | Data Backup | Daily at 2:00 AM, 90-day retention |
| NFR-R-004 | Graceful Degradation | Continue with partial data on source failure |
| NFR-R-005 | Health Monitoring | Prometheus metrics, Grafana dashboards |

### 5.4 Scalability Requirements

| ID | Requirement | Approach |
|----|-------------|----------|
| NFR-SC-001 | Horizontal Scaling | Celery workers: 1-10 configurable |
| NFR-SC-002 | API Scaling | Load-balanced behind Nginx |
| NFR-SC-003 | Data Volume | 1 year history per stock (~180MB/year) |
| NFR-SC-004 | Queue Capacity | Redis with 10000 task limit |

### 5.5 Maintainability Requirements

| ID | Requirement | Standard |
|----|-------------|----------|
| NFR-M-001 | Code Documentation | Docstrings for public functions |
| NFR-M-002 | Logging Format | JSON structured (timestamp, service, message) |
| NFR-M-003 | Configuration | All settings via environment variables |
| NFR-M-004 | Containerization | Docker images for all services |
| NFR-M-005 | Version Control | Git with conventional commits |

---

## 6. Data Requirements

### 6.1 Database Schema (PostgreSQL)

```sql
-- Stocks table: Master list of tradable securities
CREATE TABLE stocks (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(255),
    exchange VARCHAR(10) CHECK (exchange IN ('HSX', 'HNX', 'UPCOM')),
    sector VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Price history: Daily OHLCV data
CREATE TABLE price_history (
    id SERIAL PRIMARY KEY,
    stock_id INTEGER REFERENCES stocks(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    open DECIMAL(15,2),
    high DECIMAL(15,2),
    low DECIMAL(15,2),
    close DECIMAL(15,2),
    volume BIGINT,
    UNIQUE(stock_id, date)
);

-- Analysis results: Stored analysis outputs
CREATE TABLE analysis_results (
    id SERIAL PRIMARY KEY,
    stock_id INTEGER REFERENCES stocks(id) ON DELETE CASCADE,
    analysis_date DATE NOT NULL,
    recommendation VARCHAR(20) CHECK (recommendation IN
        ('STRONG BUY', 'BUY', 'HOLD', 'SELL', 'STRONG SELL')),
    confidence DECIMAL(5,4),
    technical_score DECIMAL(5,2),
    sentiment_score DECIMAL(5,4),
    risk_level VARCHAR(10) CHECK (risk_level IN ('LOW', 'MEDIUM', 'HIGH')),
    raw_data JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(stock_id, analysis_date)
);

-- News articles: Scraped news with sentiment
CREATE TABLE news_articles (
    id SERIAL PRIMARY KEY,
    source VARCHAR(50) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    url TEXT UNIQUE,
    published_at TIMESTAMP,
    sentiment VARCHAR(20),
    sentiment_score DECIMAL(5,4),
    mentioned_symbols TEXT[],
    is_spam BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Alerts: Generated system alerts
CREATE TABLE alerts (
    id SERIAL PRIMARY KEY,
    stock_id INTEGER REFERENCES stocks(id) ON DELETE CASCADE,
    alert_type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    severity VARCHAR(20) CHECK (severity IN ('INFO', 'WARNING', 'CRITICAL')),
    is_sent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    sent_at TIMESTAMP
);

-- Users: Telegram and web users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE,
    email VARCHAR(255) UNIQUE,
    username VARCHAR(100),
    watchlist TEXT[] DEFAULT '{}',
    alert_preferences JSONB DEFAULT '{}',
    tier VARCHAR(20) DEFAULT 'free' CHECK (tier IN ('free', 'basic', 'premium')),
    created_at TIMESTAMP DEFAULT NOW(),
    last_active TIMESTAMP
);

-- User subscriptions: Stock alert subscriptions
CREATE TABLE user_subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    stock_id INTEGER REFERENCES stocks(id) ON DELETE CASCADE,
    alert_types TEXT[] DEFAULT '{"HIGH_RISK", "STRONG_SIGNAL"}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, stock_id)
);

-- Indexes for performance
CREATE INDEX idx_price_history_stock_date ON price_history(stock_id, date DESC);
CREATE INDEX idx_analysis_results_stock_date ON analysis_results(stock_id, analysis_date DESC);
CREATE INDEX idx_news_articles_published ON news_articles(published_at DESC);
CREATE INDEX idx_alerts_stock_created ON alerts(stock_id, created_at DESC);
```

### 6.2 S3 Bucket Structure

```
vnstock-data/
├── raw-data/
│   ├── news/
│   │   └── YYYY-MM-DD/
│   │       ├── vneconomy.json
│   │       ├── cafef.json
│   │       └── vietstock.json
│   ├── market-data/
│   │   └── YYYY-MM-DD/
│   │       ├── VNM.json
│   │       ├── HPG.json
│   │       └── ...
│   └── social/
│       └── YYYY-MM-DD/
│           ├── facebook.json
│           └── telegram.json
├── processed/
│   ├── sentiment/
│   │   └── YYYY-MM-DD/
│   │       └── batch_results.json
│   ├── technical/
│   │   └── YYYY-MM-DD/
│   │       ├── VNM.json
│   │       └── ...
│   └── combined/
│       └── YYYY-MM-DD/
│           └── synthesis.json
├── reports/
│   ├── daily/
│   │   └── YYYY-MM-DD/
│   │       ├── summary.json
│   │       ├── summary.md
│   │       └── summary.html
│   ├── weekly/
│   │   └── YYYY-WW/
│   └── alerts/
│       └── YYYY-MM-DD/
└── backups/
    └── database/
        └── YYYY-MM-DD/
```

### 6.3 Data Retention Policies

| Data Type | Retention Period | Action After Expiry |
|-----------|-----------------|---------------------|
| Raw news data | 90 days | Delete |
| Raw market data | 365 days | Archive to Glacier |
| Processed data | 365 days | Delete |
| Reports | 730 days (2 years) | Archive |
| Database backups | 90 days | Delete |
| User data | Indefinite | Delete on request |

---

## 7. System Constraints

### 7.1 Vietnamese Market Specifics

**Trading Hours:**
| Session | Time (Vietnam, UTC+7) |
|---------|----------------------|
| Morning | 9:00 AM - 11:30 AM |
| Lunch Break | 11:30 AM - 1:00 PM |
| Afternoon | 1:00 PM - 3:00 PM |
| ATC (At-the-close) | 2:30 PM - 2:45 PM |

**Trading Days:** Monday - Friday (excluding Vietnamese holidays)

**Vietnamese Holidays (System Downtime Expected):**
- Tet Nguyen Dan (Lunar New Year): ~7 days
- Hung Kings' Temple Festival: 1 day
- Reunification Day (April 30)
- Labor Day (May 1)
- National Day (September 2)

**Symbol Format:**
- 3 uppercase Latin letters (e.g., VNM, HPG, FPT)
- No numbers or special characters

**Price Units:**
- Currency: Vietnamese Dong (VND)
- Tick sizes vary by price level:
  - < 10,000 VND: 10 VND
  - 10,000 - 49,950 VND: 50 VND
  - >= 50,000 VND: 100 VND

**Circuit Breakers:**
- Daily price limit: +/- 7% for most HSX stocks
- Trading halt triggers on extreme volatility

**Foreign Ownership:**
- Maximum 49% for most stocks
- Some sectors restricted to 0-30%

### 7.2 Regulatory Constraints

1. **No Financial Advice License:** System must clearly state it does not provide licensed financial advice

2. **Disclaimer Requirement:** Per Circular 96/2020/TT-BTC, all outputs must include investment risk warnings

3. **Data Protection:** Comply with Vietnam Cybersecurity Law for user data

4. **No Market Manipulation:** System must not spread false information or coordinate trading

5. **Content Responsibility:** Aggregated news must cite original sources

### 7.3 Technical Constraints

1. **vnstock Library Dependency:** Data availability depends on third-party library maintenance

2. **PhoBERT Memory:** Requires 4GB+ RAM for model loading; GPU optional but recommended

3. **Data Delay:** Free data sources have 15-minute delay; real-time requires paid APIs

4. **Rate Limits:**
   - vnstock: ~100 requests/minute (unofficial)
   - RSS feeds: Respect Robots.txt
   - Telegram: 30 messages/second
   - Facebook scraping: Use official APIs or respect ToS

5. **Language Processing:** PhoBERT trained on general Vietnamese; may need fine-tuning for financial domain

---

## 8. Appendices

### Appendix A: Vietnamese Stock Market Slang Dictionary

| Vietnamese | English Translation | Usage Context |
|------------|---------------------|---------------|
| Cây thông | Bullish candlestick pattern | Technical analysis |
| Cây súng | Bearish candlestick pattern | Technical analysis |
| Múa bên trăng | Price manipulation | Warning indicator |
| Lùa gà | Pump and dump scheme | Fraud detection |
| Con tép | Small retail investor | User type |
| Cá mập | Large institutional investor | User type |
| FOMO | Fear of missing out | Sentiment |
| Chốt lời | Take profit | Action |
| Cắt lỗ | Stop loss / Cut loss | Action |
| All in | Invest entire capital | Risk indicator |
| Sideway | Sideways price movement | Trend |
| Breakout | Price breakout | Signal |
| Hốt | Buy opportunity | Bullish |
| Ôm | Hold long-term | Strategy |
| Bơm | Pumping prices | Manipulation |
| Xả | Dumping shares | Manipulation |
| Tạo đáy | Creating bottom | Pattern |
| Tạo đỉnh | Creating top | Pattern |
| Bắt đáy | Bottom fishing | Strategy |
| Bắt dao rơi | Catching falling knife | Risk warning |
| T+ | Settlement period (T+2) | Trading |
| Margin call | Margin call | Risk |
| Thanh khoản | Liquidity | Volume |
| Vốn hóa | Market capitalization | Metric |
| Room ngoại | Foreign ownership room | Constraint |

### Appendix B: API Response Schemas

**Technical Analysis Response:**
```json
{
  "symbol": "VNM",
  "timestamp": "2026-01-28T08:00:00+07:00",
  "current_price": 75000,
  "recommendation": "BUY",
  "confidence": 0.72,
  "technical_score": 0.45,
  "indicators": {
    "rsi": 35.2,
    "macd": 0.52,
    "macd_signal": 0.38,
    "sma_20": 73500,
    "sma_50": 71200,
    "ema_12": 74200,
    "ema_26": 72800,
    "adx": 28.5,
    "stochastic_k": 25.3,
    "stochastic_d": 28.1,
    "atr": 1850,
    "volume_ratio": 1.35
  },
  "signals": [
    "RSI indicates oversold",
    "MACD bullish crossover",
    "Price above SMA20"
  ],
  "support_resistance": {
    "resistance": [78000, 80500, 85000],
    "support": [72000, 70000, 68500]
  },
  "price_targets": {
    "short_term": 78500,
    "stop_loss": 71000,
    "fibonacci": {
      "0.236": 73800,
      "0.382": 75200,
      "0.500": 76500,
      "0.618": 77800
    }
  }
}
```

**Sentiment Analysis Response:**
```json
{
  "symbol": "VNM",
  "analysis_date": "2026-01-28",
  "articles_analyzed": 25,
  "aggregate": {
    "positive_ratio": 0.48,
    "negative_ratio": 0.20,
    "neutral_ratio": 0.32,
    "overall_score": 0.28
  },
  "top_positive": [
    {"title": "...", "score": 0.92}
  ],
  "top_negative": [
    {"title": "...", "score": 0.85}
  ],
  "rumors": {
    "detected": false,
    "symbols": []
  }
}
```

### Appendix C: Requirements Traceability Matrix

| Requirement ID | Source | Test Case | Priority |
|----------------|--------|-----------|----------|
| FR-DC-001 | Architecture 3.1 | TC-DC-001 | High |
| FR-DC-002 | Architecture 3.2 | TC-DC-002 | High |
| FR-TA-001 | Architecture 4.1 | TC-TA-001 | High |
| FR-TA-009 | Architecture 4.9 | TC-TA-009 | High |
| FR-SA-002 | Architecture 5.2 | TC-SA-002 | High |
| FR-FC-001 | Architecture 6.1 | TC-FC-001 | High |
| FR-MA-003 | Architecture 7.3 | TC-MA-003 | High |
| NFR-P-001 | Performance Req | TC-PERF-001 | High |
| NFR-SEC-001 | Security Req | TC-SEC-001 | Critical |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | January 2026 | System | Initial SRS document |

---

**End of Document**
