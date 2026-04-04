"""
Crawl Agent — Vietnamese financial news aggregation service.

Collects news from RSS feeds, web crawlers (aiohttp), Firecrawl (self-hosted),
and Telegram channels. Processes through dedup filter, symbol detection,
and source scoring before pushing to sentiment service for analysis.
"""
