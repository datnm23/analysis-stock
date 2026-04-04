"""Unit tests for API endpoints."""

import pytest
from unittest.mock import Mock, patch, AsyncMock, MagicMock
from fastapi.testclient import TestClient

# Mock the analyzer before importing app
mock_analyzer_instance = Mock()
mock_analyzer_instance.analyze = Mock(return_value={
    "sentiment": "positive",
    "confidence": 75.0,
    "symbols": ["VNM"],
    "keywords": ["tăng"]
})
mock_analyzer_instance.analyze_batch = Mock(return_value=[
    {"sentiment": "positive", "confidence": 75.0, "symbols": ["VNM"], "keywords": []},
    {"sentiment": "negative", "confidence": 70.0, "symbols": ["HPG"], "keywords": []},
])

# Patch before importing app
with patch.dict('sys.modules', {'torch': MagicMock(), 'transformers': MagicMock()}):
    from app.main import app


@pytest.fixture
def client():
    """Create a test client."""
    return TestClient(app)


class TestHealthEndpoint:
    """Test cases for health endpoint."""

    def test_health_check(self, client):
        """Test health check returns 200."""
        response = client.get("/health")
        assert response.status_code == 200
        assert "status" in response.json()

    def test_ready_check(self, client):
        """Test readiness check returns 200."""
        response = client.get("/ready")
        assert response.status_code == 200


class TestAnalyzeEndpoint:
    """Test cases for analyze endpoints."""

    def test_analyze_text(self, client):
        """Test single text analysis."""
        from app.routers import analyze
        # Set analyzer on the router module
        analyze._analyzer = mock_analyzer_instance
        response = client.post(
            "/analyze",
            json={
                "texts": [
                    {"id": "1", "content": "Cổ phiếu VNM tăng mạnh"}
                ]
            }
        )

        assert response.status_code == 200
        data = response.json()
        assert "results" in data

    def test_analyze_batch(self, client):
        """Test batch analysis."""
        from app.routers import analyze
        analyze._analyzer = mock_analyzer_instance
        response = client.post(
            "/analyze",
            json={
                "texts": [
                    {"id": "1", "content": "Cổ phiếu VNM tăng"},
                    {"id": "2", "content": "Thị trường giảm điểm"}
                ]
            }
        )

        assert response.status_code == 200
        data = response.json()
        assert "results" in data
        assert len(data["results"]) == 2


class TestAPIKeyAuth:
    """Test cases for API key authentication."""

    def test_missing_api_key_when_enabled(self, client, monkeypatch):
        """Test request without API key when auth is enabled."""
        monkeypatch.setenv("ENABLE_AUTH", "true")
        monkeypatch.setenv("API_KEY", "test-secret-key")
        # Clear settings cache so env vars take effect
        from app.config import get_settings
        get_settings.cache_clear()

        response = client.post(
            "/analyze",
            json={"texts": [{"id": "1", "content": "test"}]}
        )
        assert response.status_code == 401

        # Restore
        get_settings.cache_clear()

    def test_invalid_api_key(self, client, monkeypatch):
        """Test request with invalid API key."""
        monkeypatch.setenv("ENABLE_AUTH", "true")
        monkeypatch.setenv("API_KEY", "test-secret-key")
        from app.config import get_settings
        get_settings.cache_clear()

        response = client.post(
            "/analyze",
            json={"texts": [{"id": "1", "content": "test"}]},
            headers={"X-API-Key": "wrong-key"}
        )
        assert response.status_code == 401

        # Restore
        get_settings.cache_clear()


class TestCORS:
    """Test cases for CORS configuration."""

    def test_cors_headers_present(self, client):
        """Test CORS headers are present in response."""
        response = client.options(
            "/api/v1/analyze",
            headers={
                "Origin": "http://localhost:3000",
                "Access-Control-Request-Method": "POST",
                "Access-Control-Request-Headers": "Content-Type"
            }
        )
        assert response.status_code == 200
