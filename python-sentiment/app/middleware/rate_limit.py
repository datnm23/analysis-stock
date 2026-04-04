"""
Rate limiting middleware using Redis for distributed rate limiting.
"""
import logging
import time
from typing import Callable, List, Optional

from fastapi import Request, Response
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.types import ASGIApp

from app.config import get_settings

logger = logging.getLogger(__name__)


class RateLimitMiddleware(BaseHTTPMiddleware):
    """
    Redis-based rate limiting middleware using sliding window algorithm.

    Limits requests per IP address within a configurable time window.
    """

    def __init__(
        self,
        app: ASGIApp,
        redis_client,
        requests_per_window: int = 100,
        window_seconds: int = 60,
        paths_to_exclude: Optional[List[str]] = None
    ):
        super().__init__(app)
        self.app = app
        self.redis_client = redis_client
        self.requests_per_window = requests_per_window
        self.window_seconds = window_seconds
        self.paths_to_exclude = paths_to_exclude or ["/health", "/docs", "/redoc", "/openapi.json"]

    def _get_client_ip(self, request: Request) -> str:
        """Extract client IP from request headers or remote address."""
        # Check for forwarded header (proxy)
        forwarded = request.headers.get("X-Forwarded-For")
        if forwarded:
            return forwarded.split(",")[0].strip()

        # Check for real IP header
        real_ip = request.headers.get("X-Real-IP")
        if real_ip:
            return real_ip

        # Fall back to client host
        if request.client:
            return request.client.host

        return "unknown"

    def _get_rate_limit_key(self, client_ip: str) -> str:
        """Generate Redis key for rate limiting."""
        return f"ratelimit:{client_ip}"

    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        """Process request with rate limiting."""
        # Skip rate limiting for excluded paths
        if any(request.url.path.startswith(path) for path in self.paths_to_exclude):
            return await call_next(request)

        # Skip if Redis is not available
        if not self.redis_client:
            return await call_next(request)

        client_ip = self._get_client_ip(request)
        rate_key = self._get_rate_limit_key(client_ip)

        try:
            # Atomically increment counter and set TTL using Lua script.
            # Prevents race condition between INCR and EXPIRE.
            lua_script = """
local count = redis.call('INCR', KEYS[1])
if count == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
"""
            current = await self.redis_client.eval(
                lua_script, 1, rate_key, self.window_seconds
            )

            # Check if over limit
            if current > self.requests_per_window:
                logger.warning(
                    f"Rate limit exceeded for {client_ip}: "
                    f"{current}/{self.requests_per_window} requests"
                )

                # Get remaining TTL for retry-after header.
                # Redis TTL returns -2 (key missing) or -1 (no expiry) on edge cases.
                ttl = await self.redis_client.ttl(rate_key)
                if ttl <= 0:
                    # Key has no TTL (Lua script may have failed) — set a safety expiry
                    await self.redis_client.expire(rate_key, self.window_seconds)
                    retry_after = self.window_seconds
                else:
                    retry_after = ttl

                return JSONResponse(
                    status_code=429,
                    content={
                        "error": "Rate limit exceeded",
                        "message": f"Maximum {self.requests_per_window} requests "
                                   f"per {self.window_seconds} seconds",
                        "retry_after": retry_after
                    },
                    headers={
                        "Retry-After": str(retry_after),
                        "X-RateLimit-Limit": str(self.requests_per_window),
                        "X-RateLimit-Remaining": "0"
                    }
                )

            # Add rate limit headers to successful responses
            response = await call_next(request)
            response.headers["X-RateLimit-Limit"] = str(self.requests_per_window)
            response.headers["X-RateLimit-Remaining"] = str(
                self.requests_per_window - current
            )

            return response

        except Exception as e:
            # If Redis fails, log and allow the request (fail open)
            logger.error(f"Rate limiting error: {e}")
            return await call_next(request)


def create_rate_limiter(redis_client) -> RateLimitMiddleware:
    """
    Create rate limiting middleware with configuration from settings.

    Args:
        redis_client: Redis client instance

    Returns:
        Configured RateLimitMiddleware instance
    """
    settings = get_settings()

    return RateLimitMiddleware(
        app=None,  # Will be set when added to app
        redis_client=redis_client,
        requests_per_window=settings.rate_limit_requests,
        window_seconds=settings.rate_limit_window
    )
