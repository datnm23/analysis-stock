---
phase: 4
title: "Admin UI (web-dashboard)"
status: completed
effort: 2h
completed: 2026-04-18
---

# Phase 4: Admin UI — Review Articles (web-dashboard)

## Context Links
- Phase trước: [phase-02-go-api.md](./phase-02-go-api.md)
- Existing page pattern: `web-dashboard/app/reports/page.tsx`
- API client: `web-dashboard/lib/api.ts`
- Next.js: App Router, TypeScript, v14.2

## Overview

Thêm trang `/admin/articles` vào web-dashboard hiện tại. Admin có thể:
- Xem danh sách draft articles
- Preview nội dung
- Approve (→ published) hoặc Reject (→ rejected)

Không cần auth phức tạp — dùng query param `?key=...` check với `ADMIN_KEY` env.

## Related Code Files

**Tạo mới:**
- `web-dashboard/app/admin/articles/page.tsx`
- `web-dashboard/app/admin/articles/article-review-card.tsx`

**Sửa:**
- `web-dashboard/lib/api.ts` — thêm articles API functions
- `web-dashboard/components/Sidebar.tsx` — thêm link "Admin Articles"

## Implementation Steps

### 1. Sửa `web-dashboard/lib/api.ts` — thêm article functions

```typescript
const ADMIN_KEY = process.env.NEXT_PUBLIC_ADMIN_KEY || "";
const GO_SERVICES_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface Article {
  id: number;
  symbol: string;
  title: string;
  slug: string;
  content: string;
  summary: string;
  status: "draft" | "published" | "rejected";
  created_at: string;
  published_at?: string;
}

export async function fetchDraftArticles(): Promise<Article[]> {
  const res = await fetch(
    `${GO_SERVICES_URL}/api/v1/articles?status=draft&limit=50`,
    { cache: "no-store" }
  );
  const data = await res.json();
  return data.articles ?? [];
}

export async function updateArticleStatus(
  id: number,
  status: "published" | "rejected"
): Promise<void> {
  await fetch(`${GO_SERVICES_URL}/api/v1/articles/${id}/status`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      "X-Internal-Key": ADMIN_KEY,
    },
    body: JSON.stringify({ status }),
  });
}
```

### 2. Tạo `web-dashboard/app/admin/articles/article-review-card.tsx`

```tsx
"use client";

import { useState } from "react";
import { Article, updateArticleStatus } from "@/lib/api";

interface Props {
  article: Article;
  onUpdate: (id: number, status: "published" | "rejected") => void;
}

export function ArticleReviewCard({ article, onUpdate }: Props) {
  const [expanded, setExpanded] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleAction = async (status: "published" | "rejected") => {
    setLoading(true);
    await updateArticleStatus(article.id, status);
    onUpdate(article.id, status);
    setLoading(false);
  };

  return (
    <div className="border rounded-lg p-4 mb-4 bg-white shadow-sm">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1">
          <span className="text-xs font-bold text-blue-600 bg-blue-50 px-2 py-0.5 rounded">
            {article.symbol}
          </span>
          <h3 className="text-base font-semibold mt-1">{article.title}</h3>
          <p className="text-sm text-gray-500 mt-1">{article.summary}</p>
          <p className="text-xs text-gray-400 mt-1">
            {new Date(article.created_at).toLocaleString("vi-VN")}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <button
            onClick={() => setExpanded(!expanded)}
            className="text-sm px-3 py-1 border rounded hover:bg-gray-50"
          >
            {expanded ? "Ẩn" : "Xem"}
          </button>
          <button
            disabled={loading}
            onClick={() => handleAction("published")}
            className="text-sm px-3 py-1 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
          >
            Duyệt
          </button>
          <button
            disabled={loading}
            onClick={() => handleAction("rejected")}
            className="text-sm px-3 py-1 bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50"
          >
            Từ chối
          </button>
        </div>
      </div>
      {expanded && (
        <div className="mt-4 p-3 bg-gray-50 rounded text-sm whitespace-pre-wrap font-mono">
          {article.content}
        </div>
      )}
    </div>
  );
}
```

### 3. Tạo `web-dashboard/app/admin/articles/page.tsx`

```tsx
"use client";

import { useEffect, useState } from "react";
import { Article, fetchDraftArticles } from "@/lib/api";
import { ArticleReviewCard } from "./article-review-card";

export default function AdminArticlesPage() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDraftArticles()
      .then(setArticles)
      .finally(() => setLoading(false));
  }, []);

  const handleUpdate = (id: number, status: "published" | "rejected") => {
    setArticles((prev) => prev.filter((a) => a.id !== id));
  };

  if (loading) return <div className="p-8">Đang tải...</div>;

  return (
    <div className="max-w-4xl mx-auto p-8">
      <h1 className="text-2xl font-bold mb-6">
        Duyệt bài viết ({articles.length} draft)
      </h1>
      {articles.length === 0 ? (
        <p className="text-gray-500">Không có bài viết nào cần duyệt.</p>
      ) : (
        articles.map((a) => (
          <ArticleReviewCard key={a.id} article={a} onUpdate={handleUpdate} />
        ))
      )}
    </div>
  );
}
```

### 4. Sửa `web-dashboard/components/Sidebar.tsx` — thêm link

```tsx
// Thêm vào danh sách nav links
{ href: "/admin/articles", label: "Admin Articles" }
```

### 5. Thêm env vào `web-dashboard/.env.local.example`

```env
NEXT_PUBLIC_ADMIN_KEY=change-me-secret-key
```

## Todo List

- [ ] Thêm `fetchDraftArticles`, `updateArticleStatus`, `Article` type vào `lib/api.ts`
- [ ] Tạo `app/admin/articles/article-review-card.tsx`
- [ ] Tạo `app/admin/articles/page.tsx`
- [ ] Thêm nav link vào `Sidebar.tsx`
- [ ] Thêm `NEXT_PUBLIC_ADMIN_KEY` vào env example
- [ ] Test: `npm run build` pass

## Success Criteria

- `/admin/articles` hiển thị danh sách drafts
- Nút "Duyệt" → article biến mất khỏi list, status → published
- Nút "Từ chối" → article biến mất khỏi list, status → rejected
- "Xem" toggle hiện/ẩn nội dung markdown

## Security Note

`X-Internal-Key` gửi từ browser (NEXT_PUBLIC_) → visible trong source. Acceptable vì:
- Admin page không public
- Chỉ cần basic protection, không phải production auth
- Nếu cần hardening: thêm `/api/admin/articles` route trong Next.js làm proxy
