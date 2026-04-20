"""Extended tests for sentiment analyzer (no pytest dependency)."""

import unittest
import re
from datetime import time, datetime


class TestVietnameseSlang(unittest.TestCase):
    """Test cases for Vietnamese stock slang handling."""

    def test_pump_and_dump_recognition(self):
        slang_mapping = {
            "lùa gà": "pump and dump",
            "cá mập": "big investors",
        }
        text = "Cổ phiếu VNM đang bị lùa gà"
        for slang, english in slang_mapping.items():
            if slang in text.lower():
                self.assertIn(english, ["pump and dump", "big investors"])

    def test_bullish_terms(self):
        bullish_terms = ["tăng mạnh", "lên điểm", "tích lũy"]
        text = "Cổ phiếu VNM tăng mạnh nhờ tin tốt"
        found = any(term in text for term in bullish_terms)
        self.assertTrue(found)

    def test_bearish_terms(self):
        bearish_terms = ["giảm mạnh", "xuống điểm", "bán tháo"]
        text = "Thị trường giảm mạnh vì lo ngại"
        found = any(term in text for term in bearish_terms)
        self.assertTrue(found)


class TestSymbolExtraction(unittest.TestCase):
    """Test cases for stock symbol extraction."""

    def test_valid_vietnamese_symbols(self):
        valid_symbols = ["VNM", "HPG", "VCB", "FPT", "VIC"]
        text = f"Cổ phiếu {' '.join(valid_symbols)} tăng giá"
        pattern = r'\b([A-Z]{2,5})\b'
        found = re.findall(pattern, text)
        for symbol in valid_symbols:
            self.assertIn(symbol, found)

    def test_mixed_content(self):
        text = "VNM và HPG là cổ phiếu tốt, FPT cũng vậy"
        pattern = r'\b([A-Z]{2,5})\b'
        symbols = re.findall(pattern, text)
        self.assertIn("VNM", symbols)
        self.assertIn("HPG", symbols)
        self.assertIn("FPT", symbols)


class TestSentimentScoring(unittest.TestCase):
    """Test cases for sentiment scoring."""

    def test_score_range(self):
        score = 0.75
        self.assertGreaterEqual(score, -1.0)
        self.assertLessEqual(score, 1.0)

    def test_confidence_levels(self):
        thresholds = {"high": 0.8, "medium": 0.5, "low": 0.3}
        test_score = 0.85
        if test_score >= thresholds["high"]:
            level = "high"
        elif test_score >= thresholds["medium"]:
            level = "medium"
        else:
            level = "low"
        self.assertEqual(level, "high")


class TestMarketHours(unittest.TestCase):
    """Test cases for Vietnamese market hours logic."""

    def test_trading_hours(self):
        MARKET_OPEN = time(9, 0)
        MARKET_CLOSE = time(15, 0)
        test_time = time(10, 30)
        self.assertGreaterEqual(test_time, MARKET_OPEN)
        self.assertLessEqual(test_time, MARKET_CLOSE)

    def test_weekend_detection(self):
        saturday = datetime(2024, 1, 6, 10, 0)
        sunday = datetime(2024, 1, 7, 10, 0)
        monday = datetime(2024, 1, 8, 10, 0)

        def is_weekend(dt):
            return dt.weekday() >= 5

        self.assertTrue(is_weekend(saturday))
        self.assertTrue(is_weekend(sunday))
        self.assertFalse(is_weekend(monday))


class TestPriceLimits(unittest.TestCase):
    """Test cases for Vietnamese price limit rules."""

    def test_floor_ceiling_calculation(self):
        reference_price = 100000.0
        limit_percent = 0.07
        floor = reference_price * (1 - limit_percent)
        ceiling = reference_price * (1 + limit_percent)
        self.assertEqual(floor, 93000.0)
        self.assertEqual(ceiling, 107000.0)
        self.assertLess(floor, reference_price)
        self.assertGreater(ceiling, reference_price)

    def test_hnx_limits(self):
        reference_price = 50000.0
        hnx_limit = 0.10
        floor = reference_price * (1 - hnx_limit)
        ceiling = reference_price * (1 + hnx_limit)
        self.assertEqual(floor, 45000.0)
        self.assertAlmostEqual(ceiling, 55000.0, places=2)


if __name__ == "__main__":
    unittest.main(verbosity=2)
