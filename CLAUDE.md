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
