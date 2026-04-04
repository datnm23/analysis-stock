"""Pytest configuration and fixtures for Python sentiment service."""

import pytest
from unittest.mock import Mock, MagicMock
from typing import Dict, Any, List

import numpy as np


# Simple mock class for PhoBERT model
class MockPhoBERTSentiment:
    """Mock PhoBERT sentiment model for testing."""

    def predict(self, text: str):
        text_lower = text.lower()
        if "tăng" in text_lower or "lợi" in text_lower or "khả quan" in text_lower:
            return "positive", 75.0
        elif "giảm" in text_lower or "lo ngại" in text_lower or "điều tra" in text_lower:
            return "negative", 70.0
        else:
            return "neutral", 65.0

    def predict_batch(self, texts: List[str]):
        return [self.predict(text) for text in texts]


# Import the actual analyzer class
from app.services.sentiment_analyzer import SentimentAnalyzer


# ============================================
# Mock Data
# ============================================

SAMPLE_NEWS_VIETNAMESE = [
    "Cổ phiếu VNM tăng mạnh nhờ kết quả kinh doanh quý 3 khả quan",
    "Thị trường chứng khoán giảm điểm do lo ngại suy thoái kinh tế",
    "HPG công bố lợi nhuận kỷ lục, cổ đông hưởng lợi",
    "Cổ phiếu ROS bị cơ quan chứng khoán điều tra về giao dịch bất thường",
    "Vietstock dự báo VN-Index sẽ tiếp tục tăng trưởng trong năm 2024",
]

EXPECTED_SENTIMENTS = [
    ("positive", 0.75),
    ("negative", 0.70),
    ("positive", 0.80),
    ("negative", 0.75),
    ("neutral", 0.65),
]

VALID_STOCK_SYMBOLS = ["VNM", "HPG", "VCB", "FPT", "MWG"]

# ============================================
# Fixtures
# ============================================


@pytest.fixture
def mock_phobert_model():
    """Create a mock PhoBERT model for testing."""
    return MockPhoBERTSentiment()


@pytest.fixture
def mock_sentiment_analyzer(mock_phobert_model):
    """Create a sentiment analyzer with mock model."""
    return SentimentAnalyzer(mock_phobert_model)


@pytest.fixture
def sample_texts():
    """Provide sample Vietnamese texts for testing."""
    return SAMPLE_NEWS_VIETNAMESE.copy()


@pytest.fixture
def valid_stock_symbols():
    """Provide valid stock symbols for testing."""
    return VALID_STOCK_SYMBOLS.copy()


@pytest.fixture
def mock_onnx_model():
    """Create a mock ONNX adapter matching PhoBERT interface."""
    from app.models.onnx_adapter import ONNXSentimentAdapter
    adapter = ONNXSentimentAdapter(model_dir="models/onnx")
    # Mock internals so _load() is skipped
    adapter._tokenizer = MagicMock()
    adapter._tokenizer.return_value = {
        "input_ids": np.array([[1, 2, 3, 0, 0]], dtype=np.int64),
        "attention_mask": np.array([[1, 1, 1, 0, 0]], dtype=np.int64),
    }
    adapter._session = MagicMock()
    adapter._session.run.return_value = [np.array([[0.1, 0.2, 0.7]])]
    adapter._model_loaded = True
    return adapter


@pytest.fixture
def mock_onnx_sentiment_analyzer(mock_onnx_model):
    """Create a sentiment analyzer with mock ONNX model."""
    return SentimentAnalyzer(mock_onnx_model)


# ============================================
# Helper Functions
# ============================================


def create_mock_response(
    sentiment: str,
    confidence: float,
    symbols: List[str] = None,
    keywords: List[str] = None
) -> Dict[str, Any]:
    """Create a mock sentiment analysis response."""
    return {
        "sentiment": sentiment,
        "confidence": confidence,
        "symbols": symbols or [],
        "keywords": keywords or []
    }


def assert_valid_sentiment_response(response: Dict[str, Any]):
    """Assert that a sentiment response has valid structure."""
    assert "sentiment" in response, "Response must contain 'sentiment'"
    assert "confidence" in response, "Response must contain 'confidence'"
    assert response["sentiment"] in ["positive", "negative", "neutral"], \
        f"Invalid sentiment: {response['sentiment']}"
    assert 0 <= response["confidence"] <= 100, \
        f"Confidence must be between 0 and 100: {response['confidence']}"


# ============================================
# Test Configuration
# ============================================


@pytest.fixture(autouse=True)
def reset_settings_cache():
    """Reset settings cache before each test."""
    # Clear lru_cache for get_settings
    from app.config import get_settings
    get_settings.cache_clear()
    yield
    get_settings.cache_clear()
