"""Image generation via Gemini + upload to S3/MinIO."""
import asyncio
import base64
import logging
import re
from typing import Optional

import httpx

logger = logging.getLogger(__name__)

_CLAUDE_URL = "https://api.anthropic.com/v1/messages"
_ANTHROPIC_VERSION = "2023-06-01"
_GEMINI_BASE = "https://generativelanguage.googleapis.com/v1beta/models"

_IMAGE_PROMPT_INSTRUCTIONS = """\
Tạo một câu mô tả ảnh minh họa ngắn gọn cho bài phân tích cổ phiếu {symbol}.
Phong cách: tối giản, màu xanh tài chính, biểu đồ abstract không hiển thị số liệu thực.
Chỉ trả về mô tả ảnh bằng tiếng Anh, tối đa 2 câu, không giải thích thêm.

Tóm tắt bài: {summary}
"""


class ImagePipeline:
    def __init__(
        self,
        gemini_api_key: str = "",
        gemini_image_model: str = "gemini-2.0-flash-preview-image-generation",
        anthropic_api_key: str = "",
        claude_model: str = "claude-haiku-4-5-20251001",
        s3_endpoint: str = "",
        s3_bucket: str = "blog-images",
        s3_access_key: str = "",
        s3_secret_key: str = "",
        s3_public_url: str = "",
    ):
        self._gemini_key = gemini_api_key
        self._gemini_image_model = gemini_image_model
        self._anthropic_key = anthropic_api_key
        self._claude_model = claude_model
        self._s3_endpoint = s3_endpoint
        self._s3_bucket = s3_bucket
        self._s3_access_key = s3_access_key
        self._s3_secret_key = s3_secret_key
        self._s3_public_url = s3_public_url.rstrip("/")

    async def build_image_prompt(self, symbol: str, summary: str) -> str:
        """Use Claude Haiku to generate a concise English image prompt."""
        if not self._anthropic_key:
            return f"Abstract stock market chart for {symbol}, minimalist blue financial style, no numbers"
        prompt = _IMAGE_PROMPT_INSTRUCTIONS.format(symbol=symbol, summary=summary[:300])
        try:
            async with httpx.AsyncClient(timeout=15) as client:
                resp = await client.post(
                    _CLAUDE_URL,
                    headers={
                        "x-api-key": self._anthropic_key,
                        "anthropic-version": _ANTHROPIC_VERSION,
                        "content-type": "application/json",
                    },
                    json={
                        "model": self._claude_model,
                        "max_tokens": 120,
                        "messages": [{"role": "user", "content": prompt}],
                    },
                )
                if resp.status_code == 200:
                    return resp.json()["content"][0]["text"].strip()
        except Exception as exc:
            logger.warning("image prompt generation failed: %s", exc)
        return f"Abstract financial chart for {symbol} stock analysis, blue gradient, minimalist"

    async def generate_image(self, image_prompt: str) -> Optional[bytes]:
        """Generate PNG bytes via Gemini image generation model."""
        if not self._gemini_key:
            logger.warning("GEMINI_API_KEY not set, skipping image generation")
            return None
        url = f"{_GEMINI_BASE}/{self._gemini_image_model}:generateContent"
        try:
            async with httpx.AsyncClient(timeout=60) as client:
                resp = await client.post(
                    url,
                    headers={"x-goog-api-key": self._gemini_key},
                    json={
                        "contents": [{"parts": [{"text": image_prompt}]}],
                        "generationConfig": {"responseModalities": ["IMAGE", "TEXT"]},
                    },
                )
                if resp.status_code == 200:
                    candidates = resp.json().get("candidates", [])
                    if not candidates:
                        logger.warning("Gemini image: no candidates returned")
                        return None
                    for part in candidates[0].get("content", {}).get("parts", []):
                        if "inlineData" in part:
                            return base64.b64decode(part["inlineData"]["data"])
                logger.warning("Gemini image API %d", resp.status_code)
        except Exception as exc:
            logger.warning("Gemini image generation failed: %s", exc)
        return None

    async def upload_image(self, key: str, image_bytes: bytes) -> Optional[str]:
        """Upload PNG bytes to S3/MinIO, return public URL."""
        if not self._s3_endpoint or not self._s3_access_key:
            logger.warning("S3 not configured, skipping image upload")
            return None
        try:
            import boto3
            from botocore.client import Config

            def _do_upload() -> None:
                s3 = boto3.client(
                    "s3",
                    endpoint_url=self._s3_endpoint,
                    aws_access_key_id=self._s3_access_key,
                    aws_secret_access_key=self._s3_secret_key,
                    config=Config(signature_version="s3v4"),
                )
                s3.put_object(
                    Bucket=self._s3_bucket,
                    Key=key,
                    Body=image_bytes,
                    ContentType="image/png",
                )

            await asyncio.to_thread(_do_upload)
            base_url = self._s3_public_url or f"{self._s3_endpoint}/{self._s3_bucket}"
            return f"{base_url}/{key}"
        except Exception as exc:
            logger.warning("S3 upload failed: %s", exc)
        return None

    @staticmethod
    def _safe_symbol(symbol: str) -> str:
        """Strip anything that's not alphanumeric to prevent path traversal in S3 keys."""
        return re.sub(r"[^a-zA-Z0-9]", "", symbol)[:10]

    async def generate_and_upload(self, symbol: str, summary: str, date_str: str) -> Optional[str]:
        """Full pipeline: Claude builds prompt → Gemini generates image → upload → URL."""
        image_prompt = await self.build_image_prompt(symbol, summary)
        image_bytes = await self.generate_image(image_prompt)
        if not image_bytes:
            return None
        safe_sym = self._safe_symbol(symbol)
        return await self.upload_image(f"articles/{safe_sym}/{date_str}.png", image_bytes)
