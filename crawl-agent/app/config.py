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

    # Redis news cache
    news_redis_max_items: int = 20
    news_redis_ttl_hours: int = 24

    # Article generation — text models
    anthropic_api_key: str = ""
    article_max_daily: int = 10
    article_hot_symbols_count: int = 10
    article_claude_model: str = "claude-haiku-4-5-20251001"
    article_model: str = "claude"  # claude | gemini | auto
    gemini_api_key: str = ""
    gemini_text_model: str = "gemini-2.0-flash"
    go_services_internal_url: str = "http://go-services:8080"
    go_services_internal_key: str = ""
    dashboard_url: str = "http://localhost:3000"
    telegram_admin_chat_id: str = ""

    # Article generation — image pipeline (optional)
    enable_image_generation: bool = False
    gemini_image_model: str = "gemini-2.0-flash-preview-image-generation"
    image_storage_backend: str = "s3"  # s3 | gdrive
    # S3/MinIO backend
    s3_endpoint: str = ""
    s3_bucket: str = "blog-images"
    s3_access_key: str = ""
    s3_secret_key: str = ""
    s3_public_url: str = ""
    # Google Drive backend
    gdrive_credentials_json: str = ""  # service account JSON as string
    gdrive_folder_id: str = ""         # folder ID to upload images into

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
