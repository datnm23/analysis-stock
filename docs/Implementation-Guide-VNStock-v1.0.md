# Implementation Guide
# VN Stock Analysis System
## Version 1.0

**Document ID:** IMPL-VNSTOCK-001
**Date:** January 2026
**Purpose:** Step-by-step development guide for building the complete system

---

## Table of Contents

1. [Introduction](#1-introduction)
2. [Phase 1: Core Infrastructure](#2-phase-1-core-infrastructure)
3. [Phase 2: Technical Analysis Agent](#3-phase-2-technical-analysis-agent)
4. [Phase 3: Sentiment Analysis Agent](#4-phase-3-sentiment-analysis-agent)
5. [Phase 4: Forecast Agent & Master Orchestrator](#5-phase-4-forecast-agent--master-orchestrator)
6. [Phase 5: Web Dashboard](#6-phase-5-web-dashboard)
7. [Phase 6: Telegram Bot](#7-phase-6-telegram-bot)
8. [Phase 7: Testing & Documentation](#8-phase-7-testing--documentation)
9. [Deployment Checklist](#9-deployment-checklist)

---

## 1. Introduction

### 1.1 Purpose

This guide provides step-by-step instructions for implementing the VN Stock Analysis System from the current state (~10% complete) to a fully functional production system.

### 1.2 Prerequisites

**Required Knowledge:**
- Python 3.9+ (FastAPI, pandas, async/await)
- JavaScript/TypeScript (Next.js, React)
- Docker and Docker Compose
- SQL (PostgreSQL)
- Redis basics
- Vietnamese stock market fundamentals

**Required Tools:**
```bash
# Verify installations
docker --version          # 20.10+
docker-compose --version  # 2.0+
python3 --version         # 3.9+
node --version            # 18+
git --version
```

### 1.3 Current State Assessment

| Component | Status | Location |
|-----------|--------|----------|
| Technical Agent | ✅ Complete | `technical_agent.py` |
| Docker Config | ✅ Complete | `docker-compose.yml` |
| n8n Workflow | ✅ Complete | `n8n-workflow-daily-analysis.json` |
| Setup Script | ✅ Complete | `quickstart.sh` |
| Env Template | ✅ Complete | `.env.example` |
| FastAPI Service | ❌ Missing | `agent-service/` |
| Sentiment Agent | ❌ Missing | - |
| Forecast Agent | ❌ Missing | - |
| Master Agent | ❌ Missing | - |
| Web Dashboard | ❌ Missing | `web-dashboard/` |
| Telegram Bot | ❌ Missing | `telegram-bot/` |
| Tests | ❌ Missing | `tests/` |

### 1.4 Target Directory Structure

```
vnstock-analysis/
├── agent-service/              # FastAPI backend
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py             # FastAPI entry point
│   │   ├── config.py           # Configuration
│   │   ├── database.py         # DB connection
│   │   ├── models/             # SQLAlchemy models
│   │   ├── schemas/            # Pydantic schemas
│   │   ├── api/
│   │   │   └── routes/         # API endpoints
│   │   ├── agents/             # AI agents
│   │   │   ├── base_agent.py
│   │   │   ├── technical_agent.py
│   │   │   ├── sentiment_agent.py
│   │   │   ├── forecast_agent.py
│   │   │   └── master_agent.py
│   │   ├── services/           # Business logic
│   │   └── utils/              # Helpers
│   ├── tests/
│   ├── Dockerfile
│   └── requirements.txt
├── web-dashboard/              # Next.js frontend
│   ├── src/
│   │   ├── app/
│   │   ├── components/
│   │   └── lib/
│   ├── Dockerfile
│   └── package.json
├── telegram-bot/               # Telegram service
│   ├── bot/
│   │   ├── main.py
│   │   └── handlers/
│   ├── Dockerfile
│   └── requirements.txt
├── init-scripts/               # Database init
│   └── 01_init_db.sql
├── workflows/                  # n8n backups
├── docs/                       # Documentation
├── tests/                      # Integration tests
├── docker-compose.yml
├── .env.example
├── CLAUDE.md
└── README.md
```

---

## 2. Phase 1: Core Infrastructure

### 2.1 Objectives

- Create FastAPI application structure
- Set up database models
- Configure Docker services
- Implement health checks

### 2.2 Files to Create

```
agent-service/
├── app/
│   ├── __init__.py
│   ├── main.py
│   ├── config.py
│   ├── database.py
│   └── models/
│       ├── __init__.py
│       ├── stock.py
│       ├── analysis.py
│       └── user.py
├── Dockerfile
└── requirements.txt

init-scripts/
└── 01_init_db.sql
```

### 2.3 Implementation

#### Step 1: Create directory structure

```bash
mkdir -p agent-service/app/{models,schemas,api/routes,agents,services,utils}
mkdir -p agent-service/tests
mkdir -p init-scripts
touch agent-service/app/__init__.py
touch agent-service/app/models/__init__.py
touch agent-service/app/schemas/__init__.py
touch agent-service/app/api/__init__.py
touch agent-service/app/api/routes/__init__.py
touch agent-service/app/agents/__init__.py
touch agent-service/app/services/__init__.py
touch agent-service/app/utils/__init__.py
```

#### Step 2: Create requirements.txt

```
# agent-service/requirements.txt
fastapi==0.109.0
uvicorn[standard]==0.27.0
sqlalchemy==2.0.25
psycopg2-binary==2.9.9
redis==5.0.1
boto3==1.34.0
pandas==2.1.4
numpy==1.26.3
ta==0.11.0
vnstock==0.0.26
pydantic==2.5.3
pydantic-settings==2.1.0
python-dotenv==1.0.0
celery[redis]==5.3.6
httpx==0.26.0
python-jose[cryptography]==3.3.0
passlib[bcrypt]==1.7.4
python-multipart==0.0.6
jinja2==3.1.3
pytz==2024.1
scipy==1.12.0
transformers==4.37.0
torch==2.1.2
```

#### Step 3: Create config.py

```python
# agent-service/app/config.py
from pydantic_settings import BaseSettings
from functools import lru_cache

class Settings(BaseSettings):
    # Application
    APP_NAME: str = "VN Stock Analysis API"
    APP_VERSION: str = "1.0.0"
    DEBUG: bool = False

    # Database
    DATABASE_URL: str = "postgresql://admin:password@postgres:5432/vnstock"

    # Redis
    REDIS_URL: str = "redis://:password@redis:6379/0"

    # AWS S3
    AWS_ACCESS_KEY_ID: str = ""
    AWS_SECRET_ACCESS_KEY: str = ""
    AWS_REGION: str = "ap-southeast-1"
    S3_BUCKET: str = "vnstock-data"

    # API Keys
    ANTHROPIC_API_KEY: str = ""

    # JWT
    JWT_SECRET_KEY: str = "your-secret-key-change-in-production"
    JWT_ALGORITHM: str = "HS256"
    JWT_EXPIRATION_HOURS: int = 24

    # Vietnamese Market
    VN_TIMEZONE: str = "Asia/Ho_Chi_Minh"
    VN_MARKET_OPEN_HOUR: int = 9
    VN_MARKET_CLOSE_HOUR: int = 15

    # Rate Limiting
    RATE_LIMIT_ANONYMOUS: int = 100
    RATE_LIMIT_AUTHENTICATED: int = 1000

    # Cache TTL (seconds)
    CACHE_TTL_TECHNICAL: int = 300
    CACHE_TTL_SENTIMENT: int = 600

    class Config:
        env_file = ".env"
        case_sensitive = True

@lru_cache()
def get_settings() -> Settings:
    return Settings()

settings = get_settings()
```

#### Step 4: Create database.py

```python
# agent-service/app/database.py
from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from app.config import settings

engine = create_engine(
    settings.DATABASE_URL,
    pool_size=10,
    max_overflow=20,
    pool_pre_ping=True
)

SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

Base = declarative_base()

def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
```

#### Step 5: Create models

```python
# agent-service/app/models/stock.py
from sqlalchemy import Column, Integer, String, Boolean, TIMESTAMP, Numeric, BigInteger, Date
from sqlalchemy.dialects.postgresql import ARRAY, JSONB
from sqlalchemy.sql import func
from app.database import Base

class Stock(Base):
    __tablename__ = "stocks"

    id = Column(Integer, primary_key=True, index=True)
    symbol = Column(String(10), unique=True, nullable=False, index=True)
    name = Column(String(255))
    exchange = Column(String(10))  # HSX, HNX, UPCOM
    sector = Column(String(100))
    is_active = Column(Boolean, default=True)
    created_at = Column(TIMESTAMP, server_default=func.now())
    updated_at = Column(TIMESTAMP, server_default=func.now(), onupdate=func.now())

class PriceHistory(Base):
    __tablename__ = "price_history"

    id = Column(Integer, primary_key=True, index=True)
    stock_id = Column(Integer, nullable=False, index=True)
    date = Column(Date, nullable=False)
    open = Column(Numeric(15, 2))
    high = Column(Numeric(15, 2))
    low = Column(Numeric(15, 2))
    close = Column(Numeric(15, 2))
    volume = Column(BigInteger)
```

```python
# agent-service/app/models/analysis.py
from sqlalchemy import Column, Integer, String, Date, Numeric, TIMESTAMP, Boolean, Text
from sqlalchemy.dialects.postgresql import JSONB, ARRAY
from sqlalchemy.sql import func
from app.database import Base

class AnalysisResult(Base):
    __tablename__ = "analysis_results"

    id = Column(Integer, primary_key=True, index=True)
    stock_id = Column(Integer, nullable=False, index=True)
    analysis_date = Column(Date, nullable=False)
    recommendation = Column(String(20))
    confidence = Column(Numeric(5, 4))
    technical_score = Column(Numeric(5, 2))
    sentiment_score = Column(Numeric(5, 4))
    risk_level = Column(String(10))
    raw_data = Column(JSONB)
    created_at = Column(TIMESTAMP, server_default=func.now())

class NewsArticle(Base):
    __tablename__ = "news_articles"

    id = Column(Integer, primary_key=True, index=True)
    source = Column(String(50), nullable=False)
    title = Column(Text, nullable=False)
    description = Column(Text)
    url = Column(Text, unique=True)
    published_at = Column(TIMESTAMP)
    sentiment = Column(String(20))
    sentiment_score = Column(Numeric(5, 4))
    mentioned_symbols = Column(ARRAY(Text))
    is_spam = Column(Boolean, default=False)
    created_at = Column(TIMESTAMP, server_default=func.now())

class Alert(Base):
    __tablename__ = "alerts"

    id = Column(Integer, primary_key=True, index=True)
    stock_id = Column(Integer, index=True)
    alert_type = Column(String(50), nullable=False)
    message = Column(Text, nullable=False)
    severity = Column(String(20))
    is_sent = Column(Boolean, default=False)
    created_at = Column(TIMESTAMP, server_default=func.now())
    sent_at = Column(TIMESTAMP)
```

```python
# agent-service/app/models/user.py
from sqlalchemy import Column, Integer, String, BigInteger, TIMESTAMP
from sqlalchemy.dialects.postgresql import JSONB, ARRAY
from sqlalchemy.sql import func
from app.database import Base

class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, index=True)
    telegram_id = Column(BigInteger, unique=True, index=True)
    email = Column(String(255), unique=True, index=True)
    username = Column(String(100))
    watchlist = Column(ARRAY(String), default=[])
    alert_preferences = Column(JSONB, default={})
    tier = Column(String(20), default="free")
    created_at = Column(TIMESTAMP, server_default=func.now())
    last_active = Column(TIMESTAMP)

class UserSubscription(Base):
    __tablename__ = "user_subscriptions"

    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(Integer, nullable=False, index=True)
    stock_id = Column(Integer, nullable=False, index=True)
    alert_types = Column(ARRAY(String), default=["HIGH_RISK", "STRONG_SIGNAL"])
    is_active = Column(Boolean, default=True)
    created_at = Column(TIMESTAMP, server_default=func.now())
```

```python
# agent-service/app/models/__init__.py
from app.models.stock import Stock, PriceHistory
from app.models.analysis import AnalysisResult, NewsArticle, Alert
from app.models.user import User, UserSubscription

__all__ = [
    "Stock", "PriceHistory",
    "AnalysisResult", "NewsArticle", "Alert",
    "User", "UserSubscription"
]
```

#### Step 6: Create main.py

```python
# agent-service/app/main.py
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
import logging

from app.config import settings
from app.database import engine, Base
from app.api.routes import analysis, reports, health

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='{"timestamp": "%(asctime)s", "level": "%(levelname)s", "message": "%(message)s"}'
)
logger = logging.getLogger(__name__)

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    logger.info("Starting VN Stock Analysis API...")
    Base.metadata.create_all(bind=engine)
    logger.info("Database tables created/verified")
    yield
    # Shutdown
    logger.info("Shutting down VN Stock Analysis API...")

app = FastAPI(
    title=settings.APP_NAME,
    version=settings.APP_VERSION,
    description="API for Vietnamese Stock Market Analysis",
    lifespan=lifespan
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(health.router, tags=["Health"])
app.include_router(analysis.router, prefix="/api", tags=["Analysis"])
app.include_router(reports.router, prefix="/api", tags=["Reports"])

@app.get("/")
async def root():
    return {
        "name": settings.APP_NAME,
        "version": settings.APP_VERSION,
        "docs": "/docs"
    }
```

#### Step 7: Create health route

```python
# agent-service/app/api/routes/health.py
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from sqlalchemy import text
import redis

from app.database import get_db
from app.config import settings

router = APIRouter()

@router.get("/health")
async def health_check(db: Session = Depends(get_db)):
    """Comprehensive health check for all services"""
    health_status = {
        "status": "healthy",
        "version": settings.APP_VERSION,
        "services": {}
    }

    # Check database
    try:
        db.execute(text("SELECT 1"))
        health_status["services"]["database"] = "connected"
    except Exception as e:
        health_status["services"]["database"] = f"error: {str(e)}"
        health_status["status"] = "degraded"

    # Check Redis
    try:
        r = redis.from_url(settings.REDIS_URL)
        r.ping()
        health_status["services"]["redis"] = "connected"
    except Exception as e:
        health_status["services"]["redis"] = f"error: {str(e)}"
        health_status["status"] = "degraded"

    return health_status
```

#### Step 8: Create placeholder routes

```python
# agent-service/app/api/routes/analysis.py
from fastapi import APIRouter

router = APIRouter()

@router.post("/analyze/technical")
async def analyze_technical():
    """Placeholder - implemented in Phase 2"""
    return {"message": "Technical analysis endpoint - coming soon"}

@router.post("/analyze/sentiment")
async def analyze_sentiment():
    """Placeholder - implemented in Phase 3"""
    return {"message": "Sentiment analysis endpoint - coming soon"}

@router.post("/synthesize")
async def synthesize():
    """Placeholder - implemented in Phase 4"""
    return {"message": "Synthesis endpoint - coming soon"}
```

```python
# agent-service/app/api/routes/reports.py
from fastapi import APIRouter

router = APIRouter()

@router.get("/reports/daily")
async def get_daily_report():
    """Placeholder - implemented in Phase 4"""
    return {"message": "Daily report endpoint - coming soon"}

@router.get("/stocks/{symbol}")
async def get_stock_analysis(symbol: str):
    """Placeholder - implemented in Phase 4"""
    return {"symbol": symbol, "message": "Stock analysis endpoint - coming soon"}
```

```python
# agent-service/app/api/routes/__init__.py
from app.api.routes import health, analysis, reports
```

#### Step 9: Create Dockerfile

```dockerfile
# agent-service/Dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    libpq-dev \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Create non-root user
RUN useradd -m appuser && chown -R appuser:appuser /app
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Run FastAPI
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

#### Step 10: Create database init script

```sql
-- init-scripts/01_init_db.sql
-- Create databases
SELECT 'CREATE DATABASE vnstock' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'vnstock')\gexec
SELECT 'CREATE DATABASE n8n' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'n8n')\gexec

\c vnstock;

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Insert initial stock data (top Vietnamese stocks)
INSERT INTO stocks (symbol, name, exchange, sector) VALUES
    ('VNM', 'Công ty CP Sữa Việt Nam', 'HSX', 'Hàng tiêu dùng'),
    ('HPG', 'Tập đoàn Hòa Phát', 'HSX', 'Vật liệu xây dựng'),
    ('VCB', 'Ngân hàng TMCP Ngoại thương Việt Nam', 'HSX', 'Ngân hàng'),
    ('VHM', 'Công ty CP Vinhomes', 'HSX', 'Bất động sản'),
    ('VIC', 'Tập đoàn Vingroup', 'HSX', 'Bất động sản'),
    ('FPT', 'Công ty CP FPT', 'HSX', 'Công nghệ'),
    ('MSN', 'Tập đoàn Masan', 'HSX', 'Hàng tiêu dùng'),
    ('MBB', 'Ngân hàng TMCP Quân đội', 'HSX', 'Ngân hàng'),
    ('TCB', 'Ngân hàng TMCP Kỹ Thương', 'HSX', 'Ngân hàng'),
    ('BID', 'Ngân hàng TMCP Đầu tư và Phát triển', 'HSX', 'Ngân hàng'),
    ('CTG', 'Ngân hàng TMCP Công Thương', 'HSX', 'Ngân hàng'),
    ('GAS', 'Tổng Công ty Khí Việt Nam', 'HSX', 'Năng lượng'),
    ('SAB', 'Tổng Công ty CP Bia - Rượu - Nước giải khát Sài Gòn', 'HSX', 'Hàng tiêu dùng'),
    ('VRE', 'Công ty CP Vincom Retail', 'HSX', 'Bất động sản'),
    ('PLX', 'Tập đoàn Xăng dầu Việt Nam', 'HSX', 'Năng lượng'),
    ('VPB', 'Ngân hàng TMCP Việt Nam Thịnh Vượng', 'HSX', 'Ngân hàng'),
    ('NVL', 'Tập đoàn Novaland', 'HSX', 'Bất động sản'),
    ('MWG', 'Công ty CP Đầu tư Thế Giới Di Động', 'HSX', 'Bán lẻ'),
    ('POW', 'Tổng Công ty Điện lực Dầu khí Việt Nam', 'HSX', 'Năng lượng'),
    ('STB', 'Ngân hàng TMCP Sài Gòn Thương Tín', 'HSX', 'Ngân hàng')
ON CONFLICT (symbol) DO NOTHING;
```

### 2.4 Acceptance Criteria

- [ ] `docker-compose up -d` starts all services without errors
- [ ] `curl http://localhost:8000/health` returns healthy status
- [ ] Database tables created automatically on startup
- [ ] Redis connectivity confirmed in health check
- [ ] API documentation available at `http://localhost:8000/docs`

---

## 3. Phase 2: Technical Analysis Agent

### 3.1 Objectives

- Refactor existing `technical_agent.py` into modular structure
- Create Pydantic schemas for request/response
- Implement Redis caching
- Create API endpoint

### 3.2 Files to Create/Modify

```
agent-service/app/
├── agents/
│   ├── base_agent.py
│   └── technical_agent.py
├── schemas/
│   └── analysis.py
├── services/
│   └── cache_service.py
└── api/routes/
    └── analysis.py (update)
```

### 3.3 Implementation

#### Step 1: Create base agent

```python
# agent-service/app/agents/base_agent.py
from abc import ABC, abstractmethod
from typing import Dict, Any
import logging
from datetime import datetime
import pytz

from app.config import settings

class BaseAgent(ABC):
    """Base class for all analysis agents"""

    def __init__(self, name: str):
        self.name = name
        self.logger = logging.getLogger(name)
        self.vn_tz = pytz.timezone(settings.VN_TIMEZONE)

    @abstractmethod
    async def analyze(self, **kwargs) -> Dict[str, Any]:
        """Perform analysis and return results"""
        pass

    def get_vietnam_time(self) -> datetime:
        """Get current time in Vietnam timezone"""
        return datetime.now(self.vn_tz)

    def is_market_open(self) -> bool:
        """Check if Vietnamese stock market is currently open"""
        now = self.get_vietnam_time()

        # Check weekday (Monday=0, Friday=4)
        if now.weekday() > 4:
            return False

        hour = now.hour
        minute = now.minute

        # Morning session: 9:00 - 11:30
        if 9 <= hour < 11:
            return True
        if hour == 11 and minute <= 30:
            return True

        # Afternoon session: 13:00 - 15:00
        if 13 <= hour < 15:
            return True

        return False

    def validate_symbol(self, symbol: str) -> bool:
        """Validate Vietnamese stock symbol format"""
        if not symbol:
            return False
        if len(symbol) != 3:
            return False
        if not symbol.isalpha():
            return False
        if not symbol.isupper():
            return False
        return True
```

#### Step 2: Create technical agent (refactored)

```python
# agent-service/app/agents/technical_agent.py
"""
Technical Analysis Agent for Vietnamese Stock Market
Refactored from standalone technical_agent.py
"""

import pandas as pd
import numpy as np
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
import ta
from scipy.signal import argrelextrema

from app.agents.base_agent import BaseAgent

class TechnicalAnalysisAgent(BaseAgent):
    """
    Technical analysis agent that calculates indicators and generates signals
    for Vietnamese stock market.
    """

    # Indicator weights for final score
    WEIGHTS = {
        'rsi': 0.15,
        'macd': 0.20,
        'bollinger': 0.15,
        'moving_average': 0.20,
        'volume': 0.10,
        'adx': 0.10,
        'stochastic': 0.10
    }

    def __init__(self, symbol: str, days: int = 90):
        super().__init__("TechnicalAnalysisAgent")

        if not self.validate_symbol(symbol):
            raise ValueError(f"Invalid symbol format: {symbol}")

        self.symbol = symbol.upper()
        self.days = min(max(days, 30), 365)  # Clamp between 30-365
        self.data: Optional[pd.DataFrame] = None
        self.indicators: Dict[str, pd.Series] = {}

    async def analyze(self, **kwargs) -> Dict[str, Any]:
        """Main analysis entry point"""
        return self.get_full_analysis()

    def fetch_data(self) -> pd.DataFrame:
        """Fetch historical data using vnstock library"""
        try:
            from vnstock import stock_historical_data

            end_date = datetime.now()
            start_date = end_date - timedelta(days=self.days + 50)  # Extra for indicator warmup

            df = stock_historical_data(
                symbol=self.symbol,
                start_date=start_date.strftime('%Y-%m-%d'),
                end_date=end_date.strftime('%Y-%m-%d'),
                resolution='1D',
                type='stock'
            )

            if df is None or df.empty:
                raise ValueError(f"No data returned for {self.symbol}")

            # Standardize column names
            df.columns = df.columns.str.lower()
            required_cols = ['open', 'high', 'low', 'close', 'volume']

            for col in required_cols:
                if col not in df.columns:
                    raise ValueError(f"Missing required column: {col}")

            self.data = df.tail(self.days)
            return self.data

        except ImportError:
            self.logger.warning("vnstock not installed, using mock data")
            return self._generate_mock_data()
        except Exception as e:
            self.logger.error(f"Error fetching data for {self.symbol}: {e}")
            raise

    def _generate_mock_data(self) -> pd.DataFrame:
        """Generate mock data for testing"""
        dates = pd.date_range(end=datetime.now(), periods=self.days, freq='D')
        np.random.seed(42)

        base_price = 50000
        prices = [base_price]
        for _ in range(self.days - 1):
            change = np.random.normal(0, 0.02) * prices[-1]
            prices.append(max(prices[-1] + change, 1000))

        df = pd.DataFrame({
            'date': dates,
            'open': prices,
            'high': [p * (1 + np.random.uniform(0, 0.03)) for p in prices],
            'low': [p * (1 - np.random.uniform(0, 0.03)) for p in prices],
            'close': [p * (1 + np.random.uniform(-0.02, 0.02)) for p in prices],
            'volume': [np.random.randint(100000, 1000000) for _ in prices]
        })

        self.data = df
        return df

    def calculate_all_indicators(self) -> Dict[str, pd.Series]:
        """Calculate all technical indicators"""
        if self.data is None:
            self.fetch_data()

        df = self.data

        # RSI
        self.indicators['rsi'] = ta.momentum.RSIIndicator(
            close=df['close'], window=14
        ).rsi()

        # MACD
        macd = ta.trend.MACD(close=df['close'])
        self.indicators['macd'] = macd.macd()
        self.indicators['macd_signal'] = macd.macd_signal()
        self.indicators['macd_histogram'] = macd.macd_diff()

        # Bollinger Bands
        bb = ta.volatility.BollingerBands(close=df['close'], window=20)
        self.indicators['bb_upper'] = bb.bollinger_hband()
        self.indicators['bb_middle'] = bb.bollinger_mavg()
        self.indicators['bb_lower'] = bb.bollinger_lband()

        # Moving Averages
        self.indicators['sma_20'] = ta.trend.SMAIndicator(
            close=df['close'], window=20
        ).sma_indicator()
        self.indicators['sma_50'] = ta.trend.SMAIndicator(
            close=df['close'], window=50
        ).sma_indicator()
        self.indicators['ema_12'] = ta.trend.EMAIndicator(
            close=df['close'], window=12
        ).ema_indicator()
        self.indicators['ema_26'] = ta.trend.EMAIndicator(
            close=df['close'], window=26
        ).ema_indicator()

        # Stochastic
        stoch = ta.momentum.StochasticOscillator(
            high=df['high'], low=df['low'], close=df['close']
        )
        self.indicators['stoch_k'] = stoch.stoch()
        self.indicators['stoch_d'] = stoch.stoch_signal()

        # ADX
        adx = ta.trend.ADXIndicator(
            high=df['high'], low=df['low'], close=df['close']
        )
        self.indicators['adx'] = adx.adx()
        self.indicators['adx_pos'] = adx.adx_pos()
        self.indicators['adx_neg'] = adx.adx_neg()

        # ATR
        self.indicators['atr'] = ta.volatility.AverageTrueRange(
            high=df['high'], low=df['low'], close=df['close']
        ).average_true_range()

        # Volume analysis
        avg_volume = df['volume'].rolling(window=20).mean()
        self.indicators['volume_ratio'] = df['volume'] / avg_volume

        return self.indicators

    def calculate_support_resistance(self) -> Dict[str, List[float]]:
        """Calculate support and resistance levels"""
        if self.data is None:
            self.fetch_data()

        df = self.data
        close = df['close'].values

        # Find local maxima (resistance) and minima (support)
        order = 5

        resistance_idx = argrelextrema(close, np.greater, order=order)[0]
        support_idx = argrelextrema(close, np.less, order=order)[0]

        resistance_levels = sorted(close[resistance_idx], reverse=True)[:3]
        support_levels = sorted(close[support_idx])[:3]

        return {
            'resistance': [round(float(r), 2) for r in resistance_levels],
            'support': [round(float(s), 2) for s in support_levels]
        }

    def calculate_fibonacci_levels(self) -> Dict[str, float]:
        """Calculate Fibonacci retracement levels"""
        if self.data is None:
            self.fetch_data()

        df = self.data
        high = df['high'].max()
        low = df['low'].min()
        diff = high - low

        return {
            '0.000': round(float(high), 2),
            '0.236': round(float(high - 0.236 * diff), 2),
            '0.382': round(float(high - 0.382 * diff), 2),
            '0.500': round(float(high - 0.500 * diff), 2),
            '0.618': round(float(high - 0.618 * diff), 2),
            '1.000': round(float(low), 2)
        }

    def generate_signals(self) -> Dict[str, Any]:
        """Generate trading signals from indicators"""
        if not self.indicators:
            self.calculate_all_indicators()

        signals = []
        scores = {}

        # Get latest values
        rsi = self.indicators['rsi'].iloc[-1]
        macd = self.indicators['macd'].iloc[-1]
        macd_signal = self.indicators['macd_signal'].iloc[-1]
        close = self.data['close'].iloc[-1]
        sma_20 = self.indicators['sma_20'].iloc[-1]
        sma_50 = self.indicators['sma_50'].iloc[-1]
        bb_upper = self.indicators['bb_upper'].iloc[-1]
        bb_lower = self.indicators['bb_lower'].iloc[-1]
        stoch_k = self.indicators['stoch_k'].iloc[-1]
        adx = self.indicators['adx'].iloc[-1]
        adx_pos = self.indicators['adx_pos'].iloc[-1]
        adx_neg = self.indicators['adx_neg'].iloc[-1]
        volume_ratio = self.indicators['volume_ratio'].iloc[-1]

        # RSI signals
        if rsi < 30:
            signals.append("RSI cho thấy quá bán (oversold)")
            scores['rsi'] = 1.0
        elif rsi > 70:
            signals.append("RSI cho thấy quá mua (overbought)")
            scores['rsi'] = -1.0
        else:
            scores['rsi'] = (50 - rsi) / 50  # Normalize to -1 to 1

        # MACD signals
        if macd > macd_signal:
            if self.indicators['macd'].iloc[-2] <= self.indicators['macd_signal'].iloc[-2]:
                signals.append("MACD vừa cắt lên đường tín hiệu (bullish crossover)")
                scores['macd'] = 1.0
            else:
                scores['macd'] = 0.5
        else:
            if self.indicators['macd'].iloc[-2] >= self.indicators['macd_signal'].iloc[-2]:
                signals.append("MACD vừa cắt xuống đường tín hiệu (bearish crossover)")
                scores['macd'] = -1.0
            else:
                scores['macd'] = -0.5

        # Bollinger Bands signals
        if close < bb_lower:
            signals.append("Giá dưới Bollinger Band dưới - có thể phục hồi")
            scores['bollinger'] = 1.0
        elif close > bb_upper:
            signals.append("Giá trên Bollinger Band trên - có thể điều chỉnh")
            scores['bollinger'] = -1.0
        else:
            bb_mid = self.indicators['bb_middle'].iloc[-1]
            scores['bollinger'] = (bb_mid - close) / (bb_upper - bb_mid) if bb_upper != bb_mid else 0

        # Moving Average signals
        if sma_20 > sma_50:
            if self.indicators['sma_20'].iloc[-2] <= self.indicators['sma_50'].iloc[-2]:
                signals.append("Golden Cross - SMA20 cắt lên SMA50")
                scores['moving_average'] = 1.0
            else:
                scores['moving_average'] = 0.5 if close > sma_20 else 0.3
        else:
            if self.indicators['sma_20'].iloc[-2] >= self.indicators['sma_50'].iloc[-2]:
                signals.append("Death Cross - SMA20 cắt xuống SMA50")
                scores['moving_average'] = -1.0
            else:
                scores['moving_average'] = -0.5 if close < sma_20 else -0.3

        # Stochastic signals
        if stoch_k < 20:
            signals.append("Stochastic cho thấy quá bán")
            scores['stochastic'] = 1.0
        elif stoch_k > 80:
            signals.append("Stochastic cho thấy quá mua")
            scores['stochastic'] = -1.0
        else:
            scores['stochastic'] = (50 - stoch_k) / 50

        # ADX signals
        if adx > 25:
            if adx_pos > adx_neg:
                signals.append(f"Xu hướng tăng mạnh (ADX: {adx:.1f})")
                scores['adx'] = 0.8
            else:
                signals.append(f"Xu hướng giảm mạnh (ADX: {adx:.1f})")
                scores['adx'] = -0.8
        else:
            signals.append(f"Xu hướng yếu/không rõ (ADX: {adx:.1f})")
            scores['adx'] = 0.0

        # Volume signals
        if volume_ratio > 1.5:
            signals.append(f"Khối lượng cao bất thường ({volume_ratio:.1f}x trung bình)")
            scores['volume'] = 0.5 if scores.get('macd', 0) > 0 else -0.5
        else:
            scores['volume'] = 0.0

        # Calculate weighted final score
        final_score = sum(
            scores.get(key, 0) * weight
            for key, weight in self.WEIGHTS.items()
        )

        # Generate recommendation
        if final_score > 0.6:
            recommendation = "STRONG BUY"
        elif final_score > 0.2:
            recommendation = "BUY"
        elif final_score > -0.2:
            recommendation = "HOLD"
        elif final_score > -0.6:
            recommendation = "SELL"
        else:
            recommendation = "STRONG SELL"

        # Calculate confidence
        confidence = min(abs(final_score) + 0.3, 1.0)

        return {
            'recommendation': recommendation,
            'confidence': round(confidence, 4),
            'technical_score': round(final_score, 4),
            'signals': signals,
            'indicator_scores': {k: round(v, 4) for k, v in scores.items()}
        }

    def get_full_analysis(self) -> Dict[str, Any]:
        """Get comprehensive technical analysis"""
        if self.data is None:
            self.fetch_data()

        self.calculate_all_indicators()
        signals = self.generate_signals()
        support_resistance = self.calculate_support_resistance()
        fibonacci = self.calculate_fibonacci_levels()

        current_price = float(self.data['close'].iloc[-1])

        # Calculate price targets
        if signals['recommendation'] in ['BUY', 'STRONG BUY']:
            short_term_target = current_price * 1.05
            stop_loss = current_price * 0.95
        elif signals['recommendation'] in ['SELL', 'STRONG SELL']:
            short_term_target = current_price * 0.95
            stop_loss = current_price * 1.05
        else:
            short_term_target = current_price
            stop_loss = current_price * 0.97

        return {
            'symbol': self.symbol,
            'timestamp': self.get_vietnam_time().isoformat(),
            'current_price': round(current_price, 2),
            'recommendation': signals['recommendation'],
            'confidence': signals['confidence'],
            'technical_score': signals['technical_score'],
            'signals': signals['signals'],
            'indicator_scores': signals['indicator_scores'],
            'indicators': {
                'rsi': round(float(self.indicators['rsi'].iloc[-1]), 2),
                'macd': round(float(self.indicators['macd'].iloc[-1]), 4),
                'macd_signal': round(float(self.indicators['macd_signal'].iloc[-1]), 4),
                'sma_20': round(float(self.indicators['sma_20'].iloc[-1]), 2),
                'sma_50': round(float(self.indicators['sma_50'].iloc[-1]), 2),
                'ema_12': round(float(self.indicators['ema_12'].iloc[-1]), 2),
                'ema_26': round(float(self.indicators['ema_26'].iloc[-1]), 2),
                'bb_upper': round(float(self.indicators['bb_upper'].iloc[-1]), 2),
                'bb_middle': round(float(self.indicators['bb_middle'].iloc[-1]), 2),
                'bb_lower': round(float(self.indicators['bb_lower'].iloc[-1]), 2),
                'stoch_k': round(float(self.indicators['stoch_k'].iloc[-1]), 2),
                'stoch_d': round(float(self.indicators['stoch_d'].iloc[-1]), 2),
                'adx': round(float(self.indicators['adx'].iloc[-1]), 2),
                'atr': round(float(self.indicators['atr'].iloc[-1]), 2),
                'volume_ratio': round(float(self.indicators['volume_ratio'].iloc[-1]), 2)
            },
            'support_resistance': support_resistance,
            'price_targets': {
                'short_term': round(short_term_target, 2),
                'stop_loss': round(stop_loss, 2),
                'fibonacci': fibonacci
            }
        }
```

#### Step 3: Create schemas

```python
# agent-service/app/schemas/analysis.py
from pydantic import BaseModel, Field, field_validator
from typing import List, Dict, Optional
from enum import Enum
from datetime import datetime

class RecommendationType(str, Enum):
    STRONG_BUY = "STRONG BUY"
    BUY = "BUY"
    HOLD = "HOLD"
    SELL = "SELL"
    STRONG_SELL = "STRONG SELL"

class RiskLevel(str, Enum):
    LOW = "LOW"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"

# Request schemas
class TechnicalAnalysisRequest(BaseModel):
    symbol: str = Field(..., min_length=3, max_length=3, description="Stock symbol (3 letters)")
    days: int = Field(default=90, ge=30, le=365, description="Number of days to analyze")

    @field_validator('symbol')
    @classmethod
    def validate_symbol(cls, v: str) -> str:
        if not v.isalpha():
            raise ValueError('Symbol must contain only letters')
        return v.upper()

class SentimentAnalysisRequest(BaseModel):
    texts: List[str] = Field(..., min_length=1, description="Texts to analyze")
    symbol: Optional[str] = Field(None, min_length=3, max_length=3)

class SynthesisRequest(BaseModel):
    technical: Dict
    sentiment: Dict
    market_context: Optional[Dict] = None

# Response schemas
class IndicatorValues(BaseModel):
    rsi: float
    macd: float
    macd_signal: float
    sma_20: float
    sma_50: float
    ema_12: float
    ema_26: float
    bb_upper: float
    bb_middle: float
    bb_lower: float
    stoch_k: float
    stoch_d: float
    adx: float
    atr: float
    volume_ratio: float

class SupportResistance(BaseModel):
    resistance: List[float]
    support: List[float]

class PriceTargets(BaseModel):
    short_term: float
    stop_loss: float
    fibonacci: Dict[str, float]

class TechnicalAnalysisResponse(BaseModel):
    symbol: str
    timestamp: str
    current_price: float
    recommendation: RecommendationType
    confidence: float = Field(..., ge=0, le=1)
    technical_score: float = Field(..., ge=-1, le=1)
    signals: List[str]
    indicator_scores: Dict[str, float]
    indicators: IndicatorValues
    support_resistance: SupportResistance
    price_targets: PriceTargets

class SentimentResult(BaseModel):
    sentiment: str
    confidence: float
    scores: Dict[str, float]

class SentimentAggregateResponse(BaseModel):
    symbol: Optional[str]
    analysis_date: str
    articles_analyzed: int
    aggregate: Dict[str, float]
    individual_results: Optional[List[SentimentResult]] = None

class ForecastResponse(BaseModel):
    symbol: str
    recommendation: RecommendationType
    confidence: float
    risk_level: RiskLevel
    reasoning: List[str]
    scores: Dict[str, float]
```

#### Step 4: Create cache service

```python
# agent-service/app/services/cache_service.py
import json
import redis
from typing import Optional, Any
from app.config import settings

class CacheService:
    def __init__(self):
        self.redis = redis.from_url(settings.REDIS_URL, decode_responses=True)

    async def get(self, key: str) -> Optional[Any]:
        """Get value from cache"""
        try:
            value = self.redis.get(key)
            if value:
                return json.loads(value)
            return None
        except Exception:
            return None

    async def set(self, key: str, value: Any, ttl: int = 300) -> bool:
        """Set value in cache with TTL"""
        try:
            self.redis.setex(key, ttl, json.dumps(value))
            return True
        except Exception:
            return False

    async def delete(self, key: str) -> bool:
        """Delete key from cache"""
        try:
            self.redis.delete(key)
            return True
        except Exception:
            return False

    async def clear_pattern(self, pattern: str) -> int:
        """Delete all keys matching pattern"""
        try:
            keys = self.redis.keys(pattern)
            if keys:
                return self.redis.delete(*keys)
            return 0
        except Exception:
            return 0

def get_cache_service() -> CacheService:
    return CacheService()
```

#### Step 5: Update analysis routes

```python
# agent-service/app/api/routes/analysis.py
from fastapi import APIRouter, HTTPException, Depends
from app.agents.technical_agent import TechnicalAnalysisAgent
from app.schemas.analysis import (
    TechnicalAnalysisRequest,
    TechnicalAnalysisResponse,
    SentimentAnalysisRequest,
    SynthesisRequest
)
from app.services.cache_service import CacheService, get_cache_service
from app.config import settings
import logging

router = APIRouter()
logger = logging.getLogger(__name__)

@router.post("/analyze/technical", response_model=TechnicalAnalysisResponse)
async def analyze_technical(
    request: TechnicalAnalysisRequest,
    cache: CacheService = Depends(get_cache_service)
):
    """
    Perform technical analysis on a Vietnamese stock.

    Returns 15+ technical indicators and buy/sell/hold recommendation.
    """
    cache_key = f"technical:{request.symbol}:{request.days}"

    # Check cache
    cached = await cache.get(cache_key)
    if cached:
        logger.info(f"Cache hit for {request.symbol}")
        return cached

    try:
        agent = TechnicalAnalysisAgent(
            symbol=request.symbol,
            days=request.days
        )
        result = agent.get_full_analysis()

        # Cache result
        await cache.set(cache_key, result, settings.CACHE_TTL_TECHNICAL)

        return result

    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.error(f"Technical analysis error for {request.symbol}: {e}")
        raise HTTPException(status_code=500, detail=f"Analysis failed: {str(e)}")

@router.post("/analyze/sentiment")
async def analyze_sentiment(request: SentimentAnalysisRequest):
    """
    Perform sentiment analysis on Vietnamese financial texts.
    (Implemented in Phase 3)
    """
    return {
        "message": "Sentiment analysis endpoint - implemented in Phase 3",
        "texts_received": len(request.texts)
    }

@router.post("/synthesize")
async def synthesize(request: SynthesisRequest):
    """
    Synthesize technical and sentiment analysis for final recommendation.
    (Implemented in Phase 4)
    """
    return {
        "message": "Synthesis endpoint - implemented in Phase 4"
    }
```

### 3.4 Acceptance Criteria

- [ ] `POST /api/analyze/technical` returns valid analysis for VNM, HPG, FPT
- [ ] Invalid symbols (lowercase, >3 chars) rejected with 400 error
- [ ] Redis caching reduces response time on repeated requests
- [ ] All 15 technical indicators calculated correctly
- [ ] Signals generated in Vietnamese language
- [ ] Support/resistance and Fibonacci levels computed

---

## 4. Phase 3: Sentiment Analysis Agent

### 4.1 Objectives

- Implement Vietnamese NLP preprocessing with slang dictionary
- Integrate PhoBERT model for sentiment classification
- Create rumor detection logic
- Build API endpoint

### 4.2 Files to Create

```
agent-service/app/
├── agents/
│   └── sentiment_agent.py
├── utils/
│   └── vietnamese_nlp.py
└── data/
    └── slang_dictionary.json
```

### 4.3 Implementation

#### Step 1: Create Vietnamese NLP utilities

```python
# agent-service/app/utils/vietnamese_nlp.py
"""
Vietnamese NLP utilities for stock market text processing
"""

import re
from typing import Dict, List, Set

class VietnameseNLPProcessor:
    """Preprocessor for Vietnamese stock market text"""

    # Vietnamese stock market slang dictionary
    SLANG_DICT = {
        # Patterns
        'cây thông': 'bullish_pattern',
        'cay thong': 'bullish_pattern',
        'cây súng': 'bearish_pattern',
        'cay sung': 'bearish_pattern',

        # Manipulation
        'múa bên trăng': 'price_manipulation',
        'mua ben trang': 'price_manipulation',
        'lùa gà': 'pump_and_dump',
        'lua ga': 'pump_and_dump',
        'bơm': 'pump_scheme',
        'xả': 'dump_scheme',

        # Investor types
        'con tép': 'retail_investor',
        'con tep': 'retail_investor',
        'cá mập': 'institutional_investor',
        'ca map': 'institutional_investor',

        # Actions
        'chốt lời': 'take_profit',
        'chot loi': 'take_profit',
        'cắt lỗ': 'stop_loss',
        'cat lo': 'stop_loss',
        'bắt đáy': 'bottom_fishing',
        'bat day': 'bottom_fishing',
        'bắt dao rơi': 'catching_falling_knife',

        # Sentiment
        'fomo': 'fear_of_missing_out',
        'all in': 'invest_all_capital',
        'ôm': 'hold_long_term',
        'hốt': 'buy_opportunity',
        'hot': 'buy_opportunity',

        # Trends
        'sideway': 'sideways_trend',
        'breakout': 'price_breakout',
        'tạo đáy': 'forming_bottom',
        'tao day': 'forming_bottom',
        'tạo đỉnh': 'forming_top',
        'tao dinh': 'forming_top',

        # Trading
        'thanh khoản': 'liquidity',
        'thanh khoan': 'liquidity',
        'vốn hóa': 'market_cap',
        'von hoa': 'market_cap',
        'room ngoại': 'foreign_ownership_room',
        'room ngoai': 'foreign_ownership_room',
        'margin call': 'margin_call',
        't+': 'settlement_period'
    }

    # Known valid Vietnamese stock symbols (extend as needed)
    VALID_SYMBOLS: Set[str] = {
        'VNM', 'HPG', 'VCB', 'VHM', 'VIC', 'FPT', 'MSN', 'MBB', 'TCB', 'BID',
        'CTG', 'GAS', 'SAB', 'VRE', 'PLX', 'VPB', 'NVL', 'MWG', 'POW', 'STB',
        'ACB', 'HDB', 'TPB', 'VJC', 'VND', 'SSI', 'HCM', 'BCM', 'PDR', 'DIG',
        'KDH', 'DGC', 'GEX', 'REE', 'VCI', 'DPM', 'DCM', 'PVD', 'PVS', 'GMD'
    }

    # Spam keywords for filtering
    SPAM_KEYWORDS = [
        'khuyến mãi', 'khuyen mai',
        'đăng ký ngay', 'dang ky ngay',
        'group vip', 'nhóm vip', 'nhom vip',
        'bảo lãi', 'bao lai',
        'cam kết lời', 'cam ket loi',
        'liên hệ zalo', 'lien he zalo',
        'inbox nhận', 'inbox nhan',
        'free signal', 'tín hiệu miễn phí'
    ]

    def __init__(self):
        pass

    def preprocess(self, text: str) -> str:
        """Normalize Vietnamese text and replace slang"""
        if not text:
            return ""

        # Lowercase
        text_lower = text.lower()

        # Replace slang terms with normalized versions
        for slang, meaning in self.SLANG_DICT.items():
            text_lower = text_lower.replace(slang, f' {meaning} ')

        # Remove extra whitespace
        text_lower = ' '.join(text_lower.split())

        return text_lower

    def extract_stock_symbols(self, text: str) -> List[str]:
        """Extract Vietnamese stock symbols from text"""
        if not text:
            return []

        # Pattern for 3-letter uppercase symbols
        pattern = r'\b([A-Z]{3})\b'
        matches = re.findall(pattern, text.upper())

        # Filter to valid symbols only
        valid_matches = [m for m in matches if m in self.VALID_SYMBOLS]

        # Remove duplicates while preserving order
        seen = set()
        result = []
        for m in valid_matches:
            if m not in seen:
                seen.add(m)
                result.append(m)

        return result

    def is_spam(self, text: str) -> bool:
        """Check if text is spam"""
        if not text:
            return True

        text_lower = text.lower()

        for keyword in self.SPAM_KEYWORDS:
            if keyword in text_lower:
                return True

        return False

    def count_symbol_mentions(self, texts: List[str]) -> Dict[str, int]:
        """Count mentions of each symbol across texts"""
        counts: Dict[str, int] = {}

        for text in texts:
            symbols = self.extract_stock_symbols(text)
            for symbol in symbols:
                counts[symbol] = counts.get(symbol, 0) + 1

        return dict(sorted(counts.items(), key=lambda x: x[1], reverse=True))

    def add_valid_symbols(self, symbols: List[str]) -> None:
        """Add symbols to valid symbols set"""
        for symbol in symbols:
            if len(symbol) == 3 and symbol.isalpha():
                self.VALID_SYMBOLS.add(symbol.upper())
```

#### Step 2: Create Sentiment Agent

```python
# agent-service/app/agents/sentiment_agent.py
"""
Sentiment Analysis Agent for Vietnamese Stock Market
Uses PhoBERT for Vietnamese NLP
"""

from typing import Dict, List, Any, Optional
import logging
from datetime import datetime

from app.agents.base_agent import BaseAgent
from app.utils.vietnamese_nlp import VietnameseNLPProcessor

class SentimentAgent(BaseAgent):
    """Vietnamese sentiment analysis agent using PhoBERT"""

    def __init__(self, use_gpu: bool = False):
        super().__init__("SentimentAgent")
        self.nlp_processor = VietnameseNLPProcessor()
        self.use_gpu = use_gpu

        # Lazy loading for model
        self._tokenizer = None
        self._model = None
        self._model_loaded = False

    def _load_model(self) -> bool:
        """Load PhoBERT model (lazy loading)"""
        if self._model_loaded:
            return True

        try:
            from transformers import AutoTokenizer, AutoModelForSequenceClassification
            import torch

            model_name = "vinai/phobert-base"

            self.logger.info(f"Loading PhoBERT model: {model_name}")

            self._tokenizer = AutoTokenizer.from_pretrained(model_name)
            self._model = AutoModelForSequenceClassification.from_pretrained(
                model_name,
                num_labels=3  # Positive, Neutral, Negative
            )

            if self.use_gpu and torch.cuda.is_available():
                self._model = self._model.cuda()
                self.logger.info("Using GPU for inference")
            else:
                self.logger.info("Using CPU for inference")

            self._model.eval()
            self._model_loaded = True
            return True

        except ImportError as e:
            self.logger.warning(f"Transformers not installed: {e}")
            return False
        except Exception as e:
            self.logger.error(f"Failed to load model: {e}")
            return False

    async def analyze(self, texts: List[str], symbol: Optional[str] = None) -> Dict[str, Any]:
        """Analyze sentiment for batch of texts"""
        if not texts:
            return self._empty_result(symbol)

        # Filter spam
        filtered_texts = [t for t in texts if not self.nlp_processor.is_spam(t)]

        if not filtered_texts:
            return self._empty_result(symbol, note="All texts filtered as spam")

        # Try loading model
        if self._load_model():
            results = [self._analyze_single_ml(t) for t in filtered_texts]
        else:
            # Fallback to rule-based
            results = [self._analyze_single_rules(t) for t in filtered_texts]

        aggregate = self._compute_aggregate(results)

        return {
            'symbol': symbol,
            'analysis_date': self.get_vietnam_time().strftime('%Y-%m-%d'),
            'articles_analyzed': len(filtered_texts),
            'aggregate': aggregate,
            'individual_results': results[:10]  # Limit for response size
        }

    def _analyze_single_ml(self, text: str) -> Dict:
        """Analyze single text using ML model"""
        import torch

        processed = self.nlp_processor.preprocess(text)

        inputs = self._tokenizer(
            processed,
            return_tensors='pt',
            truncation=True,
            max_length=256,
            padding=True
        )

        if self.use_gpu and torch.cuda.is_available():
            inputs = {k: v.cuda() for k, v in inputs.items()}

        with torch.no_grad():
            outputs = self._model(**inputs)
            probabilities = torch.softmax(outputs.logits, dim=1)

        probs = probabilities[0].cpu().numpy()
        sentiment_map = {0: 'negative', 1: 'neutral', 2: 'positive'}
        predicted = int(probabilities.argmax())

        return {
            'sentiment': sentiment_map[predicted],
            'confidence': float(probs[predicted]),
            'scores': {
                'negative': float(probs[0]),
                'neutral': float(probs[1]),
                'positive': float(probs[2])
            }
        }

    def _analyze_single_rules(self, text: str) -> Dict:
        """Fallback rule-based sentiment analysis"""
        processed = self.nlp_processor.preprocess(text)

        # Simple keyword-based scoring
        positive_keywords = [
            'tăng', 'tang', 'tích cực', 'tich cuc', 'khả quan', 'kha quan',
            'bullish', 'buy', 'mua', 'lợi nhuận', 'loi nhuan', 'tăng trưởng',
            'breakout', 'buy_opportunity', 'bullish_pattern', 'take_profit'
        ]

        negative_keywords = [
            'giảm', 'giam', 'tiêu cực', 'tieu cuc', 'rủi ro', 'rui ro',
            'bearish', 'sell', 'bán', 'lỗ', 'lo', 'sụt giảm',
            'pump_and_dump', 'price_manipulation', 'bearish_pattern', 'stop_loss'
        ]

        pos_count = sum(1 for kw in positive_keywords if kw in processed)
        neg_count = sum(1 for kw in negative_keywords if kw in processed)

        total = pos_count + neg_count + 1  # +1 to avoid division by zero

        if pos_count > neg_count:
            sentiment = 'positive'
            confidence = min(pos_count / total + 0.3, 0.9)
        elif neg_count > pos_count:
            sentiment = 'negative'
            confidence = min(neg_count / total + 0.3, 0.9)
        else:
            sentiment = 'neutral'
            confidence = 0.5

        return {
            'sentiment': sentiment,
            'confidence': round(confidence, 4),
            'scores': {
                'negative': round(neg_count / total, 4),
                'neutral': round(1 - (pos_count + neg_count) / total, 4),
                'positive': round(pos_count / total, 4)
            }
        }

    def _compute_aggregate(self, results: List[Dict]) -> Dict:
        """Compute aggregate sentiment metrics"""
        if not results:
            return {
                'positive_ratio': 0.0,
                'negative_ratio': 0.0,
                'neutral_ratio': 0.0,
                'overall_score': 0.0
            }

        total = len(results)
        positive_count = sum(1 for r in results if r['sentiment'] == 'positive')
        negative_count = sum(1 for r in results if r['sentiment'] == 'negative')
        neutral_count = sum(1 for r in results if r['sentiment'] == 'neutral')

        # Overall score: weighted average of positive - negative
        overall_score = sum(
            r['scores']['positive'] - r['scores']['negative']
            for r in results
        ) / total

        return {
            'positive_ratio': round(positive_count / total, 4),
            'negative_ratio': round(negative_count / total, 4),
            'neutral_ratio': round(neutral_count / total, 4),
            'overall_score': round(overall_score, 4)
        }

    async def detect_rumors(
        self,
        social_texts: List[str],
        official_texts: List[str]
    ) -> Dict[str, Any]:
        """Detect potential rumors by comparing social vs official sources"""
        social_symbols = set()
        official_symbols = set()

        for text in social_texts:
            social_symbols.update(self.nlp_processor.extract_stock_symbols(text))

        for text in official_texts:
            official_symbols.update(self.nlp_processor.extract_stock_symbols(text))

        # Symbols only in social media = potential rumors
        rumor_candidates = social_symbols - official_symbols

        rumors = []
        symbol_counts = self.nlp_processor.count_symbol_mentions(social_texts)

        for symbol in rumor_candidates:
            mentions = symbol_counts.get(symbol, 0)
            rumors.append({
                'symbol': symbol,
                'mentions': mentions,
                'risk_level': 'HIGH' if mentions >= 10 else 'MEDIUM',
                'warning': 'Chưa có xác nhận từ nguồn chính thống'
            })

        # Sort by mentions
        rumors.sort(key=lambda x: x['mentions'], reverse=True)

        return {
            'detected_rumors': rumors,
            'verified_symbols': list(official_symbols),
            'analysis_timestamp': self.get_vietnam_time().isoformat()
        }

    def _empty_result(self, symbol: Optional[str], note: str = None) -> Dict:
        """Return empty result structure"""
        return {
            'symbol': symbol,
            'analysis_date': self.get_vietnam_time().strftime('%Y-%m-%d'),
            'articles_analyzed': 0,
            'aggregate': {
                'positive_ratio': 0.0,
                'negative_ratio': 0.0,
                'neutral_ratio': 0.0,
                'overall_score': 0.0
            },
            'note': note or 'No texts provided'
        }
```

#### Step 3: Update analysis routes for sentiment

```python
# Update agent-service/app/api/routes/analysis.py
# Add to existing file:

from app.agents.sentiment_agent import SentimentAgent

# ... existing code ...

@router.post("/analyze/sentiment")
async def analyze_sentiment(
    request: SentimentAnalysisRequest,
    cache: CacheService = Depends(get_cache_service)
):
    """
    Perform sentiment analysis on Vietnamese financial texts.
    Uses PhoBERT model with fallback to rule-based analysis.
    """
    cache_key = f"sentiment:{hash(tuple(request.texts[:5]))}:{request.symbol}"

    cached = await cache.get(cache_key)
    if cached:
        return cached

    try:
        agent = SentimentAgent(use_gpu=False)
        result = await agent.analyze(
            texts=request.texts,
            symbol=request.symbol
        )

        await cache.set(cache_key, result, settings.CACHE_TTL_SENTIMENT)

        return result

    except Exception as e:
        logger.error(f"Sentiment analysis error: {e}")
        raise HTTPException(status_code=500, detail=f"Analysis failed: {str(e)}")
```

### 4.4 Acceptance Criteria

- [ ] `POST /api/analyze/sentiment` returns valid sentiment scores
- [ ] Vietnamese slang properly normalized
- [ ] PhoBERT model loads correctly (or graceful fallback to rules)
- [ ] Batch processing handles 100+ texts
- [ ] Spam texts filtered out
- [ ] Rumor detection identifies social-only symbols

---

## 5. Phase 4: Forecast Agent & Master Orchestrator

### 5.1 Objectives

- Create Forecast Agent with weighted synthesis
- Build Master Agent for orchestration
- Implement daily report generation
- Create report templates

### 5.2 Files to Create

```
agent-service/app/
├── agents/
│   ├── forecast_agent.py
│   └── master_agent.py
├── services/
│   ├── report_service.py
│   └── storage_service.py
└── templates/
    ├── daily_report.md.jinja2
    └── daily_report.html.jinja2
```

### 5.3 Implementation

Refer to the full code in `vnstock-analysis-architecture.md` sections 6.1-7.3 for complete implementation of:

1. **ForecastAgent**: Multi-signal synthesis with 40/30/30 weighting
2. **MasterAgent**: Orchestrates all agents, generates daily reports
3. **ReportService**: Generates Markdown/HTML/JSON reports
4. **StorageService**: S3 integration for data persistence

### 5.4 Acceptance Criteria

- [ ] Master Agent coordinates all sub-agents
- [ ] Daily report includes market overview, hot stocks, recommendations
- [ ] Alerts generated for HIGH risk or STRONG signals
- [ ] Reports saved to S3 in JSON, Markdown, HTML formats
- [ ] Legal disclaimer included in all reports

---

## 6. Phase 5: Web Dashboard

### 6.1 Objectives

- Create Next.js web dashboard
- Implement real-time updates
- Build Vietnamese UI components

### 6.2 Directory Structure

```
web-dashboard/
├── src/
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   ├── dashboard/page.tsx
│   │   └── stocks/[symbol]/page.tsx
│   ├── components/
│   │   ├── MarketOverview.tsx
│   │   ├── StockCard.tsx
│   │   └── RecommendationBadge.tsx
│   └── lib/
│       └── api.ts
├── Dockerfile
├── package.json
└── next.config.js
```

### 6.3 Key Components

Refer to `vnstock-analysis-architecture.md` sections 8.1-8.4 for component implementations.

### 6.4 Acceptance Criteria

- [ ] Dashboard loads within 3 seconds
- [ ] Stock cards display recommendations with color coding
- [ ] Vietnamese language throughout UI
- [ ] Responsive design for mobile/tablet/desktop

---

## 7. Phase 6: Telegram Bot

### 7.1 Objectives

- Create Telegram bot for alerts
- Implement command handlers
- Set up channel notifications

### 7.2 Directory Structure

```
telegram-bot/
├── bot/
│   ├── main.py
│   └── handlers/
│       ├── start.py
│       ├── analyze.py
│       └── subscribe.py
├── Dockerfile
└── requirements.txt
```

### 7.3 Acceptance Criteria

- [ ] /start displays Vietnamese help
- [ ] /analyze returns formatted analysis
- [ ] /subscribe manages watchlist
- [ ] Alerts pushed to configured channels

---

## 8. Phase 7: Testing & Documentation

### 8.1 Test Structure

```
tests/
├── unit/
│   ├── test_technical_agent.py
│   ├── test_sentiment_agent.py
│   └── test_vietnamese_nlp.py
├── integration/
│   ├── test_api_endpoints.py
│   └── test_database.py
└── conftest.py
```

### 8.2 Example Tests

```python
# tests/unit/test_technical_agent.py
import pytest
from app.agents.technical_agent import TechnicalAnalysisAgent

class TestTechnicalAgent:
    def test_valid_symbol_accepted(self):
        agent = TechnicalAnalysisAgent(symbol='VNM', days=90)
        assert agent.symbol == 'VNM'

    def test_invalid_symbol_rejected(self):
        with pytest.raises(ValueError):
            TechnicalAnalysisAgent(symbol='invalid', days=90)

    def test_days_clamped(self):
        agent = TechnicalAnalysisAgent(symbol='VNM', days=1000)
        assert agent.days == 365
```

### 8.3 Acceptance Criteria

- [ ] Unit test coverage > 80%
- [ ] Integration tests pass for all endpoints
- [ ] OpenAPI documentation complete
- [ ] Performance benchmarks documented

---

## 9. Deployment Checklist

### 9.1 Pre-Deployment

```
[ ] All tests passing
[ ] Environment variables configured
[ ] Database migrations ready
[ ] SSL certificates obtained
[ ] Monitoring dashboards configured
[ ] Legal disclaimer verified
[ ] Rate limiting configured
```

### 9.2 Deployment Commands

```bash
# 1. Configure environment
cp .env.example .env
# Edit .env with production values

# 2. Build and start
docker-compose --profile production up -d --build

# 3. Verify health
curl http://localhost:8000/health

# 4. Import n8n workflow
# Access n8n at http://localhost:5678
# Import n8n-workflow-daily-analysis.json

# 5. Verify Telegram bot
# Send /start to your bot
```

### 9.3 Post-Deployment

```
[ ] All services healthy (docker-compose ps)
[ ] API endpoints responding
[ ] n8n workflow scheduled
[ ] Telegram bot responding
[ ] First daily report generated
[ ] Monitoring alerts configured
```

---

## Appendix: Quick Reference

### Common Commands

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f agent-service

# Restart service
docker-compose restart agent-service

# Run tests
docker-compose exec agent-service pytest

# Access database
docker-compose exec postgres psql -U admin -d vnstock
```

### API Endpoints Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /health | Health check |
| POST | /api/analyze/technical | Technical analysis |
| POST | /api/analyze/sentiment | Sentiment analysis |
| POST | /api/synthesize | Combined forecast |
| GET | /api/reports/daily | Daily report |
| GET | /api/stocks/{symbol} | Stock analysis |

---

**End of Implementation Guide**
