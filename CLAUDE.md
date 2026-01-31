# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Vietnamese Stock Market Analysis System - an AI-powered automated stock analysis platform for HSX/HNX/UPCOM exchanges. Combines technical analysis, Vietnamese NLP sentiment analysis (PhoBERT), and real-time Telegram alerts.

**Stack:** Python 3.9+ (FastAPI backend), Next.js (dashboard), n8n (workflow orchestration), PostgreSQL, Redis, Docker Compose

## Common Commands

```bash
# Start all services
docker-compose up -d

# Start with specific profiles
docker-compose --profile monitoring up -d    # With Prometheus/Grafana
docker-compose --profile development up -d   # With MinIO
docker-compose --profile production up -d    # With Nginx

# Initial setup
./quickstart.sh

# Run technical agent standalone
python technical_agent.py

# View logs
docker-compose logs -f agent-service
docker-compose logs -f n8n
```

## Architecture

### Multi-Agent System

Data flows through a pipeline of specialized agents:

1. **Technical Analysis Agent** (`technical_agent.py`) - Calculates RSI, MACD, Bollinger Bands, SMA/EMA, Stochastic, ADX, ATR, VWAP; generates buy/sell/hold signals with confidence scores
2. **Sentiment Analysis Agent** - Vietnamese NLP using PhoBERT model with custom stock market slang dictionary
3. **Forecast Agent** - Synthesizes technical (40%) + sentiment (30%) + market context (30%) for final recommendations
4. **Master Orchestrator** - Coordinates sub-agents, generates daily reports, manages alerts

### Services (docker-compose.yml)

| Service | Port | Purpose |
|---------|------|---------|
| n8n | 5678 | Workflow automation |
| agent-service | 8000 | FastAPI AI backend |
| postgres | 5432 | Database |
| redis | 6379 | Cache/message broker |
| web-dashboard | 3000 | Next.js frontend |
| telegram-bot | - | Telegram integration |
| celery-worker/beat | - | Async task processing |

### Data Flow

```
RSS/APIs/Social → n8n Workflow → S3 (raw/processed/reports) → AI Agents → Telegram/Web/Email
```

### S3 Bucket Structure

```
vnstock-data/
├── raw-data/news/{date}/, market-data/{date}/, social/{date}/
├── processed/sentiment/, technical/, combined/
└── reports/daily/, weekly/, alerts/
```

### Daily Workflow (n8n-workflow-daily-analysis.json)

Triggers at 8:00 AM weekdays:
1. RSS scraping from VnEconomy, CafeF, VietStock
2. Spam filtering with keyword blacklist
3. Hot stock detection (most-mentioned symbols)
4. Technical/sentiment analysis via agent-service
5. Report generation and distribution

## Key APIs

```
POST /api/analyze/technical    - Technical analysis
POST /api/analyze/sentiment    - Sentiment analysis
POST /api/synthesize           - Master agent synthesis
GET  /api/reports/daily        - Daily report
GET  /health                   - Health check
```

## Vietnamese Market Specifics

- **Market hours:** 9:00 AM - 3:00 PM Vietnam time
- **Data sources:** VnEconomy, CafeF, VietStock RSS feeds
- **NLP:** PhoBERT model with Vietnamese stock slang dictionary (e.g., "lùa gà" = pump and dump)
- **Library:** `vnstock` for market data fetching

## Configuration

See `.env.example` for all environment variables including:
- AWS S3 credentials and bucket config
- Anthropic Claude / OpenAI API keys
- Telegram bot token and channel IDs
- FiinGroup/VietStock API credentials
- Feature flags (`ENABLE_RUMOR_DETECTION`, `ENABLE_ML_FORECAST`, etc.)

---

## Coding Standards

### Go Code Standards

**File Organization:**
- One package per directory
- Keep main.go files under 200 lines (extract logic to internal packages)
- Group imports: stdlib, external, internal (separated by blank lines)

**Naming Conventions:**
```go
// Interfaces: end with -er suffix when possible
type Analyzer interface { ... }

// Structs: PascalCase
type TechnicalAgent struct { ... }

// Functions/Methods: camelCase (exported) or camelCase (unexported)
func CalculateRSI(...) float64 { ... }
func parseSymbol(...) string { ... }

// Constants: PascalCase or SCREAMING_SNAKE_CASE for exported constants
const DefaultTimeout = 30 * time.Second
const MAX_RETRIES = 3
```

**Error Handling:**
```go
// Always check errors immediately
data, err := fetchData(symbol)
if err != nil {
    return nil, fmt.Errorf("failed to fetch data for %s: %w", symbol, err)
}

// Use %w for error wrapping to preserve error chains
// Custom errors for business logic
var ErrInvalidSymbol = errors.New("invalid stock symbol")

// Sentinel errors should be exported
```

**Context Usage:**
```go
// Always pass context.Context as first parameter
func AnalyzeStock(ctx context.Context, symbol string) (*Analysis, error) {
    // Check context cancellation in long-running operations
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // continue processing
    }
}

// Set reasonable timeouts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

**Struct Tags:**
```go
type StockAnalysis struct {
    Symbol     string    `json:"symbol" db:"symbol" validate:"required"`
    Price      float64   `json:"price" db:"current_price"`
    Timestamp  time.Time `json:"timestamp" db:"created_at"`
}
```

**Testing:**
```go
// Test files: *_test.go
// Test functions: TestXxx
// Benchmark functions: BenchmarkXxx
// Table-driven tests preferred

func TestCalculateRSI(t *testing.T) {
    tests := []struct {
        name     string
        prices   []float64
        period   int
        expected float64
        wantErr  bool
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

### Python Code Standards

**Style Guide:**
- Follow PEP 8 strictly
- Use Black formatter (line length: 100)
- Use isort for import sorting
- Type hints required for all public functions

**Import Order:**
```python
# Standard library
import os
from datetime import datetime
from typing import Dict, List, Optional

# Third-party
import pandas as pd
import numpy as np
from fastapi import FastAPI

# Local application
from app.models.phobert import PhoBERTModel
from app.services.sentiment_analyzer import SentimentAnalyzer
```

**Type Hints:**
```python
def analyze_sentiment(
    text: str,
    language: str = "vi",
    confidence_threshold: float = 0.7
) -> Dict[str, float]:
    """
    Analyze sentiment of Vietnamese text.

    Args:
        text: Input text to analyze
        language: Language code (default: "vi")
        confidence_threshold: Minimum confidence score

    Returns:
        Dictionary with sentiment scores

    Raises:
        ValueError: If text is empty
    """
    pass
```

**Error Handling:**
```python
# Custom exceptions
class AnalysisError(Exception):
    """Base exception for analysis errors"""
    pass

class InvalidSymbolError(AnalysisError):
    """Raised when stock symbol is invalid"""
    pass

# Use logging instead of print
import logging
logger = logging.getLogger(__name__)

try:
    result = process_data(data)
except ValueError as e:
    logger.error(f"Failed to process data: {e}")
    raise AnalysisError(f"Data processing failed: {e}") from e
```

**Async/Await:**
```python
# Use async for I/O operations
async def fetch_stock_data(symbol: str) -> pd.DataFrame:
    async with httpx.AsyncClient() as client:
        response = await client.get(f"/api/stocks/{symbol}")
        return parse_response(response)
```

### TypeScript/Next.js Standards

**Component Structure:**
```typescript
// Use functional components with TypeScript
interface StockCardProps {
  symbol: string;
  price: number;
  change: number;
}

export const StockCard: React.FC<StockCardProps> = ({ symbol, price, change }) => {
  // Component logic
  return (
    <div className="stock-card">
      {/* JSX */}
    </div>
  );
};
```

**API Types:**
```typescript
// Define API response types
export interface TechnicalAnalysis {
  symbol: string;
  recommendation: 'BUY' | 'SELL' | 'HOLD';
  confidence: number;
  indicators: {
    rsi: number;
    macd: number;
    // ...
  };
}

// Use consistent API client
const analyzeStock = async (symbol: string): Promise<TechnicalAnalysis> => {
  const response = await fetch(`/api/analyze/technical`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ symbol })
  });
  return response.json();
};
```

### SQL Standards

```sql
-- Use snake_case for tables and columns
-- Add created_at, updated_at to all tables
-- Use BIGSERIAL for IDs

CREATE TABLE stock_analysis (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    analysis_date DATE NOT NULL,
    recommendation VARCHAR(20) NOT NULL CHECK (recommendation IN ('BUY', 'SELL', 'HOLD')),
    confidence DECIMAL(3,2) CHECK (confidence BETWEEN 0 AND 1),
    raw_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(symbol, analysis_date)
);

-- Always add indexes for foreign keys and frequently queried columns
CREATE INDEX idx_stock_analysis_symbol_date ON stock_analysis(symbol, analysis_date DESC);
CREATE INDEX idx_stock_analysis_created_at ON stock_analysis(created_at DESC);
```

---

## Architectural Decisions

### Why Hybrid Go/Python Architecture?

**Go Services (Performance-Critical):**
- **API Gateway**: High-throughput request routing (handles 10k+ req/s)
- **Technical Analysis Agent**: CPU-intensive indicator calculations
- **Forecast Agent**: Real-time aggregation and synthesis
- **Master Orchestrator**: Concurrent workflow coordination

**Python Services (ML/NLP):**
- **Sentiment Analysis**: PhoBERT model (PyTorch ecosystem)
- **Advanced ML**: LSTM, Transformer models (TensorFlow/PyTorch)

**Rationale:**
- Go: 10-20x faster for technical indicators, better concurrency
- Python: Superior ML/NLP library ecosystem, faster prototyping
- Trade-off: Network overhead vs. development velocity

### Service Communication Patterns

**Synchronous HTTP (Primary):**
```
API Gateway → Technical Agent: POST /analyze
Forecast Agent → Sentiment Service: POST /analyze/sentiment
```

**Use Cases:**
- Request-response workflows
- Real-time user queries
- Health checks

**Asynchronous Pub/Sub (Future):**
```
n8n → Pub/Sub Topic "hot-stocks" → Master Orchestrator
Master Orchestrator → Topic "alerts" → Telegram Bot
```

**Use Cases:**
- Batch processing (100+ stocks)
- Event-driven alerts
- Decoupling services

**Decision:** Start with HTTP for simplicity, migrate high-volume workflows to Pub/Sub in Phase 4.

### Caching Strategy (Redis)

**Cache Layers:**
1. **Market Data** (TTL: 60s): Real-time prices, volume
2. **Technical Indicators** (TTL: 5min): RSI, MACD calculations
3. **Sentiment Analysis** (TTL: 10min): News sentiment scores
4. **Daily Reports** (TTL: 24h): Generated reports

**Cache Keys Pattern:**
```
stock:{symbol}:price           → "45000"
stock:{symbol}:technical:{date} → JSON
news:{article_id}:sentiment    → JSON
report:daily:{date}            → JSON
```

**Invalidation Strategy:**
- Time-based (TTL) for volatile data
- Event-based for static data (new analysis triggers cache clear)

### Database Schema Design

**Principles:**
- Separate tables for time-series data (analysis_results) vs. reference data (stocks)
- JSONB for flexible data (indicators, raw responses)
- Partitioning by date for large tables (analysis_results, news_articles)
- Materialized views for complex aggregations

**Example:**
```sql
-- Partitioned by month for scalability
CREATE TABLE analysis_results (
    id BIGSERIAL,
    symbol VARCHAR(10),
    analysis_date DATE,
    -- ... other columns
    PRIMARY KEY (id, analysis_date)
) PARTITION BY RANGE (analysis_date);

-- Create partitions
CREATE TABLE analysis_results_2024_01 PARTITION OF analysis_results
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

---

## Project-Specific Requirements

### Vietnamese Market Specifics

**Market Hours:**
```python
MARKET_OPEN = datetime.time(9, 0)   # 9:00 AM Vietnam time (UTC+7)
MARKET_CLOSE = datetime.time(15, 0)  # 3:00 PM Vietnam time

# Pre-market: 8:45 - 9:00
# Lunch break: None (continuous trading)
# After-market: 15:00 - 15:30 (matching only)
```

**Trading Symbols:**
- HSX (HCM Stock Exchange): 3-letter codes (e.g., VNM, HPG, VCB)
- HNX (Hanoi Stock Exchange): 3-letter codes (e.g., VCS, PVS)
- UPCOM (Unlisted Public Company Market): 3-letter codes

**Price Limits:**
- Floor/Ceiling: ±7% from reference price (HSX/HNX)
- Reference price: Previous day's closing price

**Vietnamese Text Processing:**
```python
# Stock slang dictionary
VIETNAMESE_STOCK_SLANG = {
    "lùa gà": "pump and dump",
    "cá mập": "big investors/sharks",
    "cá nhỏ": "retail investors",
    "trụ": "blue chip stocks",
    "penny": "penny stocks",
    "room ngoại": "foreign ownership limit",
    "dư mua/dư bán": "buy/sell surplus",
    "khớp lệnh": "order matching"
}
```

### Technical Indicator Configurations

**Default Parameters:**
```go
const (
    RSI_PERIOD          = 14
    MACD_FAST_PERIOD    = 12
    MACD_SLOW_PERIOD    = 26
    MACD_SIGNAL_PERIOD  = 9
    BOLLINGER_PERIOD    = 20
    BOLLINGER_STD_DEV   = 2.0
    SMA_SHORT_PERIOD    = 20
    SMA_MEDIUM_PERIOD   = 50
    SMA_LONG_PERIOD     = 200
    STOCHASTIC_K_PERIOD = 14
    STOCHASTIC_D_PERIOD = 3
    ADX_PERIOD          = 14
    ATR_PERIOD          = 14
)
```

**Signal Thresholds:**
```go
// RSI
const (
    RSI_OVERSOLD  = 30.0  // Buy signal
    RSI_OVERBOUGHT = 70.0  // Sell signal
)

// MACD
// Signal: MACD line crosses signal line

// Bollinger Bands
// Signal: Price touches upper/lower band
```

### Sentiment Analysis Model

**Model Selection:**
- **Primary**: PhoBERT (vinai/phobert-base) - 135M parameters
- **Fallback**: ViSoBERT (alternative Vietnamese BERT)
- **Fine-tuning**: Train on 10k+ Vietnamese stock news articles

**Sentiment Categories:**
```python
SENTIMENT_LABELS = {
    "POSITIVE": 1,      # Bullish news
    "NEUTRAL": 0,       # Factual/informative
    "NEGATIVE": -1,     # Bearish news
}

# Multi-label classification
TOPICS = [
    "earnings",         # Kết quả kinh doanh
    "dividends",        # Trả cổ tức
    "expansion",        # Mở rộng đầu tư
    "management",       # Thay đổi nhân sự
    "legal",            # Pháp lý
    "market_rumor",     # Tin đồn thị trường
]
```

**Confidence Thresholds:**
```python
MIN_CONFIDENCE_ANALYSIS = 0.7  # Minimum to include in report
MIN_CONFIDENCE_ALERT = 0.85    # Minimum to trigger alert
```

### API Rate Limiting

**Rate Limits by Service:**
```yaml
vnstock_api:
  rate_limit: 100 requests/hour
  burst: 10

fiingroup_api:
  rate_limit: 1000 requests/day
  burst: 50

vietstock_api:
  rate_limit: 500 requests/hour
  burst: 20

# Internal API (api-gateway)
public_endpoints:
  rate_limit: 100 requests/hour per IP
  burst: 20

authenticated_endpoints:
  rate_limit: 1000 requests/hour per user
  burst: 50
```

**Implementation:**
```go
// Use token bucket algorithm with Redis
rateLimiter := redis_rate.NewLimiter(redisClient)

// Per-IP limiting
rate := redis_rate.PerHour(100).Burst(20)
res, err := rateLimiter.Allow(ctx, "ip:"+clientIP, rate)
if err != nil || !res.Allowed {
    return ErrRateLimitExceeded
}
```

### Security Requirements

**API Authentication:**
- **Public endpoints**: No auth (rate-limited by IP)
- **Internal services**: JWT tokens (service-to-service)
- **Admin endpoints**: API key + JWT

**Data Protection:**
```go
// Sensitive data encryption
// Use AES-256-GCM for user data
// Store encryption keys in environment variables or secret manager

// SQL injection prevention
// Always use parameterized queries
db.Query("SELECT * FROM stocks WHERE symbol = $1", symbol)  // ✓ Good
db.Query("SELECT * FROM stocks WHERE symbol = '" + symbol + "'")  // ✗ Bad
```

**Telegram Bot Security:**
```python
# Verify webhook requests
def verify_telegram_signature(request_body: bytes, signature: str) -> bool:
    expected = hmac.new(
        key=TELEGRAM_BOT_TOKEN.encode(),
        msg=request_body,
        digestmod=hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(expected, signature)

# Rate limit bot commands per user
MAX_COMMANDS_PER_MINUTE = 10
```

**Environment Variables:**
```bash
# Never commit these to git
.env
.env.local
.env.production

# Use .env.example as template
# Secrets in GCP Secret Manager for production
```

---

## Development Workflow

### Branch Naming Conventions

```
main                    # Production-ready code
develop                 # Integration branch

# Feature branches
feature/forecast-agent
feature/telegram-bot
feature/web-dashboard

# Bug fixes
bugfix/rsi-calculation-error
bugfix/sentiment-timeout

# Hot fixes (production)
hotfix/api-gateway-crash

# Release branches
release/v1.0.0
release/v1.1.0
```

### Git Commit Messages

**Format (Conventional Commits):**
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, no logic change)
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `test`: Adding tests
- `chore`: Build process, dependencies

**Examples:**
```
feat(forecast-agent): add LSTM prediction model

Implement LSTM neural network for stock price forecasting.
Combines technical indicators with historical price patterns.

Closes #42

---

fix(sentiment): handle Vietnamese special characters

PhoBERT tokenizer was failing on text with Unicode characters.
Added proper UTF-8 encoding before tokenization.

Fixes #38

---

perf(technical-agent): optimize RSI calculation

Reduced RSI calculation time by 40% using sliding window.
Benchmarked with 1000 stocks over 90 days.

Before: 2.3s, After: 1.4s
```

### Pull Request Process

**PR Title Format:**
```
[TYPE] Brief description (#issue-number)

Example:
[FEAT] Implement Telegram bot with stock alerts (#45)
[FIX] Resolve sentiment analysis timeout (#52)
```

**PR Description Template:**
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] New feature
- [ ] Bug fix
- [ ] Breaking change
- [ ] Documentation update

## Related Issues
Closes #45

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings generated
```

**Review Requirements:**
- Minimum 1 approval required
- All CI checks must pass
- Code coverage must not decrease
- No merge conflicts with develop

### Testing Requirements

**Before Creating PR:**
```bash
# Go services
cd go-services
go test ./... -v -cover
go vet ./...
golangci-lint run

# Python services
cd python-sentiment
pytest tests/ -v --cov=app --cov-report=html
black --check app/
isort --check app/
mypy app/

# Integration tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

**Test Coverage Requirements:**
- Unit tests: 80%+ coverage
- Integration tests: Critical paths covered
- E2E tests: Happy path + error scenarios

**Test Naming:**
```go
// Go: TestFunctionName_Scenario_ExpectedResult
func TestCalculateRSI_ValidData_ReturnsCorrectValue(t *testing.T) { }
func TestCalculateRSI_EmptyData_ReturnsError(t *testing.T) { }
```

```python
# Python: test_function_name_scenario_expected_result
def test_analyze_sentiment_valid_text_returns_scores(): ...
def test_analyze_sentiment_empty_text_raises_error(): ...
```

### Deployment Procedures

**Development Environment:**
```bash
# Start services locally
docker-compose up -d

# Verify all services healthy
./scripts/check_health.sh

# Run smoke tests
pytest tests/smoke/
```

**Staging Environment:**
```bash
# Deploy to GCP staging
gcloud builds submit --config cloudbuild.yaml --substitutions=_ENV=staging

# Run integration tests against staging
pytest tests/integration/ --env=staging

# Manual QA verification
```

**Production Deployment:**
```bash
# Tag release
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# Deploy to production (requires approval)
gcloud builds submit --config cloudbuild.yaml --substitutions=_ENV=production

# Monitor for errors
kubectl logs -f deployment/api-gateway -n production
# Check Grafana dashboards

# Rollback if needed
kubectl rollout undo deployment/api-gateway -n production
```

**Deployment Checklist:**
- [ ] All tests passing
- [ ] PR approved and merged
- [ ] Release notes updated
- [ ] Database migrations applied (if any)
- [ ] Environment variables updated
- [ ] Monitoring dashboards configured
- [ ] Rollback plan documented
- [ ] Stakeholders notified

---

## Monitoring & Debugging

### Logging Standards

**Log Levels:**
```go
// Go (using zerolog or logrus)
logger.Debug().Str("symbol", symbol).Msg("Starting analysis")
logger.Info().Str("symbol", symbol).Float64("rsi", rsi).Msg("RSI calculated")
logger.Warn().Str("symbol", symbol).Msg("Cache miss, fetching from API")
logger.Error().Err(err).Str("symbol", symbol).Msg("Failed to fetch data")
logger.Fatal().Err(err).Msg("Database connection failed")
```

**Structured Logging:**
```json
{
  "timestamp": "2024-01-29T10:30:00Z",
  "level": "info",
  "service": "technical-agent",
  "symbol": "VNM",
  "operation": "calculate_rsi",
  "duration_ms": 234,
  "message": "RSI calculation completed"
}
```

### Metrics to Track

**Prometheus Metrics:**
```go
// Request duration histogram
http_request_duration_seconds{service="api-gateway", endpoint="/analyze", method="POST"}

// Request count
http_requests_total{service="api-gateway", status="200"}

// Error rate
http_requests_errors_total{service="forecast-agent", error_type="timeout"}

// Business metrics
stock_analysis_completed_total{symbol="VNM", recommendation="BUY"}
cache_hit_rate{service="technical-agent", cache_type="redis"}
```

**Grafana Dashboards:**
1. System Overview: CPU, Memory, Network, Request rate
2. Service Health: Error rates, latencies, success rate
3. Business Metrics: Analyses per day, alert distribution
4. Cost Monitoring: API call counts, cloud resource usage

---

## Common Pitfalls to Avoid

1. **Don't hardcode Vietnamese timezone** - Use `TZ=Asia/Ho_Chi_Minh` env var
2. **Don't trust external API data** - Always validate and sanitize
3. **Don't block on I/O in Go** - Use goroutines for concurrent operations
4. **Don't cache API credentials** - Fetch fresh from Secret Manager
5. **Don't ignore context cancellation** - Check `ctx.Done()` in loops
6. **Don't use print() in Python** - Use logging module
7. **Don't commit .env files** - Use .env.example templates only
8. **Don't skip error handling** - Even in "impossible" scenarios
9. **Don't use SELECT * in queries** - Explicitly list columns
10. **Don't forget to close resources** - Use defer in Go, context managers in Python
