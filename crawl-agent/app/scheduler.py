"""
Async news scraping scheduler.

Runs periodic scraping tasks during market hours (8:30-16:00 ICT, Mon-Fri).
Stores results in JSON files and pushes to sentiment service.
"""

import asyncio
import json
import logging
import os
import time
from datetime import datetime, timezone, timedelta
from pathlib import Path
from typing import Dict, Optional

logger = logging.getLogger(__name__)

# Vietnam timezone (ICT = UTC+7)
ICT = timezone(timedelta(hours=7))

# Market hours
MARKET_OPEN_HOUR = 8
MARKET_OPEN_MINUTE = 30
MARKET_CLOSE_HOUR = 16
MARKET_CLOSE_MINUTE = 0


class ScrapeScheduler:
    """Async scheduler for periodic news scraping."""

    def __init__(
        self,
        interval_minutes: int = 30,
        data_dir: str = "data/scheduled",
        market_hours_only: bool = True,
        redis_client=None,
    ):
        self.interval = interval_minutes * 60
        self.data_dir = Path(data_dir)
        self.data_dir.mkdir(parents=True, exist_ok=True)
        self.market_hours_only = market_hours_only
        self._task: Optional[asyncio.Task] = None
        self._running = False
        self._last_run: Optional[datetime] = None
        self._stats: Dict = {"runs": 0, "total_items": 0, "errors": 0}

        self._news_publisher = None
        self._article_generator = None
        if redis_client is not None:
            from app.config import get_settings
            from app.services.news_publisher import NewsPublisher
            from app.services.article_generator import ArticleGenerator
            s = get_settings()
            self._news_publisher = NewsPublisher(
                redis_client,
                max_items=s.news_redis_max_items,
                ttl_hours=s.news_redis_ttl_hours,
            )
            if s.anthropic_api_key or s.gemini_api_key or s.xai_api_key:
                self._article_generator = ArticleGenerator(
                    redis_client=redis_client,
                    anthropic_api_key=s.anthropic_api_key,
                    go_services_url=s.go_services_internal_url,
                    go_services_key=s.go_services_internal_key,
                    telegram_bot_token=os.environ.get("TELEGRAM_BOT_TOKEN", ""),
                    telegram_chat_id=s.telegram_admin_chat_id,
                    dashboard_url=s.dashboard_url,
                    claude_model=s.article_claude_model,
                    max_daily=s.article_max_daily,
                    hot_symbols_count=s.article_hot_symbols_count,
                    article_model=s.article_model,
                    gemini_api_key=s.gemini_api_key,
                    gemini_text_model=s.gemini_text_model,
                    xai_api_key=s.xai_api_key,
                    grok_model=s.grok_model,
                    enable_image_generation=s.enable_image_generation,
                    vertex_credentials_json=s.vertex_credentials_json,
                    vertex_project_id=s.vertex_project_id,
                    vertex_location=s.vertex_location,
                    vertex_image_model=s.vertex_image_model,
                    image_storage_backend=s.image_storage_backend,
                    s3_endpoint=s.s3_endpoint,
                    s3_bucket=s.s3_bucket,
                    s3_access_key=s.s3_access_key,
                    s3_secret_key=s.s3_secret_key,
                    s3_public_url=s.s3_public_url,
                    gdrive_credentials_json=s.gdrive_credentials_json or s.vertex_credentials_json,
                    gdrive_folder_id=s.gdrive_folder_id,
                )

    def _is_market_hours(self) -> bool:
        now = datetime.now(ICT)
        if now.weekday() >= 5:
            return False
        market_open = now.replace(hour=MARKET_OPEN_HOUR, minute=MARKET_OPEN_MINUTE, second=0)
        market_close = now.replace(hour=MARKET_CLOSE_HOUR, minute=MARKET_CLOSE_MINUTE, second=0)
        return market_open <= now <= market_close

    async def _run_scrape(self) -> Dict:
        """Execute one scraping cycle."""
        from app.scrapers.rss_scraper import scrape_all_feeds
        from app.scrapers.web_scraper import scrape_all_websites

        start = time.time()
        all_items = []
        stats = {}
        seen_hashes = set()

        # Tier 1: RSS feeds
        try:
            rss_items = await scrape_all_feeds(delay_between=1.0)
            for item in rss_items:
                if item.content_hash not in seen_hashes:
                    seen_hashes.add(item.content_hash)
                    all_items.append(item.to_dict())
            stats["rss"] = len(rss_items)
        except Exception as e:
            stats["rss_error"] = str(e)
            logger.error("Scheduler RSS error: %s", e)

        # Tier 1: Web crawl
        try:
            web_articles = await scrape_all_websites(delay_between=2.0)
            for article in web_articles:
                if article.content_hash not in seen_hashes:
                    seen_hashes.add(article.content_hash)
                    all_items.append(article.to_dict())
            stats["web_crawl"] = len(web_articles)
        except Exception as e:
            stats["web_crawl_error"] = str(e)
            logger.error("Scheduler web crawl error: %s", e)

        # Tier 2: Firecrawl (if available)
        try:
            from app.scrapers.firecrawl_client import FirecrawlClient
            from app.config import get_settings
            settings = get_settings()
            if settings.enable_firecrawl:
                fc = FirecrawlClient(base_url=settings.firecrawl_api_url)
                tier2_urls = [
                    "https://dantri.com.vn/kinh-doanh/chung-khoan.htm",
                    "https://thanhnien.vn/tai-chinh-kinh-doanh/chung-khoan.htm",
                    "https://tuoitre.vn/kinh-doanh.htm",
                ]
                results = await fc.scrape_batch(tier2_urls, max_concurrent=2)
                fc_count = sum(1 for r in results if r.success)
                stats["firecrawl"] = fc_count
        except Exception as e:
            stats["firecrawl_error"] = str(e)

        # Tier 3: Telegram
        try:
            from app.config import get_settings
            settings = get_settings()
            if settings.telegram_app_api_id and settings.telegram_app_api_hash:
                from app.scrapers.telegram_scraper import TelegramScraper
                scraper = TelegramScraper(
                    api_id=settings.telegram_app_api_id,
                    api_hash=settings.telegram_app_api_hash,
                    session_name=settings.telegram_session_name,
                    channels=settings.telegram_channel_list,
                )
                messages = await scraper.scrape_channels(
                    limit_per_channel=20, hours_back=2,
                )
                for msg in messages:
                    d = msg.to_dict()
                    h = d.get("id", "")
                    if h and h not in seen_hashes:
                        seen_hashes.add(h)
                        all_items.append(d)
                stats["telegram"] = len(messages)
                await scraper.close()
        except Exception as e:
            stats["telegram_error"] = str(e)

        elapsed = round(time.time() - start, 2)
        stats["elapsed_seconds"] = elapsed
        stats["total_items"] = len(all_items)

        # Extract symbols
        all_symbols = set()
        noise = {"THE", "AND", "FOR", "VND", "USD", "CEO", "IPO", "ETF", "GDP", "FDI"}
        for item in all_items:
            all_symbols.update(s for s in item.get("symbols", []) if s not in noise)

        # Publish to Redis per symbol (non-blocking — failure does not abort scrape)
        if self._news_publisher and all_items:
            pub_stats = await self._news_publisher.publish_batch(all_items)
            stats["redis_published"] = pub_stats
            logger.info(
                "Redis publish: %d items → %d symbol keys",
                pub_stats["items_published"],
                pub_stats["symbol_keys_updated"],
            )

        # Generate article drafts for hot symbols (non-blocking)
        if self._article_generator and all_items:
            try:
                article_stats = await self._article_generator.run()
                stats["articles_generated"] = article_stats
                logger.info(
                    "Articles: %d created from %d hot symbols",
                    article_stats["articles_created"],
                    article_stats["symbols_processed"],
                )
            except Exception as e:
                logger.warning("Article generation error: %s", e)

        # Save to JSON
        timestamp = datetime.now(ICT).strftime("%Y%m%d_%H%M%S")
        output_file = self.data_dir / f"scrape_{timestamp}.json"
        output = {
            "timestamp": datetime.now(ICT).isoformat(),
            "stats": stats,
            "symbols_found": sorted(all_symbols),
            "items": all_items,
        }
        with open(output_file, "w", encoding="utf-8") as f:
            json.dump(output, f, ensure_ascii=False, indent=2, default=str)

        latest_file = self.data_dir / "latest.json"
        with open(latest_file, "w", encoding="utf-8") as f:
            json.dump(output, f, ensure_ascii=False, indent=2, default=str)

        logger.info(
            "Scheduler run #%d: %d items, %d symbols in %.1fs → %s",
            self._stats["runs"] + 1, len(all_items), len(all_symbols), elapsed, output_file.name,
        )
        return stats

    async def _loop(self):
        logger.info(
            "Scrape scheduler started (interval=%dm, market_hours_only=%s)",
            self.interval // 60, self.market_hours_only,
        )
        while self._running:
            try:
                if self.market_hours_only and not self._is_market_hours():
                    await asyncio.sleep(300)
                    continue

                stats = await self._run_scrape()
                self._stats["runs"] += 1
                self._stats["total_items"] += stats.get("total_items", 0)
                self._last_run = datetime.now(ICT)
                await asyncio.sleep(self.interval)

            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error("Scheduler error: %s", e)
                self._stats["errors"] += 1
                await asyncio.sleep(60)

    def start(self):
        if self._running:
            return
        self._running = True
        self._task = asyncio.create_task(self._loop())

    def stop(self):
        self._running = False
        if self._task:
            self._task.cancel()

    @property
    def status(self) -> Dict:
        return {
            "running": self._running,
            "last_run": self._last_run.isoformat() if self._last_run else None,
            "is_market_hours": self._is_market_hours(),
            "interval_minutes": self.interval // 60,
            "stats": self._stats,
        }


_scheduler: Optional[ScrapeScheduler] = None


def get_scheduler(redis_client=None) -> ScrapeScheduler:
    global _scheduler
    if _scheduler is None:
        interval = int(os.environ.get("SCRAPE_INTERVAL_MINUTES", "30"))
        market_only = os.environ.get("SCRAPE_MARKET_HOURS_ONLY", "true").lower() == "true"
        _scheduler = ScrapeScheduler(
            interval_minutes=interval,
            market_hours_only=market_only,
            redis_client=redis_client,
        )
    return _scheduler
