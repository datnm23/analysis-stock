from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    # Server
    host: str = "0.0.0.0"
    port: int = 8000
    debug: bool = False

    # Model
    model_name: str = "vinai/phobert-base"
    model_cache_dir: str = "./.cache"
    max_sequence_length: int = 256

    # Redis
    redis_host: str = "localhost"
    redis_port: int = 6379
    redis_password: str = ""
    redis_db: int = 0

    # GCP (for Pub/Sub)
    gcp_project_id: str = ""
    pubsub_subscription: str = "sentiment-requests-sub"
    pubsub_result_topic: str = "sentiment-results"

    class Config:
        env_file = ".env"


@lru_cache()
def get_settings() -> Settings:
    return Settings()
