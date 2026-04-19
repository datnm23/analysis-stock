"""Generate candlestick + volume charts from KBS Securities OHLCV data."""
import asyncio
import io
import logging
import os
from datetime import datetime, timedelta
from pathlib import Path
from typing import Optional

import httpx
import matplotlib
matplotlib.use("Agg")  # headless rendering
import matplotlib.pyplot as plt
import mplfinance as mpf
import pandas as pd

logger = logging.getLogger(__name__)

_KBS_BASE = "https://kbbuddywts.kbsec.com.vn/iis-server/investment/stocks"
_CHART_STYLE = mpf.make_mpf_style(
    base_mpf_style="nightclouds",
    rc={
        "font.family": "DejaVu Sans",
        "axes.facecolor": "#1a1a2e",
        "figure.facecolor": "#0f0f1a",
        "axes.labelcolor": "#c8c8d4",
        "xtick.color": "#c8c8d4",
        "ytick.color": "#c8c8d4",
        "grid.color": "#2a2a3e",
        "axes.edgecolor": "#2a2a3e",
    },
    marketcolors=mpf.make_marketcolors(
        up="#00d4aa", down="#ff4444",
        edge="inherit", wick="inherit",
        volume={"up": "#00d4aa66", "down": "#ff444466"},
        ohlc="inherit",
    ),
)


async def _fetch_ohlcv(symbol: str, days: int = 90) -> Optional[pd.DataFrame]:
    """Fetch OHLCV from KBS API and return a DataFrame indexed by date."""
    end = datetime.now()
    start = end - timedelta(days=days)
    url = (
        f"{_KBS_BASE}/{symbol}/data_day"
        f"?sdate={start.strftime('%d-%m-%Y')}&edate={end.strftime('%d-%m-%Y')}"
    )
    try:
        async with httpx.AsyncClient(timeout=15) as client:
            resp = await client.get(url, headers={"User-Agent": "Mozilla/5.0"})
            if resp.status_code != 200:
                logger.warning("KBS API %d for %s", resp.status_code, symbol)
                return None
            data = resp.json().get("data_day", [])
            if not data:
                return None
        rows = []
        for bar in data:
            t = bar.get("t", "")
            try:
                dt = datetime.strptime(t[:10], "%Y-%m-%d")
            except ValueError:
                continue
            rows.append({
                "Date": dt,
                "Open": float(bar["o"]),
                "High": float(bar["h"]),
                "Low": float(bar["l"]),
                "Close": float(bar["c"]),
                "Volume": int(bar["v"]),
            })
        if not rows:
            return None
        df = pd.DataFrame(rows).sort_values("Date").set_index("Date")
        df.index = pd.DatetimeIndex(df.index)
        return df
    except Exception as exc:
        logger.warning("fetch_ohlcv failed for %s: %s", symbol, exc)
        return None


def _render_chart(df: pd.DataFrame, symbol: str) -> bytes:
    """Render candlestick+volume chart to PNG bytes."""
    # Add SMA overlays
    sma20 = mpf.make_addplot(
        df["Close"].rolling(20).mean(),
        color="#4fc3f7", width=1.2, label="SMA20",
    )
    sma50 = mpf.make_addplot(
        df["Close"].rolling(50).mean(),
        color="#ffb74d", width=1.2, label="SMA50",
    )

    buf = io.BytesIO()
    mpf.plot(
        df,
        type="candle",
        style=_CHART_STYLE,
        addplot=[sma20, sma50],
        volume=True,
        title=f"\n  {symbol} — Biểu đồ giá & khối lượng (90 ngày)",
        ylabel="Giá (VNĐ)",
        ylabel_lower="KL",
        figsize=(12, 6),
        tight_layout=True,
        savefig=dict(fname=buf, format="png", dpi=120, bbox_inches="tight"),
    )
    buf.seek(0)
    return buf.read()


def _resolve_blog_public() -> Path:
    """Find blog-site/public directory relative to this file."""
    env_path = os.environ.get("BLOG_STATIC_PATH", "")
    if env_path and Path(env_path).exists():
        return Path(env_path)
    candidates = [
        Path(__file__).parents[3] / "blog-site" / "public",
        Path(__file__).parents[2] / ".." / "blog-site" / "public",
    ]
    return next(
        (p.resolve() for p in candidates if p.resolve().exists()),
        Path("public"),
    )


async def generate_chart(symbol: str, date_str: str) -> Optional[str]:
    """
    Fetch OHLCV, render chart, save to blog-site/public/images/charts/.
    Returns relative URL like /images/charts/VCB/20260418.png or None on failure.
    """
    df = await _fetch_ohlcv(symbol)
    if df is None or len(df) < 10:
        logger.warning("Not enough data for %s chart", symbol)
        return None

    try:
        png_bytes = await asyncio.to_thread(_render_chart, df, symbol)
    except Exception as exc:
        logger.warning("Chart render failed for %s: %s", symbol, exc)
        return None

    blog_public = _resolve_blog_public()
    dest = blog_public / "images" / "charts" / symbol / f"{date_str}.png"
    dest.parent.mkdir(parents=True, exist_ok=True)
    dest.write_bytes(png_bytes)
    logger.info("Chart saved: %s", dest)
    return f"/images/charts/{symbol}/{date_str}.png"
