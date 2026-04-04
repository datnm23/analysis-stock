# VN Stock Analysis System - API Documentation

This document provides comprehensive API documentation for the VN Stock Analysis System.

## Base URLs

- **Go API Gateway**: `http://localhost:8080`
- **Python Sentiment Service**: `http://localhost:8000`

## Authentication

All API endpoints require authentication except for health checks. See [Authentication Guide](authentication.md) for details.

### Authentication Methods

1. **JWT Token** (for user authentication)
2. **API Key** (for service-to-service communication)

## Endpoints

### Health Check

#### GET `/health`

Health check endpoint - no authentication required.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-29T10:30:00Z"
}
```

#### GET `/ready`

Readiness check endpoint - no authentication required.

---

### Technical Analysis

#### GET `/api/v1/technical/:symbol`

Get technical analysis for a single stock symbol.

**Parameters:**
- `symbol` (path): Stock symbol (e.g., VNM, HPG)

**Response:**
```json
{
  "symbol": "VNM",
  "price": 85000,
  "indicators": {
    "rsi": 65.5,
    "macd": 1200,
    "bollinger": {
      "upper": 88000,
      "middle": 85000,
      "lower": 82000
    },
    "sma": {
      "sma20": 84500,
      "sma50": 83000,
      "sma200": 81000
    }
  },
  "signals": {
    "recommendation": "HOLD",
    "confidence": 0.75
  }
}
```

#### POST `/api/v1/technical/batch`

Get technical analysis for multiple stock symbols.

**Request Body:**
```json
{
  "symbols": ["VNM", "HPG", "VCB"],
  "indicators": ["rsi", "macd", "bollinger"]
}
```

**Response:**
```json
{
  "results": [
    {
      "symbol": "VNM",
      "indicators": {...},
      "signals": {...}
    }
  ]
}
```

---

### Sentiment Analysis

#### POST `/api/v1/sentiment`

Analyze sentiment of Vietnamese news text.

**Headers:**
- `X-API-Key`: API key for authentication (when enabled)

**Request Body:**
```json
{
  "text": "Cổ phiếu VNM tăng mạnh nhờ kết quả kinh doanh khả quan",
  "language": "vi"
}
```

**Response:**
```json
{
  "sentiment": "positive",
  "confidence": 75.0,
  "symbols": ["VNM"],
  "keywords": ["tăng", "lợi nhuận"]
}
```

#### POST `/api/v1/analyze/batch`

Analyze sentiment for multiple texts.

**Request Body:**
```json
{
  "texts": [
    "Cổ phiếu VNM tăng mạnh",
    "Thị trường giảm điểm"
  ]
}
```

**Response:**
```json
{
  "results": [
    {
      "sentiment": "positive",
      "confidence": 75.0,
      "symbols": ["VNM"],
      "keywords": []
    }
  ]
}
```

---

### Combined Analysis

#### POST `/api/v1/analyze`

Get combined technical and sentiment analysis.

**Request Body:**
```json
{
  "symbol": "VNM",
  "include_sentiment": true
}
```

**Response:**
```json
{
  "symbol": "VNM",
  "technical": {
    "recommendation": "BUY",
    "confidence": 0.8,
    "indicators": {...}
  },
  "sentiment": {
    "sentiment": "positive",
    "confidence": 75.0
  },
  "combined": {
    "recommendation": "BUY",
    "confidence": 0.77
  }
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request parameters",
  "code": "BAD_REQUEST"
}
```

### 401 Unauthorized
```json
{
  "error": "Invalid or missing authentication token",
  "code": "UNAUTHORIZED"
}
```

### 403 Forbidden
```json
{
  "error": "Insufficient permissions",
  "code": "FORBIDDEN"
}
```

### 429 Too Many Requests
```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMITED",
  "retry_after": 60
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error",
  "code": "INTERNAL_ERROR"
}
```

---

## Rate Limits

| Endpoint | Limit |
|----------|-------|
| Technical Analysis | 100 requests/minute |
| Sentiment Analysis | 200 requests/minute |
| Combined Analysis | 50 requests/minute |

---

## WebSocket (Future)

Real-time updates will be available via WebSocket at `ws://localhost:8080/ws`
