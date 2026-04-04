"""Health check router."""

import logging
from fastapi import APIRouter, Request

from app.config import get_settings

logger = logging.getLogger(__name__)
router = APIRouter(tags=["health"])


@router.get("/health")
async def health(request: Request):
    """Health check endpoint."""
    settings = get_settings()
    redis_ok = False
    firecrawl_ok = False

    # Check Redis
    redis_client = getattr(request.app.state, "redis_client", None)
    if redis_client:
        try:
            await redis_client.ping()
            redis_ok = True
        except Exception:
            pass

    # Check Firecrawl
    firecrawl = getattr(request.app.state, "firecrawl", None)
    if firecrawl and settings.enable_firecrawl:
        firecrawl_ok = await firecrawl.health_check()

    return {
        "status": "ok",
        "service": "crawl-agent",
        "redis": "connected" if redis_ok else "disconnected",
        "firecrawl": "healthy" if firecrawl_ok else "unreachable",
        "features": {
            "firecrawl_enabled": settings.enable_firecrawl,
            "telegram_enabled": settings.enable_telegram,
            "source_chasing_enabled": settings.enable_source_chasing,
        },
    }
