---
title: Phase 03 – Nav link layout.tsx
status: pending
file: blog-site/app/layout.tsx
---

# Phase 03 – Thêm nav link "Thị trường" vào layout

## Change

Thêm link `/market` vào nav bar, cùng style với link "Screener" hiện có.

## Implementation

Trong `blog-site/app/layout.tsx`, sau link Screener, thêm:

```tsx
<a
  href="/market"
  className="text-white/70 hover:text-yellow font-semibold text-sm px-3 py-1.5 uppercase tracking-wide transition-colors duration-100"
>
  Thị trường
</a>
```

Đặt giữa "Bài viết" và "Screener" (thứ tự logic: Trang chủ → Bài viết → Thị trường → Screener).

## Final Nav Order

```
Trang chủ | Bài viết | Thị trường | Screener
```

## Success Criteria

- Link "Thị trường" hiển thị trên header mọi trang
- Click → navigate đến `/market`
- Style nhất quán với các link khác
