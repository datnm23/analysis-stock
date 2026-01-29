import json
import logging
import re
from pathlib import Path
from typing import List, Dict, Any, Optional

from app.models.phobert import PhoBERTSentiment

logger = logging.getLogger(__name__)


class SentimentAnalyzer:
    def __init__(self, model: PhoBERTSentiment, slang_dict_path: Optional[str] = None):
        self.model = model
        self.slang_mappings = {}
        self.positive_keywords = []
        self.negative_keywords = []

        # Load slang dictionary
        dict_path = slang_dict_path or Path(__file__).parent.parent / "data" / "slang_dictionary.json"
        self._load_slang_dictionary(dict_path)

        # Symbol pattern for Vietnamese stocks (3 uppercase letters)
        self.symbol_pattern = re.compile(r'\b([A-Z]{3})\b')

    def _load_slang_dictionary(self, path: str):
        try:
            with open(path, 'r', encoding='utf-8') as f:
                data = json.load(f)
                self.slang_mappings = data.get("slang_mappings", {})
                self.positive_keywords = data.get("positive_keywords", [])
                self.negative_keywords = data.get("negative_keywords", [])
                logger.info(f"Loaded {len(self.slang_mappings)} slang mappings")
        except Exception as e:
            logger.warning(f"Failed to load slang dictionary: {e}")

    def preprocess_text(self, text: str) -> str:
        """Preprocess Vietnamese text for sentiment analysis."""
        # Normalize whitespace
        text = " ".join(text.split())

        # Replace slang with standard terms (optional, for logging)
        for slang, info in self.slang_mappings.items():
            if slang in text.lower():
                logger.debug(f"Found slang: {slang} -> {info['meaning']}")

        return text

    def extract_symbols(self, text: str) -> List[str]:
        """Extract stock symbols from text."""
        symbols = self.symbol_pattern.findall(text)
        # Filter to only valid Vietnamese stock symbols (simple validation)
        valid_symbols = [s for s in symbols if len(s) == 3]
        return list(set(valid_symbols))

    def extract_keywords(self, text: str) -> List[str]:
        """Extract relevant keywords from text."""
        keywords = []
        text_lower = text.lower()

        for kw in self.positive_keywords + self.negative_keywords:
            if kw in text_lower:
                keywords.append(kw)

        for slang in self.slang_mappings:
            if slang in text_lower:
                keywords.append(slang)

        return keywords

    def adjust_confidence(self, base_sentiment: str, base_confidence: float, text: str) -> tuple:
        """Adjust sentiment based on slang and keywords."""
        text_lower = text.lower()
        adjustment = 0.0

        # Apply slang modifiers
        for slang, info in self.slang_mappings.items():
            if slang in text_lower:
                adjustment += info.get("sentiment_modifier", 0) * 100

        # Apply keyword modifiers
        pos_count = sum(1 for kw in self.positive_keywords if kw in text_lower)
        neg_count = sum(1 for kw in self.negative_keywords if kw in text_lower)
        adjustment += (pos_count - neg_count) * 5

        # Adjust confidence
        new_confidence = min(100, max(0, base_confidence + adjustment))

        # Potentially flip sentiment if adjustment is strong
        if adjustment < -20 and base_sentiment == "positive":
            return "neutral", new_confidence
        elif adjustment > 20 and base_sentiment == "negative":
            return "neutral", new_confidence

        return base_sentiment, new_confidence

    def analyze(self, text: str) -> Dict[str, Any]:
        """Analyze sentiment of a single text."""
        processed_text = self.preprocess_text(text)
        base_sentiment, base_confidence = self.model.predict(processed_text)

        # Adjust based on domain knowledge
        sentiment, confidence = self.adjust_confidence(
            base_sentiment, base_confidence, text
        )

        return {
            "sentiment": sentiment,
            "confidence": round(confidence, 2),
            "symbols": self.extract_symbols(text),
            "keywords": self.extract_keywords(text)
        }

    def analyze_batch(self, texts: List[str]) -> List[Dict[str, Any]]:
        """Analyze sentiment of multiple texts."""
        processed_texts = [self.preprocess_text(t) for t in texts]
        predictions = self.model.predict_batch(processed_texts)

        results = []
        for i, (sentiment, confidence) in enumerate(predictions):
            adj_sentiment, adj_confidence = self.adjust_confidence(
                sentiment, confidence, texts[i]
            )
            results.append({
                "sentiment": adj_sentiment,
                "confidence": round(adj_confidence, 2),
                "symbols": self.extract_symbols(texts[i]),
                "keywords": self.extract_keywords(texts[i])
            })

        return results
