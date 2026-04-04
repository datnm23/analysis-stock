"""
Telegram channel scraper for stock market sentiment.

Uses Telethon to connect to public Telegram channels and extract
recent messages for sentiment analysis.
"""

import asyncio
import logging
import re
from dataclasses import dataclass, field
from datetime import datetime, timedelta, timezone
from typing import List, Optional

logger = logging.getLogger(__name__)

# Default channels to monitor
DEFAULT_CHANNELS = [
    "chungkhoanvietnam",
    "stockvietnam",
    "vnstock_analysis",
]


@dataclass
class TelegramMessage:
    """A parsed Telegram message."""
    text: str
    channel: str
    message_id: int
    date: datetime
    views: int = 0
    source: str = "telegram_unknown"
    symbols: List[str] = field(default_factory=list)

    def __post_init__(self):
        if not self.symbols:
            self.symbols = re.findall(r'\b[A-Z]{3}\b', self.text or "")

    def to_dict(self) -> dict:
        return {
            "id": f"tg_{self.channel}_{self.message_id}",
            "content": self.text,
            "source": self.source,
            "channel": self.channel,
            "views": self.views,
            "symbols": self.symbols,
            "published_at": self.date.isoformat() if self.date else None,
        }


class TelegramScraper:
    """Scrape public Telegram channels for stock-related messages.

    Requires Telethon and a Telegram API ID/hash. These can be obtained
    from https://my.telegram.org.

    Usage:
        scraper = TelegramScraper(api_id=12345, api_hash="abc123")
        messages = await scraper.scrape_channels()
    """

    def __init__(
        self,
        api_id: int,
        api_hash: str,
        session_name: str = "vnstock_scraper",
        channels: Optional[List[str]] = None,
    ):
        self.api_id = api_id
        self.api_hash = api_hash
        self.session_name = session_name
        self.channels = channels or DEFAULT_CHANNELS
        self._client = None

    async def _get_client(self):
        """Lazy-initialize Telethon client."""
        if self._client is None:
            try:
                from telethon import TelegramClient
                self._client = TelegramClient(
                    self.session_name, self.api_id, self.api_hash
                )
                await self._client.start()
                logger.info("Telegram client connected")
            except ImportError:
                logger.error(
                    "Telethon not installed. Install with: pip install telethon"
                )
                raise
        return self._client

    async def scrape_channel(
        self,
        channel: str,
        limit: int = 50,
        hours_back: int = 24,
    ) -> List[TelegramMessage]:
        """Scrape recent messages from a single channel.

        Args:
            channel: Channel username (without @)
            limit: Max messages to fetch
            hours_back: Only fetch messages from the last N hours
        """
        client = await self._get_client()
        cutoff = datetime.now(timezone.utc) - timedelta(hours=hours_back)
        messages: List[TelegramMessage] = []

        try:
            async for msg in client.iter_messages(channel, limit=limit):
                if not msg.text:
                    continue
                if msg.date and msg.date < cutoff:
                    break  # Messages are chronological, stop when too old

                messages.append(TelegramMessage(
                    text=msg.text,
                    channel=channel,
                    message_id=msg.id,
                    date=msg.date,
                    views=msg.views or 0,
                    source=f"telegram_group",
                ))
            logger.info("Scraped %d messages from @%s", len(messages), channel)
        except Exception as e:
            logger.warning("Failed to scrape @%s: %s", channel, e)

        return messages

    async def scrape_channels(
        self,
        limit_per_channel: int = 50,
        hours_back: int = 24,
        delay_between: float = 2.0,
    ) -> List[TelegramMessage]:
        """Scrape all configured channels with rate limiting."""
        all_messages: List[TelegramMessage] = []

        for channel in self.channels:
            msgs = await self.scrape_channel(channel, limit_per_channel, hours_back)
            all_messages.extend(msgs)
            await asyncio.sleep(delay_between)

        logger.info("Total Telegram messages scraped: %d", len(all_messages))
        return all_messages

    async def close(self):
        """Disconnect the client."""
        if self._client:
            await self._client.disconnect()
            self._client = None
