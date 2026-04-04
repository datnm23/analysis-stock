"""
LSTM-based price forecast model for Vietnamese stock market.

Architecture:
    2-layer LSTM (hidden=64) → Linear → 5-day close price predictions

Usage:
    model = LSTMForecast(weights_path="models/lstm_forecast.pt")
    preds = model.predict(history_closes)   # list of ~60 close prices
    # → {"predictions": [...], "method": "lstm"|"ema_fallback"}

When weights_path does not exist the model gracefully falls back to
an Exponential Moving Average (EMA) extrapolation so the endpoint is
immediately usable before training data is available.
"""

from __future__ import annotations

import logging
import math
import os
from typing import List, Optional

import numpy as np
import torch
import torch.nn as nn

logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Neural network architecture
# ---------------------------------------------------------------------------

class _LSTMNet(nn.Module):
    """Two-layer LSTM that maps a sequence of prices to 5-step predictions."""

    def __init__(
        self,
        input_size: int = 1,
        hidden_size: int = 64,
        num_layers: int = 2,
        output_size: int = 5,
        dropout: float = 0.2,
    ) -> None:
        super().__init__()
        self.lstm = nn.LSTM(
            input_size,
            hidden_size,
            num_layers=num_layers,
            batch_first=True,
            dropout=dropout if num_layers > 1 else 0.0,
        )
        self.fc = nn.Linear(hidden_size, output_size)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        # x: (batch, seq_len, input_size)
        out, _ = self.lstm(x)
        return self.fc(out[:, -1, :])   # take last time-step


# ---------------------------------------------------------------------------
# Public wrapper
# ---------------------------------------------------------------------------

class LSTMForecast:
    """
    Load (or skip) an LSTM model and expose a predict() interface.

    Parameters
    ----------
    weights_path : str
        Path to a .pt file saved with torch.save(model.state_dict(), ...).
        If the file does not exist, EMA fallback is used automatically.
    sequence_length : int
        Number of historical close prices fed into the LSTM.
    forecast_days : int
        Number of future days to predict.
    device : str or None
        "cuda" / "cpu" / None (auto-detect).
    """

    def __init__(
        self,
        weights_path: str = "models/lstm_forecast.pt",
        sequence_length: int = 60,
        forecast_days: int = 5,
        device: Optional[str] = None,
    ) -> None:
        self.sequence_length = sequence_length
        self.forecast_days = forecast_days
        self.weights_path = weights_path
        self.model_loaded = False

        self.device = torch.device(
            device if device else ("cuda" if torch.cuda.is_available() else "cpu")
        )

        self._net = _LSTMNet(output_size=forecast_days).to(self.device)

        if os.path.exists(weights_path):
            try:
                state = torch.load(weights_path, map_location=self.device)
                self._net.load_state_dict(state)
                self._net.eval()
                self.model_loaded = True
                logger.info("LSTM forecast model loaded from %s", weights_path)
            except Exception as exc:
                logger.warning("Failed to load LSTM weights (%s) — using EMA fallback: %s", weights_path, exc)
        else:
            logger.info(
                "LSTM weights not found at %s — using EMA fallback until model is trained",
                weights_path,
            )

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def predict(self, history: List[float]) -> dict:
        """
        Predict the next `forecast_days` close prices.

        Parameters
        ----------
        history : list[float]
            Recent close prices (at least sequence_length values recommended).

        Returns
        -------
        dict with keys:
            predictions  – list[float] of length forecast_days
            confidence   – float 0–1 (lower for EMA fallback)
            method       – "lstm" | "ema_fallback"
        """
        if len(history) < 5:
            raise ValueError(f"Need at least 5 historical prices, got {len(history)}")

        if self.model_loaded and len(history) >= self.sequence_length:
            return self._lstm_predict(history)
        return self._ema_predict(history)

    # ------------------------------------------------------------------
    # Private helpers
    # ------------------------------------------------------------------

    def _lstm_predict(self, history: List[float]) -> dict:
        seq = np.array(history[-self.sequence_length:], dtype=np.float32)

        # Min-max normalise within the window to stabilise inference
        lo, hi = seq.min(), seq.max()
        if hi == lo:
            hi = lo + 1.0
        norm = (seq - lo) / (hi - lo)

        tensor = torch.tensor(norm, dtype=torch.float32).unsqueeze(0).unsqueeze(-1).to(self.device)

        with torch.no_grad():
            out = self._net(tensor).squeeze().cpu().numpy()

        # Denormalise
        predictions = (out * (hi - lo) + lo).tolist()

        # Confidence heuristic: higher when recent volatility is low
        recent_std = float(np.std(history[-20:]))
        recent_mean = float(np.mean(history[-20:]))
        cv = recent_std / (recent_mean + 1e-9)
        confidence = max(0.40, min(0.90, 1.0 - cv * 2))

        return {"predictions": predictions, "confidence": round(confidence, 3), "method": "lstm"}

    def _ema_predict(self, history: List[float]) -> dict:
        """
        Simple EMA-based extrapolation used when LSTM weights are not available.
        Computes a short EMA trend and extrapolates linearly.
        """
        span = min(20, len(history))
        alpha = 2.0 / (span + 1)

        ema = history[0]
        for price in history:
            ema = alpha * price + (1 - alpha) * ema

        # Estimate recent trend (slope over last 5 prices)
        recent = history[-5:]
        slope = (recent[-1] - recent[0]) / max(len(recent) - 1, 1)

        predictions = [ema + slope * (i + 1) for i in range(self.forecast_days)]

        return {"predictions": [round(p, 2) for p in predictions], "confidence": 0.35, "method": "ema_fallback"}
