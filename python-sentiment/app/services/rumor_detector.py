"""
Rumor detection module.

Compares social media mentions against official/verified news sources
to identify unverified claims and potential manipulation campaigns.
"""

import logging
import re
from collections import Counter
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Tuple

from app.services.source_scorer import SourceScorer, MIN_DIVERSE_SOURCES

logger = logging.getLogger(__name__)

# Vietnamese stock symbol pattern: 3 uppercase letters
_SYMBOL_RE = re.compile(r'\b([A-Z]{3})\b')

# Known valid exchange-listed symbols can be loaded from DB;
# for now we just validate length.
_MIN_MENTIONS_FOR_RUMOR = 3


@dataclass
class RumorCandidate:
    """A stock symbol that appears in social media but not in official news."""
    symbol: str
    social_mentions: int
    official_mentions: int
    risk_level: str  # LOW, MEDIUM, HIGH
    sources: List[str] = field(default_factory=list)
    sample_texts: List[str] = field(default_factory=list)

    def to_dict(self) -> dict:
        return {
            "symbol": self.symbol,
            "social_mentions": self.social_mentions,
            "official_mentions": self.official_mentions,
            "risk_level": self.risk_level,
            "unique_sources": len(set(self.sources)),
            "warning": self._warning_text(),
            "sample_texts": self.sample_texts[:3],  # max 3 samples
        }

    def _warning_text(self) -> str:
        if self.risk_level == "HIGH":
            return "⚠️ Nhiều đề cập trên MXH nhưng KHÔNG có nguồn chính thống xác nhận"
        elif self.risk_level == "MEDIUM":
            return "⚠️ Đề cập chủ yếu từ nguồn không xác minh"
        return "Theo dõi — xuất hiện trên MXH nhiều hơn bình thường"


@dataclass
class TextWithSource:
    """A text item with its source information."""
    text: str
    source: str  # e.g. "vnexpress", "telegram_unknown", "facebook_group"
    symbol_hint: Optional[str] = None  # pre-extracted symbol if available


class RumorDetector:
    """Detect potential rumors by comparing social vs official mentions."""

    def __init__(self, source_scorer: Optional[SourceScorer] = None):
        self.scorer = source_scorer or SourceScorer()

    def detect(self, items: List[TextWithSource]) -> Dict:
        """Analyze a batch of texts to identify rumor candidates.

        Splits items into 'official' (source_weight >= 0.7) and 'social'
        (source_weight < 0.7), then finds symbols that appear heavily
        in social but not in official news.

        Returns:
            Dict with 'rumors', 'verified_symbols', and 'stats'.
        """
        official_mentions: Counter = Counter()
        social_mentions: Counter = Counter()
        social_sources: dict[str, list[str]] = {}
        social_samples: dict[str, list[str]] = {}

        for item in items:
            symbols = self._extract_symbols(item.text, item.symbol_hint)
            weight = self.scorer.get_weight(item.source)

            for sym in symbols:
                if weight >= 0.7:
                    official_mentions[sym] += 1
                else:
                    social_mentions[sym] += 1
                    social_sources.setdefault(sym, []).append(item.source)
                    social_samples.setdefault(sym, []).append(
                        item.text[:200]  # truncate for storage
                    )

        # Find rumor candidates: high social mentions, low/no official
        rumors: List[RumorCandidate] = []
        for sym, social_count in social_mentions.most_common():
            official_count = official_mentions.get(sym, 0)

            # Only flag if social mentions are significant
            if social_count < _MIN_MENTIONS_FOR_RUMOR:
                continue

            # Ratio check: social >> official
            ratio = social_count / max(official_count, 1)
            if ratio < 2.0 and official_count > 0:
                continue  # balanced coverage, not a rumor

            risk = self._assess_risk(social_count, official_count, ratio)
            rumors.append(RumorCandidate(
                symbol=sym,
                social_mentions=social_count,
                official_mentions=official_count,
                risk_level=risk,
                sources=social_sources.get(sym, []),
                sample_texts=social_samples.get(sym, []),
            ))

        # Verified: symbols with adequate official coverage
        verified = [
            sym for sym, count in official_mentions.items()
            if count >= 2
        ]

        return {
            "rumors": [r.to_dict() for r in rumors],
            "verified_symbols": sorted(verified),
            "stats": {
                "total_items": len(items),
                "official_items": sum(
                    1 for i in items if self.scorer.get_weight(i.source) >= 0.7
                ),
                "social_items": sum(
                    1 for i in items if self.scorer.get_weight(i.source) < 0.7
                ),
                "unique_symbols_found": len(
                    set(official_mentions.keys()) | set(social_mentions.keys())
                ),
                "rumors_detected": len(rumors),
            },
        }

    @staticmethod
    def _extract_symbols(text: str, hint: Optional[str] = None) -> List[str]:
        """Extract stock symbols from text."""
        symbols = set(_SYMBOL_RE.findall(text))
        if hint:
            symbols.add(hint.upper())
        return list(symbols)

    @staticmethod
    def _assess_risk(
        social_count: int, official_count: int, ratio: float
    ) -> str:
        """Determine risk level from mention counts."""
        if official_count == 0 and social_count >= 10:
            return "HIGH"
        if ratio >= 5.0:
            return "HIGH"
        if ratio >= 3.0 or (official_count == 0 and social_count >= 5):
            return "MEDIUM"
        return "LOW"
