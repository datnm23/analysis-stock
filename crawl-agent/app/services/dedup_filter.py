"""
Near-duplicate detection using SimHash for filtering copy-paste spam.

Maintains a sliding window of recent SimHash fingerprints in Redis.
Texts with Hamming distance ≤ threshold are flagged as near-duplicates.
"""

import hashlib
import logging
import struct
from typing import Optional, Tuple

logger = logging.getLogger(__name__)

# Number of bits in the SimHash fingerprint
HASH_BITS = 64
# Maximum Hamming distance to consider two texts as near-duplicates
DEFAULT_HAMMING_THRESHOLD = 3
# How many recent fingerprints to keep per symbol in Redis
MAX_RECENT_HASHES = 10_000
# TTL for fingerprint entries (24 hours)
HASH_TTL_SECONDS = 86_400


def _tokenize(text: str) -> list[str]:
    """Simple character n-gram tokenizer (3-grams) for SimHash.

    Character n-grams are language-agnostic and work well for
    near-duplicate detection without requiring a word segmenter.
    """
    text = text.lower().strip()
    n = 3
    if len(text) < n:
        return [text]
    return [text[i:i + n] for i in range(len(text) - n + 1)]


def simhash(text: str) -> int:
    """Compute a 64-bit SimHash fingerprint for a text string."""
    tokens = _tokenize(text)
    v = [0] * HASH_BITS

    for token in tokens:
        token_hash = int(hashlib.md5(token.encode("utf-8")).hexdigest(), 16)
        for i in range(HASH_BITS):
            if token_hash & (1 << i):
                v[i] += 1
            else:
                v[i] -= 1

    fingerprint = 0
    for i in range(HASH_BITS):
        if v[i] > 0:
            fingerprint |= (1 << i)
    return fingerprint


def hamming_distance(a: int, b: int) -> int:
    """Count the number of differing bits between two integers."""
    return bin(a ^ b).count("1")


class DedupFilter:
    """Near-duplicate filter backed by Redis for distributed state."""

    def __init__(
        self,
        redis_client=None,
        hamming_threshold: int = DEFAULT_HAMMING_THRESHOLD,
    ):
        self.redis = redis_client
        self.threshold = hamming_threshold
        # In-memory fallback when Redis is unavailable
        self._local_hashes: list[int] = []

    async def is_duplicate(self, text: str, symbol: str = "global") -> Tuple[bool, int]:
        """Check if text is a near-duplicate of a recently seen text.

        Returns (is_dup, fingerprint).
        """
        fp = simhash(text)

        recent = await self._get_recent_hashes(symbol)
        for stored_fp in recent:
            if hamming_distance(fp, stored_fp) <= self.threshold:
                logger.info(
                    "Near-duplicate detected (hamming=%d, symbol=%s)",
                    hamming_distance(fp, stored_fp),
                    symbol,
                )
                return True, fp

        await self._store_hash(symbol, fp)
        return False, fp

    async def _get_recent_hashes(self, symbol: str) -> list[int]:
        """Retrieve recent fingerprints from Redis or local cache."""
        key = f"dedup:{symbol}:hashes"

        if self.redis:
            try:
                raw = await self.redis.lrange(key, 0, MAX_RECENT_HASHES - 1)
                return [int(h) for h in raw]
            except Exception as e:
                logger.warning("Redis dedup read failed: %s", e)

        return list(self._local_hashes)

    async def _store_hash(self, symbol: str, fingerprint: int) -> None:
        """Store a fingerprint in Redis or local cache."""
        key = f"dedup:{symbol}:hashes"

        if self.redis:
            try:
                pipe = self.redis.pipeline()
                pipe.lpush(key, str(fingerprint))
                pipe.ltrim(key, 0, MAX_RECENT_HASHES - 1)
                pipe.expire(key, HASH_TTL_SECONDS)
                await pipe.execute()
                return
            except Exception as e:
                logger.warning("Redis dedup write failed: %s", e)

        # Local fallback
        self._local_hashes.append(fingerprint)
        if len(self._local_hashes) > MAX_RECENT_HASHES:
            self._local_hashes = self._local_hashes[-MAX_RECENT_HASHES:]
