# World Monitor — Architectural System Overview

> **Document Version**: 1.0  
> **Date**: February 24, 2026  
> **Author**: Senior Software Architect Review  
> **Project Version**: 2.5.4

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [System Design & Architectural Pattern](#2-system-design--architectural-pattern)
3. [Technology Stack](#3-technology-stack)
4. [Core Components](#4-core-components)
5. [Data Flow Architecture](#5-data-flow-architecture)
6. [Security Architecture](#6-security-architecture)
7. [Scalability & Performance](#7-scalability--performance)
8. [Deployment Topology](#8-deployment-topology)
9. [API Contract Architecture](#9-api-contract-architecture)
10. [Observability & Error Handling](#10-observability--error-handling)

---

## 1. Executive Summary

**World Monitor** is a real-time global intelligence dashboard that aggregates 150+ data sources (RSS feeds, military tracking, market data, satellite imagery, AI models) into a unified situational awareness interface. It delivers three specialized variants (Geopolitical, Tech, Finance) from a single codebase, deployable as a Web SPA, Progressive Web App, and native Desktop application (macOS, Windows, Linux).

### Business Value

| Value Proposition | Description |
|---|---|
| **OSINT Democratization** | Replaces expensive commercial intelligence tools ($10K+/yr) with a 100% free, open-source alternative |
| **Multi-Domain Fusion** | Correlates geopolitics, military, markets, infrastructure, and cyber threats in one view |
| **Local-First Privacy** | AI summarization runs entirely on local hardware (Ollama/LM Studio) — no data leaves the machine |
| **Analyst Productivity** | 35+ map layers, AI-synthesized briefs, and automated anomaly detection reduce analysis time from hours to minutes |

---

## 2. System Design & Architectural Pattern

### Primary Pattern: **Hybrid Edge-Serverless + Client-Heavy SPA**

World Monitor is **not** a traditional monolithic or microservices application. It employs a **hybrid architecture** combining:

| Pattern | Implementation |
|---|---|
| **Edge-Serverless** | 60+ Vercel Edge Functions act as a lightweight, stateless API gateway — no monolithic backend |
| **Client-Heavy SPA** | Core intelligence processing (clustering, instability scoring, surge detection, ML inference) runs **in the browser**, reducing server dependency |
| **Proto-First Contract-Driven** | 17 API service domains defined in Protocol Buffers with auto-generated TypeScript clients, servers, and OpenAPI documentation |
| **Tri-Variant Monorepo** | A single codebase produces three specialized dashboards via build-time configuration (`VITE_VARIANT`) |
| **Multi-Runtime** | Same application runs on Vercel (web), Tauri (desktop with Rust + Node.js sidecar), and as an installable PWA |

### Architectural Diagram (Textual)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        CLIENT TIER (Browser / Tauri WebView)           │
│  ┌──────────┐ ┌───────────┐ ┌────────────┐ ┌───────────┐ ┌──────────┐ │
│  │ MapLibre │ │  deck.gl  │ │ Panels (50+│ │ ML Worker │ │ Analysis │ │
│  │ GL (2D)  │ │ (3D Globe)│ │ components)│ │(Transformer│ │  Worker  │ │
│  └──────────┘ └───────────┘ └────────────┘ │  s.js)    │ └──────────┘ │
│                                             └───────────┘              │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │           TypeScript App Core (App.ts — 4,600+ lines)          │   │
│  │  Services: 90+ modules | Config: 25+ data files | i18n: 14 lang│   │
│  └─────────────────────────────────────────────────────────────────┘   │
└───────────────────────────┬───────────────────────────────────────────┘
                            │  HTTPS / WSS
                            ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                     EDGE / API TIER (Vercel Edge Functions)             │
│  ┌─────────────────────┐  ┌──────────────────────────────────────────┐ │
│  │ Legacy Endpoints     │  │ Proto Gateway (api/[domain]/v1/[rpc])   │ │
│  │ (api/*.js — 15+)    │  │ 17 typed services, Map-based router     │ │
│  │ RSS Proxy, Download, │  │ CORS enforcement, error boundary        │ │
│  │ OG Images, Version   │  │ Rate limiting, field validation         │ │
│  └─────────────────────┘  └──────────────────────────────────────────┘ │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ Middleware: Bot detection, social crawler allowlist              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└───────────────────────────┬───────────────────────────────────────────┘
                            │
              ┌─────────────┼──────────────┐
              ▼             ▼              ▼
┌──────────────────┐ ┌────────────┐ ┌─────────────────┐
│ Redis (Upstash)  │ │  Railway   │ │ External APIs   │
│ Cache + Baseline │ │  Relay     │ │ (30+ providers) │
│ AI Deduplication │ │  WebSocket │ │ GDELT, ACLED,   │
│ Temporal State   │ │  AIS, RSS  │ │ Yahoo, OpenSky, │
│                  │ │  OpenSky   │ │ NASA, USGS ...  │
└──────────────────┘ └────────────┘ └─────────────────┘

              ┌───── DESKTOP RUNTIME (Optional) ─────┐
              │  Tauri 2 (Rust shell)                 │
              │  ├── Node.js Sidecar (port 46123)     │
              │  │   └── 60+ local API handlers       │
              │  ├── OS Keychain (credentials vault)  │
              │  ├── Token-authenticated IPC           │
              │  └── Cloud fallback on local failure   │
              └───────────────────────────────────────┘
```

---

## 3. Technology Stack

### 3.1 Languages

| Language | Role | LOC Estimate |
|---|---|---|
| **TypeScript** | Frontend SPA, services, components, API handlers, config, generated code | ~90% of codebase |
| **JavaScript (ESM)** | Legacy API edge functions (`api/*.js`), build scripts | ~5% |
| **Rust** | Tauri desktop shell, keychain integration, native TLS bridge | ~3% |
| **Protocol Buffers** | API contract definitions (92 `.proto` files across 18 domains) | ~2% |

### 3.2 Frameworks & Libraries

| Category | Technology | Purpose |
|---|---|---|
| **Build System** | Vite 6 | Dev server, HMR, production bundling, tree-shaking |
| **3D Rendering** | deck.gl 9 + MapLibre GL 5 | WebGL globe, 35+ data layers, 60fps rendering |
| **Desktop** | Tauri 2 | Native window, IPC, keychain, sidecar management |
| **PWA** | vite-plugin-pwa (Workbox) | Service worker, offline caching, install prompt |
| **AI/ML (Browser)** | Transformers.js (@xenova) + ONNX Runtime | NER, embeddings, T5 summarization — zero server dependency |
| **AI/ML (Cloud)** | Groq (Llama 3.1 8B), OpenRouter | LLM summarization, threat classification |
| **AI/ML (Local)** | Ollama / LM Studio | Air-gapped AI inference on local hardware |
| **Visualization** | D3.js 7 | SVG charts, timelines, score rings, sparklines |
| **Localization** | i18next | 14 languages, RTL support, lazy-loaded bundles |
| **Analytics** | PostHog, Vercel Analytics, Sentry | Usage tracking, error monitoring, performance |
| **Backend DB** | Convex | User registration (email interest list) |

### 3.3 Data Stores

| Store | Type | Role |
|---|---|---|
| **Redis (Upstash)** | Key-Value (serverless) | 3-tier cache, AI deduplication, temporal baselines (Welford state), rate limiting |
| **Convex** | Document DB (serverless) | User registration data |
| **IndexedDB** | Browser-side | Historical snapshots, playback state, persistent cache |
| **localStorage** | Browser-side | Panel layout, theme, feature toggles, variant preference |
| **OS Keychain** | System credential store | API keys (macOS Keychain / Windows Credential Manager) |

### 3.4 Infrastructure & Deployment

| Component | Provider | Role |
|---|---|---|
| **Web Hosting** | Vercel | Static SPA + 60+ Edge Functions + CDN |
| **Relay Server** | Railway | WebSocket relay (AIS), OpenSky OAuth2 proxy, blocked-domain RSS proxy |
| **Caching** | Upstash Redis | Cross-user cache, AI dedup, temporal state |
| **DNS/CDN** | Vercel Edge Network | Global distribution, TLS termination |
| **Desktop Distribution** | GitHub Releases | macOS (.dmg), Windows (.exe/.msi), Linux (.AppImage) |
| **API Contracts** | buf.build ecosystem | Protobuf linting, breaking-change detection, code generation |

---

## 4. Core Components

### 4.1 Frontend Layer

| Component | File(s) | Responsibility |
|---|---|---|
| **App Core** | `src/App.ts` (4,600+ lines) | Central orchestrator — initializes all services, manages refresh cycles, coordinates 50+ panels |
| **Map Engine** | `src/components/Map.ts`, `MapContainer.ts`, `DeckGLMap.ts` | MapLibre GL (2D) + deck.gl (3D globe), 35+ toggleable layers, cluster rendering |
| **Panel System** | `src/components/*.ts` (53 panel components) | Modular UI panels — each is self-contained with its own data fetching and rendering |
| **Virtual List** | `src/components/VirtualList.ts` | DOM-recycling scroll renderer for high-volume news panels (15+ items) |
| **Search** | `src/components/SearchModal.ts` | Cmd+K fuzzy search across 20+ result types |
| **Country Briefs** | `src/components/CountryBriefPage.ts` | Full-page intelligence dossier with CII ring, AI analysis, timeline, export |
| **Story Sharing** | `src/components/StoryModal.ts` | Social sharing with canvas-rendered PNG, QR codes, deep links |

### 4.2 Services Layer (90+ Modules)

| Domain | Key Modules | Responsibility |
|---|---|---|
| **Intelligence** | `country-instability.ts`, `focal-point-detector.ts`, `hotspot-escalation.ts` | CII scoring (22 countries), focal point detection, escalation analysis |
| **Signal Processing** | `signal-aggregator.ts`, `geo-convergence.ts`, `temporal-baseline.ts` | Multi-source fusion, geographic convergence, anomaly detection (Welford's) |
| **Military** | `military-flights.ts`, `military-vessels.ts`, `military-surge.ts`, `usni-fleet.ts` | ADS-B tracking, AIS monitoring, surge detection, theater posture |
| **AI Pipeline** | `summarization.ts`, `threat-classifier.ts`, `ml-worker.ts`, `ml-capabilities.ts` | 4-tier LLM chain, hybrid threat classification, browser-side ML |
| **Data Feeds** | `rss.ts`, `conflict/`, `displacement/`, `climate/`, `cyber/`, `wildfires/` | 150+ RSS feeds, ACLED/UCDP, UNHCR/HAPI, climate anomalies, IOC feeds, VIIRS |
| **Markets** | `market/`, `economic/`, `prediction/` | Yahoo Finance, CoinGecko, Polymarket, FRED, EIA, macro signals |
| **Infrastructure** | `cable-health.ts`, `infrastructure-cascade.ts`, `related-assets.ts` | Cable monitoring, cascade modeling, proximity correlation |
| **Geospatial** | `country-geometry.ts`, `geo-hub-index.ts`, `geo-activity.ts` | Browser-side ray-casting, hub matching, geo-location |
| **Desktop** | `tauri-bridge.ts`, `runtime.ts`, `runtime-config.ts` | IPC bridge, platform detection, feature toggle management |

### 4.3 Workers

| Worker | File | Responsibility |
|---|---|---|
| **ML Worker** | `src/workers/ml.worker.ts` | Runs Transformers.js models (embeddings, NER, T5) in a Web Worker to avoid blocking the main thread |
| **Analysis Worker** | `src/workers/analysis.worker.ts` | Background computation for clustering, correlation, and scoring |

### 4.4 API Layer (Edge Functions)

| Generation | Location | Count | Pattern |
|---|---|---|---|
| **Legacy** | `api/*.js` | ~15 | One file per concern (RSS proxy, download, OG images, version) |
| **Proto-First** | `api/[domain]/v1/` | 17 services | Single gateway function, Map-based routing, typed handlers |
| **Middleware** | `middleware.ts` | 1 | Bot detection, social crawler allowlist |

### 4.5 Configuration Data

| Category | Location | Content |
|---|---|---|
| **Geographic** | `src/config/geo.ts`, `bases-expanded.ts`, `pipelines.ts`, `ports.ts` | 220+ military bases, 200+ pipelines, 83 ports, undersea cables, nuclear facilities |
| **Finance** | `src/config/finance-geo.ts`, `gulf-fdi.ts` | 92 stock exchanges, 19 financial centers, 13 central banks, 64 Gulf FDI investments |
| **Tech** | `src/config/tech-companies.ts`, `ai-research-labs.ts`, `ai-datacenters.ts` | Tech HQs, AI labs, 111 datacenter clusters, startup ecosystems |
| **Feeds** | `src/config/feeds.ts` | 150+ RSS feeds with tier classification and propaganda risk ratings |
| **Entities** | `src/config/entities.ts`, `countries.ts` | Structured entity registry with alias/keyword indices |

---

## 5. Data Flow Architecture

### 5.1 Primary Data Flow

```
External APIs (30+)
    │
    ▼
Edge Functions (60+) ──── Redis Cache (Upstash) ──── CDN Cache (Vercel)
    │                         ▲
    │    cache miss            │ cache hit
    ▼                         │
Browser Client ───────────────┘
    │
    ├── Signal Aggregator (country + region clustering)
    ├── Geo-Convergence Engine (1°×1° cell binning)
    ├── Temporal Baseline (Welford's z-score anomaly detection)
    ├── ML Worker (NER, embeddings, T5 fallback)
    ├── CII Calculator (22 countries, 4-component weighted score)
    ├── Focal Point Detector (multi-source entity convergence)
    └── UI Panels (50+ components) → User
```

### 5.2 AI Summarization Flow

```
Headlines (150+ feeds) → Jaccard Deduplication (>60% overlap merged)
    │
    ▼
Redis Check (composite key: mode:variant:lang:hash)
    │
    ├── HIT → Return cached summary (24h TTL)
    │
    └── MISS → 4-Tier Provider Chain:
         ├── Tier 1: Ollama/LM Studio (local, 5s timeout)
         ├── Tier 2: Groq (Llama 3.1 8B, 5s timeout)
         ├── Tier 3: OpenRouter (multi-model, 5s timeout)
         └── Tier 4: Browser T5 (Transformers.js, no network)
              │
              ▼
         Redis Write → Return to UI
```

### 5.3 Threat Classification Pipeline

```
News Item → Keyword Classifier (~120 patterns, instant)
                │                         │
                ▼                         ▼ (async, background)
         UI renders immediately     LLM Classifier (Groq, temp=0)
                                         │
                                         ▼
                                   Redis Cache (24h TTL)
                                         │
                                         ▼
                                   Override keyword result
                                   (only if higher confidence)
```

### 5.4 Caching Architecture (3-Tier)

```
Request
  │
  ▼
[Tier 1] In-Memory Cache (per edge function instance, 60s–900s TTL)
  │ miss
  ▼
[Tier 2] Redis / Upstash (cross-user, cross-instance, 120s–900s TTL)
  │ miss
  ▼
[Tier 3] Upstream API (source of truth)
  │ error
  ▼
Stale data served from Tier 2 (stale-on-error fallback)
```

---

## 6. Security Architecture

### 6.1 Defense-in-Depth Model

```
Layer 1: Edge Middleware
  ├── Bot/crawler detection (UA regex, 30+ patterns)
  ├── Social crawler allowlist (Twitter, Facebook, etc. on OG routes only)
  └── Short/missing UA rejection

Layer 2: CORS & Domain Allowlists
  ├── Origin allowlist: worldmonitor.app, tech.*, finance.*, localhost:*
  ├── RSS domain allowlist: ~90+ explicitly listed domains
  └── Railway relay: separate smaller domain allowlist

Layer 3: API Key Isolation
  ├── Web: Vercel environment variables (never exposed to browser)
  ├── Desktop: OS keychain (consolidated JSON vault, 1 auth prompt)
  └── Runtime hot-patching via IPC (no restart required)

Layer 4: Input Validation & Sanitization
  ├── escapeHtml() — XSS prevention
  ├── sanitizeUrl() — blocks javascript: and data: URIs
  ├── escapeAttr() — attribute context encoding
  ├── Query param regex validation (e.g., [a-z0-9-]+ for coin IDs)
  └── Proto field constraints (buf.validate: lat ∈ [-90, 90])

Layer 5: Rate Limiting & Circuit Breakers
  ├── Upstash Redis-backed IP rate limiting on AI endpoints
  ├── Per-feed circuit breakers (5-minute cooldowns)
  └── AI queue pause on 500 errors (quota protection)

Layer 6: Desktop Authentication
  ├── Per-session 32-char hex token (Rust RandomState)
  ├── Bearer token on every local sidecar request
  └── Health check endpoints exempt from auth
```

### 6.2 Privacy Levels

| Level | Mode | Data Leaves Machine |
|---|---|---|
| **Level 1** | Web App (Vercel) | Yes — all processing server-side |
| **Level 2** | Desktop + Cloud APIs | Partially — API keys local, some cloud calls |
| **Level 3** | Desktop + Ollama | No — full local AI pipeline, zero cloud dependency |

---

## 7. Scalability & Performance

### 7.1 Frontend Performance

| Technique | Implementation |
|---|---|
| **WebGL Rendering** | deck.gl + MapLibre GL for 60fps with thousands of concurrent markers |
| **Virtual Scrolling** | Custom DOM-recycling list with 3-item overscan, `requestAnimationFrame` batching |
| **Smart Clustering** | Supercluster adapts thresholds to zoom level; progressive disclosure for detail layers |
| **Idle-Aware Resources** | Animations pause after 2min inactivity; video streams destroyed from DOM; polling pauses on hidden tabs |
| **Tree-Shaking** | Variant-specific builds exclude unused data files (finance build drops military base data) |
| **Lazy Loading** | i18n bundles loaded on-demand; ML models loaded only when needed; webcam iframes via IntersectionObserver |
| **Brotli Pre-Compression** | Build-time `.br` files for assets >1KB (20-30% smaller than gzip) |

### 7.2 Backend Scalability

| Technique | Implementation |
|---|---|
| **Stateless Edge Functions** | Each function scales independently; no shared state beyond Redis |
| **CDN Caching** | `s-maxage` + `stale-while-revalidate` absorb repeated requests before origin |
| **AI Deduplication** | Content-hash Redis keys ensure 1,000 concurrent users trigger 1 LLM call |
| **Staggered Polling** | Panels refresh at different intervals (10s–5min) to prevent synchronized API storms |
| **Circuit Breakers** | Per-feed and per-API breakers with cooldowns prevent cascading failures |
| **Railway Relay** | Dedicated relay for WebSocket-heavy workloads (AIS vessel multiplexing) |

### 7.3 Bandwidth Optimization

| Layer | Savings |
|---|---|
| Brotli pre-compression (build-time) | 20-30% vs gzip |
| Gzip on all Railway relay responses | ~80% reduction |
| Content-hash assets with 1-year immutable cache | Zero re-download |
| CDN s-maxage on API responses | Eliminated repeated origin hits |
| Service worker tile caching (CacheFirst, 500 tiles, 30-day TTL) | Offline map support |

---

## 8. Deployment Topology

### 8.1 Multi-Platform Deployment Matrix

| Platform | Target | Build Command | Distribution |
|---|---|---|---|
| **Vercel (Web)** | worldmonitor.app, tech.*, finance.* | `npm run build:{full\|tech\|finance}` | Vercel Edge Network (global CDN) |
| **Railway (Relay)** | WebSocket relay | `node scripts/ais-relay.cjs` | Railway container |
| **Tauri (Desktop)** | macOS, Windows, Linux | `npm run desktop:build:{full\|tech\|finance}` | GitHub Releases (.dmg, .exe, .msi, .AppImage) |
| **PWA** | Installable web app | Same as Vercel build | Service worker + web manifest |

### 8.2 Tri-Variant Build System

```
              ┌── VITE_VARIANT=full ──── worldmonitor.app (44 panels)
              │
Source Code ──┼── VITE_VARIANT=tech ──── tech.worldmonitor.app (31 panels)
              │
              └── VITE_VARIANT=finance ── finance.worldmonitor.app (30 panels)
```

Each variant tree-shakes unused configuration data, RSS feeds, map layers, panels, and SEO metadata at build time.

---

## 9. API Contract Architecture

### 9.1 Proto-First Pipeline

```
.proto files (92 files, 17 service domains)
    │
    ▼ buf generate (Makefile)
    │
    ├── protoc-gen-ts-client → src/generated/client/ (typed fetch clients)
    ├── protoc-gen-ts-server → src/generated/server/ (handler stubs + route descriptors)
    └── protoc-gen-openapiv3 → docs/api/ (OpenAPI 3.1.0 YAML + JSON)
```

### 9.2 Service Domains (17)

`aviation` · `climate` · `conflict` · `cyber` · `displacement` · `economic` · `infrastructure` · `intelligence` · `maritime` · `market` · `military` · `news` · `prediction` · `research` · `seismology` · `unrest` · `wildfire`

### 9.3 Edge Gateway Pattern

All proto-first endpoints route through a single Vercel Edge Function (`api/[domain]/v1/[rpc].ts`) that:

1. Imports all 17 `createServiceRoutes()` into a flat `Map<string, handler>`
2. Matches `POST /api/{domain}/v1/{rpc}` to the correct handler
3. Enforces CORS, error boundary (hides internals on 5xx), and rate limiting
4. Returns `retryAfter` on 429 responses

---

## 10. Observability & Error Handling

### 10.1 Error Tracking (Sentry)

- Environment-aware routing (production, preview, disabled on localhost/Tauri)
- 30+ `ignoreErrors` patterns (third-party injections, WebGL context loss, iOS quirks)
- Custom `beforeSend` hook for second-stage filtering
- 10% transaction sampling
- Release tracking (`worldmonitor@{version}`)

### 10.2 Resilience Patterns

| Pattern | Implementation |
|---|---|
| **4-Tier AI Fallback** | Ollama → Groq → OpenRouter → Browser T5 (never blocked) |
| **Circuit Breakers** | Per-feed, 5-minute cooldown; AI queue pause on 500 errors |
| **Stale-on-Error** | Redis serves cached data when upstream APIs fail |
| **Chunk Reload Guard** | One-shot page reload on stale-asset 404s after deployments |
| **Storage Quota Management** | Graceful degradation on exhausted localStorage/IndexedDB |
| **Intelligence Gap Reporting** | Explicitly shows what data sources are down (never silently hides) |

### 10.3 Desktop Observability

- Traffic log: ring buffer of last 200 requests with timing
- Dual log files: `desktop.log` (Rust) + `local-api.log` (Node.js)
- Verbose debug mode: toggle via API, persists across restarts
- DevTools: `Cmd+Alt+I` opens embedded web inspector

---

*This document provides a comprehensive architectural overview of World Monitor v2.5.4. For implementation details, see the [full documentation](./DOCUMENTATION.md).*
