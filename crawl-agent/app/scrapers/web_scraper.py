"""
Hybrid web crawl engine for Vietnamese financial news.

Architecture (as recommended):
- Group 1: Static HTML (VnExpress, CafeF) — aiohttp + BeautifulSoup4
- Group 2: Static HTML w/ complex DOM (DanTri, ThanhNien, TuoiTre) — aiohttp + BeautifulSoup4
  (All sites serve articles in HTML; Playwright not needed)

Features:
- BeautifulSoup4 + lxml for robust parsing
- User-Agent rotation via fake-useragent
- ETag/Last-Modified caching to skip unchanged pages
- Domain-aware rate limiting
- Content deduplication

Usage:
    from app.scrapers.web_scraper import scrape_all_websites
    articles = await scrape_all_websites()
"""

import asyncio
import hashlib
import logging
import re
import time
from dataclasses import dataclass, field
from datetime import datetime
from typing import Any, Callable, Dict, List, Optional, Set
from urllib.parse import urlparse

import aiohttp
from bs4 import BeautifulSoup

logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# User-Agent rotation
# ---------------------------------------------------------------------------
try:
    from fake_useragent import UserAgent
    _ua = UserAgent()
    def random_ua() -> str:
        return _ua.random
except ImportError:
    _AGENTS = [
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/605.1.15 Safari/605.1.15",
        "Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
    ]
    _idx = [0]
    def random_ua() -> str:
        _idx[0] = (_idx[0] + 1) % len(_AGENTS)
        return _AGENTS[_idx[0]]


# Stock symbol detection (company name mapping + multi-strategy extraction)
from app.scrapers.symbol_detector import extract_symbols


# ---------------------------------------------------------------------------
# Article dataclass
# ---------------------------------------------------------------------------

@dataclass
class WebArticle:
    """A scraped web article."""
    title: str
    description: str
    link: str
    source: str
    published_at: Optional[str] = None
    content_hash: str = ""
    symbols: List[str] = field(default_factory=list)

    def __post_init__(self):
        if not self.content_hash:
            text = f"{self.title}|{self.link}"
            self.content_hash = hashlib.md5(text.encode("utf-8")).hexdigest()
        if not self.symbols:
            self.symbols = extract_symbols(f"{self.title} {self.description}")

    @property
    def full_text(self) -> str:
        return f"{self.title}. {self.description}" if self.description else self.title

    def to_dict(self) -> dict:
        return {
            "id": self.content_hash,
            "content": self.full_text,
            "title": self.title,
            "source": self.source,
            "link": self.link,
            "symbols": self.symbols,
            "published_at": self.published_at,
        }


# ---------------------------------------------------------------------------
# ETag cache (in-memory, per-process)
# ---------------------------------------------------------------------------

_etag_cache: Dict[str, Dict[str, str]] = {}


# ---------------------------------------------------------------------------
# BeautifulSoup parsers for each site
# ---------------------------------------------------------------------------

def parse_vnexpress(soup: BeautifulSoup, base_url: str = "https://vnexpress.net") -> List[WebArticle]:
    """Parse VnExpress pages using BeautifulSoup."""
    articles = []
    seen = set()

    # Primary: article title links
    for tag in soup.select("h2.title-news a, h3.title-news a, .title-news a"):
        link = tag.get("href", "")
        title = tag.get_text(strip=True)
        if not title or len(title) < 15 or link in seen:
            continue
        if not link.startswith("http"):
            link = base_url + link
        seen.add(link)

        # Get description from nearby <p class="description">
        desc = ""
        parent = tag.find_parent("article") or tag.find_parent("div")
        if parent:
            desc_tag = parent.find("p", class_=re.compile(r"description|sapo"))
            if desc_tag:
                desc = desc_tag.get_text(strip=True)

        articles.append(WebArticle(title=title, description=desc, link=link, source="vnexpress"))

    # Fallback: any article links with .html extension
    if len(articles) < 5:
        for a in soup.find_all("a", href=re.compile(r"vnexpress\.net/.*-\d+\.html")):
            link = a.get("href", "")
            title = a.get_text(strip=True) or a.get("title", "")
            if not title or len(title) < 20 or link in seen:
                continue
            seen.add(link)
            articles.append(WebArticle(title=title, description="", link=link, source="vnexpress"))

    return articles


def parse_dantri(soup: BeautifulSoup, base_url: str = "https://dantri.com.vn") -> List[WebArticle]:
    """Parse DanTri pages using BeautifulSoup."""
    articles = []
    seen = set()

    # DanTri uses <h3 class="article-title"><a href="...">Title</a></h3>
    for tag in soup.select("h3.article-title a, h2.article-title a, .article-title a"):
        link = tag.get("href", "")
        title = tag.get_text(strip=True) or tag.get("title", "")
        if not title or len(title) < 15 or link in seen:
            continue
        if not link.startswith("http"):
            link = base_url + link
        seen.add(link)

        desc = ""
        parent = tag.find_parent("article") or tag.find_parent("div")
        if parent:
            desc_tag = parent.find("div", class_=re.compile(r"article-excerpt|excerpt|sapo"))
            if desc_tag:
                desc = desc_tag.get_text(strip=True)

        articles.append(WebArticle(title=title, description=desc, link=link, source="dantri"))

    # Fallback: article links
    if len(articles) < 5:
        for a in soup.find_all("a", href=re.compile(r"/kinh-doanh/.*\.htm")):
            link = a.get("href", "")
            title = a.get_text(strip=True)
            if not title or len(title) < 20 or link in seen:
                continue
            if not link.startswith("http"):
                link = base_url + link
            seen.add(link)
            articles.append(WebArticle(title=title, description="", link=link, source="dantri"))

    return articles


def parse_cafef(soup: BeautifulSoup, base_url: str = "https://cafef.vn") -> List[WebArticle]:
    """Parse CafeF pages using BeautifulSoup."""
    articles = []
    seen = set()

    # CafeF: various article link patterns
    for a in soup.find_all("a", href=re.compile(r".*\.chn$")):
        link = a.get("href", "")
        title = a.get_text(strip=True) or a.get("title", "")
        if not title or len(title) < 15 or link in seen:
            continue
        if not link.startswith("http"):
            link = base_url + link
        # Skip category/navigation links
        if link.count("/") <= 3 and not re.search(r"-\d+", link):
            continue
        seen.add(link)

        desc = ""
        parent = tag_parent = a.find_parent("div") or a.find_parent("li")
        if tag_parent:
            desc_tag = tag_parent.find("p", class_=re.compile(r"sapo|desc"))
            if desc_tag:
                desc = desc_tag.get_text(strip=True)

        articles.append(WebArticle(title=title, description=desc, link=link, source="cafef"))

    return articles


def parse_thanhnien(soup: BeautifulSoup, base_url: str = "https://thanhnien.vn") -> List[WebArticle]:
    """Parse ThanhNien pages using BeautifulSoup."""
    articles = []
    seen = set()

    # ThanhNien: article links with specific patterns
    for a in soup.find_all("a", href=re.compile(r"thanhnien\.vn/.*-\d+\.htm")):
        link = a.get("href", "")
        title = a.get_text(strip=True) or a.get("title", "")
        if not title or len(title) < 20 or link in seen:
            continue
        if not link.startswith("http"):
            link = base_url + link
        seen.add(link)
        articles.append(WebArticle(title=title, description="", link=link, source="thanhnien"))

    # Also check h2/h3 article titles
    for tag in soup.select("h2 a, h3 a"):
        link = tag.get("href", "")
        title = tag.get_text(strip=True)
        if not link or not title or len(title) < 20 or link in seen:
            continue
        if not link.startswith("http"):
            link = base_url + link
        if not re.search(r"-\d+\.htm", link):
            continue
        seen.add(link)
        articles.append(WebArticle(title=title, description="", link=link, source="thanhnien"))

    return articles


def parse_tuoitre(soup: BeautifulSoup, base_url: str = "https://tuoitre.vn") -> List[WebArticle]:
    """Parse TuoiTre pages using BeautifulSoup."""
    articles = []
    seen = set()

    for a in soup.find_all("a", href=re.compile(r"tuoitre\.vn/.*\.htm")):
        link = a.get("href", "")
        title = a.get_text(strip=True) or a.get("title", "")
        if not title or len(title) < 20 or link in seen:
            continue
        if not link.startswith("http"):
            link = base_url + link
        seen.add(link)
        articles.append(WebArticle(title=title, description="", link=link, source="tuoitre"))

    for tag in soup.select("h2 a, h3 a"):
        link = tag.get("href", "")
        title = tag.get_text(strip=True)
        if not link or not title or len(title) < 20 or link in seen:
            continue
        if not link.startswith("http"):
            link = base_url + link
        if not re.search(r"\.htm", link):
            continue
        seen.add(link)
        articles.append(WebArticle(title=title, description="", link=link, source="tuoitre"))

    return articles


# ---------------------------------------------------------------------------
# Site configuration
# ---------------------------------------------------------------------------

SITES: List[Dict[str, Any]] = [
    # Group 1: VnExpress (static, fast)
    {"name": "vnexpress_chungkhoan", "url": "https://vnexpress.net/kinh-doanh/chung-khoan", "parser": parse_vnexpress},
    {"name": "vnexpress_kinhdoanh", "url": "https://vnexpress.net/kinh-doanh", "parser": parse_vnexpress},
    # Group 1: CafeF (static, needs UA rotation)
    {"name": "cafef_chungkhoan", "url": "https://cafef.vn/thi-truong-chung-khoan.chn", "parser": parse_cafef},
    {"name": "cafef_home", "url": "https://cafef.vn/", "parser": parse_cafef},
    # Group 2: DanTri (static HTML, complex DOM)
    {"name": "dantri_chungkhoan", "url": "https://dantri.com.vn/kinh-doanh/chung-khoan.htm", "parser": parse_dantri},
    {"name": "dantri_kinhdoanh", "url": "https://dantri.com.vn/kinh-doanh.htm", "parser": parse_dantri},
    # Group 2: ThanhNien
    {"name": "thanhnien_chungkhoan", "url": "https://thanhnien.vn/tai-chinh-kinh-doanh/chung-khoan.htm", "parser": parse_thanhnien},
    # Group 2: TuoiTre
    {"name": "tuoitre_kinhdoanh", "url": "https://tuoitre.vn/kinh-doanh.htm", "parser": parse_tuoitre},
]


# ---------------------------------------------------------------------------
# Scraping engine
# ---------------------------------------------------------------------------

async def _fetch_with_cache(
    session: aiohttp.ClientSession,
    url: str,
    timeout: int = 15,
) -> Optional[str]:
    """Fetch URL with ETag/Last-Modified caching."""
    headers = {"User-Agent": random_ua()}

    # Add conditional headers if we have cached values
    cached = _etag_cache.get(url, {})
    if cached.get("etag"):
        headers["If-None-Match"] = cached["etag"]
    if cached.get("last_modified"):
        headers["If-Modified-Since"] = cached["last_modified"]

    try:
        async with session.get(url, headers=headers, timeout=aiohttp.ClientTimeout(total=timeout)) as resp:
            # 304 Not Modified → use cached content
            if resp.status == 304:
                logger.debug("Cache hit (304): %s", url)
                return cached.get("content")

            if resp.status != 200:
                logger.warning("HTTP %d from %s", resp.status, url)
                return None

            html = await resp.text()

            # Cache ETag and Last-Modified
            _etag_cache[url] = {
                "etag": resp.headers.get("ETag", ""),
                "last_modified": resp.headers.get("Last-Modified", ""),
                "content": html,
                "fetched_at": time.time(),
            }

            return html
    except asyncio.TimeoutError:
        logger.warning("Timeout: %s", url)
        return None
    except Exception as e:
        logger.warning("Failed to fetch %s: %s", url, e)
        return None


async def scrape_website(
    session: aiohttp.ClientSession,
    site: Dict[str, Any],
) -> List[WebArticle]:
    """Scrape a single website."""
    html = await _fetch_with_cache(session, site["url"])
    if not html:
        return []

    soup = BeautifulSoup(html, "lxml")
    articles = site["parser"](soup)
    logger.info("Crawled %d articles from %s", len(articles), site["name"])
    return articles


async def scrape_all_websites(
    sites: Optional[List[Dict[str, Any]]] = None,
    delay_between: float = 2.0,
    max_concurrent: int = 3,
) -> List[WebArticle]:
    """Crawl all configured websites with rate limiting.

    Args:
        sites: Override site config (default: SITES)
        delay_between: Seconds between requests to same domain
        max_concurrent: Max concurrent requests across domains

    Returns:
        Combined list of WebArticle, deduped by content_hash and title.
    """
    sites = sites or SITES
    all_articles: List[WebArticle] = []
    seen_hashes: Set[str] = set()
    seen_titles: Set[str] = set()

    # Group by domain
    domain_groups: Dict[str, List[Dict]] = {}
    for site in sites:
        domain = urlparse(site["url"]).netloc
        domain_groups.setdefault(domain, []).append(site)

    semaphore = asyncio.Semaphore(max_concurrent)

    async def scrape_domain(domain: str, domain_sites: List[Dict]):
        """Scrape all pages for one domain, respecting rate limits."""
        async with aiohttp.ClientSession() as session:
            for site in domain_sites:
                async with semaphore:
                    articles = await scrape_website(session, site)
                    for article in articles:
                        title_key = article.title[:50].lower()
                        if (article.content_hash not in seen_hashes
                                and title_key not in seen_titles):
                            seen_hashes.add(article.content_hash)
                            seen_titles.add(title_key)
                            all_articles.append(article)
                    await asyncio.sleep(delay_between)

    start = time.time()
    await asyncio.gather(*[
        scrape_domain(domain, domain_sites)
        for domain, domain_sites in domain_groups.items()
    ])
    elapsed = time.time() - start

    logger.info(
        "Web crawl complete: %d articles from %d pages across %d domains in %.1fs",
        len(all_articles), len(sites), len(domain_groups), elapsed,
    )
    return all_articles
