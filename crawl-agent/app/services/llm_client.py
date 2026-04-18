"""LLM client supporting Claude (Anthropic) and Gemini (Google) via httpx."""
import logging
from typing import Optional

import httpx

logger = logging.getLogger(__name__)

_CLAUDE_URL = "https://api.anthropic.com/v1/messages"
_ANTHROPIC_VERSION = "2023-06-01"
_GEMINI_BASE = "https://generativelanguage.googleapis.com/v1beta/models"


class LLMClient:
    def __init__(
        self,
        anthropic_api_key: str = "",
        gemini_api_key: str = "",
        claude_model: str = "claude-haiku-4-5-20251001",
        gemini_text_model: str = "gemini-2.0-flash",
        article_model: str = "claude",
    ):
        self._anthropic_key = anthropic_api_key
        self._gemini_key = gemini_api_key
        self._claude_model = claude_model
        self._gemini_text_model = gemini_text_model
        self._article_model = article_model  # claude | gemini | auto

    async def call_claude(self, prompt: str, max_tokens: int = 1024) -> Optional[str]:
        if not self._anthropic_key:
            logger.warning("ANTHROPIC_API_KEY not set")
            return None
        try:
            async with httpx.AsyncClient(timeout=30) as client:
                resp = await client.post(
                    _CLAUDE_URL,
                    headers={
                        "x-api-key": self._anthropic_key,
                        "anthropic-version": _ANTHROPIC_VERSION,
                        "content-type": "application/json",
                    },
                    json={
                        "model": self._claude_model,
                        "max_tokens": max_tokens,
                        "messages": [{"role": "user", "content": prompt}],
                    },
                )
                if resp.status_code == 200:
                    return resp.json()["content"][0]["text"]
                logger.warning("Claude API %d: %s", resp.status_code, resp.text[:200])
        except Exception as exc:
            logger.warning("Claude call failed: %s", exc)
        return None

    async def call_gemini(self, prompt: str) -> Optional[str]:
        if not self._gemini_key:
            logger.warning("GEMINI_API_KEY not set")
            return None
        url = f"{_GEMINI_BASE}/{self._gemini_text_model}:generateContent"
        try:
            async with httpx.AsyncClient(timeout=30) as client:
                resp = await client.post(
                    url,
                    headers={"x-goog-api-key": self._gemini_key},
                    json={"contents": [{"parts": [{"text": prompt}]}]},
                )
                if resp.status_code == 200:
                    candidates = resp.json().get("candidates", [])
                    if not candidates:
                        logger.warning("Gemini returned no candidates (content filtered?)")
                        return None
                    parts = candidates[0].get("content", {}).get("parts", [])
                    return parts[0].get("text") if parts else None
                logger.warning("Gemini API %d", resp.status_code)
        except Exception as exc:
            logger.warning("Gemini call failed: %s", exc)
        return None

    async def call_llm(self, prompt: str) -> Optional[str]:
        """Route to Claude, Gemini, or auto-fallback based on config."""
        if self._article_model == "gemini":
            return await self.call_gemini(prompt)
        if self._article_model == "claude":
            return await self.call_claude(prompt)
        # auto: Claude primary, Gemini fallback
        result = await self.call_claude(prompt)
        if result is None:
            result = await self.call_gemini(prompt)
        return result
