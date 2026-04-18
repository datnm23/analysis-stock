"""
Generates stock analysis articles by combining scraped news with forecast data.

Pipeline per symbol:
  hot symbols (Redis LLEN) → GET /forecast/{symbol} → news from Redis
  → Claude Haiku rewrite → POST /api/v1/articles (draft)
  → Telegram notify admin
"""

import json
import logging
import os
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional

import httpx

logger = logging.getLogger(__name__)

_CLAUDE_API_URL = "https://api.anthropic.com/v1/messages"
_ANTHROPIC_VERSION = "2023-06-01"

_PROMPT_TEMPLATE = """\
Bạn là chuyên gia phân tích chứng khoán Việt Nam. Viết bài phân tích về cổ phiếu {symbol}.

DỮ LIỆU PHÂN TÍCH KỸ THUẬT:
- Khuyến nghị: {recommendation}
- Độ tin cậy: {confidence:.0%}
- Điểm kỹ thuật: {technical_score:.2f}/1.0
- Điểm sentiment: {sentiment_score:.2f}/1.0
- Lý do: {reasoning}

TIN TỨC GẦN ĐÂY ({news_count} bài):
{news_list}

YÊU CẦU:
- Bài ~400-500 từ, tiếng Việt, định dạng Markdown
- Cấu trúc: ## Tóm tắt | ## Phân tích kỹ thuật | ## Diễn biến thị trường | ## Kết luận
- Chỉ phân tích từ dữ liệu trên, kết luận phải nhất quán với khuyến nghị: {recommendation}
- Cuối bài thêm: *Bài viết chỉ mang tính tham khảo, không phải lời khuyên đầu tư.*\
"""


class ArticleGenerator:
    def __init__(
        self,
        redis_client,
        anthropic_api_key: str,
        go_services_url: str,
        go_services_key: str,
        telegram_bot_token: str = "",
        telegram_chat_id: str = "",
        dashboard_url: str = "",
        claude_model: str = "claude-haiku-4-5-20251001",
        max_daily: int = 10,
        hot_symbols_count: int = 10,
    ):
        self._redis = redis_client
        self._api_key = anthropic_api_key
        self._go_url = go_services_url.rstrip("/")
        self._go_key = go_services_key
        self._telegram_token = telegram_bot_token
        self._telegram_chat_id = telegram_chat_id
        self._dashboard_url = dashboard_url
        self._model = claude_model
        self._max_daily = max_daily
        self._hot_symbols_count = hot_symbols_count

    async def get_hot_symbols(self) -> List[str]:
        """Return symbols sorted by Redis list length (most news = hottest)."""
        try:
            keys = await self._redis.keys("news:*:recent")
            pairs = []
            for key in keys:
                key_str = key.decode() if isinstance(key, bytes) else key
                parts = key_str.split(":")
                if len(parts) == 3:
                    length = await self._redis.llen(key_str)
                    pairs.append((parts[1], length))
            pairs.sort(key=lambda x: x[1], reverse=True)
            return [sym for sym, _ in pairs[: self._hot_symbols_count]]
        except Exception as exc:
            logger.warning("get_hot_symbols failed: %s", exc)
            return []

    async def _fetch_forecast(self, symbol: str) -> Optional[Dict[str, Any]]:
        try:
            async with httpx.AsyncClient(timeout=10) as client:
                resp = await client.get(f"{self._go_url}/api/v1/forecast/{symbol}")
                if resp.status_code == 200:
                    return resp.json()
        except Exception as exc:
            logger.warning("forecast fetch failed for %s: %s", symbol, exc)
        return None

    async def _fetch_news(self, symbol: str, limit: int = 5) -> List[Dict]:
        try:
            raw_list = await self._redis.lrange(f"news:{symbol}:recent", 0, limit - 1)
            items = []
            for raw in raw_list:
                try:
                    items.append(json.loads(raw))
                except Exception:
                    pass
            return items
        except Exception as exc:
            logger.warning("news fetch failed for %s: %s", symbol, exc)
            return []

    def _build_prompt(self, symbol: str, forecast: Dict, news: List[Dict]) -> str:
        reasoning = "; ".join(forecast.get("reasoning", [])) or "Không có dữ liệu."
        news_lines = [
            f"{i}. [{n.get('source', 'unknown')}] {n.get('title', '')}"
            for i, n in enumerate(news, 1)
        ]
        return _PROMPT_TEMPLATE.format(
            symbol=symbol,
            recommendation=forecast.get("recommendation", "HOLD"),
            confidence=forecast.get("confidence", 0),
            technical_score=forecast.get("technical_score", 0),
            sentiment_score=forecast.get("sentiment_score", 0),
            reasoning=reasoning,
            news_count=len(news),
            news_list="\n".join(news_lines) if news_lines else "Không có tin tức.",
        )

    async def _call_claude(self, prompt: str) -> Optional[str]:
        if not self._api_key:
            logger.warning("ANTHROPIC_API_KEY not set, skipping article generation")
            return None
        try:
            async with httpx.AsyncClient(timeout=30) as client:
                resp = await client.post(
                    _CLAUDE_API_URL,
                    headers={
                        "x-api-key": self._api_key,
                        "anthropic-version": _ANTHROPIC_VERSION,
                        "content-type": "application/json",
                    },
                    json={
                        "model": self._model,
                        "max_tokens": 1024,
                        "messages": [{"role": "user", "content": prompt}],
                    },
                )
                if resp.status_code == 200:
                    return resp.json()["content"][0]["text"]
                logger.warning("Claude API %d: %s", resp.status_code, resp.text[:200])
        except Exception as exc:
            logger.warning("Claude API call failed: %s", exc)
        return None

    async def _post_article(
        self, symbol: str, title: str, content: str, forecast: Dict, source_urls: List[str]
    ) -> bool:
        summary = next(
            (line.lstrip("#").strip() for line in content.splitlines() if line.strip()),
            content[:200],
        )
        payload = {
            "symbol": symbol,
            "title": title,
            "content": content,
            "summary": summary[:300],
            "source_urls": source_urls,
            "forecast_data": json.dumps(forecast, ensure_ascii=False),
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
            logger.warning("post article failed for %s: %s", symbol, exc)
        return False

    async def generate_for_symbol(self, symbol: str) -> Optional[Dict[str, str]]:
        """Generate one draft article. Returns {symbol, title} on success, None on failure."""
        forecast = await self._fetch_forecast(symbol)
        if not forecast:
            return None

        news = await self._fetch_news(symbol)
        prompt = self._build_prompt(symbol, forecast, news)
        content = await self._call_claude(prompt)
        if not content:
            return None

        title = f"Phân tích {symbol} ngày {datetime.now(timezone.utc).strftime('%d/%m/%Y')}"
        source_urls = [n.get("url", "") for n in news if n.get("url")]

        if await self._post_article(symbol, title, content, forecast, source_urls):
            logger.info("Article draft created: %s", symbol)
            return {"symbol": symbol, "title": title}
        return None

    async def _notify_telegram(self, created: List[Dict[str, str]]) -> None:
        if not self._telegram_token or not self._telegram_chat_id or not created:
            return
        lines = [f"\U0001f4dd {len(created)} bài viết draft mới:\n"]
        for a in created:
            lines.append(f"\u2022 {a['symbol']} \u2014 {a['title']}")
        if self._dashboard_url:
            lines.append(f"\nDuyệt tại: {self._dashboard_url}/admin/articles")
        text = "\n".join(lines)
        try:
            async with httpx.AsyncClient(timeout=10) as client:
                await client.post(
                    f"https://api.telegram.org/bot{self._telegram_token}/sendMessage",
                    json={"chat_id": self._telegram_chat_id, "text": text},
                )
        except Exception as exc:
            logger.warning("Telegram notify failed: %s", exc)

    async def run(self) -> Dict[str, int]:
        """Generate articles for top hot symbols. Returns generation stats."""
        hot = await self.get_hot_symbols()
        created: List[Dict[str, str]] = []
        for symbol in hot[: self._max_daily]:
            result = await self.generate_for_symbol(symbol)
            if result:
                created.append(result)

        await self._notify_telegram(created)
        return {"symbols_processed": len(hot), "articles_created": len(created)}
