# Codebase Summary

**Last Updated**: 2026-04-20
**Version**: 1.0.0
**Project**: Vietnamese Stock Market Analysis System

## Overview

VNStock Analysis System là hệ thống phân tích chứng khoán Việt Nam tự động sử dụng AI Agents, n8n workflows, và real-time data processing cho thị trường HSX/HNX/UPCOM.

## Project Structure

```
vnstock-analysis/
├── blog-site/                # Next.js 14 Frontend (Blog + Market Data)
│   ├── app/                  # App Router pages
│   │   ├── api/             # API routes
│   │   ├── articles/        # Articles listing & detail
│   │   ├── market/          # Market data page
│   │   ├── screener/        # Stock screener
│   │   └── symbols/          # Symbol detail pages
│   ├── components/           # UI components (12 components)
│   └── lib/                  # API clients & utilities
├── go-services/              # Go Backend Services
│   ├── cmd/                  # Command entrypoints
│   └── internal/              # Internal packages
│       ├── handlers/          # HTTP handlers
│       ├── indicators/         # Technical indicators (RSI, MACD, etc.)
│       ├── services/           # Business logic
│       └── models/            # Data models
├── python-sentiment/          # Python ML/NLP Service
│   ├── app/                  # FastAPI application
│   ├── models/               # ML models (PhoBERT)
│   └── tests/                # Unit tests
├── crawl-agent/              # Python Data Crawling Agent
├── telegram-bot/             # Telegram Bot Service
├── web-dashboard/            # Legacy Next.js Dashboard
├── n8n-workflow-*.json       # n8n Workflow Definitions
├── docker-compose.yml         # Docker Orchestration (15+ services)
├── docs/                      # Documentation
├── plans/                     # Implementation plans
└── .claude/                   # Claude Code Configuration
    ├── skills/                # Claude Kit Skills (82 skills)
    ├── agents/                # Claude Agents (13 agents)
    └── hooks/                 # Development hooks
```

## Core Technologies

### Frontend (blog-site)
- **Next.js 14** - App Router, Server Components, ISR
- **TypeScript** - Full type safety
- **Tailwind CSS** - Neo-brutalism styling
- **API Routes** - Server-side API handlers

### Backend (go-services)
- **Go** - High-performance API Gateway & Agents
- **Gin** - HTTP framework
- **PostgreSQL** - Primary database
- **Redis** - Caching & message broker

### ML/NLP (python-sentiment)
- **Python 3.9+** - ML service
- **FastAPI** - REST API
- **PhoBERT** - Vietnamese NLP model
- **ONNX** - Model optimization

### Orchestration
- **n8n** - Workflow automation
- **Docker Compose** - Container orchestration (15+ services)
- **Nginx** - Reverse proxy & load balancer

### Monitoring
- **Prometheus** - Metrics collection
- **Grafana** - Dashboards & visualization

### AI Integration
- **Claude API** - AI-powered analysis
- **Gemini API** - Image generation
- **Grok API** - Text analysis

## Key Components

### 1. Go Services (go-services/)

**Core Services:**
- `api-gateway` (port 8080) - Request routing, rate limiting
- `technical-agent` (port 8081) - Technical analysis (RSI, MACD, etc.)
- `forecast-agent` (port 8082) - ML-based forecasting

**Technical Indicators:**
```go
// internal/indicators/
rsi.go, macd.go, bollinger.go, ema.go, sma.go
stochastic.go, adx.go, atr.go, vwap.go
```

### 2. Python Sentiment Service (python-sentiment/)

**Features:**
- Vietnamese NLP with PhoBERT
- Stock slang dictionary ("lùa gà", "cá mập")
- Symbol extraction from text
- Sentiment scoring (-1 to +1)

**Structure:**
```
app/
├── routers/       # FastAPI endpoints
├── services/      # Business logic
├── models/        # ONNX models
└── tests/         # Unit tests
```

### 3. Blog Frontend (blog-site/)

**Routes:**
| Path | Description |
|------|-------------|
| `/` | Homepage + Market Board |
| `/articles` | Article listing |
| `/articles/[slug]` | Article detail |
| `/market` | Live market data |
| `/screener` | Stock filter/sort |
| `/symbols/[symbol]` | Symbol analysis |

**Components:**
- `MarketIndexBar` - Real-time VN30/HNX30/UPCOM
- `MarketBoardTable` - Live price board
- `StockChart` - Interactive TradingView chart
- `TrendingStocks` - Top mentioned stocks
- `ScreenerTable` - Filterable stock table

### 4. Crawl Agent (crawl-agent/)

**Data Sources:**
- VnEconomy RSS
- CafeF RSS
- VietStock RSS
- Market data APIs

### 5. Telegram Bot (telegram-bot/)

**Features:**
- Daily reports
- Alert subscriptions
- Symbol lookup
- Portfolio tracking

### 6. Claude Kit Integration (.claude/)

**Skills (82 total):**
- Development: plan, cook, debug, test, code-review
- Research: ask, brainstorm, research, scout
- Design: ui-ux-pro-max, frontend-design, web-design-guidelines
- DevOps: deploy, devops, docker

**Agents (13 total):**
- planner, researcher, fullstack-developer
- code-reviewer, tester, debugger
- docs-manager, git-manager
- brainstormer, project-manager, ui-ux-designer
- mcp-manager, code-simplifier

**Commands:**
- `/plan` - Research and planning
- `/ask` - Technical consultation
- `/debug` - Issue debugging
- `/scout` - Codebase exploration

### 7. Workflows (n8n)

**Primary Workflows:**
1. **n8n-workflow-daily-analysis.json** - Daily analysis pipeline
   - RSS scraping (VnEconomy, CafeF, VietStock)
   - Spam filtering
   - Hot stock detection
   - AI analysis trigger
   - Report generation

2. **n8n-workflow-error-handler.json** - Error handling
   - Alert on service failures
   - Retry logic
   - Notification dispatch

## Entry Points

### For Users
- **README.md**: Project overview and quick start
- **docs/blog-features.md**: Blog system documentation

### For Developers
- **blog-site/package.json**: Frontend dependencies
- **go-services/go.mod**: Backend dependencies
- **docker-compose.yml**: Service orchestration

### For Agents
- **CLAUDE.md**: Primary agent instructions
- **.claude/rules/**: Development rules and protocols

## Development Principles

### YAGNI (You Aren't Gonna Need It)
Avoid over-engineering and unnecessary features

### KISS (Keep It Simple, Stupid)
Prefer simple, straightforward solutions

### DRY (Don't Repeat Yourself)
Eliminate code duplication

### File Size Management
- Keep files under 500 lines
- Split large files into focused components
- Extract utilities into separate modules

### Security First
- Try-catch error handling
- Security standards coverage
- No secrets in commits
- Confidential info protection

## Agent Communication Protocol

**Report Format**: Markdown files in `./plans/<plan-name>/reports/`
**Naming Convention**: `{date}-from-[agent]-to-[agent]-[task]-report.md`

**Communication Patterns**:
- Sequential: Task dependencies require ordered execution
- Parallel: Independent tasks run simultaneously
- Query Fan-Out: Multiple researchers explore different approaches

## Git Workflow

**Commit Message Format**: Conventional Commits
```
type(scope): description

Types:
- feat: Features (minor bump)
- fix: Bug fixes (patch bump)
- docs: Documentation (patch bump)
- refactor: Code refactoring (patch bump)
- test: Tests (patch bump)
- ci: CI changes (patch bump)
- BREAKING CHANGE: Major version bump
```

**Automated Release**:
- Every push to `main` triggers release check
- Semantic versioning (MAJOR.MINOR.PATCH)
- Automated changelog generation
- GitHub releases with generated notes

## Testing Strategy

- Comprehensive unit tests required
- High code coverage mandatory
- Error scenario testing
- Performance validation
- Tests must pass before push
- No ignoring failed tests

## Documentation Standards

**Required Docs** (`./docs/`):
- `project-overview-pdr.md` - Project overview and PDR
- `code-standards.md` - Coding standards
- `codebase-summary.md` - This file
- `system-architecture.md` - Architecture documentation
- `blog-features.md` - Blog system documentation
- `project-roadmap.md` - Development roadmap

## File Statistics

**Code Files:**
| Language | Files |
|----------|-------|
| Go | 62 |
| Python | 295 |
| TypeScript | 48 |
| Markdown | 25+ |

**Services:** 15+ Docker containers

## Version History

**Current**: v1.0.0 (2026-04-20)
**License**: MIT
**Stack**: Go + Python + Next.js + n8n

## Related Documentation

- [Blog Features](./blog-features.md) - Detailed blog system docs
- [System Architecture](./system-architecture.md) - Architecture overview
- [Project Roadmap](./project-roadmap.md) - Development plan
