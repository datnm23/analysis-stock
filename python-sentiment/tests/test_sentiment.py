"""Unit tests for SentimentAnalyzer."""

import pytest
from unittest.mock import Mock, patch, MagicMock
from app.services.sentiment_analyzer import SentimentAnalyzer


class TestSentimentAnalyzer:
    """Test cases for SentimentAnalyzer class."""

    def test_init(self, mock_phobert_model):
        """Test analyzer initialization."""
        analyzer = SentimentAnalyzer(mock_phobert_model)
        assert analyzer.model is not None
        assert analyzer.slang_mappings is not None

    def test_preprocess_text(self, mock_sentiment_analyzer):
        """Test text preprocessing."""
        text = "  Cổ phiếu  VNM   tăng  mạnh  "
        result = mock_sentiment_analyzer.preprocess_text(text)
        assert "  " not in result
        assert result == "Cổ phiếu VNM tăng mạnh"

    def test_extract_symbols(self, mock_sentiment_analyzer):
        """Test stock symbol extraction."""
        text = "Cổ phiếu VNM và HPG tăng giá, VCB giảm"
        symbols = mock_sentiment_analyzer.extract_symbols(text)
        assert "VNM" in symbols
        assert "HPG" in symbols
        assert "VCB" in symbols

    def test_extract_symbols_no_symbols(self, mock_sentiment_analyzer):
        """Test symbol extraction with no valid symbols."""
        text = "Thị trường chứng khoán tăng điểm"
        symbols = mock_sentiment_analyzer.extract_symbols(text)
        assert len(symbols) == 0

    def test_extract_keywords(self, mock_sentiment_analyzer):
        """Test keyword extraction."""
        text = "Cổ phiếu VNM tăng mạnh nhờ lợi nhuận kỷ lục"
        keywords = mock_sentiment_analyzer.extract_keywords(text)
        # Should find positive keywords if configured
        assert isinstance(keywords, list)

    def test_analyze_single_text(self, mock_sentiment_analyzer):
        """Test single text sentiment analysis."""
        text = "Cổ phiếu VNM tăng mạnh nhờ kết quả kinh doanh khả quan"
        result = mock_sentiment_analyzer.analyze(text)

        assert "sentiment" in result
        assert "confidence" in result
        assert "symbols" in result
        assert "keywords" in result
        assert result["sentiment"] in ["positive", "negative", "neutral"]
        assert 0 <= result["confidence"] <= 100

    def test_analyze_batch(self, mock_sentiment_analyzer):
        """Test batch sentiment analysis."""
        texts = [
            "Cổ phiếu VNM tăng mạnh",
            "Thị trường giảm điểm",
            "Kết quả kinh doanh ổn định"
        ]
        results = mock_sentiment_analyzer.analyze_batch(texts)

        assert len(results) == len(texts)
        for result in results:
            assert "sentiment" in result
            assert "confidence" in result

    def test_adjust_confidence_positive(self, mock_sentiment_analyzer):
        """Test confidence adjustment for positive sentiment."""
        sentiment, confidence = mock_sentiment_analyzer.adjust_confidence(
            "positive", 70.0, "Cổ phiếu tăng mạnh"
        )
        assert isinstance(sentiment, str)
        assert 0 <= confidence <= 100

    def test_adjust_confidence_negative(self, mock_sentiment_analyzer):
        """Test confidence adjustment for negative sentiment."""
        sentiment, confidence = mock_sentiment_analyzer.adjust_confidence(
            "negative", 70.0, "Cổ phiếu giảm mạnh"
        )
        assert isinstance(sentiment, str)
        assert 0 <= confidence <= 100

    def test_extract_symbols_with_slang(self, mock_sentiment_analyzer):
        """Test symbol extraction with slang text."""
        text = "Cá mập đang lùa gà VNM, chuẩn bị pump cổ phiếu"
        symbols = mock_sentiment_analyzer.extract_symbols(text)
        assert "VNM" in symbols


class TestSentimentAnalyzerIntegration:
    """Integration tests that require actual model loading (slow)."""

    @pytest.mark.slow
    def test_analyze_with_actual_model(self):
        """Test analyzer with actual model (requires model to be downloaded)."""
        # This test is marked as slow and skipped by default
        # Run with: pytest -m slow
        pytest.skip("Requires actual model download")

        # Uncomment to run with actual model:
        # from app.models.phobert import PhoBERTSentiment
        # model = PhoBERTSentiment()
        # analyzer = SentimentAnalyzer(model)
        # result = analyzer.analyze("Cổ phiếu VNM tăng mạnh")
        # assert result["sentiment"] in ["positive", "negative", "neutral"]
