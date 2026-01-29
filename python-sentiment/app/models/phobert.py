import logging
from typing import List, Tuple

import torch
from transformers import AutoTokenizer, AutoModelForSequenceClassification

logger = logging.getLogger(__name__)


class PhoBERTSentiment:
    """Vietnamese sentiment analysis using PhoBERT."""

    def __init__(self, model_name: str = "vinai/phobert-base", cache_dir: str = "./.cache"):
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        logger.info(f"Using device: {self.device}")

        # Load tokenizer and model
        logger.info(f"Loading model: {model_name}")
        self.tokenizer = AutoTokenizer.from_pretrained(
            model_name,
            cache_dir=cache_dir
        )

        # For sentiment, we use a fine-tuned version or base model with classification head
        # In production, use a fine-tuned sentiment model
        self.model = AutoModelForSequenceClassification.from_pretrained(
            model_name,
            num_labels=3,  # positive, negative, neutral
            cache_dir=cache_dir,
            ignore_mismatched_sizes=True  # Allow adding classification head
        )
        self.model.to(self.device)
        self.model.eval()

        self.labels = ["negative", "neutral", "positive"]
        self._model_loaded = True
        logger.info("Model loaded successfully")

    @property
    def is_loaded(self) -> bool:
        return self._model_loaded

    def predict(self, text: str, max_length: int = 256) -> Tuple[str, float]:
        """Predict sentiment for a single text."""
        with torch.no_grad():
            inputs = self.tokenizer(
                text,
                return_tensors="pt",
                truncation=True,
                max_length=max_length,
                padding=True
            ).to(self.device)

            outputs = self.model(**inputs)
            probabilities = torch.softmax(outputs.logits, dim=1)

            predicted_idx = torch.argmax(probabilities, dim=1).item()
            confidence = probabilities[0][predicted_idx].item() * 100

            return self.labels[predicted_idx], confidence

    def predict_batch(self, texts: List[str], max_length: int = 256, batch_size: int = 16) -> List[Tuple[str, float]]:
        """Predict sentiment for multiple texts."""
        results = []

        for i in range(0, len(texts), batch_size):
            batch_texts = texts[i:i + batch_size]

            with torch.no_grad():
                inputs = self.tokenizer(
                    batch_texts,
                    return_tensors="pt",
                    truncation=True,
                    max_length=max_length,
                    padding=True
                ).to(self.device)

                outputs = self.model(**inputs)
                probabilities = torch.softmax(outputs.logits, dim=1)

                for j in range(len(batch_texts)):
                    predicted_idx = torch.argmax(probabilities[j]).item()
                    confidence = probabilities[j][predicted_idx].item() * 100
                    results.append((self.labels[predicted_idx], confidence))

        return results

    def get_memory_usage(self) -> float:
        """Get current GPU memory usage in MB."""
        if torch.cuda.is_available():
            return torch.cuda.memory_allocated() / 1024 / 1024
        return 0.0
