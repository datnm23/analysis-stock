"""
Crawl Agent — FastAPI application.

Multi-tier Vietnamese financial news aggregation service:
- Tier 1: RSS + aiohttp (VnExpress, CafeF)
- Tier 2: Firecrawl self-hosted (DanTri, ThanhNien, TuoiTre)
- Tier 3: Telegram channels

Post-processing: SimHash dedup → Symbol detection → Source scoring.
"""

import logging
import sys
from contextlib import asynccontextmanager

import redis.asyncio as redis
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import get_settings

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)],
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Startup and shutdown events."""
    settings = get_settings()

    # Initialize Redis client
    redis_client = None
    try:
        redis_client = redis.Redis(
            host=settings.redis_host,
            port=settings.redis_port,
            password=settings.redis_password if settings.redis_password else None,
            db=settings.redis_db,
            decode_responses=False,
        )
        await redis_client.ping()
        logger.info("Redis connected: %s:%d", settings.redis_host, settings.redis_port)
    except Exception as e:
        logger.warning("Redis not available: %s", e)
        redis_client = None

    app.state.redis_client = redis_client

    # Initialize dedup filter
    from app.services.dedup_filter import DedupFilter
    app.state.dedup_filter = DedupFilter(redis_client=redis_client)

    # Initialize Firecrawl client
    from app.scrapers.firecrawl_client import FirecrawlClient
    app.state.firecrawl = FirecrawlClient(
        base_url=settings.firecrawl_api_url,
        timeout=settings.firecrawl_timeout,
    )
    if settings.enable_firecrawl:
        is_healthy = await app.state.firecrawl.health_check()
        logger.info("Firecrawl status: %s", "healthy" if is_healthy else "unreachable")

    # Start scheduler (if not debug mode)
    scheduler = None
    if not settings.debug:
        from app.scheduler import get_scheduler
        scheduler = get_scheduler(redis_client=redis_client)
        scheduler.start()

    logger.info("Crawl agent ready on port %d", settings.port)

    yield

    # Cleanup
    if scheduler:
        scheduler.stop()
    if redis_client:
        await redis_client.close()
    logger.info("Crawl agent shutting down")


def create_app() -> FastAPI:
    settings = get_settings()

    app = FastAPI(
        title="VN Stock Crawl Agent",
        description="Multi-tier Vietnamese financial news aggregation",
        version="1.0.0",
        lifespan=lifespan,
    )

    # CORS
    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.cors_origins,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # Routers
    from app.routers import scrape, health
    app.include_router(scrape.router)
    app.include_router(health.router)

    return app


app = create_app()

if __name__ == "__main__":
    import uvicorn
    settings = get_settings()
    uvicorn.run(
        "app.main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
    )
