"""
Source credibility scoring for sentiment analysis.

Assigns a trust weight to each content source so that unverified
Telegram/Facebook posts carry far less influence than institutional
research or established financial news outlets.
"""

import logging
from typing import Optional

logger = logging.getLogger(__name__)

# Weight 0.0–1.0: how much we trust sentiment from this source.
# Tier 1 — Institutional research & major financial media
# Tier 2 — Verified community platforms
# Tier 3 — Unverified social media
SOURCE_WEIGHTS: dict[str, float] = {
    # Tier 1: trusted financial media
    "vnexpress":       1.0,
    "cafef":           1.0,
    "vietstock":       1.0,
    "ssi_research":    1.0,
    "vps_research":    0.95,
    "vcbs_research":   0.95,
    "hsc_research":    0.95,
    "vndirect":        0.90,
    "tinnhanhchungkhoan": 0.90,
    "ndh":             0.85,
    "thanhnien":       0.85,
    "tuoitre":         0.85,

    # Tier 2: community / aggregators with moderation
    "fireant":         0.70,
    "simplize":        0.65,
    "stockbiz":        0.65,
    "telegram_verified": 0.50,
    "zalo_official":   0.50,

    # Tier 3: unverified / personal social
    "telegram_group":  0.20,
    "telegram_unknown": 0.15,
    "facebook_group":  0.15,
    "facebook_personal": 0.10,
    "tiktok":          0.10,
    "youtube_comment":  0.10,
    "unknown":         0.10,
}

# Minimum number of distinct sources required for full trust
MIN_DIVERSE_SOURCES = 3


class SourceScorer:
    """Score the credibility of a content source."""

    def __init__(self, custom_weights: Optional[dict[str, float]] = None):
        self.weights = {**SOURCE_WEIGHTS, **(custom_weights or {})}

    def get_weight(self, source: Optional[str]) -> float:
        """Return credibility weight for a given source identifier.

        Performs case-insensitive prefix matching so that
        'vnexpress_kinhdoanh' still maps to 'vnexpress'.
        """
        if not source:
            return self.weights["unknown"]

        source_lower = source.lower().strip()

        # Exact match first
        if source_lower in self.weights:
            return self.weights[source_lower]

        # Prefix match (e.g. "cafef_something" → "cafef")
        for key, weight in self.weights.items():
            if source_lower.startswith(key):
                return weight

        logger.debug("Unknown source '%s', assigning minimum weight", source)
        return self.weights["unknown"]

    def compute_diversity_factor(self, sources: list[str]) -> float:
        """Return a multiplier [0.1, 1.0] based on source diversity.

        - ≥ MIN_DIVERSE_SOURCES unique credible sources → 1.0
        - 2 sources → 0.6
        - 1 source  → 0.3
        - 0 sources → 0.1
        """
        unique = set()
        for s in sources:
            key = self._canonical_source(s)
            if self.get_weight(s) >= 0.5:  # only count credible sources
                unique.add(key)

        n = len(unique)
        if n >= MIN_DIVERSE_SOURCES:
            return 1.0
        if n == 2:
            return 0.6
        if n == 1:
            return 0.3
        return 0.1

    @staticmethod
    def _canonical_source(source: str) -> str:
        """Normalize source to canonical key for dedup counting."""
        return source.lower().strip().split("_")[0] if source else "unknown"
