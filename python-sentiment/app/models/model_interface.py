"""
Unified model interface for sentiment analysis backends.

Both PhoBERTSentiment (PyTorch) and ONNXSentimentAdapter implement
this protocol, allowing main.py to swap backends transparently.
"""

from typing import List, Protocol, Tuple, runtime_checkable


@runtime_checkable
class SentimentModel(Protocol):
    """Protocol that all sentiment model backends must implement."""

    def predict(self, text: str, max_length: int = 256) -> Tuple[str, float]:
        """Predict sentiment for a single text.

        Returns:
            Tuple of (label, confidence) where label is one of
            "negative", "neutral", "positive" and confidence is 0-100.
        """
        ...

    def predict_batch(
        self, texts: List[str], max_length: int = 256
    ) -> List[Tuple[str, float]]:
        """Predict sentiment for multiple texts.

        Returns:
            List of (label, confidence) tuples.
        """
        ...

    @property
    def is_loaded(self) -> bool:
        """Whether the model has been loaded and is ready for inference."""
        ...
