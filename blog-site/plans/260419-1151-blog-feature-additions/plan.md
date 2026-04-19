---
title: Blog Feature Additions
status: pending
created: 2026-04-19
priority: high
blockedBy: []
blocks: []
---

# Blog Feature Additions Plan

**Mục tiêu:** Bổ sung 6 nhóm tính năng để tăng engagement, SEO, retention cho VietStock AI blog.  
**Nguồn:** Brainstorm + research competitive analysis (Vietstock, CafeF, Seeking Alpha, TradingView).  
**Stack:** Next.js 14 App Router + TypeScript + Tailwind (neo-brutalism) + Go API gateway.

## Phases

| Phase | Tính năng | Effort | Status |
|-------|-----------|--------|--------|
| [Phase 1](phase-01-related-articles-sidebar.md) | Related Articles + Quick Data Sidebar | ~1 ngày | pending |
| [Phase 2](phase-02-search-filter.md) | Search + Filter bài viết | ~1 ngày | pending |
| [Phase 3](phase-03-symbol-page.md) | Symbol Page `/symbols/[symbol]` | ~1 ngày | pending |
| [Phase 4](phase-04-newsletter-signup.md) | Newsletter Signup (email capture) | ~0.5 ngày | pending |
| [Phase 5](phase-05-trending-screener.md) | Trending Widget + Screener | ~1.5 ngày | pending |

## Key Dependencies

- Go API: cần thêm endpoints `/api/v1/articles?symbol=X` và `/api/v1/articles/related/:slug`
- Newsletter: cần email service (Resend free tier hoặc mailto đơn giản) — Phase 4 độc lập
- Screener (Phase 5): cần Go endpoint tổng hợp symbol metrics

## Execution Order

Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5

Phase 1 có thể làm song song với Phase 4 (independent).
