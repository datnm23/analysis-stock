"""
Source Chaser — Phase 2: Primary Source Deep Crawl.

Given aggregated news items from Tier 1/2/3 scrapers, this module:
1. Extracts outbound links from each article
2. Identifies "primary source" URLs (press releases, SEC filings, official reports)
3. Uses Firecrawl to fetch clean Markdown from the primary source
4. Returns enriched articles with the original source content attached

This is the n8n "Source Chasing" pattern from the reference workflow:
  - Article mentions "theo thông cáo của HOSE" → chase HOSE link
  - Article links to company IR page → chase and extract PR content
  - Article cites a government decree → chase decree URL

Use cases:
- Verify rumored news against official sources
- Get the full, unedited text of press releases
- Build a corpus of high-quality primary documents for LLM analysis
"""

import asyncio
import logging
import re
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Set
from urllib.parse import urlparse

logger = logging.getLogger(__name__)


# ---------------------------------------------------------------------------
# Configuration: Which domains are considered "primary sources"
# ---------------------------------------------------------------------------

# Official market regulators and exchanges
REGULATORY_DOMAINS = {
    "hnx.vn",             # Hanoi Stock Exchange
    "hsx.vn",             # Ho Chi Minh Stock Exchange  (HOSE)
    "ssc.gov.vn",         # State Securities Commission
    "sbv.gov.vn",         # State Bank of Vietnam
    "gso.gov.vn",         # General Statistics Office
    "mof.gov.vn",         # Ministry of Finance
    "mpi.gov.vn",         # Ministry of Planning & Investment
    "customs.gov.vn",     # General Dept. of Customs
}

# Company IR pages and data providers
IR_DOMAINS = {
    "ir.vingroup.net",
    "ir.fpt.com.vn",
    "investor.vinhomes.vn",
    "www.hoaphat.com.vn",
    "www.vcbs.com.vn",
    "www.vndirect.com.vn",
    "www.ssi.com.vn",
    "www.vietcombank.com.vn",
    "www.bidv.com.vn",
    "www.agribank.com.vn",
}

# Financial data portals (structured data)
DATA_PORTALS = {
    "finance.vietstock.vn",
    "s.cafef.vn",
    "stockbiz.vn",
    "fireant.vn",
    "simplize.vn",
}

# Combine all priority source domains
PRIMARY_SOURCE_DOMAINS = REGULATORY_DOMAINS | IR_DOMAINS | DATA_PORTALS

# Patterns that indicate a link is likely a primary source
PRIMARY_SOURCE_PATTERNS = [
    r"thong-?bao",           # Thông báo (announcement)
    r"cong-?bo",             # Công bố (disclosure)
    r"quyet-?dinh",          # Quyết định (decision)
    r"nghi-?dinh",           # Nghị định (decree)
    r"bao-?cao",             # Báo cáo (report)
    r"press-?release",       # English press release
    r"investor-?relation",   # IR pages
    r"/ir/",                 # IR path
    r"/disclosure/",         # Disclosure path
    r"\.pdf$",               # PDF documents
    r"\.doc[x]?$",           # Word documents
]

# Domains to always skip (ads, tracking, social)
SKIP_DOMAINS = {
    "facebook.com", "fb.com", "twitter.com", "x.com",
    "youtube.com", "youtu.be", "tiktok.com",
    "google.com", "googleapis.com", "gstatic.com",
    "doubleclick.net", "googlesyndication.com",
    "zalo.me", "zalo.vn",
}


@dataclass
class PrimarySource:
    """A primary source document extracted from a news article link."""
    url: str
    title: str = ""
    markdown: str = ""
    source_type: str = "unknown"  # regulatory, ir, data_portal, inferred
    domain: str = ""
    confidence: float = 0.0  # 0-1, how confident are we this is a real primary source
    linked_from: str = ""    # URL of the article that linked here
    error: str = ""

    def to_dict(self) -> dict:
        return {
            "url": self.url,
            "title": self.title,
            "markdown_preview": self.markdown[:500] if self.markdown else "",
            "markdown_length": len(self.markdown),
            "source_type": self.source_type,
            "domain": self.domain,
            "confidence": round(self.confidence, 2),
            "linked_from": self.linked_from,
        }


@dataclass
class ChaseResult:
    """Result of chasing primary sources for an article."""
    article_url: str
    primary_sources: List[PrimarySource] = field(default_factory=list)
    links_analyzed: int = 0
    links_chased: int = 0

    def to_dict(self) -> dict:
        return {
            "article_url": self.article_url,
            "primary_sources": [ps.to_dict() for ps in self.primary_sources],
            "links_analyzed": self.links_analyzed,
            "links_chased": self.links_chased,
        }


class SourceChaser:
    """Chases primary sources from aggregated news articles.

    Given a list of articles (from RSS/web scrape/Firecrawl), extracts
    outbound links and fetches primary source documents using Firecrawl.
    """

    def __init__(
        self,
        firecrawl_client,  # FirecrawlClient instance
        max_chase_per_article: int = 3,
        max_concurrent: int = 2,
        delay_between: float = 3.0,
    ):
        self.firecrawl = firecrawl_client
        self.max_chase_per_article = max_chase_per_article
        self.max_concurrent = max_concurrent
        self.delay_between = delay_between
        self._chased_urls: Set[str] = set()  # Global dedup within one session

    def classify_url(self, url: str) -> Optional[PrimarySource]:
        """Classify a URL as a potential primary source.

        Returns PrimarySource stub if it's likely a primary source, None otherwise.
        """
        try:
            parsed = urlparse(url)
            domain = parsed.netloc.lower().lstrip("www.")
            path = parsed.path.lower()
        except Exception:
            return None

        # Skip known non-primary domains
        if any(skip in domain for skip in SKIP_DOMAINS):
            return None

        # Skip same-site links from news sites (these are just other articles)
        news_domains = {"vnexpress.net", "cafef.vn", "dantri.com.vn",
                        "thanhnien.vn", "tuoitre.vn", "tienphong.vn",
                        "vietnamnet.vn", "laodong.vn", "nld.com.vn"}
        if domain in news_domains:
            return None

        # Check regulatory domains (highest confidence)
        if domain in REGULATORY_DOMAINS:
            return PrimarySource(
                url=url, domain=domain,
                source_type="regulatory", confidence=0.95,
            )

        # Check company IR domains
        if domain in IR_DOMAINS:
            return PrimarySource(
                url=url, domain=domain,
                source_type="ir", confidence=0.90,
            )

        # Check data portals
        if domain in DATA_PORTALS:
            return PrimarySource(
                url=url, domain=domain,
                source_type="data_portal", confidence=0.80,
            )

        # Check URL patterns
        for pattern in PRIMARY_SOURCE_PATTERNS:
            if re.search(pattern, path, re.IGNORECASE):
                return PrimarySource(
                    url=url, domain=domain,
                    source_type="inferred", confidence=0.60,
                )

        return None

    def extract_candidate_links(
        self,
        article: Dict[str, Any],
    ) -> List[PrimarySource]:
        """Extract candidate primary source links from an article.

        Looks at:
        - Explicit 'links' field (from Firecrawl)
        - URLs embedded in content/markdown
        - URLs in description
        """
        candidates: List[PrimarySource] = []
        seen_urls: Set[str] = set()

        # From explicit links field (Firecrawl output)
        for link in article.get("links", []):
            if isinstance(link, str) and link not in seen_urls:
                seen_urls.add(link)
                ps = self.classify_url(link)
                if ps:
                    ps.linked_from = article.get("link", article.get("url", ""))
                    candidates.append(ps)

        # From markdown/content (URL regex)
        for field_name in ("markdown", "content", "description"):
            text = article.get(field_name, "")
            if text:
                urls = re.findall(
                    r'https?://[^\s\)\]\"\'>]+',
                    text, re.IGNORECASE,
                )
                for url in urls:
                    # Clean trailing punctuation
                    url = url.rstrip(".,;:!?)")
                    if url not in seen_urls:
                        seen_urls.add(url)
                        ps = self.classify_url(url)
                        if ps:
                            ps.linked_from = article.get("link", article.get("url", ""))
                            candidates.append(ps)

        # Sort by confidence (highest first)
        candidates.sort(key=lambda ps: ps.confidence, reverse=True)
        return candidates[:self.max_chase_per_article]

    async def chase_article(self, article: Dict[str, Any]) -> ChaseResult:
        """Chase primary sources for a single article.

        1. Extract candidate links
        2. Fetch each via Firecrawl
        3. Return enriched primary sources
        """
        article_url = article.get("link", article.get("url", "unknown"))
        candidates = self.extract_candidate_links(article)
        result = ChaseResult(
            article_url=article_url,
            links_analyzed=len(article.get("links", [])),
        )

        for candidate in candidates:
            if candidate.url in self._chased_urls:
                continue
            self._chased_urls.add(candidate.url)

            # Fetch via Firecrawl
            try:
                scrape_result = await self.firecrawl.scrape_to_markdown(candidate.url)
                if scrape_result.success and scrape_result.markdown:
                    candidate.title = scrape_result.title or candidate.title
                    candidate.markdown = scrape_result.markdown
                    result.primary_sources.append(candidate)
                    result.links_chased += 1
                    logger.info(
                        "Source chased: %s (%s, confidence=%.2f)",
                        candidate.url, candidate.source_type, candidate.confidence,
                    )
                else:
                    candidate.error = scrape_result.error
                    logger.debug("Chase failed for %s: %s", candidate.url, scrape_result.error)
            except Exception as e:
                candidate.error = str(e)
                logger.debug("Chase error for %s: %s", candidate.url, e)

            await asyncio.sleep(self.delay_between)

        return result

    async def chase_batch(
        self,
        articles: List[Dict[str, Any]],
    ) -> List[ChaseResult]:
        """Chase primary sources for multiple articles.

        Processes sequentially to be polite to target servers.
        """
        results: List[ChaseResult] = []
        total_chased = 0

        for i, article in enumerate(articles):
            result = await self.chase_article(article)
            results.append(result)
            total_chased += result.links_chased

            if result.primary_sources:
                logger.info(
                    "Article %d/%d: %d primary sources found",
                    i + 1, len(articles), len(result.primary_sources),
                )

        logger.info(
            "Source chasing complete: %d articles → %d primary sources",
            len(articles), total_chased,
        )
        return results

    def get_stats(self) -> Dict:
        """Return chasing session stats."""
        return {
            "unique_urls_chased": len(self._chased_urls),
        }
