from typing import List
from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    # Server
    host: str = "0.0.0.0"
    port: int = 8000
    debug: bool = False

    # TLS/HTTPS
    enable_https: bool = False
    tls_cert_file: str = "certs/server.crt"
    tls_key_file: str = "certs/server.key"

    # Model
    model_name: str = "vinai/phobert-base"
    model_cache_dir: str = "./.cache"
    max_sequence_length: int = 256

    # ONNX Runtime (reduces memory from ~4GB to ~1GB)
    use_onnx: bool = False
    onnx_model_dir: str = "models/onnx"

    # Redis
    redis_host: str = "localhost"
    redis_port: int = 6379
    redis_password: str = ""
    redis_db: int = 0

    # GCP (for Pub/Sub)
    gcp_project_id: str = ""
    pubsub_subscription: str = "sentiment-requests-sub"
    pubsub_result_topic: str = "sentiment-results"

    # ML Forecast (LSTM)
    enable_ml_forecast: bool = False
    lstm_weights_path: str = "models/lstm_forecast.pt"
    lstm_sequence_length: int = 60
    lstm_forecast_days: int = 5

    # Security
    api_key: str = ""  # API key for service-to-service auth
    enable_auth: bool = False

    # NOTE: Telegram/scraper config moved to crawl-agent service

    # Rate limiting
    rate_limit_requests: int = 100  # requests per window
    rate_limit_window: int = 60  # window in seconds

    # CORS
    cors_allowed_origins: str = "*"
    cors_allowed_methods: str = "GET,POST,PUT,DELETE,OPTIONS"
    cors_allowed_headers: str = "Authorization,Content-Type,X-API-Key"

    # OpenTelemetry tracing
    otel_enabled: bool = False
    # host:port of OTLP HTTP collector — matches Jaeger container in docker-compose
    otel_endpoint: str = "http://jaeger:4318"
    otel_service_name: str = "sentiment-service"
    otel_sample_rate: float = 1.0

    class Config:
        env_file = ".env"

    @property
    def cors_origins(self) -> List[str]:
        if self.cors_allowed_origins == "*":
            return ["*"]
        return [origin.strip() for origin in self.cors_allowed_origins.split(",")]

    @property
    def cors_methods(self) -> List[str]:
        return [method.strip() for method in self.cors_allowed_methods.split(",")]

    @property
    def cors_headers(self) -> List[str]:
        return [header.strip() for header in self.cors_allowed_headers.split(",")]


@lru_cache()
def get_settings() -> Settings:
    return Settings()
