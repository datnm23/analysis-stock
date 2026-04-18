---
phase: 3
title: "Article Generator (crawl-agent)"
status: completed
effort: 2.5h
completed: 2026-04-18
---

# Phase 3: Article Generator (crawl-agent)

## Context Links
- Phase trước: [phase-02-go-api.md](./phase-02-go-api.md)
- Scheduler: `crawl-agent/app/scheduler.py`
- Config: `crawl-agent/app/config.py`
- NewsPublisher (tham khảo pattern): `crawl-agent/app/services/news_publisher.py`
- Redis keys: `news:{symbol}:recent` (list, max 20 items)

## Overview

Thêm `ArticleGenerator` service vào crawl-agent:
1. Sau mỗi scrape cycle, lấy top N hot symbols (symbols có nhiều news nhất trong Redis)
2. Với mỗi symbol: fetch ForecastResult từ go-services + news từ Redis
3. Gọi Claude API (Haiku) để rewrite thành bài viết
4. POST draft lên go-services `/api/v1/articles`

## Related Code Files

**Tạo mới:**
- `crawl-agent/app/services/article_generator.py`

**Sửa:**
- `crawl-agent/app/config.py` — thêm 5 settings
- `crawl-agent/app/scheduler.py` — gọi ArticleGenerator sau publish_batch

## Implementation Steps

### 1. Thêm settings vào `crawl-agent/app/config.py`

```python
# Article generation settings
anthropic_api_key: str = ""
article_max_daily: int = 10
article_hot_symbols_count: int = 10
article_claude_model: str = "claude-haiku-4-5-20251001"
go_services_internal_url: str = "http://localhost:8080"
go_services_internal_key: str = ""
```

### 2. Tạo `crawl-agent/app/services/article_generator.py`

```python
"""
Generates stock analysis articles by combining scraped news with forecast data.

Pipeline: hot symbols (Redis) → forecast (go-services) → Claude rewrite → POST draft
"""
import json
import logging
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional

import httpx

logger = logging.getLogger(__name__)

CLAUDE_API_URL = "https://api.anthropic.com/v1/messages"

ARTICLE_PROMPT_TEMPLATE = """Bạn là chuyên gia phân tích chứng khoán Việt Nam. Viết bài phân tích chuyên sâu về cổ phiếu {symbol}.

DỮ LIỆU PHÂN TÍCH KỸ THUẬT:
- Khuyến nghị: {recommendation}
- Độ tin cậy: {confidence:.0%}
- RSI: {rsi}
- MACD: {macd}
- Điểm kỹ thuật: {technical_score:.2f}/1.0
- Điểm sentiment: {sentiment_score:.2f}/1.0
- Lý do phân tích: {reasoning}

TIN TỨC GẦN ĐÂY ({news_count} bài):
{news_list}

YÊU CẦU:
- Viết bài ~400-500 từ bằng tiếng Việt, định dạng Markdown
- Cấu trúc: ## Tóm tắt | ## Phân tích kỹ thuật | ## Diễn biến thị trường | ## Kết luận
- Chỉ phân tích từ dữ liệu trên, không tự suy diễn thêm
- Kết luận phải nhất quán với khuyến nghị: {recommendation}
- Thêm disclaimer cuối bài: *Bài viết chỉ mang tính tham khảo, không phải lời khuyên đầu tư.*"""


class ArticleGenerator:
    def __init__(
        self,
        redis_client,
        anthropic_api_key: str,
        go_services_url: str,
        go_services_key: str,
        claude_model: str = "claude-haiku-4-5-20251001",
        max_daily: int = 10,
        hot_symbols_count: int = 10,
    ):
        self._redis = redis_client
        self._api_key = anthropic_api_key
        self._go_url = go_services_url.rstrip("/")
        self._go_key = go_services_key
        self._model = claude_model
        self._max_daily = max_daily
        self._hot_symbols_count = hot_symbols_count

    async def get_hot_symbols(self) -> List[str]:
        """Return top N symbols by Redis list length (most recent news = hottest)."""
        try:
            keys = await self._redis.keys("news:*:recent")
            symbol_lengths = []
            for key in keys:
                key_str = key.decode() if isinstance(key, bytes) else key
                parts = key_str.split(":")
                if len(parts) == 3:
                    symbol = parts[1]
                    length = await self._redis.llen(key_str)
                    symbol_lengths.append((symbol, length))
            symbol_lengths.sort(key=lambda x: x[1], reverse=True)
            return [s for s, _ in symbol_lengths[: self._hot_symbols_count]]
        except Exception as exc:
            logger.warning("Failed to get hot symbols: %s", exc)
            return []

    async def _fetch_forecast(self, symbol: str) -> Optional[Dict[str, Any]]:
        try:
            async with httpx.AsyncClient(timeout=10) as client:
                resp = await client.get(f"{self._go_url}/api/v1/forecast/{symbol}")
                if resp.status_code == 200:
                    return resp.json()
        except Exception as exc:
            logger.warning("Failed to fetch forecast for %s: %s", symbol, exc)
        return None

    async def _fetch_news_from_redis(self, symbol: str, limit: int = 5) -> List[Dict]:
        try:
            raw_list = await self._redis.lrange(f"news:{symbol}:recent", 0, limit - 1)
            items = []
            for raw in raw_list:
                try:
                    items.append(json.loads(raw))
                except Exception:
                    continue
            return items
        except Exception as exc:
            logger.warning("Failed to fetch news for %s: %s", symbol, exc)
            return []

    def _build_prompt(self, symbol: str, forecast: Dict, news: List[Dict]) -> str:
        reasoning_str = "; ".join(forecast.get("reasoning", []))
        news_lines = []
        for i, n in enumerate(news, 1):
            title = n.get("title", "")
            source = n.get("source", "unknown")
            news_lines.append(f"{i}. [{source}] {title}")
        news_list = "\n".join(news_lines) if news_lines else "Không có tin tức."

        indicators = forecast.get("indicators", {})
        return ARTICLE_PROMPT_TEMPLATE.format(
            symbol=symbol,
            recommendation=forecast.get("recommendation", "HOLD"),
            confidence=forecast.get("confidence", 0),
            rsi=indicators.get("rsi", "N/A"),
            macd=indicators.get("macd", "N/A"),
            technical_score=forecast.get("technical_score", 0),
            sentiment_score=forecast.get("sentiment_score", 0),
            reasoning=reasoning_str or "Không có dữ liệu.",
            news_count=len(news),
            news_list=news_list,
        )

    async def _call_claude(self, prompt: str) -> Optional[str]:
        if not self._api_key:
            logger.warning("ANTHROPIC_API_KEY not set, skipping article generation")
            return None
        try:
            async with httpx.AsyncClient(timeout=30) as client:
                resp = await client.post(
                    CLAUDE_API_URL,
                    headers={
                        "x-api-key": self._api_key,
                        "anthropic-version": "2023-06-01",
                        "content-type": "application/json",
                    },
                    json={
                        "model": self._model,
                        "max_tokens": 1024,
                        "messages": [{"role": "user", "content": prompt}],
                    },
                )
                if resp.status_code == 200:
                    data = resp.json()
                    return data["content"][0]["text"]
                logger.warning("Claude API error %d: %s", resp.status_code, resp.text[:200])
        except Exception as exc:
            logger.warning("Claude API call failed: %s", exc)
        return None

    async def _post_article(self, symbol: str, content: str, forecast: Dict, source_urls: List[str]) -> bool:
        title = f"Phân tích {symbol} ngày {datetime.now(timezone.utc).strftime('%d/%m/%Y')}"
        summary = content[:300].split("\n")[0].lstrip("#").strip()
        payload = {
            "symbol": symbol,
            "title": title,
            "content": content,
            "summary": summary,
            "source_urls": source_urls,
            "forecast_data": json.dumps(forecast, ensure_ascii=False).encode(),
        }
        try:
            async with httpx.AsyncClient(timeout=10) as client:
                resp = await client.post(
                    f"{self._go_url}/api/v1/articles",
                    json=payload,
                    headers={"X-Internal-Key": self._go_key},
                )
                return resp.status_code == 201
        except Exception as exc:
            logger.warning("Failed to post article for %s: %s", symbol, exc)
        return False

    async def generate_for_symbol(self, symbol: str) -> bool:
        """Generate one article draft for a symbol. Returns True if successful."""
        forecast = await self._fetch_forecast(symbol)
        if not forecast:
            logger.info("No forecast for %s, skipping", symbol)
            return False

        news = await self._fetch_news_from_redis(symbol)
        prompt = self._build_prompt(symbol, forecast, news)
        content = await self._call_claude(prompt)
        if not content:
            return False

        source_urls = [n.get("url", "") for n in news if n.get("url")]
        success = await self._post_article(symbol, content, forecast, source_urls)
        if success:
            logger.info("Article draft created for %s", symbol)
        return success

    async def run(self) -> Dict[str, int]:
        """Generate articles for top hot symbols. Returns stats."""
        hot = await self.get_hot_symbols()
        generated = 0
        for symbol in hot[: self._max_daily]:
            if await self.generate_for_symbol(symbol):
                generated += 1
        return {"symbols_processed": len(hot), "articles_created": generated}
```

### 3. Thêm `anthropic` (không cần SDK, dùng httpx trực tiếp)

`requirements.txt` — `httpx` đã có sẵn trong crawl-agent. **Không cần thêm dependency.**

### 4. Sửa `crawl-agent/app/scheduler.py` — gọi ArticleGenerator

```python
# Trong __init__, sau _news_publisher
from app.services.article_generator import ArticleGenerator

self._article_generator = (
    ArticleGenerator(
        redis_client=redis_client,
        anthropic_api_key=settings.anthropic_api_key,
        go_services_url=settings.go_services_internal_url,
        go_services_key=settings.go_services_internal_key,
        claude_model=settings.article_claude_model,
        max_daily=settings.article_max_daily,
        hot_symbols_count=settings.article_hot_symbols_count,
    )
    if redis_client and settings.anthropic_api_key
    else None
)
```

Trong `_run_scrape()`, sau block publish Redis:

```python
# Generate article drafts for hot symbols
if self._article_generator:
    article_stats = await self._article_generator.run()
    stats["articles_generated"] = article_stats
    logger.info(
        "Generated %d article drafts from %d hot symbols",
        article_stats["articles_created"],
        article_stats["symbols_processed"],
    )
```

## Todo List

- [ ] Thêm 6 settings vào `crawl-agent/app/config.py`
- [ ] Tạo `crawl-agent/app/services/article_generator.py`
- [ ] Sửa `ScrapeScheduler.__init__()` — khởi tạo `_article_generator`
- [ ] Thêm generate call trong `_run_scrape()` sau publish block
- [ ] Thêm env vars vào `.env.example`

## Environment Variables

```env
ANTHROPIC_API_KEY=sk-ant-api03-...
ARTICLE_MAX_DAILY=10
ARTICLE_HOT_SYMBOLS_COUNT=10
ARTICLE_CLAUDE_MODEL=claude-haiku-4-5-20251001
GO_SERVICES_INTERNAL_URL=http://go-services:8080
GO_SERVICES_INTERNAL_KEY=change-me-secret-key
```

## Success Criteria

- Sau scheduler run, có article drafts trong PostgreSQL (status=draft)
- `ANTHROPIC_API_KEY` không set → skip gracefully, không crash
- Log: "Generated N article drafts from M hot symbols"
- stats response có field `articles_generated`

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Claude API rate limit | Low | Max 10 calls/day, sequential (không parallel) |
| go-services down | Medium | `_fetch_forecast()` trả None → skip symbol |
| Vietnamese → slug rỗng | Medium | Slug generator ở Phase 2 handle |
| API cost | Low | Haiku ~$0.00025/1K tokens, 10 articles/ngày ≈ $0.05/ngày |
