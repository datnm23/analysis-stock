import logging
import sys
from contextlib import asynccontextmanager

import redis.asyncio as redis
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import get_settings
from app.models.phobert import PhoBERTSentiment
from app.models.onnx_adapter import ONNXSentimentAdapter
from app.services.sentiment_analyzer import SentimentAnalyzer
from app.routers import analyze, health
from app.middleware import APIKeyMiddleware, RateLimitMiddleware
from app.models.lstm_forecast import LSTMForecast
from app.routers import forecast as forecast_router

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

    # Initialize Redis client for rate limiting
    redis_client = None
    try:
        redis_client = redis.Redis(
            host=settings.redis_host,
            port=settings.redis_port,
            password=settings.redis_password if settings.redis_password else None,
            db=settings.redis_db,
            decode_responses=False
        )
        # Test connection
        await redis_client.ping()
        logger.info(f"Redis connected for rate limiting: {settings.redis_host}:{settings.redis_port}")
    except Exception as e:
        logger.warning(f"Redis not available for rate limiting: {e}")
        redis_client = None

    # Add rate limiting middleware (after app is created)
    if redis_client:
        app.add_middleware(RateLimitMiddleware, redis_client=redis_client)

    # Load model on startup — auto-select ONNX or PyTorch backend
    if settings.use_onnx and ONNXSentimentAdapter.is_available(settings.onnx_model_dir):
        logger.info("Loading ONNX Runtime backend (low memory mode)...")
        model = ONNXSentimentAdapter(model_dir=settings.onnx_model_dir)
        # Trigger eager load so health check reflects real state
        model._load()
    else:
        if settings.use_onnx:
            logger.warning(
                "USE_ONNX=true but ONNX model not found at %s. "
                "Falling back to PyTorch. Run: python -m scripts.export_onnx --quantize",
                settings.onnx_model_dir,
            )
        logger.info("Loading PyTorch PhoBERT model...")
        model = PhoBERTSentiment(
            model_name=settings.model_name,
            cache_dir=settings.model_cache_dir
        )
    analyzer = SentimentAnalyzer(model)

    # Store analyzer in app state for dependency injection (used by routers)
    app.state.analyzer = analyzer
    health.set_model(model)

    # Optionally load LSTM forecast model
    if settings.enable_ml_forecast:
        logger.info("Loading LSTM forecast model (ENABLE_ML_FORECAST=true)...")
        forecaster = LSTMForecast(
            weights_path=settings.lstm_weights_path,
            sequence_length=settings.lstm_sequence_length,
            forecast_days=settings.lstm_forecast_days,
        )
        forecast_router.set_forecaster(forecaster)
        logger.info("LSTM forecaster ready (model_loaded=%s)", forecaster.model_loaded)

    logger.info("Model loaded, service ready")

    # NOTE: Dedup filter moved to crawl-agent service
    # Kept for backward compatibility with analyze router
    app.state.dedup_filter = None

    yield

    # Cleanup on shutdown
    if redis_client:
        await redis_client.close()
    logger.info("Shutting down sentiment service")


def _init_otel(settings) -> None:
    """Initialize OpenTelemetry TracerProvider pointing at Jaeger via OTLP HTTP.

    Called once at import time (before create_app) so the provider is ready
    when FastAPIInstrumentor instruments the app.  Non-fatal: if the Jaeger
    container is unreachable the exporter will retry silently in the background.
    """
    if not settings.otel_enabled:
        return

    from opentelemetry import trace
    from opentelemetry.sdk.resources import Resource
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor
    from opentelemetry.sdk.trace.sampling import TraceIdRatioBased
    from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter

    resource = Resource.create({"service.name": settings.otel_service_name})
    sampler = TraceIdRatioBased(settings.otel_sample_rate)
    provider = TracerProvider(resource=resource, sampler=sampler)

    exporter = OTLPSpanExporter(
        endpoint=f"{settings.otel_endpoint}/v1/traces"
    )
    provider.add_span_processor(BatchSpanProcessor(exporter))
    trace.set_tracer_provider(provider)
    logger.info("OTel tracing enabled → %s (service=%s)", settings.otel_endpoint, settings.otel_service_name)


def create_app() -> FastAPI:
    settings = get_settings()

    # Initialize OTel before app creation so FastAPIInstrumentor can attach
    _init_otel(settings)

    app = FastAPI(
        title="VN Stock Sentiment Service",
        description="Vietnamese stock market sentiment analysis using PhoBERT",
        version="1.0.0",
        docs_url="/docs" if not settings.debug else "/docs",
        redoc_url="/redoc" if not settings.debug else "/redoc",
        lifespan=lifespan
    )

    # CORS middleware - use configured origins instead of wildcard
    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.cors_origins,
        allow_credentials=True,
        allow_methods=settings.cors_methods,
        allow_headers=settings.cors_headers,
    )

    # API Key authentication middleware
    if settings.enable_auth and settings.api_key:
        app.add_middleware(APIKeyMiddleware)

    # Include routers
    app.include_router(analyze.router)
    app.include_router(health.router)

    # Forecast router only registered when feature flag is enabled
    if settings.enable_ml_forecast:
        app.include_router(forecast_router.router)

    # Instrument FastAPI: auto-creates spans for every request/response
    if settings.otel_enabled:
        from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
        FastAPIInstrumentor.instrument_app(app)
        logger.info("FastAPI OTel instrumentation active")

    return app


app = create_app()

if __name__ == "__main__":
    import uvicorn
    settings = get_settings()

    # Configure uvicorn with TLS if enabled
    uvicorn_kwargs = {
        "app": "app.main:app",
        "host": settings.host,
        "port": settings.port,
        "reload": settings.debug,
    }

    if settings.enable_https:
        uvicorn_kwargs["ssl_certfile"] = settings.tls_cert_file
        uvicorn_kwargs["ssl_keyfile"] = settings.tls_key_file
        logger.info(f"Starting with HTTPS enabled: {settings.tls_cert_file}")

    uvicorn.run(**uvicorn_kwargs)
