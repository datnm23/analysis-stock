from typing import List, Optional
from datetime import datetime

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

router = APIRouter(tags=["analyze"])

# These will be set by main.py
_analyzer = None


def set_analyzer(analyzer):
    global _analyzer
    _analyzer = analyzer


def get_analyzer():
    return _analyzer


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
    symbols: List[str]
    keywords: List[str]


class AnalyzeResponse(BaseModel):
    results: List[SentimentResultItem]
    processing_time_ms: float
    model_version: str = "phobert-base-v1"


@router.post("/analyze", response_model=AnalyzeResponse)
async def analyze_sentiment(request: AnalyzeRequest):
    """Analyze sentiment of Vietnamese texts."""
    analyzer = get_analyzer()
    if analyzer is None:
        raise HTTPException(status_code=503, detail="Model not loaded")

    start_time = datetime.now()

    texts = [item.content for item in request.texts]
    analyses = analyzer.analyze_batch(texts)

    results = []
    for i, analysis in enumerate(analyses):
        results.append(SentimentResultItem(
            id=request.texts[i].id,
            sentiment=analysis["sentiment"],
            confidence=analysis["confidence"],
            symbols=analysis["symbols"],
            keywords=analysis["keywords"]
        ))

    processing_time = (datetime.now() - start_time).total_seconds() * 1000

    return AnalyzeResponse(
        results=results,
        processing_time_ms=round(processing_time, 2)
    )


@router.post("/analyze/single")
async def analyze_single(text: str):
    """Quick endpoint for single text analysis."""
    analyzer = get_analyzer()
    if analyzer is None:
        raise HTTPException(status_code=503, detail="Model not loaded")

    result = analyzer.analyze(text)
    return result
