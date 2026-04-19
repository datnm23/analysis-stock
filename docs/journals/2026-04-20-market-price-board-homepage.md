# Market Price Board & Homepage Redesign

**Date**: 2026-04-20 04:30
**Severity**: High
**Component**: Frontend (Next.js), Backend (Go API Gateway, Python crawl-agent)
**Status**: Resolved

## What Happened

Completed a comprehensive redesign of the blog homepage into a full market overview page modeled after VietStock Finance. The page now displays live market data across 1,739 stock symbols with real-time pricing, sector filtering, and AI screener integration. Successfully integrated multiple data sources (VCI listing API, KBS ISS real-time pricing) and bridged architectural gaps between Go API gateway and Python crawl-agent services.

## The Brutal Truth

This was exhausting because we discovered that **vnstock 3.5.1 silently dropped the `price_board` endpoint** — the entire foundation of our initial plan evaporated mid-implementation. We wasted 2+ hours chasing a dead API before realizing it was gone. Then came port conflicts, duplicate symbol errors, and API discovery that consumed another 3 hours. The frustrating part is that none of these issues are rocket science, but they all cascaded because we didn't validate external dependencies upfront.

The real kick in the teeth: we had to expose crawl-agent on port 8085 instead of the "standard" 8000 because `boq_backend` was occupying it. This created configuration friction and potential for environment-specific bugs downstream. For a system this distributed, API surface area is critical — and we have three ports flying around now (crawl-agent: 8085, go-gateway: 8080, web dashboard: 3000).

## Technical Details

**vnstock API Issue:**
```python
# Expected (vnstock 3.5.1 documentation)
data = vnstock.price_board()  # ❌ AttributeError: module has no attribute 'price_board'

# Root cause: TCBS removed endpoint, vnstock wrapped it without deprecation notice
# Discovery: Only found via runtime error, no changelog mention in 3.5.0 → 3.5.1
```

**KBS ISS API Details:**
```
GET https://kbbuddywts.kbsec.com.vn/iis-server/investment/stock/iss
Response: ~1739 symbols with real-time price, bid/ask, volume, change%
Rate limit: Unknown (batched async with httpx, 60s total wall time)
Reliability: High (confirmed stability over 5 test runs)
```

**VCI API Details:**
```
GET https://trading.vietcap.com.vn/api/price/symbols/getAll
Response: Symbol list with sector mapping
Rate limit: Unknown
Reliability: Good (source of truth for symbol coverage)
```

**Port Conflict:**
```
Port 8000: Occupied by boq_backend service (not our control)
Port 8085: Assigned to crawl-agent GET /market/board
Port 8080: Go API Gateway
Impact: MARKET_SERVICE_URL must be explicitly set in .env; localhost:8000 fails silently
```

**TypeScript Duplicate Symbol Error:**
```typescript
// sector-mapping.ts line ~450
"PVT": "Energy",  // First definition
"PVT": "Materials",  // Second definition → error: "Object literal may only specify known properties"
// Fix: Removed duplicate, kept first occurrence
```

## What We Tried

1. **Direct vnstock approach** (30min) - Called `vnstock.price_board()`, got AttributeError. Checked docs, saw it listed. Tried pinning vnstock 3.5.0, same error. Realized it never existed in public API.

2. **Fallback to vnstock price fetching** (45min) - Attempted per-symbol fetch loop. Quickly realized 1739 symbols × 5+ API calls each = impossible latency for homepage load.

3. **KBS ISS async batching** (90min) - Initially tried sequential requests, hit timeout. Pivoted to httpx async with semaphore (max 50 concurrent), achieved <60s for full board. Tested against rate limits, appeared stable.

4. **API Gateway caching strategy** (45min) - Implemented 5min Redis cache with graceful degradation: returns empty board instead of 502 on cache miss. Frontend handles empty state via skeleton loading.

5. **MarketIndexBar indices** (30min) - Initially tried Redis cache for index data (VNIndex, VN30, HNX, UPCOM). Switched to in-memory 2min cache (4 values, negligible memory, simpler lifecycle).

## Root Cause Analysis

**Why vnstock failed:**
- External API deprecation without library update is a classic dependency trap
- We trusted library documentation over actual runtime behavior
- No validation step before committing to architecture
- Should have spiked vnstock in isolation first

**Why port conflicts happened:**
- No centralized port registry or conflict detection in docker-compose
- boq_backend service brought into environment without coordination
- MARKET_SERVICE_URL hardcoded in places instead of using env-driven defaults

**Why duplicate symbol slipped through:**
- sector-mapping.ts hand-written without validation. No linter caught duplicate object keys (TypeScript allowed it until we tried to use it)
- Should have generated from API response or used a Set + validation

**Why API discovery took so long:**
- No systematic approach: searched for alternatives instead of analyzing what we actually needed (1739 symbols, real-time prices)
- Should have documented API requirements first, then matched sources

## Lessons Learned

1. **Validate external dependencies before architecture decisions** - Spike critical APIs in isolation. Test at target scale. Don't trust docs; test actual endpoints.

2. **Port management needs governance** - Create a central registry file (docs/port-registry.md) listing all ports and services. Check it before adding new services. Prevents cascading config pain.

3. **Explicit over implicit for service URLs** - MARKET_SERVICE_URL should have been required from day one, with .env.example showing all three ports. Defaults to localhost:8085, not localhost:8000.

4. **Data generation beats hand-written mappings** - For 200+ symbols, generate sector-mapping.ts from API response. Include validation: duplicate check, coverage report, last-updated timestamp.

5. **In-memory cache for tiny datasets** - Don't force Redis for 4 values refreshed every 2min. Reduces deployment complexity and failure modes. Redis is for heavy lifting (1739 symbols), not light.

6. **Graceful degradation over hard failures** - API Gateway returning empty board on cache miss > returning 502. Frontend shows loading state, user sees "market data unavailable" instead of error. Recovers automatically on next request.

## Next Steps

1. **Port registry documentation** (Owner: Infra lead, Timeline: this week)
   - Create `docs/port-registry.md` with all service ports
   - Add validation to docker-compose/scripts to prevent conflicts
   - Document why each port choice (e.g., "8085 avoids boq_backend collision")

2. **API validation layer** (Owner: Backend lead, Timeline: next sprint)
   - Add health check endpoint to crawl-agent that tests KBS ISS connectivity
   - Implement circuit breaker pattern for degraded KBS API (fallback to cached data)
   - Monitor API response times; alert if >90s

3. **Sector mapping validation** (Owner: Frontend lead, Timeline: before next release)
   - Generate sector-mapping.ts from VCI API response at build time
   - Add TypeScript validation script: duplicate check, coverage check, version stamp
   - Include in pre-commit hooks to catch mismatches early

4. **Configuration hardening** (Owner: DevOps, Timeline: this week)
   - Audit all service-to-service URLs for hardcoded defaults
   - Ensure .env.example matches actual defaults
   - Add startup check: validate MARKET_SERVICE_URL is reachable, fail fast if not

5. **Load testing** (Owner: QA/Performance team, Timeline: before production)
   - Benchmark KBS ISS with concurrent load: 100 requests/min, verify <60s response
   - Test API Gateway cache invalidation: confirm stale data doesn't serve >5min
   - Test graceful degradation: kill KBS ISS, verify empty board response, measure recovery time

## Unresolved Questions

- **KBS ISS rate limits**: We've never hit one in testing, but actual limits are undocumented. Should add monitoring to detect if we breach them.
- **Symbol coverage drift**: VCI says ~1739, KBS returns variable count. Should we validate coverage parity? Could miss new symbols.
- **Sector mapping timeliness**: Currently static. How often do companies change sectors? Should we refresh weekly or per-release?
