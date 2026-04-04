import asyncio
import logging
from typing import List, Optional
from datetime import datetime

from fastapi import APIRouter, HTTPException, Request
from pydantic import BaseModel, Field

from app.services.source_scorer import SourceScorer
from app.services.dedup_filter import DedupFilter

logger = logging.getLogger(__name__)
router = APIRouter(tags=["analyze"])

# Module-level singletons (stateless / lightweight)
_source_scorer = SourceScorer()


def get_analyzer(request: Request):
    """FastAPI dependency: retrieve analyzer from app state (set in main.py)."""
    analyzer = getattr(request.app.state, "analyzer", None)
    if analyzer is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    return analyzer


def get_dedup(request: Request) -> Optional[DedupFilter]:
    """Retrieve optional dedup filter from app state."""
    return getattr(request.app.state, "dedup_filter", None)


# ---------------------------------------------------------------------------
# Request / Response models
# ---------------------------------------------------------------------------

class TextItem(BaseModel):
    id: str
    content: str
    source: Optional[str] = None
    published_at: Optional[datetime] = None


class AnalyzeRequest(BaseModel):
    texts: List[TextItem] = Field(..., min_length=1, max_length=100)


class SentimentResultItem(BaseModel):
    id: str
    sentiment: str
    confidence: float
    source_weight: float = 1.0
    is_duplicate: bool = False
    symbols: List[str]
    keywords: List[str]


class AnalyzeResponse(BaseModel):
    results: List[SentimentResultItem]
    processing_time_ms: float
    model_version: str = "phobert-base-v1"
    duplicates_filtered: int = 0


# ---------------------------------------------------------------------------
# Endpoints
# ---------------------------------------------------------------------------

@router.post("/analyze", response_model=AnalyzeResponse)
async def analyze_sentiment(request: AnalyzeRequest, http_request: Request):
    """Analyze sentiment of Vietnamese texts with source credibility and dedup."""
    analyzer = get_analyzer(http_request)
    dedup = get_dedup(http_request)

    start_time = datetime.now()

    # --- Phase 2a: Dedup filter ---
    dup_flags: list[bool] = []
    for item in request.texts:
        if dedup:
            is_dup, _ = await dedup.is_duplicate(item.content, symbol=item.id)
            dup_flags.append(is_dup)
        else:
            dup_flags.append(False)

    # Only analyze non-duplicate texts
    texts_to_analyze = [
        item.content for item, is_dup in zip(request.texts, dup_flags) if not is_dup
    ]
    idx_map = [i for i, is_dup in enumerate(dup_flags) if not is_dup]

    # --- ML Inference ---
    analyses = []
    if texts_to_analyze:
        loop = asyncio.get_event_loop()
        analyses = await loop.run_in_executor(None, analyzer.analyze_batch, texts_to_analyze)

    # --- Build results ---
    analysis_iter = iter(analyses)
    results = []
    duplicates_filtered = 0

    for i, item in enumerate(request.texts):
        if dup_flags[i]:
            # Duplicate → return neutral with zero weight
            duplicates_filtered += 1
            results.append(SentimentResultItem(
                id=item.id,
                sentiment="neutral",
                confidence=0.0,
                source_weight=0.0,
                is_duplicate=True,
                symbols=[],
                keywords=[],
            ))
        else:
            analysis = next(analysis_iter)
            source_weight = _source_scorer.get_weight(item.source)

            # Apply source weight to confidence
            adjusted_confidence = analysis["confidence"] * source_weight

            results.append(SentimentResultItem(
                id=item.id,
                sentiment=analysis["sentiment"],
                confidence=round(adjusted_confidence, 2),
                source_weight=round(source_weight, 2),
                is_duplicate=False,
                symbols=analysis["symbols"],
                keywords=analysis["keywords"],
            ))

    processing_time = (datetime.now() - start_time).total_seconds() * 1000

    return AnalyzeResponse(
        results=results,
        processing_time_ms=round(processing_time, 2),
        duplicates_filtered=duplicates_filtered,
    )


@router.post("/analyze/single")
async def analyze_single(text: str, request: Request, source: Optional[str] = None):
    """Quick endpoint for single text analysis with source scoring."""
    analyzer = get_analyzer(request)

    loop = asyncio.get_event_loop()
    result = await loop.run_in_executor(None, analyzer.analyze, text)

    source_weight = _source_scorer.get_weight(source)
    result["source_weight"] = round(source_weight, 2)
    result["confidence"] = round(result["confidence"] * source_weight, 2)

    return result


# ---------------------------------------------------------------------------
# Rumor Detection
# ---------------------------------------------------------------------------

class RumorRequest(BaseModel):
    texts: List[TextItem] = Field(..., min_length=1, max_length=200)


@router.post("/analyze/rumors")
async def detect_rumors(request: RumorRequest, http_request: Request):
    """Detect potential stock rumors by comparing social vs official source mentions."""
    from app.services.rumor_detector import RumorDetector, TextWithSource

    detector = RumorDetector(_source_scorer)
    items = [
        TextWithSource(
            text=item.content,
            source=item.source or "unknown",
        )
        for item in request.texts
    ]

    loop = asyncio.get_event_loop()
    result = await loop.run_in_executor(None, detector.detect, items)
    return result


# ---------------------------------------------------------------------------
# NOTE: All /scrape/* and /scheduler/* endpoints have been moved to
# the crawl-agent service (port 8085). See crawl-agent/app/routers/scrape.py.
#
# Removed endpoints:
#   POST /scrape/rss       → crawl-agent:8085/scrape/rss
#   POST /scrape/telegram  → crawl-agent:8085/scrape/telegram
#   POST /scrape/cafef     → crawl-agent:8085/scrape/web
#   POST /scrape/web       → crawl-agent:8085/scrape/web
#   POST /scrape/all       → crawl-agent:8085/scrape/all
#   POST /scrape/analyze   → crawl-agent:8085/scrape/pipeline
#   POST /scheduler/start  → crawl-agent:8085/scheduler/start
#   POST /scheduler/stop   → crawl-agent:8085/scheduler/stop
#   GET  /scheduler/status → crawl-agent:8085/scheduler/status
# ---------------------------------------------------------------------------
