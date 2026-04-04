"""API Key authentication middleware for Python sentiment service."""

import hmac

from fastapi import Request, HTTPException, status
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware

from app.config import get_settings


class APIKeyMiddleware(BaseHTTPMiddleware):
    """Middleware to validate API key for service-to-service communication."""

    def __init__(self, app, skip_paths: list = None):
        super().__init__(app)
        self.skip_paths = skip_paths or ["/health", "/ready", "/docs", "/redoc", "/openapi.json"]

    async def dispatch(self, request: Request, call_next):
        settings = get_settings()

        # Skip auth if disabled or no API key configured
        if not settings.enable_auth or not settings.api_key:
            return await call_next(request)

        # Skip auth for certain paths
        if request.url.path in self.skip_paths or request.url.path.startswith("/docs"):
            return await call_next(request)

        # Check API key
        api_key = request.headers.get("X-API-Key")
        if not api_key:
            return JSONResponse(
                status_code=status.HTTP_401_UNAUTHORIZED,
                content={"error": "missing API key", "code": "MISSING_API_KEY"}
            )

        if not hmac.compare_digest(api_key, settings.api_key):
            return JSONResponse(
                status_code=status.HTTP_401_UNAUTHORIZED,
                content={"error": "invalid API key", "code": "INVALID_API_KEY"}
            )

        return await call_next(request)
