"""
POST /forecast/price — 5-day price prediction endpoint.

Gated by ENABLE_ML_FORECAST=true (default: false).
When LSTM weights are missing, falls back to EMA extrapolation transparently.
"""

from __future__ import annotations

import logging
from typing import List, Optional

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

from app.models.lstm_forecast import LSTMForecast

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/forecast", tags=["forecast"])

# Module-level reference set during app startup (see main.py)
_forecaster: Optional[LSTMForecast] = None


def set_forecaster(f: LSTMForecast) -> None:
    global _forecaster
    _forecaster = f


# ---------------------------------------------------------------------------
# Request / Response models
# ---------------------------------------------------------------------------

class PriceForecastRequest(BaseModel):
    symbol: str = Field(..., description="Stock symbol, e.g. VNM", pattern=r"^[A-Z]{2,10}$")
    history: List[float] = Field(
        ...,
        description="Recent close prices (newest last). At least 5 required; 60+ recommended for LSTM.",
        min_length=5,
        max_length=500,
    )


class PriceForecastResponse(BaseModel):
    symbol: str
    predictions: List[float] = Field(..., description="Predicted close prices for the next N days")
    forecast_days: int
    confidence: float = Field(..., description="Model confidence 0–1")
    method: str = Field(..., description="'lstm' or 'ema_fallback'")
    model_loaded: bool


# ---------------------------------------------------------------------------
# Endpoints
# ---------------------------------------------------------------------------

@router.post("/price", response_model=PriceForecastResponse, summary="5-day price forecast")
async def forecast_price(request: PriceForecastRequest) -> PriceForecastResponse:
    """
    Predict the next 5 trading days' close prices for a given symbol.

    - Uses a 2-layer LSTM (64 hidden units) when pre-trained weights are available.
    - Automatically falls back to EMA extrapolation when the model has not been trained yet.
    - History should be in chronological order (oldest → newest).
    """
    if _forecaster is None:
        raise HTTPException(status_code=503, detail="Forecast model not initialised")

    try:
        result = _forecaster.predict(request.history)
    except ValueError as exc:
        raise HTTPException(status_code=422, detail=str(exc))
    except Exception as exc:
        logger.exception("Forecast error for %s", request.symbol)
        raise HTTPException(status_code=500, detail="Forecast failed: " + str(exc))

    return PriceForecastResponse(
        symbol=request.symbol,
        predictions=result["predictions"],
        forecast_days=len(result["predictions"]),
        confidence=result["confidence"],
        method=result["method"],
        model_loaded=_forecaster.model_loaded,
    )


@router.get("/status", summary="Forecast model status")
async def forecast_status() -> dict:
    """Returns whether the LSTM weights are loaded or EMA fallback is active."""
    if _forecaster is None:
        return {"enabled": False, "model_loaded": False}
    return {
        "enabled": True,
        "model_loaded": _forecaster.model_loaded,
        "method": "lstm" if _forecaster.model_loaded else "ema_fallback",
        "sequence_length": _forecaster.sequence_length,
        "forecast_days": _forecaster.forecast_days,
        "weights_path": _forecaster.weights_path,
    }
