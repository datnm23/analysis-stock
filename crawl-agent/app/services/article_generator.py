"""
Generates stock analysis articles by combining scraped news with forecast data.

Pipeline per symbol:
  hot symbols (Redis LLEN) → GET /forecast/{symbol} → news from Redis
  → LLM rewrite (Claude/Gemini via config) → optional image generation
  → POST /api/v1/articles (draft) → Telegram notify admin
"""

import asyncio
import json
import logging
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional

import httpx

from .image_pipeline import ImagePipeline
from .llm_client import LLMClient

logger = logging.getLogger(__name__)

_PROMPT_TEMPLATE = """\
Bạn là chuyên gia phân tích chứng khoán Việt Nam với hơn 10 năm kinh nghiệm. Viết bài phân tích chi tiết về cổ phiếu {symbol}.

DỮ LIỆU PHÂN TÍCH KỸ THUẬT:
- Khuyến nghị: {recommendation} | Độ tin cậy: {confidence:.0%}
- Điểm kỹ thuật: {technical_score:.2f}/1.0 | Điểm sentiment: {sentiment_score:.2f}/1.0
- Nhận định: {reasoning}

TIN TỨC GẦN ĐÂY ({news_count} bài):
{news_list}

YÊU CẦU BÀI VIẾT:
- Độ dài 700-900 từ, tiếng Việt chuyên nghiệp, định dạng Markdown
- Cấu trúc bắt buộc:
  ## Tổng quan cổ phiếu
  (2-3 câu tóm tắt tình hình hiện tại và lý do quan tâm)

  ## Phân tích kỹ thuật
  (Giải thích chi tiết các chỉ số RSI, MACD, đường hỗ trợ/kháng cự, xu hướng ngắn và trung hạn)

  ## Phân tích cơ bản & Tin tức
  (Tổng hợp các tin tức quan trọng, đánh giá tác động đến giá cổ phiếu)

  ## Rủi ro cần lưu ý
  (Liệt kê 2-3 rủi ro chính bằng danh sách gạch đầu dòng)

  ## Kết luận & Khuyến nghị
  (Kết luận rõ ràng nhất quán với khuyến nghị **{recommendation}**, mức giá mục tiêu nếu có, điều kiện vào/ra lệnh)

- KHÔNG bịa đặt số liệu cụ thể nếu không có trong dữ liệu
- Kết luận phải nhất quán với khuyến nghị: **{recommendation}**
- Dòng cuối cùng: *Bài viết chỉ mang tính tham khảo, không phải lời khuyên đầu tư.*\
"""

_TITLE_PROMPT_TEMPLATE = """\
Tạo tiêu đề bài viết phân tích chứng khoán hấp dẫn và chuyên nghiệp cho cổ phiếu {symbol}.

Thông tin:
- Khuyến nghị: {recommendation}
- Độ tin cậy: {confidence:.0%}
- Nhận định chính: {reasoning}
- Tin tức nổi bật: {top_news}

Yêu cầu tiêu đề:
- Ngắn gọn 10-15 từ, tiếng Việt
- Phải thể hiện được khuyến nghị {recommendation} hoặc tín hiệu kỹ thuật
- Có tên mã cổ phiếu {symbol}
- Mang tính thông tin, thu hút (ví dụ: "VNM tạo đáy kỹ thuật, RSI quá bán — cơ hội mua vào?" hoặc "HPG bứt phá kháng cự, tín hiệu BUY rõ ràng")
- Chỉ trả về tiêu đề, không giải thích thêm\
"""


def _inject_image_into_content(content: str, image_url: str, symbol: str) -> str:
    """Insert image markdown between Tổng quan and Phân tích kỹ thuật sections."""
    img_md = f"\n![{symbol} - Biểu đồ kỹ thuật]({image_url})\n"
    lines = content.split("\n")
    h2_indices = [i for i, line in enumerate(lines) if line.startswith("## ")]
    if len(h2_indices) >= 2:
        lines.insert(h2_indices[1], img_md)
        return "\n".join(lines)
    return content + img_md


def _inject_chart_into_content(content: str, chart_url: str, symbol: str) -> str:
    """Insert stock price/volume chart right after the ## Phân tích kỹ thuật heading."""
    chart_md = (
        f"\n![{symbol} — Biểu đồ giá & khối lượng giao dịch 90 ngày]({chart_url})\n"
        f"*Nguồn: KB Securities | SMA20 (xanh) · SMA50 (cam)*\n"
    )
    lines = content.split("\n")
    for i, line in enumerate(lines):
        if line.strip().startswith("## ") and "kỹ thuật" in line.lower():
            lines.insert(i + 1, chart_md)
            return "\n".join(lines)
    # Fallback: insert after second ## heading
    h2_indices = [i for i, line in enumerate(lines) if line.startswith("## ")]
    if h2_indices:
        lines.insert(h2_indices[0] + 1, chart_md)
        return "\n".join(lines)
    return content + chart_md


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
        # Multi-model text generation
        article_model: str = "auto",
        gemini_api_key: str = "",
        gemini_text_model: str = "gemini-2.0-flash",
        xai_api_key: str = "",
        grok_model: str = "grok-3",
        # Image generation via Vertex AI
        enable_image_generation: bool = False,
        vertex_credentials_json: str = "",
        vertex_project_id: str = "",
        vertex_location: str = "us-central1",
        vertex_image_model: str = "imagen-3.0-generate-001",
        image_storage_backend: str = "s3",
        s3_endpoint: str = "",
        s3_bucket: str = "blog-images",
        s3_access_key: str = "",
        s3_secret_key: str = "",
        s3_public_url: str = "",
        gdrive_credentials_json: str = "",
        gdrive_folder_id: str = "",
    ):
        self._redis = redis_client
        self._go_url = go_services_url.rstrip("/")
        self._go_key = go_services_key
        self._telegram_token = telegram_bot_token
        self._telegram_chat_id = telegram_chat_id
        self._dashboard_url = dashboard_url
        self._max_daily = max_daily
        self._hot_symbols_count = hot_symbols_count

        self._llm = LLMClient(
            anthropic_api_key=anthropic_api_key,
            gemini_api_key=gemini_api_key,
            xai_api_key=xai_api_key,
            claude_model=claude_model,
            gemini_text_model=gemini_text_model,
            grok_model=grok_model,
            article_model=article_model,
        )
        self._images: Optional[ImagePipeline] = (
            ImagePipeline(
                anthropic_api_key=anthropic_api_key,
                claude_model=claude_model,
                vertex_credentials_json=vertex_credentials_json,
                vertex_project_id=vertex_project_id,
                vertex_location=vertex_location,
                vertex_image_model=vertex_image_model,
                storage_backend=image_storage_backend,
                s3_endpoint=s3_endpoint,
                s3_bucket=s3_bucket,
                s3_access_key=s3_access_key,
                s3_secret_key=s3_secret_key,
                s3_public_url=s3_public_url,
                gdrive_credentials_json=gdrive_credentials_json,
                gdrive_folder_id=gdrive_folder_id,
            )
            if enable_image_generation
            else None
        )

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

    async def _generate_title(self, symbol: str, forecast: Dict, news: List[Dict]) -> str:
        """Generate a descriptive headline via LLM; fall back to generic title."""
        reasoning = "; ".join(forecast.get("reasoning", [])) or "Không có dữ liệu"
        top_news = news[0].get("title", "") if news else "Không có tin tức"
        prompt = _TITLE_PROMPT_TEMPLATE.format(
            symbol=symbol,
            recommendation=forecast.get("recommendation", "HOLD"),
            confidence=forecast.get("confidence", 0),
            reasoning=reasoning,
            top_news=top_news,
        )
        title = await self._llm.call_llm(prompt)
        if title:
            title = title.strip().strip('"').strip("'").strip('*').strip()
        if not title or len(title) > 120:
            date_str = datetime.now(timezone.utc).strftime("%d/%m/%Y")
            title = f"{symbol}: Phân tích kỹ thuật và khuyến nghị {forecast.get('recommendation', 'HOLD')} — {date_str}"
        return title

    async def _post_article(
        self,
        symbol: str,
        title: str,
        content: str,
        forecast: Dict,
        source_urls: List[str],
        image_url: Optional[str] = None,
    ) -> bool:
        summary = next(
            (line.lstrip("#").strip() for line in content.splitlines() if line.strip()),
            content[:200],
        )
        payload: Dict[str, Any] = {
            "symbol": symbol,
            "title": title,
            "content": content,
            "summary": summary[:300],
            "source_urls": source_urls,
            "forecast_data": json.dumps(forecast, ensure_ascii=False),
        }
        if image_url:
            payload["image_url"] = image_url
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
        from .chart_generator import generate_chart

        forecast = await self._fetch_forecast(symbol)
        if not forecast:
            return None

        news = await self._fetch_news(symbol)
        date_str = datetime.now(timezone.utc).strftime("%Y%m%d%H%M")

        # Generate title, content, and stock chart in parallel
        title, content, chart_url = await asyncio.gather(
            self._generate_title(symbol, forecast, news),
            self._llm.call_llm(self._build_prompt(symbol, forecast, news)),
            generate_chart(symbol, date_str),
        )
        if not content:
            return None
        source_urls = [n.get("url", "") for n in news if n.get("url")]

        # Inject stock chart after "## Phân tích kỹ thuật" section
        if chart_url:
            content = _inject_chart_into_content(content, chart_url, symbol)

        # Hero thumbnail: Vertex AI abstract image (if enabled), else use chart
        image_url: Optional[str] = chart_url
        if self._images:
            summary = next(
                (line.lstrip("#").strip() for line in content.splitlines() if line.strip()),
                content[:200],
            )
            ai_image_url = await self._images.generate_and_upload(symbol, summary, date_str)
            if ai_image_url:
                image_url = ai_image_url  # AI image as hero thumbnail

        success = await self._post_article(symbol, title, content, forecast, source_urls, image_url)
        if success:
            logger.info(
                "Article draft created: %s (chart=%s, hero=%s)",
                symbol, bool(chart_url), bool(image_url),
            )
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
        try:
            async with httpx.AsyncClient(timeout=10) as client:
                await client.post(
                    f"https://api.telegram.org/bot{self._telegram_token}/sendMessage",
                    json={"chat_id": self._telegram_chat_id, "text": "\n".join(lines)},
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
