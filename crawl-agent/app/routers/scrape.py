"""
Scrape router — HTTP endpoints for triggering crawl operations.

Endpoints:
- POST /scrape/rss        — Trigger RSS feed scraping (Tier 1)
- POST /scrape/web        — Trigger web crawling via aiohttp (Tier 1)
- POST /scrape/firecrawl  — Trigger Firecrawl scraping (Tier 2)
- POST /scrape/telegram   — Trigger Telegram scraping (Tier 3)
- POST /scrape/all        — Unified: all tiers combined
- POST /scrape/pipeline   — Full pipeline: scrape → dedup → sentiment
- POST /scrape/chase      — Source chasing: follow links to primary sources
- GET  /scheduler/status  — Scheduler status
- POST /scheduler/start   — Start scheduler
- POST /scheduler/stop    — Stop scheduler
"""

import asyncio
import logging
import time
from typing import List, Optional

from fastapi import APIRouter, HTTPException, Request
from pydantic import BaseModel, Field

from app.config import get_settings
from app.services.source_scorer import SourceScorer

logger = logging.getLogger(__name__)
router = APIRouter(tags=["scrape"])

_source_scorer = SourceScorer()


# ---------------------------------------------------------------------------
# Request / Response models
# ---------------------------------------------------------------------------

class FirecrawlRequest(BaseModel):
    urls: List[str] = Field(..., min_length=1, max_length=20)
    max_concurrent: int = Field(default=3, ge=1, le=10)


class TelegramScrapeRequest(BaseModel):
    channels: Optional[List[str]] = None
    limit_per_channel: int = Field(default=50, ge=1, le=200)
    hours_back: int = Field(default=24, ge=1, le=168)


# ---------------------------------------------------------------------------
# Tier 1: RSS Feeds
# ---------------------------------------------------------------------------

@router.post("/scrape/rss")
async def scrape_rss():
    """Scrape all configured Vietnamese financial RSS feeds."""
    from app.scrapers.rss_scraper import scrape_all_feeds

    items = await scrape_all_feeds()
    return {
        "tier": 1,
        "source": "rss",
        "items": [item.to_dict() for item in items],
        "total": len(items),
    }


# ---------------------------------------------------------------------------
# Tier 1: Web Crawl (aiohttp + BS4)
# ---------------------------------------------------------------------------

@router.post("/scrape/web")
async def scrape_web():
    """Crawl VnExpress + CafeF via aiohttp (Tier 1 static HTML)."""
    from app.scrapers.web_scraper import scrape_all_websites

    articles = await scrape_all_websites()

    by_source = {}
    for a in articles:
        by_source.setdefault(a.source, 0)
        by_source[a.source] += 1

    return {
        "tier": 1,
        "source": "web_crawl",
        "articles": [a.to_dict() for a in articles],
        "total": len(articles),
        "by_source": by_source,
    }


# ---------------------------------------------------------------------------
# Tier 2: Firecrawl (JS-heavy sites)
# ---------------------------------------------------------------------------

TIER2_URLS = [
    "https://dantri.com.vn/kinh-doanh/chung-khoan.htm",
    "https://dantri.com.vn/kinh-doanh.htm",
    "https://thanhnien.vn/tai-chinh-kinh-doanh/chung-khoan.htm",
    "https://tuoitre.vn/kinh-doanh.htm",
]


@router.post("/scrape/firecrawl")
async def scrape_firecrawl(request: Request, body: Optional[FirecrawlRequest] = None):
    """Scrape JS-heavy sites via self-hosted Firecrawl (Tier 2).

    Returns clean Markdown content for each URL.
    """
    settings = get_settings()
    if not settings.enable_firecrawl:
        return {"error": "Firecrawl not enabled", "hint": "Set ENABLE_FIRECRAWL=true"}

    firecrawl = request.app.state.firecrawl
    urls = body.urls if body else TIER2_URLS
    max_concurrent = body.max_concurrent if body else 3

    results = await firecrawl.scrape_batch(urls, max_concurrent=max_concurrent)

    return {
        "tier": 2,
        "source": "firecrawl",
        "results": [r.to_dict() for r in results],
        "total": len(results),
        "success_count": sum(1 for r in results if r.success),
    }


# ---------------------------------------------------------------------------
# Tier 3: Telegram
# ---------------------------------------------------------------------------

@router.post("/scrape/telegram")
async def scrape_telegram(body: Optional[TelegramScrapeRequest] = None):
    """Scrape Telegram channels (Tier 3)."""
    settings = get_settings()

    if not settings.telegram_app_api_id or not settings.telegram_app_api_hash:
        return {
            "error": "Telegram API credentials not configured",
            "hint": "Set TELEGRAM_APP_API_ID and TELEGRAM_APP_API_HASH",
        }

    from app.scrapers.telegram_scraper import TelegramScraper

    channels = (body.channels if body and body.channels
                else settings.telegram_channel_list)

    scraper = TelegramScraper(
        api_id=settings.telegram_app_api_id,
        api_hash=settings.telegram_app_api_hash,
        session_name=settings.telegram_session_name,
        channels=channels,
    )

    try:
        messages = await scraper.scrape_channels(
            limit_per_channel=body.limit_per_channel if body else 50,
            hours_back=body.hours_back if body else 24,
        )
        return {
            "tier": 3,
            "source": "telegram",
            "messages": [msg.to_dict() for msg in messages],
            "total": len(messages),
            "channels_scraped": channels,
        }
    finally:
        await scraper.close()


# ---------------------------------------------------------------------------
# Unified: All Tiers
# ---------------------------------------------------------------------------

@router.post("/scrape/all")
async def scrape_all_sources(request: Request):
    """Aggregate news from ALL tiers in one call.

    Pipeline:
    1. Tier 1: RSS feeds + VnExpress/CafeF web crawl
    2. Tier 2: DanTri/ThanhNien/TuoiTre via Firecrawl
    3. Tier 3: Telegram channels
    4. Cross-source dedup
    5. Symbol extraction + Source scoring
    """
    settings = get_settings()
    start = time.time()
    all_items = []
    stats = {}
    seen_hashes = set()

    def _add_items(items: list, source_key: str):
        added = 0
        for item in items:
            h = item.get("id") or item.get("content_hash", "")
            if h and h not in seen_hashes:
                seen_hashes.add(h)
                # Add source weight
                source = item.get("source", "unknown")
                item["source_weight"] = round(_source_scorer.get_weight(source), 2)
                all_items.append(item)
                added += 1
        stats[source_key] = added

    # Tier 1: RSS
    try:
        from app.scrapers.rss_scraper import scrape_all_feeds
        rss_items = await scrape_all_feeds(delay_between=1.0)
        _add_items([item.to_dict() for item in rss_items], "rss")
    except Exception as e:
        stats["rss_error"] = str(e)
        logger.error("RSS error: %s", e)

    # Tier 1: Web crawl (aiohttp)
    try:
        from app.scrapers.web_scraper import scrape_all_websites
        web_articles = await scrape_all_websites(delay_between=2.0)
        _add_items([a.to_dict() for a in web_articles], "web_crawl")
    except Exception as e:
        stats["web_crawl_error"] = str(e)
        logger.error("Web crawl error: %s", e)

    # Tier 2: Firecrawl
    if settings.enable_firecrawl:
        try:
            firecrawl = request.app.state.firecrawl
            fc_results = await firecrawl.scrape_batch(TIER2_URLS, max_concurrent=2)
            fc_items = []
            for r in fc_results:
                if r.success and r.markdown:
                    from app.scrapers.symbol_detector import extract_symbols
                    symbols = extract_symbols(f"{r.title} {r.markdown[:500]}")
                    fc_items.append({
                        "id": r.url,
                        "content": r.markdown[:1000],  # Truncate for API response
                        "title": r.title,
                        "source": _infer_source(r.url),
                        "link": r.url,
                        "symbols": symbols,
                        "markdown_length": len(r.markdown),
                        "has_full_content": True,
                    })
            _add_items(fc_items, "firecrawl")
        except Exception as e:
            stats["firecrawl_error"] = str(e)
            logger.error("Firecrawl error: %s", e)

    # Tier 3: Telegram
    if settings.enable_telegram and settings.telegram_app_api_id:
        try:
            from app.scrapers.telegram_scraper import TelegramScraper
            scraper = TelegramScraper(
                api_id=settings.telegram_app_api_id,
                api_hash=settings.telegram_app_api_hash,
                session_name=settings.telegram_session_name,
                channels=settings.telegram_channel_list,
            )
            messages = await scraper.scrape_channels(
                limit_per_channel=30, hours_back=24,
            )
            _add_items([msg.to_dict() for msg in messages], "telegram")
            await scraper.close()
        except Exception as e:
            stats["telegram_error"] = str(e)

    # Extract all symbols
    all_symbols = set()
    noise = {"THE", "AND", "FOR", "VND", "USD", "CEO", "IPO", "ETF", "GDP", "FDI"}
    for item in all_items:
        all_symbols.update(s for s in item.get("symbols", []) if s not in noise)

    elapsed = round(time.time() - start, 2)
    stats["elapsed_seconds"] = elapsed

    return {
        "items": all_items,
        "total": len(all_items),
        "stats": stats,
        "symbols_found": sorted(all_symbols),
    }


# ---------------------------------------------------------------------------
# Full Pipeline: Scrape → Sentiment
# ---------------------------------------------------------------------------

@router.post("/scrape/pipeline")
async def scrape_pipeline(request: Request):
    """Full pipeline: Scrape all sources → Push to sentiment for analysis.

    This is the endpoint n8n workflow calls.
    """
    import httpx

    settings = get_settings()

    # Step 1: Scrape all sources
    scrape_result = await scrape_all_sources(request)
    items = scrape_result["items"]

    if not items:
        return {**scrape_result, "sentiment": {"error": "No items to analyze"}}

    # Step 2: Push to sentiment service
    texts_payload = [
        {
            "id": item.get("id", ""),
            "content": item.get("content", ""),
            "source": item.get("source", "unknown"),
        }
        for item in items[:100]  # Limit to 100 items per batch
    ]

    sentiment_results = []
    try:
        async with httpx.AsyncClient(timeout=120) as client:
            resp = await client.post(
                f"{settings.sentiment_service_url}/analyze",
                json={"texts": texts_payload},
            )
            if resp.status_code == 200:
                sentiment_results = resp.json().get("results", [])
            else:
                logger.warning("Sentiment API returned %d", resp.status_code)
    except Exception as e:
        logger.error("Sentiment service error: %s", e)

    # Step 3: Merge results
    sentiment_map = {r["id"]: r for r in sentiment_results}
    enriched = []
    for item in items:
        item_id = item.get("id", "")
        sentiment = sentiment_map.get(item_id, {})
        enriched.append({
            **item,
            "sentiment": sentiment.get("sentiment", "unknown"),
            "sentiment_confidence": sentiment.get("confidence", 0),
        })

    # Sort by sentiment confidence
    enriched.sort(key=lambda x: x.get("sentiment_confidence", 0), reverse=True)

    return {
        "total": len(enriched),
        "stats": scrape_result["stats"],
        "symbols_found": scrape_result["symbols_found"],
        "items": enriched,
    }


# ---------------------------------------------------------------------------
# Source Chasing (Phase 2): Follow links to primary sources
# ---------------------------------------------------------------------------

class ChaseRequest(BaseModel):
    """Request body for /scrape/chase — provide articles to chase."""
    articles: Optional[List[dict]] = None  # If None, runs /scrape/all first
    max_chase_per_article: int = Field(default=3, ge=1, le=10)


@router.post("/scrape/chase")
async def chase_primary_sources(request: Request, body: Optional[ChaseRequest] = None):
    """Chase primary sources from aggregated news articles.

    Given articles (or auto-scrapes via /scrape/all), follows outbound links
    to regulatory bodies (HOSE, HNX, SSC, SBV), company IR pages, and
    financial data portals. Returns the original source content as Markdown.

    This endpoint enables verification of rumors against official sources.
    """
    settings = get_settings()
    if not settings.enable_source_chasing:
        return {
            "error": "Source chasing not enabled",
            "hint": "Set ENABLE_SOURCE_CHASING=true",
        }
    if not settings.enable_firecrawl:
        return {
            "error": "Source chasing requires Firecrawl",
            "hint": "Set ENABLE_FIRECRAWL=true",
        }

    from app.scrapers.source_chaser import SourceChaser

    firecrawl = request.app.state.firecrawl
    max_chase = body.max_chase_per_article if body else 3

    # Get articles — either from request body or scrape fresh
    articles = (body.articles if body and body.articles else None)
    if not articles:
        scrape_result = await scrape_all_sources(request)
        articles = scrape_result.get("items", [])

    chaser = SourceChaser(
        firecrawl_client=firecrawl,
        max_chase_per_article=max_chase,
    )

    results = await chaser.chase_batch(articles)

    # Flatten primary sources for summary
    all_primary = []
    for result in results:
        all_primary.extend(result.primary_sources)

    return {
        "articles_processed": len(articles),
        "primary_sources_found": len(all_primary),
        "results": [r.to_dict() for r in results],
        "stats": chaser.get_stats(),
        "source_types": {
            "regulatory": sum(1 for ps in all_primary if ps.source_type == "regulatory"),
            "ir": sum(1 for ps in all_primary if ps.source_type == "ir"),
            "data_portal": sum(1 for ps in all_primary if ps.source_type == "data_portal"),
            "inferred": sum(1 for ps in all_primary if ps.source_type == "inferred"),
        },
    }


# ---------------------------------------------------------------------------
# Scheduler Control
# ---------------------------------------------------------------------------

@router.get("/scheduler/status")
async def scheduler_status():
    from app.scheduler import get_scheduler
    return get_scheduler().status


@router.post("/scheduler/start")
async def scheduler_start():
    from app.scheduler import get_scheduler
    scheduler = get_scheduler()
    scheduler.start()
    return {"status": "started", **scheduler.status}


@router.post("/scheduler/stop")
async def scheduler_stop():
    from app.scheduler import get_scheduler
    scheduler = get_scheduler()
    scheduler.stop()
    return {"status": "stopped", **scheduler.status}


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _infer_source(url: str) -> str:
    """Infer source name from URL domain."""
    if "dantri" in url:
        return "dantri"
    if "thanhnien" in url:
        return "thanhnien"
    if "tuoitre" in url:
        return "tuoitre"
    if "vnexpress" in url:
        return "vnexpress"
    if "cafef" in url:
        return "cafef"
    return "unknown"
