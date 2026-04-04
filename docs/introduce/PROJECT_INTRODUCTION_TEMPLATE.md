# Universal Project Introduction Template / Mẫu Giới Thiệu Dự Án Tổng Quát

> **Purpose / Mục đích**: A reusable Markdown template for introducing any technical project — covering business value, architecture, tech stack, and setup. Derived from the architectural analysis of a complex, multi-platform intelligence dashboard.
>
> **Cách sử dụng**: Thay thế tất cả các placeholder `[...]` bằng thông tin cụ thể của dự án. Xóa các phần không liên quan.

---

<!-- ============================================================ -->
<!-- ENGLISH VERSION -->
<!-- ============================================================ -->

# 🇬🇧 ENGLISH VERSION

---

# [Project Name]

> **[One-line tagline describing the project's core value]**

| Metadata | Value |
|---|---|
| **Version** | [x.y.z] |
| **License** | [License Type, e.g., MIT, AGPL-3.0, Apache-2.0] |
| **Status** | [Production / Beta / Alpha / Prototype] |
| **Last Updated** | [Date] |

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Business Value & Objectives](#business-value--objectives)
3. [Key Features](#key-features)
4. [System Architecture](#system-architecture)
5. [Technology Stack](#technology-stack)
6. [Core Components](#core-components)
7. [Data Flow](#data-flow)
8. [Security Model](#security-model)
9. [Scalability & Performance](#scalability--performance)
10. [Deployment Architecture](#deployment-architecture)
11. [Getting Started](#getting-started)
12. [Configuration](#configuration)
13. [Contributing](#contributing)
14. [Roadmap](#roadmap)
15. [License](#license)

---

## Executive Summary

[2-3 paragraphs describing what the project does, who it's for, and why it exists. Focus on the problem space and the solution approach.]

**Target Users**: [List primary user personas — e.g., Developers, Analysts, Operations Teams, End Users]

**Core Problem**: [What specific problem does this project solve?]

**Solution Approach**: [How does it solve it differently from existing alternatives?]

---

## Business Value & Objectives

| Problem | Solution |
|---|---|
| [Problem 1 — e.g., Data scattered across multiple sources] | [Solution 1 — e.g., Unified dashboard with 50+ integrated feeds] |
| [Problem 2 — e.g., High cost of commercial alternatives] | [Solution 2 — e.g., 100% open source, free to use] |
| [Problem 3 — e.g., No real-time processing capability] | [Solution 3 — e.g., Edge-computed real-time analytics] |
| [Problem 4 — e.g., Cloud-dependent, privacy concerns] | [Solution 4 — e.g., Local-first architecture with optional cloud] |

### Key Metrics / KPIs

| Metric | Target | Current |
|---|---|---|
| [e.g., Response Time] | [< 200ms] | [150ms] |
| [e.g., Data Sources Integrated] | [50+] | [35] |
| [e.g., Uptime] | [99.9%] | [99.95%] |

---

## Key Features

### [Feature Category 1 — e.g., Data Visualization]

- **[Feature Name]** — [Brief description of what it does and why it matters]
- **[Feature Name]** — [Brief description]

### [Feature Category 2 — e.g., AI/ML Capabilities]

- **[Feature Name]** — [Brief description]
- **[Feature Name]** — [Brief description]

### [Feature Category 3 — e.g., Real-Time Processing]

- **[Feature Name]** — [Brief description]
- **[Feature Name]** — [Brief description]

### [Feature Category 4 — e.g., Integration & Export]

- **[Feature Name]** — [Brief description]
- **[Feature Name]** — [Brief description]

---

## System Architecture

### Architectural Pattern

**Primary Pattern**: [e.g., Microservices / Monolithic / Serverless / Hybrid Edge-Serverless / Event-Driven]

| Aspect | Pattern | Rationale |
|---|---|---|
| **Overall Structure** | [e.g., Hybrid Edge-Serverless] | [e.g., Minimize backend complexity while keeping API keys server-side] |
| **API Design** | [e.g., Contract-First (Proto/OpenAPI)] | [e.g., Eliminate schema drift between frontend and backend] |
| **Data Processing** | [e.g., Client-Heavy SPA] | [e.g., Reduce server dependency; enable offline capabilities] |
| **Deployment** | [e.g., Multi-Runtime (Web + Desktop + PWA)] | [e.g., Reach users on all platforms from a single codebase] |

### Architecture Diagram

```
┌──────────────────────────────────────────────┐
│              CLIENT TIER                      │
│  [Component 1] [Component 2] [Component 3]   │
│  [Workers / Background Processes]             │
└──────────────────┬───────────────────────────┘
                   │ [Protocol: HTTPS / WSS / gRPC]
                   ▼
┌──────────────────────────────────────────────┐
│            API / GATEWAY TIER                 │
│  [API Gateway / Edge Functions / BFF]        │
│  [Middleware: Auth, CORS, Rate Limiting]      │
└──────────┬──────────┬──────────┬─────────────┘
           │          │          │
           ▼          ▼          ▼
┌──────────────┐ ┌─────────┐ ┌──────────────────┐
│ [Data Store] │ │ [Cache] │ │ [External APIs]  │
│ [DB Type]    │ │ [Redis] │ │ [Provider List]  │
└──────────────┘ └─────────┘ └──────────────────┘
```

> **Instructions**: Replace the diagram above with your actual architecture. Use ASCII art for Markdown compatibility. For complex systems, consider linking to an external diagram tool (Mermaid, draw.io, Excalidraw).

---

## Technology Stack

### Languages

| Language | Role | Percentage |
|---|---|---|
| [e.g., TypeScript] | [e.g., Frontend + Backend services] | [~80%] |
| [e.g., Rust] | [e.g., Desktop native shell, performance-critical paths] | [~10%] |
| [e.g., Python] | [e.g., Data pipelines, ML training] | [~10%] |

### Frameworks & Libraries

| Category | Technology | Purpose |
|---|---|---|
| **Frontend** | [e.g., React 18, Vue 3, Vanilla TS] | [e.g., Component-based UI] |
| **Backend** | [e.g., Node.js, FastAPI, Spring Boot] | [e.g., API services] |
| **Build** | [e.g., Vite, Webpack, esbuild] | [e.g., Dev server + production bundling] |
| **Testing** | [e.g., Playwright, Jest, Vitest] | [e.g., E2E + unit tests] |
| **AI/ML** | [e.g., Transformers.js, TensorFlow, PyTorch] | [e.g., Browser-side inference / model training] |
| **Visualization** | [e.g., D3.js, deck.gl, Three.js] | [e.g., Charts, maps, 3D rendering] |
| **Localization** | [e.g., i18next, vue-i18n] | [e.g., Multi-language support] |

### Data Stores

| Store | Type | Role |
|---|---|---|
| [e.g., PostgreSQL] | Relational | [e.g., Primary application data] |
| [e.g., Redis] | Key-Value | [e.g., Caching, session state, rate limiting] |
| [e.g., IndexedDB] | Browser-side | [e.g., Offline data, client-side cache] |

### Infrastructure

| Component | Provider | Role |
|---|---|---|
| [e.g., Hosting] | [e.g., Vercel, AWS, GCP] | [e.g., Edge-deployed SPA + API functions] |
| [e.g., CI/CD] | [e.g., GitHub Actions, GitLab CI] | [e.g., Automated testing + deployment] |
| [e.g., CDN] | [e.g., Cloudflare, Vercel Edge] | [e.g., Global asset distribution] |
| [e.g., Monitoring] | [e.g., Sentry, Datadog] | [e.g., Error tracking, performance monitoring] |

---

## Core Components

### [Component 1 — e.g., Frontend Application]

| Module | Responsibility |
|---|---|
| [e.g., App Core] | [e.g., Central orchestrator, service initialization, state management] |
| [e.g., UI Components (50+)] | [e.g., Self-contained panels with own data fetching and rendering] |
| [e.g., Web Workers] | [e.g., Background ML inference, heavy computation off main thread] |

### [Component 2 — e.g., API Layer]

| Module | Responsibility |
|---|---|
| [e.g., Edge Functions (60+)] | [e.g., Stateless API handlers for proxying, caching, transformation] |
| [e.g., Middleware] | [e.g., Authentication, CORS enforcement, bot detection] |
| [e.g., Proto Gateway] | [e.g., Single entry point routing to typed service handlers] |

### [Component 3 — e.g., Services / Business Logic]

| Domain | Modules | Responsibility |
|---|---|---|
| [e.g., Intelligence] | [e.g., scoring, detection, analysis] | [e.g., Real-time scoring and anomaly detection] |
| [e.g., Data Feeds] | [e.g., RSS, conflict, climate] | [e.g., External data ingestion and normalization] |
| [e.g., AI Pipeline] | [e.g., summarization, classification] | [e.g., Multi-tier LLM chain with fallback] |

---

## Data Flow

### Primary Data Flow

```
[Data Source(s)]
    │
    ▼
[Ingestion Layer — e.g., API Gateway / Edge Functions]
    │
    ├── [Cache Layer — check cache first]
    │
    ▼
[Processing Layer — e.g., normalization, enrichment, scoring]
    │
    ▼
[Storage Layer — e.g., database, cache write-back]
    │
    ▼
[Presentation Layer — e.g., UI components, API response]
```

### Caching Strategy

| Tier | Scope | TTL | Purpose |
|---|---|---|---|
| [Tier 1 — e.g., In-Memory] | [Per instance] | [60s–900s] | [Eliminate repeated remote calls] |
| [Tier 2 — e.g., Redis] | [Cross-user] | [120s–24h] | [Deduplicate across all visitors] |
| [Tier 3 — e.g., CDN] | [Global edge] | [Varies] | [Absorb repeated requests at edge] |

### Key Data Pipelines

| Pipeline | Input | Processing | Output |
|---|---|---|---|
| [e.g., AI Summarization] | [Raw headlines] | [Dedup → LLM chain → cache] | [Synthesized brief] |
| [e.g., Anomaly Detection] | [Event streams] | [Welford's online stats → z-score] | [Anomaly alerts] |
| [e.g., Threat Classification] | [News items] | [Keyword match (instant) + LLM (async)] | [Severity + category labels] |

---

## Security Model

### Defense Layers

| Layer | Mechanism | Details |
|---|---|---|
| **Network** | [e.g., CORS allowlist, TLS, WAF] | [e.g., Only allowed origins can call API endpoints] |
| **Authentication** | [e.g., JWT, API key, OAuth2, session tokens] | [e.g., Per-session tokens for desktop IPC; server-side API key isolation] |
| **Authorization** | [e.g., RBAC, ABAC, scope-based] | [e.g., Role-based access to admin endpoints] |
| **Input Validation** | [e.g., Schema validation, sanitization, regex guards] | [e.g., XSS prevention, SQL injection protection, proto field constraints] |
| **Rate Limiting** | [e.g., Redis-backed IP limiting, per-user quotas] | [e.g., AI endpoint abuse prevention] |
| **Secrets Management** | [e.g., Env vars, OS keychain, vault] | [e.g., Credentials stored in OS keychain, never in plaintext] |
| **Bot Protection** | [e.g., UA detection, CAPTCHA, fingerprinting] | [e.g., Middleware blocks known bot patterns on API routes] |

### Privacy Architecture

| Level | Mode | Data Locality |
|---|---|---|
| [Level 1 — e.g., Full Cloud] | [e.g., Web app, server-side processing] | [Data leaves machine] |
| [Level 2 — e.g., Hybrid] | [e.g., Desktop + cloud APIs] | [Partially local] |
| [Level 3 — e.g., Air-Gapped] | [e.g., Desktop + local AI] | [Zero cloud dependency] |

---

## Scalability & Performance

### Frontend Optimization

| Technique | Description |
|---|---|
| [e.g., Virtual Scrolling] | [DOM recycling for large lists — only visible items rendered] |
| [e.g., Code Splitting] | [Route-based lazy loading reduces initial bundle size] |
| [e.g., Web Workers] | [Heavy computation moved off main thread] |
| [e.g., Idle-Aware] | [Animations/polling pause when tab is hidden or user inactive] |
| [e.g., Compression] | [Brotli/gzip pre-compression for static assets] |

### Backend Scaling

| Technique | Description |
|---|---|
| [e.g., Stateless Functions] | [Each function scales independently; no shared mutable state] |
| [e.g., CDN Caching] | [Edge cache absorbs repeated requests before hitting origin] |
| [e.g., Request Deduplication] | [Content-hash keys ensure N concurrent users trigger 1 API call] |
| [e.g., Circuit Breakers] | [Per-source breakers with cooldowns prevent cascading failures] |
| [e.g., Staggered Polling] | [Different refresh intervals prevent synchronized API storms] |

---

## Deployment Architecture

### Environments

| Environment | URL | Purpose |
|---|---|---|
| **Production** | [e.g., https://app.example.com] | [Live user-facing deployment] |
| **Staging** | [e.g., https://staging.example.com] | [Pre-production validation] |
| **Development** | [e.g., http://localhost:3000] | [Local development] |

### Platform Matrix

| Platform | Build Command | Distribution |
|---|---|---|
| [e.g., Web (Vercel)] | `npm run build` | [Vercel Edge Network] |
| [e.g., Desktop (Tauri)] | `npm run desktop:build` | [GitHub Releases — .dmg, .exe, .AppImage] |
| [e.g., Mobile (Capacitor)] | `npm run mobile:build` | [App Store, Play Store] |
| [e.g., Docker] | `docker build .` | [Docker Hub / GHCR] |

---

## Getting Started

### Prerequisites

| Requirement | Version | Notes |
|---|---|---|
| [e.g., Node.js] | [≥ 18.0] | [LTS recommended] |
| [e.g., npm / pnpm / yarn] | [≥ 9.0] | [Package manager] |
| [e.g., Rust] | [≥ 1.70] | [Only for desktop builds] |
| [e.g., Docker] | [≥ 24.0] | [Optional — for containerized deployment] |

### Quick Start

```bash
# Clone the repository
git clone [https://github.com/org/project.git]
cd [project]

# Install dependencies
npm install

# Start development server
npm run dev

# Open in browser
# http://localhost:[port]
```

### Environment Variables

```bash
# Copy example environment file
cp .env.example .env.local
```

| Group | Variables | Required | Free Tier |
|---|---|---|---|
| [e.g., Database] | `DATABASE_URL` | Yes | [e.g., Supabase free tier] |
| [e.g., Auth] | `AUTH_SECRET`, `OAUTH_CLIENT_ID` | Yes | N/A |
| [e.g., AI] | `OPENAI_API_KEY`, `LOCAL_LLM_URL` | Optional | [e.g., Ollama — free, local] |
| [e.g., Cache] | `REDIS_URL`, `REDIS_TOKEN` | Optional | [e.g., Upstash free tier] |
| [e.g., Monitoring] | `SENTRY_DSN` | Optional | [e.g., Sentry free tier] |

---

## Configuration

### Build Variants (if applicable)

| Variant | Command | Description |
|---|---|---|
| [e.g., Full] | `npm run build:full` | [All features enabled] |
| [e.g., Lite] | `npm run build:lite` | [Reduced feature set for lightweight deployment] |
| [e.g., Enterprise] | `npm run build:enterprise` | [Additional enterprise features] |

### Feature Toggles

| Toggle | Default | Description |
|---|---|---|
| [e.g., ENABLE_AI] | `true` | [Enable AI-powered features] |
| [e.g., ENABLE_ANALYTICS] | `true` | [Enable usage analytics] |
| [e.g., DEBUG_MODE] | `false` | [Verbose logging and debug UI] |

---

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed guidelines.

```bash
# Development workflow
npm run dev          # Start dev server
npm run test         # Run tests
npm run lint         # Lint code
npm run typecheck    # TypeScript type checking
npm run build        # Production build
```

---

## Roadmap

- [x] [Completed Feature 1]
- [x] [Completed Feature 2]
- [ ] [Planned Feature 1]
- [ ] [Planned Feature 2]
- [ ] [Planned Feature 3]

---

## License

[License Type] — see [LICENSE](./LICENSE) for details.

---

## Author

**[Author Name]** — [GitHub](https://github.com/[username]) · [Website](https://[website])

---

<!-- ============================================================ -->
<!-- VIETNAMESE VERSION -->
<!-- ============================================================ -->

# 🇻🇳 PHIÊN BẢN TIẾNG VIỆT

---

# [Tên Dự Án]

> **[Một dòng mô tả giá trị cốt lõi của dự án]**

| Thông tin | Giá trị |
|---|---|
| **Phiên bản** | [x.y.z] |
| **Giấy phép** | [Loại giấy phép, ví dụ: MIT, AGPL-3.0, Apache-2.0] |
| **Trạng thái** | [Production / Beta / Alpha / Prototype] |
| **Cập nhật lần cuối** | [Ngày] |

---

## Mục Lục

1. [Tóm Tắt Tổng Quan](#tóm-tắt-tổng-quan)
2. [Giá Trị Kinh Doanh & Mục Tiêu](#giá-trị-kinh-doanh--mục-tiêu)
3. [Tính Năng Chính](#tính-năng-chính)
4. [Kiến Trúc Hệ Thống](#kiến-trúc-hệ-thống)
5. [Công Nghệ Sử Dụng](#công-nghệ-sử-dụng)
6. [Các Thành Phần Cốt Lõi](#các-thành-phần-cốt-lõi)
7. [Luồng Dữ Liệu](#luồng-dữ-liệu)
8. [Mô Hình Bảo Mật](#mô-hình-bảo-mật)
9. [Khả Năng Mở Rộng & Hiệu Năng](#khả-năng-mở-rộng--hiệu-năng)
10. [Kiến Trúc Triển Khai](#kiến-trúc-triển-khai)
11. [Bắt Đầu Nhanh](#bắt-đầu-nhanh)
12. [Cấu Hình](#cấu-hình)
13. [Đóng Góp](#đóng-góp)
14. [Lộ Trình Phát Triển](#lộ-trình-phát-triển)
15. [Giấy Phép](#giấy-phép)

---

## Tóm Tắt Tổng Quan

[2-3 đoạn mô tả dự án làm gì, phục vụ ai, và tại sao nó tồn tại. Tập trung vào không gian vấn đề và cách tiếp cận giải pháp.]

**Đối tượng sử dụng**: [Liệt kê các persona người dùng chính — ví dụ: Nhà phát triển, Phân tích viên, Đội vận hành, Người dùng cuối]

**Vấn đề cốt lõi**: [Dự án giải quyết vấn đề cụ thể gì?]

**Cách tiếp cận giải pháp**: [Giải quyết vấn đề khác với các giải pháp hiện có như thế nào?]

---

## Giá Trị Kinh Doanh & Mục Tiêu

| Vấn đề | Giải pháp |
|---|---|
| [Vấn đề 1 — ví dụ: Dữ liệu phân tán qua nhiều nguồn] | [Giải pháp 1 — ví dụ: Dashboard thống nhất với 50+ nguồn tích hợp] |
| [Vấn đề 2 — ví dụ: Chi phí cao của các công cụ thương mại] | [Giải pháp 2 — ví dụ: 100% mã nguồn mở, miễn phí sử dụng] |
| [Vấn đề 3 — ví dụ: Không có khả năng xử lý thời gian thực] | [Giải pháp 3 — ví dụ: Phân tích thời gian thực tại edge] |
| [Vấn đề 4 — ví dụ: Phụ thuộc cloud, lo ngại quyền riêng tư] | [Giải pháp 4 — ví dụ: Kiến trúc ưu tiên local với cloud tùy chọn] |

### Chỉ Số Đo Lường / KPIs

| Chỉ số | Mục tiêu | Hiện tại |
|---|---|---|
| [ví dụ: Thời gian phản hồi] | [< 200ms] | [150ms] |
| [ví dụ: Số nguồn dữ liệu tích hợp] | [50+] | [35] |
| [ví dụ: Uptime] | [99.9%] | [99.95%] |

---

## Tính Năng Chính

### [Danh mục tính năng 1 — ví dụ: Trực Quan Hóa Dữ Liệu]

- **[Tên tính năng]** — [Mô tả ngắn gọn chức năng và tầm quan trọng]
- **[Tên tính năng]** — [Mô tả ngắn gọn]

### [Danh mục tính năng 2 — ví dụ: Khả Năng AI/ML]

- **[Tên tính năng]** — [Mô tả ngắn gọn]
- **[Tên tính năng]** — [Mô tả ngắn gọn]

### [Danh mục tính năng 3 — ví dụ: Xử Lý Thời Gian Thực]

- **[Tên tính năng]** — [Mô tả ngắn gọn]
- **[Tên tính năng]** — [Mô tả ngắn gọn]

### [Danh mục tính năng 4 — ví dụ: Tích Hợp & Xuất Dữ Liệu]

- **[Tên tính năng]** — [Mô tả ngắn gọn]
- **[Tên tính năng]** — [Mô tả ngắn gọn]

---

## Kiến Trúc Hệ Thống

### Mô Hình Kiến Trúc

**Mô hình chính**: [ví dụ: Microservices / Monolithic / Serverless / Hybrid Edge-Serverless / Event-Driven]

| Khía cạnh | Mô hình | Lý do |
|---|---|---|
| **Cấu trúc tổng thể** | [ví dụ: Hybrid Edge-Serverless] | [ví dụ: Giảm thiểu phức tạp backend, giữ API key phía server] |
| **Thiết kế API** | [ví dụ: Contract-First (Proto/OpenAPI)] | [ví dụ: Loại bỏ lệch schema giữa frontend và backend] |
| **Xử lý dữ liệu** | [ví dụ: Client-Heavy SPA] | [ví dụ: Giảm phụ thuộc server; hỗ trợ hoạt động offline] |
| **Triển khai** | [ví dụ: Multi-Runtime (Web + Desktop + PWA)] | [ví dụ: Tiếp cận người dùng mọi nền tảng từ một codebase] |

### Sơ Đồ Kiến Trúc

```
┌──────────────────────────────────────────────┐
│              TẦNG CLIENT                      │
│  [Thành phần 1] [Thành phần 2] [Thành phần 3]│
│  [Workers / Tiến trình nền]                   │
└──────────────────┬───────────────────────────┘
                   │ [Giao thức: HTTPS / WSS / gRPC]
                   ▼
┌──────────────────────────────────────────────┐
│          TẦNG API / GATEWAY                   │
│  [API Gateway / Edge Functions / BFF]        │
│  [Middleware: Xác thực, CORS, Rate Limiting] │
└──────────┬──────────┬──────────┬─────────────┘
           │          │          │
           ▼          ▼          ▼
┌──────────────┐ ┌─────────┐ ┌──────────────────┐
│ [Kho Dữ Liệu]│ │ [Cache] │ │ [API Bên Ngoài] │
│ [Loại DB]    │ │ [Redis] │ │ [Danh sách]     │
└──────────────┘ └─────────┘ └──────────────────┘
```

> **Hướng dẫn**: Thay thế sơ đồ trên bằng kiến trúc thực tế. Sử dụng ASCII art cho tương thích Markdown. Với hệ thống phức tạp, cân nhắc liên kết đến công cụ vẽ bên ngoài (Mermaid, draw.io, Excalidraw).

---

## Công Nghệ Sử Dụng

### Ngôn Ngữ Lập Trình

| Ngôn ngữ | Vai trò | Tỷ lệ |
|---|---|---|
| [ví dụ: TypeScript] | [ví dụ: Frontend + Backend services] | [~80%] |
| [ví dụ: Rust] | [ví dụ: Desktop native shell, đường dẫn hiệu năng cao] | [~10%] |
| [ví dụ: Python] | [ví dụ: Pipeline dữ liệu, huấn luyện ML] | [~10%] |

### Framework & Thư Viện

| Danh mục | Công nghệ | Mục đích |
|---|---|---|
| **Frontend** | [ví dụ: React 18, Vue 3, Vanilla TS] | [ví dụ: UI dựa trên component] |
| **Backend** | [ví dụ: Node.js, FastAPI, Spring Boot] | [ví dụ: Dịch vụ API] |
| **Build** | [ví dụ: Vite, Webpack, esbuild] | [ví dụ: Dev server + bundling production] |
| **Testing** | [ví dụ: Playwright, Jest, Vitest] | [ví dụ: E2E + unit test] |
| **AI/ML** | [ví dụ: Transformers.js, TensorFlow, PyTorch] | [ví dụ: Suy luận phía browser / huấn luyện mô hình] |
| **Trực quan hóa** | [ví dụ: D3.js, deck.gl, Three.js] | [ví dụ: Biểu đồ, bản đồ, kết xuất 3D] |
| **Đa ngôn ngữ** | [ví dụ: i18next, vue-i18n] | [ví dụ: Hỗ trợ đa ngôn ngữ] |

### Kho Dữ Liệu

| Kho | Loại | Vai trò |
|---|---|---|
| [ví dụ: PostgreSQL] | Quan hệ | [ví dụ: Dữ liệu ứng dụng chính] |
| [ví dụ: Redis] | Key-Value | [ví dụ: Cache, trạng thái session, rate limiting] |
| [ví dụ: IndexedDB] | Phía browser | [ví dụ: Dữ liệu offline, cache phía client] |

### Hạ Tầng

| Thành phần | Nhà cung cấp | Vai trò |
|---|---|---|
| [ví dụ: Hosting] | [ví dụ: Vercel, AWS, GCP] | [ví dụ: SPA + API functions triển khai edge] |
| [ví dụ: CI/CD] | [ví dụ: GitHub Actions, GitLab CI] | [ví dụ: Testing + deployment tự động] |
| [ví dụ: CDN] | [ví dụ: Cloudflare, Vercel Edge] | [ví dụ: Phân phối asset toàn cầu] |
| [ví dụ: Giám sát] | [ví dụ: Sentry, Datadog] | [ví dụ: Theo dõi lỗi, giám sát hiệu năng] |

---

## Các Thành Phần Cốt Lõi

### [Thành phần 1 — ví dụ: Ứng Dụng Frontend]

| Module | Trách nhiệm |
|---|---|
| [ví dụ: App Core] | [ví dụ: Bộ điều phối trung tâm, khởi tạo service, quản lý state] |
| [ví dụ: UI Components (50+)] | [ví dụ: Panel tự chứa với tự fetch và render dữ liệu] |
| [ví dụ: Web Workers] | [ví dụ: Suy luận ML nền, tính toán nặng tách khỏi main thread] |

### [Thành phần 2 — ví dụ: Tầng API]

| Module | Trách nhiệm |
|---|---|
| [ví dụ: Edge Functions (60+)] | [ví dụ: API handler stateless cho proxy, cache, biến đổi dữ liệu] |
| [ví dụ: Middleware] | [ví dụ: Xác thực, CORS, phát hiện bot] |
| [ví dụ: Proto Gateway] | [ví dụ: Điểm vào duy nhất routing đến typed service handler] |

### [Thành phần 3 — ví dụ: Service / Logic Nghiệp Vụ]

| Miền | Modules | Trách nhiệm |
|---|---|---|
| [ví dụ: Intelligence] | [ví dụ: scoring, detection, analysis] | [ví dụ: Chấm điểm thời gian thực và phát hiện bất thường] |
| [ví dụ: Data Feeds] | [ví dụ: RSS, conflict, climate] | [ví dụ: Thu thập và chuẩn hóa dữ liệu bên ngoài] |
| [ví dụ: AI Pipeline] | [ví dụ: summarization, classification] | [ví dụ: Chuỗi LLM đa tầng với cơ chế fallback] |

---

## Luồng Dữ Liệu

### Luồng Dữ Liệu Chính

```
[Nguồn Dữ Liệu]
    │
    ▼
[Tầng Thu Thập — ví dụ: API Gateway / Edge Functions]
    │
    ├── [Tầng Cache — kiểm tra cache trước]
    │
    ▼
[Tầng Xử Lý — ví dụ: chuẩn hóa, làm giàu, chấm điểm]
    │
    ▼
[Tầng Lưu Trữ — ví dụ: database, ghi ngược cache]
    │
    ▼
[Tầng Trình Bày — ví dụ: UI components, phản hồi API]
```

### Chiến Lược Cache

| Tầng | Phạm vi | TTL | Mục đích |
|---|---|---|---|
| [Tầng 1 — ví dụ: In-Memory] | [Mỗi instance] | [60s–900s] | [Loại bỏ gọi remote lặp lại] |
| [Tầng 2 — ví dụ: Redis] | [Xuyên người dùng] | [120s–24h] | [Loại bỏ trùng lặp qua tất cả visitors] |
| [Tầng 3 — ví dụ: CDN] | [Edge toàn cầu] | [Tùy biến] | [Hấp thụ request lặp lại tại edge] |

### Pipeline Dữ Liệu Chính

| Pipeline | Đầu vào | Xử lý | Đầu ra |
|---|---|---|---|
| [ví dụ: AI Tóm Tắt] | [Tiêu đề thô] | [Loại trùng → Chuỗi LLM → Cache] | [Bản tóm tắt tổng hợp] |
| [ví dụ: Phát Hiện Bất Thường] | [Luồng sự kiện] | [Welford online stats → z-score] | [Cảnh báo bất thường] |
| [ví dụ: Phân Loại Mối Đe Dọa] | [Tin tức] | [Keyword (tức thì) + LLM (bất đồng bộ)] | [Nhãn mức độ + danh mục] |

---

## Mô Hình Bảo Mật

### Các Tầng Phòng Thủ

| Tầng | Cơ chế | Chi tiết |
|---|---|---|
| **Mạng** | [ví dụ: CORS allowlist, TLS, WAF] | [ví dụ: Chỉ origin được phép mới có thể gọi API] |
| **Xác thực** | [ví dụ: JWT, API key, OAuth2, session token] | [ví dụ: Token mỗi phiên cho IPC desktop; cô lập API key phía server] |
| **Phân quyền** | [ví dụ: RBAC, ABAC, scope-based] | [ví dụ: Truy cập dựa trên vai trò cho endpoint admin] |
| **Xác thực đầu vào** | [ví dụ: Schema validation, sanitization, regex] | [ví dụ: Chống XSS, chống SQL injection, ràng buộc trường proto] |
| **Giới hạn tốc độ** | [ví dụ: Redis IP limiting, per-user quotas] | [ví dụ: Ngăn lạm dụng endpoint AI] |
| **Quản lý bí mật** | [ví dụ: Env vars, OS keychain, vault] | [ví dụ: Credentials lưu trong OS keychain, không bao giờ plaintext] |
| **Bảo vệ Bot** | [ví dụ: Phát hiện UA, CAPTCHA, fingerprinting] | [ví dụ: Middleware chặn bot trên các route API] |

### Kiến Trúc Quyền Riêng Tư

| Cấp độ | Chế độ | Vị trí dữ liệu |
|---|---|---|
| [Cấp 1 — ví dụ: Full Cloud] | [ví dụ: Web app, xử lý server] | [Dữ liệu rời máy] |
| [Cấp 2 — ví dụ: Hybrid] | [ví dụ: Desktop + cloud APIs] | [Một phần tại local] |
| [Cấp 3 — ví dụ: Air-Gapped] | [ví dụ: Desktop + local AI] | [Không phụ thuộc cloud] |

---

## Khả Năng Mở Rộng & Hiệu Năng

### Tối Ưu Frontend

| Kỹ thuật | Mô tả |
|---|---|
| [ví dụ: Virtual Scrolling] | [Tái chế DOM cho danh sách lớn — chỉ render item hiển thị] |
| [ví dụ: Code Splitting] | [Lazy loading theo route giảm kích thước bundle ban đầu] |
| [ví dụ: Web Workers] | [Tính toán nặng chuyển khỏi main thread] |
| [ví dụ: Nhận biết Idle] | [Animation/polling tạm dừng khi tab ẩn hoặc user không hoạt động] |
| [ví dụ: Nén] | [Brotli/gzip nén trước cho asset tĩnh] |

### Mở Rộng Backend

| Kỹ thuật | Mô tả |
|---|---|
| [ví dụ: Stateless Functions] | [Mỗi function scale độc lập; không có shared mutable state] |
| [ví dụ: CDN Caching] | [Edge cache hấp thụ request lặp trước khi đến origin] |
| [ví dụ: Loại bỏ Request Trùng Lặp] | [Content-hash key đảm bảo N user đồng thời chỉ trigger 1 API call] |
| [ví dụ: Circuit Breakers] | [Breaker mỗi nguồn với cooldown ngăn lỗi lan truyền] |
| [ví dụ: Polling Xen Kẽ] | [Interval refresh khác nhau ngăn API storm đồng bộ] |

---

## Kiến Trúc Triển Khai

### Môi Trường

| Môi trường | URL | Mục đích |
|---|---|---|
| **Production** | [ví dụ: https://app.example.com] | [Triển khai phục vụ người dùng] |
| **Staging** | [ví dụ: https://staging.example.com] | [Xác thực trước production] |
| **Development** | [ví dụ: http://localhost:3000] | [Phát triển local] |

### Ma Trận Nền Tảng

| Nền tảng | Lệnh build | Phân phối |
|---|---|---|
| [ví dụ: Web (Vercel)] | `npm run build` | [Vercel Edge Network] |
| [ví dụ: Desktop (Tauri)] | `npm run desktop:build` | [GitHub Releases — .dmg, .exe, .AppImage] |
| [ví dụ: Mobile (Capacitor)] | `npm run mobile:build` | [App Store, Play Store] |
| [ví dụ: Docker] | `docker build .` | [Docker Hub / GHCR] |

---

## Bắt Đầu Nhanh

### Yêu Cầu Tiên Quyết

| Yêu cầu | Phiên bản | Ghi chú |
|---|---|---|
| [ví dụ: Node.js] | [≥ 18.0] | [Khuyến nghị LTS] |
| [ví dụ: npm / pnpm / yarn] | [≥ 9.0] | [Trình quản lý package] |
| [ví dụ: Rust] | [≥ 1.70] | [Chỉ cho bản desktop] |
| [ví dụ: Docker] | [≥ 24.0] | [Tùy chọn — cho triển khai container] |

### Khởi Động Nhanh

```bash
# Clone repository
git clone [https://github.com/org/project.git]
cd [project]

# Cài đặt dependencies
npm install

# Khởi chạy development server
npm run dev

# Mở trình duyệt
# http://localhost:[port]
```

### Biến Môi Trường

```bash
# Sao chép file môi trường mẫu
cp .env.example .env.local
```

| Nhóm | Biến | Bắt buộc | Free Tier |
|---|---|---|---|
| [ví dụ: Database] | `DATABASE_URL` | Có | [ví dụ: Supabase free tier] |
| [ví dụ: Auth] | `AUTH_SECRET`, `OAUTH_CLIENT_ID` | Có | N/A |
| [ví dụ: AI] | `OPENAI_API_KEY`, `LOCAL_LLM_URL` | Tùy chọn | [ví dụ: Ollama — miễn phí, local] |
| [ví dụ: Cache] | `REDIS_URL`, `REDIS_TOKEN` | Tùy chọn | [ví dụ: Upstash free tier] |
| [ví dụ: Giám sát] | `SENTRY_DSN` | Tùy chọn | [ví dụ: Sentry free tier] |

---

## Cấu Hình

### Biến Thể Build (nếu áp dụng)

| Biến thể | Lệnh | Mô tả |
|---|---|---|
| [ví dụ: Full] | `npm run build:full` | [Tất cả tính năng được bật] |
| [ví dụ: Lite] | `npm run build:lite` | [Bộ tính năng rút gọn cho triển khai nhẹ] |
| [ví dụ: Enterprise] | `npm run build:enterprise` | [Thêm tính năng enterprise] |

### Feature Toggles

| Toggle | Mặc định | Mô tả |
|---|---|---|
| [ví dụ: ENABLE_AI] | `true` | [Bật tính năng AI] |
| [ví dụ: ENABLE_ANALYTICS] | `true` | [Bật phân tích sử dụng] |
| [ví dụ: DEBUG_MODE] | `false` | [Logging chi tiết và UI debug] |

---

## Đóng Góp

Xem [CONTRIBUTING.md](./CONTRIBUTING.md) để biết hướng dẫn chi tiết.

```bash
# Quy trình phát triển
npm run dev          # Khởi chạy dev server
npm run test         # Chạy test
npm run lint         # Kiểm tra lint
npm run typecheck    # Kiểm tra kiểu TypeScript
npm run build        # Build production
```

---

## Lộ Trình Phát Triển

- [x] [Tính năng đã hoàn thành 1]
- [x] [Tính năng đã hoàn thành 2]
- [ ] [Tính năng dự kiến 1]
- [ ] [Tính năng dự kiến 2]
- [ ] [Tính năng dự kiến 3]

---

## Giấy Phép

[Loại giấy phép] — xem [LICENSE](./LICENSE) để biết chi tiết.

---

## Tác Giả

**[Tên Tác Giả]** — [GitHub](https://github.com/[username]) · [Website](https://[website])

---

<!-- ============================================================ -->
<!-- USAGE NOTES -->
<!-- ============================================================ -->

# 📋 Template Usage Notes / Hướng Dẫn Sử Dụng Template

## How to Use This Template / Cách Sử Dụng Template Này

### English

1. **Copy** the English or Vietnamese section (or both) into your project's `README.md` or `docs/PROJECT_INTRO.md`.
2. **Replace** all `[placeholder]` values with your project-specific information.
3. **Remove** sections that don't apply to your project (e.g., remove "Desktop Application" if web-only).
4. **Add** project-specific sections as needed (e.g., "Machine Learning Pipeline" for ML projects).
5. **Update** the architecture diagram to match your actual system topology.

### Tiếng Việt

1. **Sao chép** phần tiếng Anh hoặc tiếng Việt (hoặc cả hai) vào `README.md` hoặc `docs/PROJECT_INTRO.md` của dự án.
2. **Thay thế** tất cả giá trị `[placeholder]` bằng thông tin cụ thể của dự án.
3. **Xóa** các phần không áp dụng cho dự án (ví dụ: xóa "Ứng dụng Desktop" nếu chỉ có web).
4. **Thêm** các phần riêng cho dự án khi cần (ví dụ: "Pipeline Machine Learning" cho dự án ML).
5. **Cập nhật** sơ đồ kiến trúc cho phù hợp với topology hệ thống thực tế.

### Sections Checklist / Danh Sách Kiểm Tra

| Section / Phần | Required / Bắt buộc | Notes / Ghi chú |
|---|---|---|
| Executive Summary / Tóm tắt | ✅ Yes | Always include |
| Business Value / Giá trị kinh doanh | ✅ Yes | Critical for stakeholders |
| Key Features / Tính năng chính | ✅ Yes | Primary selling points |
| Architecture / Kiến trúc | ✅ Yes | Core technical documentation |
| Tech Stack / Công nghệ | ✅ Yes | Helps contributors onboard |
| Core Components / Thành phần cốt lõi | ✅ Yes | Navigational guide for codebase |
| Data Flow / Luồng dữ liệu | ⚡ Recommended | Essential for complex systems |
| Security / Bảo mật | ⚡ Recommended | Required for production systems |
| Scalability / Khả năng mở rộng | ⚡ Recommended | Important for high-traffic systems |
| Deployment / Triển khai | ⚡ Recommended | Multi-platform projects |
| Getting Started / Bắt đầu | ✅ Yes | Critical for adoption |
| Configuration / Cấu hình | ⚡ Recommended | When non-trivial setup required |
| Contributing / Đóng góp | ✅ Yes | For open-source projects |
| Roadmap / Lộ trình | ⚡ Recommended | Shows project health and direction |

---

*Template v1.0 — Derived from architectural analysis of World Monitor (v2.5.4), a complex multi-platform intelligence dashboard with 60+ API endpoints, 90+ service modules, and 50+ UI components.*
