---
phase: 5
title: "Blog Site (Next.js mới)"
status: completed
effort: 3h
completed: 2026-04-18
---

# Phase 5: Blog Site (Next.js riêng biệt)

## Context Links
- Phase trước: [phase-02-go-api.md](./phase-02-go-api.md)
- Go API: `GET /api/v1/articles?status=published`, `GET /api/v1/articles/:slug`
- Tham khảo structure: `web-dashboard/` (Next.js 14 App Router)

## Overview

Tạo Next.js app mới tại `blog-site/` với:
- `/` — trang chủ: latest 6 articles
- `/articles` — list tất cả published articles (pagination)
- `/articles/[slug]` — bài viết chi tiết, SSG + ISR

**Không cần auth, không cần admin** — đây là public blog.

## Related Code Files

**Tạo mới:**
- `blog-site/` — toàn bộ Next.js app

## Implementation Steps

### 1. Khởi tạo Next.js app

```bash
cd /media/datnm/Data/Java/analysis-stock
npx create-next-app@14 blog-site \
  --typescript \
  --tailwind \
  --app \
  --no-src-dir \
  --import-alias "@/*"
```

Sau đó xoá boilerplate (page.tsx default content, globals.css default styles).

### 2. Cấu trúc files cần tạo

```
blog-site/
├── app/
│   ├── layout.tsx          # Root layout, nav header
│   ├── page.tsx            # Home: latest 6 articles
│   ├── articles/
│   │   ├── page.tsx        # Article list with pagination
│   │   └── [slug]/
│   │       └── page.tsx    # Article detail (SSG + ISR)
├── lib/
│   └── articles-api.ts     # API fetch functions
├── components/
│   ├── article-card.tsx    # Card component for list
│   └── article-content.tsx # Markdown renderer
├── .env.local.example
└── next.config.js
```

### 3. `blog-site/lib/articles-api.ts`

```typescript
const API_URL = process.env.API_URL || "http://localhost:8080";

export interface Article {
  id: number;
  symbol: string;
  title: string;
  slug: string;
  summary: string;
  content: string;
  published_at: string;
  created_at: string;
}

export async function getArticles(limit = 20, offset = 0): Promise<{ articles: Article[]; total: number }> {
  const res = await fetch(
    `${API_URL}/api/v1/articles?status=published&limit=${limit}&offset=${offset}`,
    { next: { revalidate: 3600 } }
  );
  if (!res.ok) return { articles: [], total: 0 };
  return res.json();
}

export async function getArticleBySlug(slug: string): Promise<Article | null> {
  const res = await fetch(`${API_URL}/api/v1/articles/${slug}`, {
    next: { revalidate: 3600 },
  });
  if (!res.ok) return null;
  return res.json();
}

export async function getLatestArticles(limit = 6): Promise<Article[]> {
  const { articles } = await getArticles(limit, 0);
  return articles;
}
```

### 4. `blog-site/app/layout.tsx`

```tsx
import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Phân tích chứng khoán Việt Nam",
  description: "Bài viết phân tích chứng khoán HSX/HNX/UPCOM tự động từ AI",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="vi">
      <body className="min-h-screen bg-gray-50">
        <header className="bg-white border-b px-6 py-4">
          <nav className="max-w-4xl mx-auto flex gap-6 items-center">
            <a href="/" className="text-lg font-bold text-blue-700">StockBlog</a>
            <a href="/articles" className="text-sm text-gray-600 hover:text-blue-600">Tất cả bài viết</a>
          </nav>
        </header>
        <main className="max-w-4xl mx-auto px-6 py-8">{children}</main>
        <footer className="border-t mt-12 py-6 text-center text-xs text-gray-400">
          Nội dung chỉ mang tính tham khảo, không phải lời khuyên đầu tư.
        </footer>
      </body>
    </html>
  );
}
```

### 5. `blog-site/components/article-card.tsx`

```tsx
import Link from "next/link";
import { Article } from "@/lib/articles-api";

export function ArticleCard({ article }: { article: Article }) {
  const date = article.published_at
    ? new Date(article.published_at).toLocaleDateString("vi-VN")
    : "";
  return (
    <Link href={`/articles/${article.slug}`} className="block group">
      <div className="bg-white rounded-lg border p-5 hover:shadow-md transition-shadow">
        <span className="text-xs font-bold text-blue-600 bg-blue-50 px-2 py-0.5 rounded">
          {article.symbol}
        </span>
        <h2 className="text-base font-semibold mt-2 group-hover:text-blue-700 line-clamp-2">
          {article.title}
        </h2>
        <p className="text-sm text-gray-500 mt-1 line-clamp-2">{article.summary}</p>
        <p className="text-xs text-gray-400 mt-2">{date}</p>
      </div>
    </Link>
  );
}
```

### 6. `blog-site/app/page.tsx` (Home)

```tsx
import { getLatestArticles } from "@/lib/articles-api";
import { ArticleCard } from "@/components/article-card";
import Link from "next/link";

export const revalidate = 3600;

export default async function HomePage() {
  const articles = await getLatestArticles(6);
  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Phân tích chứng khoán mới nhất</h1>
      {articles.length === 0 ? (
        <p className="text-gray-400">Chưa có bài viết nào.</p>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {articles.map((a) => <ArticleCard key={a.id} article={a} />)}
        </div>
      )}
      <div className="mt-8 text-center">
        <Link href="/articles" className="text-blue-600 hover:underline text-sm">
          Xem tất cả bài viết →
        </Link>
      </div>
    </div>
  );
}
```

### 7. `blog-site/app/articles/page.tsx` (List)

```tsx
import { getArticles } from "@/lib/articles-api";
import { ArticleCard } from "@/components/article-card";

export const revalidate = 3600;

export default async function ArticlesPage({
  searchParams,
}: {
  searchParams: { page?: string };
}) {
  const page = parseInt(searchParams.page ?? "1");
  const limit = 12;
  const offset = (page - 1) * limit;
  const { articles, total } = await getArticles(limit, offset);
  const totalPages = Math.ceil(total / limit);

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Tất cả bài viết ({total})</h1>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {articles.map((a) => <ArticleCard key={a.id} article={a} />)}
      </div>
      {totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-8">
          {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
            <a
              key={p}
              href={`/articles?page=${p}`}
              className={`px-3 py-1 rounded border text-sm ${p === page ? "bg-blue-600 text-white border-blue-600" : "hover:bg-gray-50"}`}
            >
              {p}
            </a>
          ))}
        </div>
      )}
    </div>
  );
}
```

### 8. `blog-site/app/articles/[slug]/page.tsx` (Detail — SSG + ISR)

```tsx
import { getArticleBySlug, getArticles } from "@/lib/articles-api";
import { notFound } from "next/navigation";

export const revalidate = 3600;

export async function generateStaticParams() {
  const { articles } = await getArticles(100, 0);
  return articles.map((a) => ({ slug: a.slug }));
}

export default async function ArticleDetailPage({ params }: { params: { slug: string } }) {
  const article = await getArticleBySlug(params.slug);
  if (!article) notFound();

  const date = article.published_at
    ? new Date(article.published_at).toLocaleDateString("vi-VN", {
        year: "numeric", month: "long", day: "numeric",
      })
    : "";

  return (
    <article className="max-w-2xl">
      <div className="mb-4">
        <span className="text-sm font-bold text-blue-600 bg-blue-50 px-2 py-0.5 rounded">
          {article.symbol}
        </span>
        <h1 className="text-2xl font-bold mt-3">{article.title}</h1>
        <p className="text-sm text-gray-400 mt-1">{date}</p>
      </div>
      {/* Render markdown as plain text — add react-markdown if needed */}
      <div className="prose prose-sm max-w-none mt-6 whitespace-pre-wrap text-gray-800">
        {article.content}
      </div>
    </article>
  );
}
```

> **Optional:** Cài `react-markdown` để render markdown đẹp hơn: `npm install react-markdown`

### 9. `blog-site/.env.local.example`

```env
API_URL=http://localhost:8080
```

### 10. Build check

```bash
cd blog-site && npm install && npm run build
```

## Todo List

- [ ] `npx create-next-app@14 blog-site` với TypeScript + Tailwind + App Router
- [ ] Tạo `lib/articles-api.ts`
- [ ] Tạo `components/article-card.tsx`
- [ ] Tạo `app/layout.tsx`
- [ ] Tạo `app/page.tsx` (home)
- [ ] Tạo `app/articles/page.tsx` (list)
- [ ] Tạo `app/articles/[slug]/page.tsx` (detail SSG)
- [ ] Thêm `.env.local.example`
- [ ] Chạy `npm run build` — pass

## Success Criteria

- `npm run build` pass, 0 TypeScript errors
- `/` hiển thị 6 bài mới nhất
- `/articles` phân trang đúng
- `/articles/{slug}` hiển thị nội dung, 404 nếu không tìm thấy
- ISR revalidate 3600s: data refresh mỗi giờ

## Docker Integration (sau khi build pass)

Thêm vào `docker-compose.yml`:
```yaml
blog-site:
  build: ./blog-site
  ports:
    - "3001:3000"
  environment:
    - API_URL=http://go-services:8080
  depends_on:
    - go-services
```
