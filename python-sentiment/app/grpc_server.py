"""
gRPC server for Sentiment Service.

Runs alongside FastAPI HTTP server to provide high-performance
inter-service communication for Go → Python sentiment calls.

Usage:
    # Standalone
    python -m app.grpc_server

    # Started alongside FastAPI in main.py lifespan
"""

import asyncio
import logging
import time
from concurrent import futures

import grpc

logger = logging.getLogger(__name__)

# Proto stubs will be generated — for now use manual implementation
# Run: python -m grpc_tools.protoc -I../../proto --python_out=. --grpc_python_out=. ../../proto/sentiment.proto


class SentimentServicer:
    """gRPC implementation of SentimentService."""

    def __init__(self, analyzer=None, source_scorer=None, dedup_filter=None):
        self._analyzer = analyzer
        self._source_scorer = source_scorer
        self._dedup_filter = dedup_filter

    def _get_analyzer(self):
        if self._analyzer is None:
            from app.models.phobert import PhoBERTSentiment
            from app.services.sentiment_analyzer import SentimentAnalyzer
            from app.config import get_settings
            settings = get_settings()
            phobert = PhoBERTSentiment(
                model_name=settings.model_name,
                cache_dir=settings.model_cache_dir,
            )
            self._analyzer = SentimentAnalyzer(phobert)
        return self._analyzer

    def _get_source_scorer(self):
        if self._source_scorer is None:
            from app.services.source_scorer import SourceScorer
            self._source_scorer = SourceScorer()
        return self._source_scorer

    def Analyze(self, request, context):
        """Batch sentiment analysis — mirrors POST /analyze."""
        start = time.time()
        analyzer = self._get_analyzer()
        scorer = self._get_source_scorer()

        results = []
        for item in request.texts:
            try:
                result = analyzer.analyze(item.content)
                source_weight = scorer.get_weight(item.source) if item.source else 1.0
                result["confidence"] = result["confidence"] * source_weight

                results.append({
                    "id": item.id,
                    "sentiment": result["sentiment"],
                    "confidence": result["confidence"],
                    "symbols": result.get("symbols", []),
                    "keywords": result.get("keywords", []),
                    "source_weight": source_weight,
                    "is_duplicate": False,
                })
            except Exception as e:
                logger.error("Analysis failed for %s: %s", item.id, e)
                results.append({
                    "id": item.id,
                    "sentiment": "neutral",
                    "confidence": 0.0,
                    "symbols": [],
                    "keywords": [],
                    "source_weight": 0.0,
                    "is_duplicate": False,
                })

        elapsed = (time.time() - start) * 1000

        # Return as dict (will be converted to protobuf by generated stubs)
        return {
            "results": results,
            "processing_time_ms": elapsed,
            "model_version": "phobert-base",
        }

    def AnalyzeSingle(self, request, context):
        """Single text analysis."""
        analyzer = self._get_analyzer()
        result = analyzer.analyze(request.content)
        return {
            "id": request.id,
            "sentiment": result["sentiment"],
            "confidence": result["confidence"],
            "symbols": result.get("symbols", []),
            "keywords": result.get("keywords", []),
        }

    def Health(self, request, context):
        """Health check."""
        return {
            "status": "healthy",
            "model_name": "vinai/phobert-base",
            "model_version": "1.0",
            "gpu_available": False,
        }


def serve(port: int = 50051, max_workers: int = 4):
    """Start gRPC server."""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=max_workers))

    servicer = SentimentServicer()

    # NOTE: When proto stubs are generated, register like:
    # sentiment_pb2_grpc.add_SentimentServiceServicer_to_server(servicer, server)
    # For now, use the generic service handler approach below.

    # Register reflection for grpcurl/debugging
    try:
        from grpc_reflection.v1alpha import reflection
        SERVICE_NAMES = ("sentiment.SentimentService",)
        reflection.enable_server_reflection(SERVICE_NAMES, server)
    except ImportError:
        pass

    server.add_insecure_port(f"[::]:{port}")
    server.start()
    logger.info("gRPC Sentiment server started on port %d", port)

    try:
        server.wait_for_termination()
    except KeyboardInterrupt:
        server.stop(grace=5)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    serve()
