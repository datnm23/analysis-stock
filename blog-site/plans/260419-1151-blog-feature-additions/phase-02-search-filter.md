# Phase 2: Search + Filter

**Priority:** High  
**Effort:** ~1 ngày  
**Status:** pending  
**Depends on:** Phase 1 (Go articles handler pattern)

## Overview

Client-side search + filter trên trang `/articles`:
- Tìm theo tên symbol hoặc tiêu đề
- Filter theo recommendation (BUY / SELL / HOLD)
- URL params để shareable (`?q=VNM&rec=BUY`)

**Approach:** Server-side filtering qua query params (`/api/v1/articles?q=VNM&recommendation=BUY`) — tốt cho SEO hơn client-side filter.

## Related Code Files

**Sửa:**
- `blog-site/app/articles/page.tsx` — thêm search bar + filter UI + đọc searchParams
- `blog-site/lib/articles-api.ts` — thêm params `q`, `recommendation` vào `getArticles()`
- `go-services/cmd/api-gateway/main.go` — pass query params xuống DB (nếu Go xử lý)

**Không cần tạo component mới** — dùng inline UI trong `articles/page.tsx`.

## Implementation Steps

### Backend

1. Cập nhật `/api/v1/articles` endpoint hỗ trợ:
   ```
   GET /api/v1/articles?status=published&q=VNM&recommendation=BUY&limit=12&offset=0
   ```
   - `q` → `WHERE symbol ILIKE '%VNM%' OR title ILIKE '%VNM%'`
   - `recommendation` → join/filter qua `forecast_data->>'recommendation'`

### Frontend

2. Cập nhật `articles-api.ts`:
   ```ts
   export async function getArticles(
     limit = 20, offset = 0,
     filters?: { q?: string; recommendation?: string }
   ): Promise<ArticleListResult>
   ```

3. Cập nhật `app/articles/page.tsx`:
   ```tsx
   // searchParams thêm: q, rec
   export default async function ArticlesPage({
     searchParams,
   }: { searchParams: { page?: string; q?: string; rec?: string } })
   ```

4. Search bar UI (thêm vào trước article grid):
   ```tsx
   // Form với method GET, action="/articles"
   // Input: search symbol/title (placeholder "Tìm mã hoặc tiêu đề...")
   // Filter pills: TẤT CẢ | MUA | GIỮ | BÁN
   // Style: brutalist input (border-3 border-ink, no border-radius)
   ```

5. Hiển thị trạng thái filter:
   ```tsx
   // Nếu có filter active: "Kết quả cho 'VNM' — 5 bài"
   // Empty state riêng: "Không tìm thấy bài nào cho 'XYZ'"
   ```

6. Pagination giữ filter params:
   ```tsx
   // href={`/articles?page=${p}&q=${q}&rec=${rec}`}
   ```

## UI Mockup

```
┌─────────────────────────────────────────────────────────┐
│  [input: Tìm mã hoặc tiêu đề...          ] [Tìm →]     │
│  [TẤT CẢ] [■ MUA] [BÁN] [GIỮ]                          │
└─────────────────────────────────────────────────────────┘
Kết quả cho "VNM" — 3 bài
[card][card][card]
```

## Success Criteria

- [ ] Search theo symbol (VNM) trả đúng bài
- [ ] Filter BUY/SELL/HOLD hoạt động
- [ ] URL shareable: `/articles?q=VNM&rec=BUY`
- [ ] Empty state hiển thị khi không có kết quả
- [ ] Pagination giữ filter state
