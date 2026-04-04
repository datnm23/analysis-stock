"""Unit tests for ONNXSentimentAdapter."""

import os
import pytest
from unittest.mock import Mock, MagicMock, patch
import numpy as np


class TestONNXSentimentAdapter:
    """Test cases for ONNXSentimentAdapter class."""

    def _make_adapter(self):
        """Create an adapter with mocked internals."""
        from app.models.onnx_adapter import ONNXSentimentAdapter
        adapter = ONNXSentimentAdapter(model_dir="models/onnx")
        return adapter

    def _make_loaded_adapter(self):
        """Create an adapter with mocked ONNX session already loaded."""
        adapter = self._make_adapter()

        # Mock tokenizer
        mock_tokenizer = MagicMock()
        mock_tokenizer.return_value = {
            "input_ids": np.array([[1, 2, 3, 0, 0]], dtype=np.int64),
            "attention_mask": np.array([[1, 1, 1, 0, 0]], dtype=np.int64),
        }
        adapter._tokenizer = mock_tokenizer

        # Mock ONNX session — returns logits for 3 classes
        mock_session = MagicMock()
        # Simulate: negative=0.1, neutral=0.2, positive=0.7
        mock_session.run.return_value = [np.array([[0.1, 0.2, 0.7]])]
        adapter._session = mock_session
        adapter._model_loaded = True

        return adapter

    def test_predict_returns_tuple(self):
        """predict() must return (label, confidence) tuple."""
        adapter = self._make_loaded_adapter()
        result = adapter.predict("Cổ phiếu VNM tăng mạnh")

        assert isinstance(result, tuple)
        assert len(result) == 2
        label, confidence = result
        assert label in ["negative", "neutral", "positive"]
        assert isinstance(confidence, float)
        assert 0 <= confidence <= 100

    def test_predict_positive_sentiment(self):
        """verify correct label selection from logits."""
        adapter = self._make_loaded_adapter()
        # Mock logits: positive has highest score
        adapter._session.run.return_value = [np.array([[0.1, 0.2, 0.9]])]

        label, confidence = adapter.predict("test text")
        assert label == "positive"

    def test_predict_negative_sentiment(self):
        """verify negative label when negative logit is highest."""
        adapter = self._make_loaded_adapter()
        adapter._session.run.return_value = [np.array([[0.9, 0.1, 0.05]])]

        label, confidence = adapter.predict("test text")
        assert label == "negative"

    def test_predict_neutral_sentiment(self):
        """verify neutral label when neutral logit is highest."""
        adapter = self._make_loaded_adapter()
        adapter._session.run.return_value = [np.array([[0.1, 0.9, 0.05]])]

        label, confidence = adapter.predict("test text")
        assert label == "neutral"

    def test_predict_batch_returns_list_of_tuples(self):
        """predict_batch() must return List[Tuple[str, float]]."""
        adapter = self._make_loaded_adapter()
        # Return 2 rows of logits
        adapter._session.run.return_value = [
            np.array([
                [0.1, 0.2, 0.7],  # positive
                [0.8, 0.1, 0.1],  # negative
            ])
        ]

        results = adapter.predict_batch(["text1", "text2"])

        assert isinstance(results, list)
        assert len(results) == 2
        for item in results:
            assert isinstance(item, tuple)
            assert len(item) == 2
            label, conf = item
            assert label in ["negative", "neutral", "positive"]
            assert 0 <= conf <= 100

    def test_predict_batch_correct_labels(self):
        """predict_batch labels match logit ordering."""
        adapter = self._make_loaded_adapter()
        adapter._session.run.return_value = [
            np.array([
                [0.1, 0.2, 0.8],  # positive
                [0.7, 0.2, 0.1],  # negative
            ])
        ]

        results = adapter.predict_batch(["positive text", "negative text"])
        assert results[0][0] == "positive"
        assert results[1][0] == "negative"

    def test_is_loaded_property(self):
        """is_loaded reflects actual state."""
        adapter = self._make_adapter()
        assert adapter.is_loaded is False

        adapter._model_loaded = True
        assert adapter.is_loaded is True

    def test_is_available_no_files(self, tmp_path):
        """is_available returns False when no model files exist."""
        from app.models.onnx_adapter import ONNXSentimentAdapter
        assert ONNXSentimentAdapter.is_available(str(tmp_path)) is False

    def test_is_available_fp32(self, tmp_path):
        """is_available returns True when FP32 model exists."""
        from app.models.onnx_adapter import ONNXSentimentAdapter
        (tmp_path / "phobert_sentiment.onnx").touch()
        assert ONNXSentimentAdapter.is_available(str(tmp_path)) is True

    def test_is_available_int8(self, tmp_path):
        """is_available returns True when INT8 model exists."""
        from app.models.onnx_adapter import ONNXSentimentAdapter
        (tmp_path / "phobert_sentiment_int8.onnx").touch()
        assert ONNXSentimentAdapter.is_available(str(tmp_path)) is True

    def test_predict_detailed_returns_dict(self):
        """predict_detailed() returns full dict with scores."""
        adapter = self._make_loaded_adapter()
        adapter._session.run.return_value = [np.array([[0.1, 0.2, 0.7]])]

        result = adapter.predict_detailed("test")

        assert isinstance(result, dict)
        assert "label" in result
        assert "confidence" in result
        assert "scores" in result
        assert "negative" in result["scores"]
        assert "neutral" in result["scores"]
        assert "positive" in result["scores"]

    def test_lazy_load_on_predict(self):
        """_load is called lazily on first predict."""
        adapter = self._make_adapter()

        with patch.object(adapter, '_load') as mock_load:
            # After mocking _load, we need session and tokenizer
            def side_effect():
                adapter._session = MagicMock()
                adapter._session.run.return_value = [np.array([[0.1, 0.2, 0.7]])]
                adapter._tokenizer = MagicMock()
                adapter._tokenizer.return_value = {
                    "input_ids": np.array([[1, 2, 3]], dtype=np.int64),
                    "attention_mask": np.array([[1, 1, 1]], dtype=np.int64),
                }

            mock_load.side_effect = side_effect
            adapter.predict("test")
            mock_load.assert_called_once()

    def test_get_memory_usage(self):
        """get_memory_usage returns 0 for ONNX (CPU-only)."""
        adapter = self._make_adapter()
        assert adapter.get_memory_usage() == 0.0

    def test_softmax(self):
        """Verify softmax computation."""
        from app.models.onnx_adapter import ONNXSentimentAdapter
        result = ONNXSentimentAdapter._softmax(np.array([1.0, 2.0, 3.0]))
        assert abs(sum(result) - 1.0) < 1e-6
        assert result[2] > result[1] > result[0]

    def test_confidence_scale_0_to_100(self):
        """Confidence must be on 0-100 scale (matching PhoBERT)."""
        adapter = self._make_loaded_adapter()
        # Set a clear softmax — softmax([0,0,5]) ≈ [0.003, 0.003, 0.993]
        adapter._session.run.return_value = [np.array([[0.0, 0.0, 5.0]])]

        _, confidence = adapter.predict("test")
        assert confidence > 90  # Should be ~99.3


class TestONNXPhoBERTAPICompatibility:
    """Verify ONNX adapter is API-compatible with PhoBERTSentiment.

    SentimentAnalyzer must work identically with either backend.
    """

    def test_protocol_compliance(self):
        """ONNXSentimentAdapter implements SentimentModel protocol."""
        from app.models.model_interface import SentimentModel
        from app.models.onnx_adapter import ONNXSentimentAdapter

        # Check that it has all required methods/properties
        adapter = ONNXSentimentAdapter()
        assert hasattr(adapter, 'predict')
        assert hasattr(adapter, 'predict_batch')
        assert hasattr(adapter, 'is_loaded')

    def test_works_with_sentiment_analyzer(self):
        """SentimentAnalyzer accepts ONNXSentimentAdapter as model."""
        from app.services.sentiment_analyzer import SentimentAnalyzer
        from app.models.onnx_adapter import ONNXSentimentAdapter

        adapter = ONNXSentimentAdapter()
        # Mock internals
        adapter._session = MagicMock()
        adapter._session.run.return_value = [np.array([[0.1, 0.2, 0.7]])]
        adapter._tokenizer = MagicMock()
        adapter._tokenizer.return_value = {
            "input_ids": np.array([[1, 2, 3]], dtype=np.int64),
            "attention_mask": np.array([[1, 1, 1]], dtype=np.int64),
        }
        adapter._model_loaded = True

        analyzer = SentimentAnalyzer(adapter)
        result = analyzer.analyze("Cổ phiếu VNM tăng mạnh")

        assert "sentiment" in result
        assert "confidence" in result
        assert result["sentiment"] in ["negative", "neutral", "positive"]
