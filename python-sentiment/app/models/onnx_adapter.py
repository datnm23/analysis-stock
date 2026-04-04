"""
ONNX Runtime inference adapter for PhoBERT sentiment model.

Drop-in replacement for PyTorch inference when ONNX model is available.
Implements the same predict/predict_batch API as PhoBERTSentiment.
Reduces memory from ~4GB to ~1GB with INT8 quantization.

Usage:
    adapter = ONNXSentimentAdapter("models/onnx")
    label, confidence = adapter.predict("Cổ phiếu VNM tăng mạnh")
    # ("positive", 87.3)
"""

import logging
import os
from typing import List, Tuple

import numpy as np

logger = logging.getLogger(__name__)


class ONNXSentimentAdapter:
    """Sentiment inference using ONNX Runtime (faster, less memory).

    Implements the SentimentModel protocol — same API as PhoBERTSentiment.
    """

    LABELS = ["negative", "neutral", "positive"]

    def __init__(self, model_dir: str = "models/onnx"):
        self.model_dir = model_dir
        self._session = None
        self._tokenizer = None
        self._model_loaded = False

    def _load(self):
        """Lazy-load ONNX model and tokenizer."""
        if self._session is not None:
            return

        try:
            import onnxruntime as ort
            from transformers import AutoTokenizer
        except ImportError as e:
            raise ImportError(
                f"Missing dependency: {e}. Install: pip install onnxruntime transformers"
            )

        # Prefer quantized model
        int8_path = os.path.join(self.model_dir, "phobert_sentiment_int8.onnx")
        fp32_path = os.path.join(self.model_dir, "phobert_sentiment.onnx")

        model_path = int8_path if os.path.exists(int8_path) else fp32_path
        if not os.path.exists(model_path):
            raise FileNotFoundError(
                f"ONNX model not found at {model_path}. "
                "Run: python -m scripts.export_onnx --quantize"
            )

        logger.info("Loading ONNX model from %s", model_path)
        self._session = ort.InferenceSession(
            model_path,
            providers=["CPUExecutionProvider"],
        )

        self._tokenizer = AutoTokenizer.from_pretrained(self.model_dir)
        model_size = os.path.getsize(model_path) / (1024 * 1024)
        self._model_loaded = True
        logger.info("ONNX model loaded: %.1f MB", model_size)

    @property
    def is_loaded(self) -> bool:
        return self._model_loaded

    def predict(self, text: str, max_length: int = 256) -> Tuple[str, float]:
        """Run sentiment prediction on a single text.

        Returns:
            Tuple of (label, confidence_percent) matching PhoBERTSentiment API.
        """
        self._load()

        inputs = self._tokenizer(
            text,
            return_tensors="np",
            max_length=max_length,
            padding="max_length",
            truncation=True,
        )

        logits = self._session.run(
            None,
            {
                "input_ids": inputs["input_ids"].astype(np.int64),
                "attention_mask": inputs["attention_mask"].astype(np.int64),
            },
        )[0]

        probs = self._softmax(logits[0])
        pred_idx = int(np.argmax(probs))
        confidence = float(probs[pred_idx]) * 100  # 0-100 scale like PhoBERT

        return self.LABELS[pred_idx], confidence

    def predict_batch(
        self, texts: List[str], max_length: int = 256, batch_size: int = 16
    ) -> List[Tuple[str, float]]:
        """Run sentiment prediction on a batch of texts.

        Returns:
            List of (label, confidence_percent) tuples.
        """
        self._load()

        results = []
        for i in range(0, len(texts), batch_size):
            batch_texts = texts[i : i + batch_size]

            inputs = self._tokenizer(
                batch_texts,
                return_tensors="np",
                max_length=max_length,
                padding="max_length",
                truncation=True,
            )

            logits = self._session.run(
                None,
                {
                    "input_ids": inputs["input_ids"].astype(np.int64),
                    "attention_mask": inputs["attention_mask"].astype(np.int64),
                },
            )[0]

            for row in logits:
                probs = self._softmax(row)
                pred_idx = int(np.argmax(probs))
                confidence = float(probs[pred_idx]) * 100
                results.append((self.LABELS[pred_idx], confidence))

        return results

    def predict_detailed(self, text: str, max_length: int = 256) -> dict:
        """Run prediction and return detailed scores (for debugging/admin).

        Returns:
            {"label": str, "confidence": float, "scores": dict}
        """
        self._load()

        inputs = self._tokenizer(
            text,
            return_tensors="np",
            max_length=max_length,
            padding="max_length",
            truncation=True,
        )

        logits = self._session.run(
            None,
            {
                "input_ids": inputs["input_ids"].astype(np.int64),
                "attention_mask": inputs["attention_mask"].astype(np.int64),
            },
        )[0]

        probs = self._softmax(logits[0])
        pred_idx = int(np.argmax(probs))

        return {
            "label": self.LABELS[pred_idx],
            "confidence": float(probs[pred_idx]),
            "scores": {
                label: float(prob)
                for label, prob in zip(self.LABELS, probs)
            },
        }

    def get_memory_usage(self) -> float:
        """ONNX runs on CPU — return 0 for GPU compat."""
        return 0.0

    @staticmethod
    def _softmax(x):
        e = np.exp(x - np.max(x))
        return e / e.sum()

    @staticmethod
    def is_available(model_dir: str = "models/onnx") -> bool:
        """Check if ONNX model files exist."""
        int8 = os.path.join(model_dir, "phobert_sentiment_int8.onnx")
        fp32 = os.path.join(model_dir, "phobert_sentiment.onnx")
        return os.path.exists(int8) or os.path.exists(fp32)
