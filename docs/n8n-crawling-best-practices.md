# n8n Data Crawling Best Practices Guide
## VN Stock Analysis System

---

## Table of Contents

1. [Overview](#1-overview)
2. [Rate Limiting](#2-rate-limiting)
3. [Error Handling & Retry Logic](#3-error-handling--retry-logic)
4. [Data Deduplication](#4-data-deduplication)
5. [Spam Filtering](#5-spam-filtering)
6. [Monitoring & Alerting](#6-monitoring--alerting)
7. [Anti-Blocking Techniques](#7-anti-blocking-techniques)
8. [Workflow Configuration](#8-workflow-configuration)
9. [Troubleshooting](#9-troubleshooting)

---

## 1. Overview

This guide covers the best practices implemented in the VN Stock Analysis n8n workflows for reliable, efficient data crawling from Vietnamese financial news sources.

### Key Principles

| Principle | Implementation |
|-----------|----------------|
| **Be Polite** | 2-second delays between requests |
| **Be Resilient** | Retry with exponential backoff |
| **Be Smart** | Deduplicate data before processing |
| **Be Vigilant** | Filter spam and validate data |
| **Be Aware** | Monitor and alert on failures |

### Workflows

| Workflow | Purpose | File |
|----------|---------|------|
| Daily Analysis | Main data crawling and analysis | `n8n-workflow-daily-analysis.json` |
| Error Handler | Centralized error handling | `n8n-workflow-error-handler.json` |

---

## 2. Rate Limiting

### Why Rate Limiting?

- Avoid getting blocked by source websites
- Reduce server load on data providers
- Stay within API quotas
- Maintain good reputation

### Implementation

#### Wait Nodes Between Requests

```
┌─────────────┐     2s      ┌─────────────┐     2s      ┌─────────────┐
│ VnEconomy   │ ──────────► │   CafeF     │ ──────────► │ VietStock   │
└─────────────┘   delay     └─────────────┘   delay     └─────────────┘
```

#### Configuration (.env)

```bash
CRAWL_DELAY_RSS_MS=2000      # 2 seconds between RSS fetches
CRAWL_DELAY_API_MS=2000      # 2 seconds between API calls
CRAWL_DELAY_SOCIAL_MS=6000   # 6 seconds for social media
CRAWL_MAX_CONCURRENT_SOURCES=2
```

#### n8n Node: Rate Limit Wait

```json
{
  "name": "Rate Limit Wait (2s)",
  "type": "n8n-nodes-base.wait",
  "parameters": {
    "amount": 2000,
    "unit": "milliseconds"
  }
}
```

### Best Practices

1. **Use Split In Batches** - Process one source at a time
2. **Add Wait nodes** - Between each HTTP request
3. **Set timeouts** - 10s for RSS, 30s for APIs, 60s for analysis
4. **Continue on fail** - Don't stop workflow on single failure

---

## 3. Error Handling & Retry Logic

### 3-Layer Error Handling

```
Layer 1: HTTP Request Level
├── Timeout: 10-60 seconds
├── continueOnFail: true
└── Graceful degradation

Layer 2: Workflow Level
├── Retry logic with exponential backoff
├── Max 3 retries per request
└── Error classification

Layer 3: System Level
├── Error Handler Workflow
├── Telegram alerts
├── Dead Letter Queue
└── Retry Queue
```

### Retry Logic Code

```javascript
const MAX_RETRIES = 3;
const BACKOFF_MS = [1000, 2000, 4000]; // exponential

if (error && retryCount < MAX_RETRIES) {
  // Schedule retry
  return {
    needsRetry: true,
    retryDelayMs: BACKOFF_MS[retryCount]
  };
}
```

### Error Classification

| Error Type | Severity | Retryable | Retry Delay |
|------------|----------|-----------|-------------|
| TIMEOUT | Warning | Yes | 5 minutes |
| RATE_LIMIT | Warning | Yes | 1 hour |
| CONNECTION | High | Yes | 10 minutes |
| PARSE_ERROR | Medium | No | - |
| AUTH_ERROR | Critical | No | - |
| SERVER_ERROR | High | Yes | 10 minutes |

### Error Handler Workflow

The `n8n-workflow-error-handler.json` workflow:

1. **Classifies** errors by type and severity
2. **Logs** to Redis with daily keys
3. **Increments** error counters for metrics
4. **Alerts** via Telegram for critical/high severity
5. **Queues** for retry or dead letter

---

## 4. Data Deduplication

### Why Deduplicate?

- Same news appears on multiple sources
- RSS feeds may return previously seen items
- Avoid duplicate processing and storage

### Redis-Based Deduplication

```
┌──────────────┐     MD5      ┌─────────────┐
│ New Article  │ ───────────► │ Generate    │
│              │              │ Hash Key    │
└──────────────┘              └──────┬──────┘
                                     │
                              ┌──────▼──────┐
                              │ Redis GET   │
                              │ dedup:news: │
                              └──────┬──────┘
                                     │
                  ┌──────────────────┼──────────────────┐
                  │ Key not found    │                  │ Key found
                  ▼                  │                  ▼
           ┌──────────────┐         │         ┌──────────────┐
           │ Process      │         │         │ Skip         │
           │ Set Key+TTL  │         │         │ (Duplicate)  │
           └──────────────┘         │         └──────────────┘
```

### Hash Generation

```javascript
function generateDedupHash(article) {
  const content = `${article.source}:${article.title}:${article.pubDate}`;
  return crypto.createHash('md5').update(content).digest('hex');
}

// Redis key format: dedup:news:{hash}
// TTL: 7 days (604800 seconds)
```

### Configuration (.env)

```bash
DEDUP_ENABLED=true
DEDUP_TTL_SECONDS=604800  # 7 days
DEDUP_KEY_PREFIX=dedup:news
```

---

## 5. Spam Filtering

### Multi-Stage Validation

```
Stage 1: Schema Validation
├── Required: title, link
├── Title: 10-500 characters
└── Date: not future, not > 7 days old

Stage 2: Spam Keyword Filter
├── Vietnamese promotional terms
├── Scam indicators
└── Low-quality content markers

Stage 3: Pattern Matching
├── "lợi nhuận X%" patterns
├── "tăng Xx trong" patterns
├── VIP group mentions
└── Zalo phone numbers
```

### Vietnamese Spam Keywords

```javascript
const spamKeywords = [
  // Promotional
  'khuyến mại', 'đăng ký ngay', 'cơ hội vàng', 'click ngay',
  'nhận quà', 'giảm giá sốc', 'miễn phí 100%',

  // Scam indicators
  'lùa gà', 'pump', 'group vip', 'room kín',
  'đánh lên', 'chắc thắng', 'x2 x3',

  // Low quality
  'theo dõi kênh', 'subscribe', 'like share'
];
```

### Spam Patterns (Regex)

```javascript
const spamPatterns = [
  /lợi nhuận \d+%/i,           // "lợi nhuận 50%"
  /tăng \d+x trong/i,          // "tăng 5x trong 1 tuần"
  /nhóm (vip|kín|riêng)/i,     // VIP group mentions
  /zalo:?\s*\d{10,11}/i        // Zalo phone numbers
];
```

---

## 6. Monitoring & Alerting

### Redis Metrics Keys

```
metrics:articles:{date}:{source}    # Article count per source
metrics:duplicates:{date}           # Duplicate count
metrics:errors:{type}:{date}        # Error count by type
crawl:lastSuccess:daily             # Last successful crawl time
```

### Alert Rules

| Alert | Condition | Severity | Channel |
|-------|-----------|----------|---------|
| CrawlFailure | success_rate < 80% for 10m | Warning | Telegram |
| SourceDown | source failed for 30m | Critical | Telegram + Email |
| NoNewArticles | 0 articles for 2h | Warning | Telegram |
| HighErrorRate | error_rate > 5% | Warning | Telegram |

### Telegram Alert Format

```
🚨 *CRITICAL ERROR*

*Workflow:* Vietnamese Stock Market Daily Analysis
*Node:* Fetch RSS Feed
*Type:* TIMEOUT
*Time:* 2026-01-29T08:05:23Z

*Error:* Request timed out after 10000ms

⚠️ Immediate attention required!
```

---

## 7. Anti-Blocking Techniques

### User-Agent Rotation

```bash
USER_AGENTS=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36,Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36
```

### Request Headers

```json
{
  "User-Agent": "Mozilla/5.0 (Windows...)",
  "Accept": "application/rss+xml, application/xml, text/xml, */*",
  "Accept-Language": "vi-VN,vi;q=0.9,en;q=0.8",
  "Accept-Encoding": "gzip, deflate, br"
}
```

### Timing Randomization

Add random jitter to delays:

```javascript
const baseDelay = 2000; // 2 seconds
const jitter = Math.random() * 1000; // 0-1 second random
const delay = baseDelay + jitter;
```

### Respect robots.txt

- Check robots.txt before crawling new sources
- Honor Crawl-delay directives
- Avoid pages marked as disallowed

---

## 8. Workflow Configuration

### Import Workflow

1. Open n8n (http://localhost:5678)
2. Go to Workflows → Import from File
3. Select `n8n-workflow-daily-analysis.json`
4. Configure credentials:
   - AWS S3
   - Redis
   - Telegram Bot

### Required Credentials

| Credential | ID | Purpose |
|------------|-----|---------|
| AWS Account | 1 | S3 storage |
| Telegram Bot | 2 | Notifications |
| Redis | 3 | Caching, dedup, metrics |

### Workflow Settings

```json
{
  "settings": {
    "executionOrder": "v1",
    "errorWorkflow": "n8n-workflow-error-handler",
    "timezone": "Asia/Ho_Chi_Minh",
    "saveManualExecutions": true,
    "saveDataErrorExecution": "all",
    "saveDataSuccessExecution": "all"
  }
}
```

### Schedule

- **Trigger:** 8:00 AM weekdays (Vietnam time)
- **Cron:** `0 8 * * 1-5`
- **Timezone:** Asia/Ho_Chi_Minh

---

## 9. Troubleshooting

### Common Issues

#### 1. RSS Fetch Timeout

**Symptom:** ETIMEDOUT errors

**Solution:**
- Increase timeout (10s → 15s)
- Check source website availability
- Add retry logic (already implemented)

#### 2. Too Many Duplicates

**Symptom:** High duplicate count in metrics

**Solution:**
- Check Redis connection
- Verify dedup hash generation
- Ensure TTL is appropriate (7 days)

#### 3. Spam Bypassing Filters

**Symptom:** Promotional content in results

**Solution:**
- Add new spam keywords to list
- Add regex patterns for new spam types
- Review source reliability scores

#### 4. Agent Service Errors

**Symptom:** 500 errors from agent-service

**Solution:**
- Check agent-service container logs
- Verify service is running: `docker-compose ps`
- Check database connection

### Debug Commands

```bash
# Check Redis metrics
redis-cli KEYS "metrics:*"
redis-cli GET "crawl:lastSuccess:daily"

# Check error logs
redis-cli LRANGE "errors:TIMEOUT:$(date +%Y-%m-%d)" 0 -1

# Check n8n logs
docker-compose logs -f n8n

# Manual workflow trigger
curl -X POST http://localhost:5678/webhook/manual-trigger
```

### Redis Key Reference

| Key Pattern | Description | TTL |
|-------------|-------------|-----|
| `dedup:news:{hash}` | Article deduplication | 7 days |
| `metrics:articles:{date}:{source}` | Article count | 30 days |
| `metrics:duplicates:{date}` | Duplicate count | 30 days |
| `metrics:errors:{type}:{date}` | Error count | 30 days |
| `crawl:lastSuccess:daily` | Last crawl time | 2 days |
| `errors:{type}:{date}` | Error details list | 7 days |
| `retry:queue` | Retry queue | No TTL |
| `deadletter:queue` | Failed items | No TTL |

---

## Summary

The enhanced n8n workflows implement:

| Feature | Status | Benefit |
|---------|--------|---------|
| Rate Limiting | ✅ 2s delays | Avoid blocks |
| Retry Logic | ✅ 3 attempts, exponential backoff | Handle transient failures |
| Deduplication | ✅ Redis-based, 7-day TTL | No duplicate processing |
| Spam Filtering | ✅ Keywords + patterns | Clean data |
| Error Handling | ✅ Dedicated workflow | Quick response |
| Monitoring | ✅ Redis metrics | Visibility |
| Alerting | ✅ Telegram | Immediate notification |

**Files Modified:**
- `n8n-workflow-daily-analysis.json` - Enhanced with best practices
- `n8n-workflow-error-handler.json` - New error handling workflow
- `.env.example` - New crawling configuration
- `docs/n8n-crawling-best-practices.md` - This documentation
