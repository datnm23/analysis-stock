---
title: "Auto Blog từ Crawl-Agent"
description: "Crawl tin tức chứng khoán → kết hợp phân tích kỹ thuật/sentiment → AI rewrite thành bài viết → admin duyệt → blog site"
status: completed
priority: P2
effort: 11h
branch: main
tags: [blog, crawl-agent, claude-api, go-services, next-js, postgresql]
blockedBy: []
blocks: []
created: 2026-04-18
completed: 2026-04-18
---

# Auto Blog từ Crawl-Agent

## Overview

**Status: COMPLETED (2026-04-18)**

Tự động tạo bài viết phân tích chứng khoán từ:
1. Tin tức crawl được (đã có trong Redis `news:{symbol}:recent`)
2. Kết quả phân tích kỹ thuật/sentiment từ `go-services /api/v1/forecast/{symbol}`
3. Claude AI rewrite thành bài viết chuyên sâu (~500 từ, markdown)

**Pipeline:**
```
crawl-agent scheduler (daily 8AM)
  → lấy hot symbols từ Redis
  → call go-services /forecast/{symbol}
  → call Claude API (anthropic SDK)
  → POST /api/v1/articles (draft)
  → Telegram notify admin

Admin duyệt qua web-dashboard /admin/articles
  → PATCH /api/v1/articles/{id}/status → published

Blog site (Next.js mới)
  → GET /api/v1/articles?status=published
  → hiển thị bài viết với SSG/ISR
```

## Key Design Decisions

- **Không dùng n8n** (không có trong codebase) → dùng crawl-agent scheduler
- **Claude Haiku** cho article generation (cost-effective, đủ chất lượng)
- **Max 10 articles/ngày** để kiểm soát chi phí API
- **Hot symbols** = symbols có nhiều news nhất trong Redis (đếm list length)
- **GORM auto-migration** tự tạo table `articles` khi start go-services
- **ISR** (revalidate 3600s) cho blog site → SEO tốt + data fresh

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│  crawl-agent/app/services/article_generator.py (MỚI)        │
│  ┌──────────────────┐   ┌────────────────────────────────┐  │
│  │ get_hot_symbols()│──▶│ ArticleGenerator.generate()    │  │
│  │ (Redis LLEN)     │   │ 1. GET /forecast/{symbol}      │  │
│  └──────────────────┘   │ 2. get news from Redis         │  │
│                         │ 3. Claude API rewrite          │  │
│                         │ 4. POST /api/v1/articles       │  │
│                         └────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────┐
│  go-services                                                 │
│  ┌──────────────────────────┐  ┌───────────────────────┐    │
│  │ models/article.go (MỚI) │  │ handlers/articles.go  │    │
│  │ Article{GORM}            │  │ GET /api/v1/articles  │    │
│  └──────────────────────────┘  │ GET .../articles/:id  │    │
│                                │ POST .../articles     │    │
│                                │ PATCH .../articles/.. │    │
│                                │   /status             │    │
│                                └───────────────────────┘    │
│  PostgreSQL: articles table (auto-migration)                 │
└─────────────────────────────────────────────────────────────┘
                    │                         │
          ┌─────────▼──────┐        ┌────────▼────────┐
          │ web-dashboard  │        │  blog-site/     │
          │ /admin/articles│        │  (Next.js MỚI)  │
          │ approve/reject │        │  /articles      │
          └────────────────┘        │  /articles/{slug}│
                                    └─────────────────┘
```

## Stack Additions

| Component | Service | Files mới |
|-----------|---------|-----------|
| Article GORM model | go-services | `internal/models/article.go` |
| Articles handler | go-services | `internal/handlers/articles.go` |
| Articles service | go-services | `internal/services/article_service.go` |
| Article generator | crawl-agent | `app/services/article_generator.py` |
| Config thêm | crawl-agent | `app/config.py` (+4 settings) |
| Scheduler update | crawl-agent | `app/scheduler.py` |
| Admin page | web-dashboard | `app/admin/articles/page.tsx` |
| Blog site | mới | `blog-site/` (Next.js app) |

## Phases

| Phase | Title | Effort | Status |
|-------|-------|--------|--------|
| [01](./phase-01-db-model.md) | DB Model & Migration | 1h | **completed** |
| [02](./phase-02-go-api.md) | Go API Endpoints | 2h | **completed** |
| [03](./phase-03-article-generator.md) | Article Generator (crawl-agent) | 2.5h | **completed** |
| [04](./phase-04-admin-ui.md) | Admin UI (web-dashboard) | 2h | **completed** |
| [05](./phase-05-blog-site.md) | Blog Site (Next.js) | 3h | **completed** |
| [06](./phase-06-telegram-notify.md) | Telegram Notify Admin | 0.5h | **completed** |

## Dependencies

```
Phase 01 → Phase 02 → Phase 03
Phase 02 → Phase 04
Phase 02 → Phase 05
Phase 01-05 → Phase 06 (optional enhancement)
```

## Environment Variables Cần Thêm

```env
# crawl-agent
ANTHROPIC_API_KEY=sk-ant-...
ARTICLE_MAX_DAILY=10
ARTICLE_CLAUDE_MODEL=claude-haiku-4-5-20251001
ARTICLE_HOT_SYMBOLS_COUNT=10
GO_SERVICES_INTERNAL_URL=http://go-services:8080
```

## Success Criteria

- [ ] Sau mỗi scheduler run, có draft articles trong PostgreSQL
- [ ] Admin có thể approve/reject qua `/admin/articles`
- [ ] Blog site hiển thị published articles tại `/articles/{slug}`
- [ ] Max 10 articles/ngày được generate
- [ ] `go build ./...` không lỗi
- [ ] Blog site build pass (`next build`)
