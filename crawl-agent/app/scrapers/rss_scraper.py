"""
RSS feed scraper for Vietnamese financial news.

Fetches and parses RSS feeds from official financial news sources,
returning structured items ready for sentiment analysis.
"""

import asyncio
import hashlib
import logging
import re
from dataclasses import dataclass, field
from datetime import datetime
from typing import List, Optional
from xml.etree import ElementTree

import aiohttp

logger = logging.getLogger(__name__)

# Vietnamese financial news RSS feeds (verified working 2026-03)
RSS_FEEDS = {
    "vnexpress_kinhdoanh": {
        "url": "https://vnexpress.net/rss/kinh-doanh.rss",
        "source": "vnexpress",
    },
    "vnexpress_latest": {
        "url": "https://vnexpress.net/rss/tin-moi-nhat.rss",
        "source": "vnexpress",
    },
    "tuoitre_kinhdoanh": {
        "url": "https://tuoitre.vn/rss/kinh-doanh.rss",
        "source": "tuoitre",
    },
    "dantri_kinhdoanh": {
        "url": "https://dantri.com.vn/rss/kinh-doanh.htm",
        "source": "dantri",
    },
    "thanhnien_taichinh": {
        "url": "https://thanhnien.vn/rss/tai-chinh-kinh-doanh/chung-khoan.rss",
        "source": "thanhnien",
    },
    "cafef_homepage": {
        "url": "https://cafef.vn/rss/home.rss",
        "source": "cafef",
    },
}

# Spam keywords to filter out non-financial content
SPAM_KEYWORDS = [
    "khuyến mại", "đăng ký ngay", "tải app", "quảng cáo",
    "affiliate", "click here", "subscribe",
]


@dataclass
class RSSItem:
    """A parsed RSS feed item."""
    title: str
    description: str
    link: str
    source: str
    published_at: Optional[datetime] = None
    content_hash: str = ""
    symbols: List[str] = field(default_factory=list)

    def __post_init__(self):
        if not self.content_hash:
            text = f"{self.title}|{self.description}"
            self.content_hash = hashlib.md5(text.encode("utf-8")).hexdigest()
        if not self.symbols:
            from app.scrapers.symbol_detector import extract_symbols
            self.symbols = extract_symbols(f"{self.title} {self.description}")

    @property
    def full_text(self) -> str:
        return f"{self.title}. {self.description}"

    def to_dict(self) -> dict:
        return {
            "id": self.content_hash,
            "content": self.full_text,
            "source": self.source,
            "link": self.link,
            "symbols": self.symbols,
            "published_at": self.published_at.isoformat() if self.published_at else None,
        }


def _is_spam(text: str) -> bool:
    """Check if text contains spam keywords."""
    text_lower = text.lower()
    return any(kw in text_lower for kw in SPAM_KEYWORDS)


def _parse_rss_xml(xml_text: str, source: str) -> List[RSSItem]:
    """Parse RSS XML into structured items."""
    items = []
    try:
        root = ElementTree.fromstring(xml_text)
        # Handle both RSS 2.0 and Atom feeds
        for item_el in root.iter("item"):
            title = (item_el.findtext("title") or "").strip()
            desc = (item_el.findtext("description") or "").strip()
            link = (item_el.findtext("link") or "").strip()
            pub_date = item_el.findtext("pubDate")

            if not title:
                continue
            if _is_spam(f"{title} {desc}"):
                continue

            # Clean HTML tags from description
            desc = re.sub(r"<[^>]+>", "", desc).strip()
            # Truncate long descriptions
            if len(desc) > 500:
                desc = desc[:500] + "..."

            parsed_date = None
            if pub_date:
                try:
                    from email.utils import parsedate_to_datetime
                    parsed_date = parsedate_to_datetime(pub_date)
                except Exception:
                    pass

            items.append(RSSItem(
                title=title,
                description=desc,
                link=link,
                source=source,
                published_at=parsed_date,
            ))
    except ElementTree.ParseError as e:
        logger.warning("Failed to parse RSS XML from %s: %s", source, e)

    return items


async def fetch_feed(
    session: aiohttp.ClientSession,
    name: str,
    url: str,
    source: str,
    timeout: int = 15,
) -> List[RSSItem]:
    """Fetch and parse a single RSS feed."""
    try:
        async with session.get(url, timeout=aiohttp.ClientTimeout(total=timeout)) as resp:
            if resp.status != 200:
                logger.warning("RSS feed %s returned HTTP %d", name, resp.status)
                return []
            xml_text = await resp.text()
            items = _parse_rss_xml(xml_text, source)
            logger.info("Fetched %d items from %s", len(items), name)
            return items
    except asyncio.TimeoutError:
        logger.warning("RSS feed %s timed out", name)
        return []
    except Exception as e:
        logger.warning("Failed to fetch RSS feed %s: %s", name, e)
        return []


async def scrape_all_feeds(
    feeds: Optional[dict] = None,
    delay_between: float = 1.0,
) -> List[RSSItem]:
    """Scrape all configured RSS feeds with rate limiting.

    Args:
        feeds: Override feed config (default: RSS_FEEDS)
        delay_between: Seconds to wait between feed requests

    Returns:
        Combined list of RSSItem from all feeds, deduped by content_hash.
    """
    feeds = feeds or RSS_FEEDS
    all_items: List[RSSItem] = []
    seen_hashes: set = set()

    async with aiohttp.ClientSession(
        headers={"User-Agent": "VNStockBot/1.0"}
    ) as session:
        for name, config in feeds.items():
            items = await fetch_feed(session, name, config["url"], config["source"])
            for item in items:
                if item.content_hash not in seen_hashes:
                    seen_hashes.add(item.content_hash)
                    all_items.append(item)
            # Rate limit between feeds
            await asyncio.sleep(delay_between)

    logger.info("Total unique RSS items scraped: %d", len(all_items))
    return all_items
