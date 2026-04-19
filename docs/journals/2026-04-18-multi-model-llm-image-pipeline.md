# Multi-Model LLM + Gemini Image Generation Pipeline Complete

**Date**: 2026-04-18 14:30  
**Severity**: Medium  
**Component**: Auto-blog pipeline (Phase 7)  
**Status**: Resolved  

## What Happened

Completed Phase 7: integrated multi-model LLM routing (Claude/Gemini) + Gemini Imagen image generation into the auto-blog pipeline. Core refactors: new `LLMClient` dispatcher, `ImagePipeline` orchestrator, and updated Go Article model for nullable images.

## The Brutal Truth

We shipped security issues that code review barely caught (6.5/10 score initially). API key exposure in query parameters would've leaked credentials across logs and error messages—a junior mistake that only surfaced during review. The path traversal vulnerability in S3 key construction was even more embarrassing: we were directly interpolating Redis `symbol` values without sanitization. Production would've been one malicious ticker symbol away from arbitrary S3 directory writes.

## Technical Details

**LLMClient** routes via `ARTICLE_MODEL=claude|gemini|auto` config:
```python
if model == "gemini":
    headers = {"x-goog-api-key": self.api_key}  # Fixed: was in URL params
    response = await client.post("https://generativelanguage.googleapis.com/v1/generateText", headers=headers)
```

**ImagePipeline** flow:
1. Claude Haiku → image prompt JSON
2. Gemini Imagen API → PNG bytes
3. boto3 → S3 with sanitized key: `f"blog-images/{_safe_symbol(symbol)}/{uuid}.png"`
4. URL → PostgreSQL `article.image_url`

**Go Article model** changed from `ImageURL string` to `ImageURL *string` (NULL semantics for missing images, not empty strings).

## What We Tried

Initial implementation piped API keys through query params for simplicity—caught by reviewer. S3 path sanitization wasn't considered until security audit flagged the symbol-to-path conversion.

## Root Cause Analysis

**API key in params**: Assumption that httpx headers weren't available (they are—we just didn't check docs).  
**Path traversal**: YAGNI thinking—"symbols are controlled input"—ignored the principle that external data is never trusted.

## Lessons Learned

1. **Headers over params for secrets**: Always. Query params live in logs/error traces.
2. **Sanitize all external data as S3 keys**: Even "controlled" inputs from Redis. Use allowlist regex: `^[A-Z0-9]{3,4}$` for Vietnamese tickers.
3. **NULL vs empty string**: Distinguish semantic meanings. Missing image ≠ empty URL.
4. **Code review catches what automation misses**: Linters won't flag these—need human eyes on auth/path logic.

## Next Steps

- [ ] Audit all S3 key construction for similar path injection vulnerabilities
- [ ] Enforce header-based auth for all third-party APIs via linter (custom rule)
- [ ] Document NULL field semantics in Go model comments
- [ ] Add integration test: malicious `symbol` values → S3 path rejection

**Commit**: 1649878

