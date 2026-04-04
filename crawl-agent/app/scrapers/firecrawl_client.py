"""
Firecrawl self-hosted API client.

Calls the local Firecrawl instance (port 3002) to:
1. Render JavaScript-heavy pages via Playwright
2. Extract main article content (strip nav/footer/ads)
3. Convert to clean Markdown

Reference: lucaswalter/n8n-ai-workflows firecrawl_scrape_url pattern.

Usage:
    client = FirecrawlClient("http://firecrawl-api:3002")
    result = await client.scrape_to_markdown("https://vnexpress.net/...")
    print(result["markdown"])
"""

import asyncio
import logging
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional

import httpx

logger = logging.getLogger(__name__)

# Default Firecrawl scrape config (based on reference workflow)
DEFAULT_SCRAPE_CONFIG = {
    "formats": ["markdown", "links"],
    "onlyMainContent": True,
    "excludeTags": ["iframe", "nav", "header", "footer", "aside"],
    "blockAds": True,
}


@dataclass
class ScrapeResult:
    """Result from Firecrawl scrape."""
    url: str
    markdown: str
    links: List[str] = field(default_factory=list)
    title: str = ""
    description: str = ""
    source_url: str = ""
    success: bool = True
    error: str = ""

    def to_dict(self) -> dict:
        return {
            "url": self.url,
            "markdown": self.markdown,
            "title": self.title,
            "description": self.description,
            "links": self.links,
            "success": self.success,
        }


class FirecrawlClient:
    """Client for self-hosted Firecrawl API.

    Firecrawl renders pages via Playwright, strips non-essential content,
    and returns clean Markdown.
    """

    def __init__(
        self,
        base_url: str = "http://firecrawl-api:3002",
        timeout: int = 60,
        max_retries: int = 3,
        retry_delay: float = 5.0,
    ):
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.max_retries = max_retries
        self.retry_delay = retry_delay

    async def scrape_to_markdown(
        self,
        url: str,
        extra_config: Optional[Dict[str, Any]] = None,
    ) -> ScrapeResult:
        """Scrape a single URL and return Markdown content.

        Args:
            url: The page URL to scrape.
            extra_config: Override/extend default scrape config.

        Returns:
            ScrapeResult with markdown content and extracted links.
        """
        config = {**DEFAULT_SCRAPE_CONFIG, **(extra_config or {})}
        config["url"] = url

        for attempt in range(self.max_retries):
            try:
                async with httpx.AsyncClient(timeout=self.timeout) as client:
                    resp = await client.post(
                        f"{self.base_url}/v1/scrape",
                        json=config,
                        headers={"Content-Type": "application/json"},
                    )

                    if resp.status_code == 200:
                        data = resp.json().get("data", {})
                        return ScrapeResult(
                            url=url,
                            markdown=data.get("markdown", ""),
                            links=data.get("links", []),
                            title=data.get("metadata", {}).get("title", ""),
                            description=data.get("metadata", {}).get("description", ""),
                            source_url=data.get("metadata", {}).get("sourceURL", url),
                        )
                    else:
                        error_msg = f"HTTP {resp.status_code}: {resp.text[:200]}"
                        logger.warning(
                            "Firecrawl scrape attempt %d/%d failed for %s: %s",
                            attempt + 1, self.max_retries, url, error_msg,
                        )

            except httpx.TimeoutException:
                logger.warning(
                    "Firecrawl timeout attempt %d/%d for %s",
                    attempt + 1, self.max_retries, url,
                )
            except Exception as e:
                logger.warning(
                    "Firecrawl error attempt %d/%d for %s: %s",
                    attempt + 1, self.max_retries, url, e,
                )

            if attempt < self.max_retries - 1:
                await asyncio.sleep(self.retry_delay)

        return ScrapeResult(
            url=url, markdown="", success=False,
            error=f"All {self.max_retries} attempts failed",
        )

    async def scrape_batch(
        self,
        urls: List[str],
        max_concurrent: int = 3,
        delay_between: float = 2.0,
    ) -> List[ScrapeResult]:
        """Scrape multiple URLs with concurrency control.

        Args:
            urls: List of URLs to scrape.
            max_concurrent: Max concurrent Firecrawl requests.
            delay_between: Seconds between requests.

        Returns:
            List of ScrapeResult for each URL.
        """
        semaphore = asyncio.Semaphore(max_concurrent)
        results: List[ScrapeResult] = []

        async def _scrape_one(url: str) -> ScrapeResult:
            async with semaphore:
                result = await self.scrape_to_markdown(url)
                await asyncio.sleep(delay_between)
                return result

        tasks = [_scrape_one(url) for url in urls]
        results = await asyncio.gather(*tasks)

        success_count = sum(1 for r in results if r.success)
        logger.info(
            "Firecrawl batch: %d/%d succeeded for %d URLs",
            success_count, len(urls), len(urls),
        )
        return list(results)

    async def health_check(self) -> bool:
        """Check if Firecrawl API is reachable."""
        try:
            async with httpx.AsyncClient(timeout=5) as client:
                resp = await client.get(f"{self.base_url}/")
                return resp.status_code == 200
        except Exception:
            return False
