"""
Crawl Agent configuration — env-based settings (12-factor).
"""

from functools import lru_cache
from typing import List, Optional

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Crawl agent settings loaded from environment variables."""

    # Server
    host: str = "0.0.0.0"
    port: int = 8085
    debug: bool = False

    # Firecrawl self-hosted
    firecrawl_api_url: str = "http://firecrawl-api:3002"
    firecrawl_timeout: int = 60  # seconds per scrape

    # Sentiment service (downstream)
    sentiment_service_url: str = "http://sentiment:8000"

    # Redis
    redis_host: str = "localhost"
    redis_port: int = 6379
    redis_password: str = ""
    redis_db: int = 0

    # Scheduler
    scrape_interval_minutes: int = 30
    scrape_market_hours_only: bool = True

    # Telegram
    telegram_app_api_id: int = 0
    telegram_app_api_hash: str = ""
    telegram_session_name: str = "vnstock_crawl"
    telegram_channels: str = "chungkhoanUG,ChungkhoanGalaxy,finbotrealtimenews"

    # Feature flags
    enable_firecrawl: bool = True
    enable_telegram: bool = True
    enable_source_chasing: bool = False  # Phase 2

    # OpenTelemetry
    otel_enabled: bool = False
    otel_endpoint: str = "http://jaeger:4318"
    otel_service_name: str = "crawl-agent"
    otel_sample_rate: float = 1.0

    # CORS
    cors_origins: List[str] = ["*"]

    @property
    def telegram_channel_list(self) -> List[str]:
        """Parse comma-separated channel string into list."""
        return [c.strip() for c in self.telegram_channels.split(",") if c.strip()]

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = False


@lru_cache()
def get_settings() -> Settings:
    return Settings()
