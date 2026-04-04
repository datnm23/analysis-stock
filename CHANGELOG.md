# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- JWT authentication middleware for Go API Gateway
- API key authentication middleware for service-to-service communication
- CORS hardening with configurable origins
- Python test infrastructure (pytest, fixtures, test files)
- GitHub Actions CI/CD pipeline
- CONTRIBUTING.md guidelines
- MIT License
- API documentation (authentication guide)

### Changed
- Updated `.env.example` with security and CORS configuration
- Updated Python config to support API key and CORS settings
- Updated Go config to support JWT and API key settings
- Updated Python sentiment service to use configurable CORS

### Fixed
- Fixed duplicate import in auth.go middleware

---

## [1.0.0] - 2024-01-29

### Added
- Technical Analysis Agent (RSI, MACD, Bollinger Bands, SMA/EMA, Stochastic, ADX, ATR, VWAP)
- Sentiment Analysis Agent with Vietnamese NLP (PhoBERT model)
- Forecast Agent combining technical and sentiment analysis
- Master Orchestrator for coordinating sub-agents
- Daily workflow automation with n8n
- Telegram bot integration for alerts
- PostgreSQL database integration
- Redis caching
- Docker Compose setup for all services

### Features
- Vietnamese stock symbol detection
- Stock market slang dictionary (lùa gà, cá mập, etc.)
- Hot stock detection from news
- Daily and weekly report generation
- RSS scraping from VnEconomy, CafeF, VietStock

---

## [0.1.0] - 2024-01-15

### Added
- Initial project setup
- Basic Go API Gateway
- Python sentiment service with PhoBERT
- Docker Compose configuration

---

## Upgrade Notes

### Upgrading to 1.1.0

1. Update your `.env` file with new security settings:
   ```bash
   ENABLE_AUTH=true
   JWT_SECRET_KEY=<your-secure-key>
   ```

2. Run database migrations if any

3. Restart all services

---

## Deprecated

- `allow_origins=["*"]` in CORS configuration - Use explicit origins instead
- Unauthenticated endpoints (except /health, /ready) - Now requires auth in production
