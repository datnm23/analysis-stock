import logging
import sys
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import get_settings
from app.models.phobert import PhoBERTSentiment
from app.services.sentiment_analyzer import SentimentAnalyzer
from app.routers import analyze, health

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)]
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Startup and shutdown events."""
    settings = get_settings()

    # Load model on startup
    logger.info("Loading PhoBERT model...")
    model = PhoBERTSentiment(
        model_name=settings.model_name,
        cache_dir=settings.model_cache_dir
    )
    analyzer = SentimentAnalyzer(model)

    # Set references in routers
    analyze.set_analyzer(analyzer)
    health.set_model(model)

    logger.info("Model loaded, service ready")

    yield

    # Cleanup on shutdown
    logger.info("Shutting down sentiment service")


def create_app() -> FastAPI:
    app = FastAPI(
        title="VN Stock Sentiment Service",
        description="Vietnamese stock market sentiment analysis using PhoBERT",
        version="1.0.0",
        lifespan=lifespan
    )

    # CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # Include routers
    app.include_router(analyze.router)
    app.include_router(health.router)

    return app


app = create_app()

if __name__ == "__main__":
    import uvicorn
    settings = get_settings()
    uvicorn.run(
        "app.main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug
    )
