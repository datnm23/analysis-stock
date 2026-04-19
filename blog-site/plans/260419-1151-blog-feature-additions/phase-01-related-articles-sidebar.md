# Phase 1: Related Articles + Quick Data Sidebar

**Priority:** Critical  
**Effort:** ~1 ngày  
**Status:** pending

## Overview

Hai widget trên article detail page (`/articles/[slug]`):
1. **Related Articles** — 3 bài cùng symbol (fallback: cùng sector) ở cuối trang
2. **Quick Data Sidebar** — sticky widget bên phải: giá, RSI, MACD, recommendation, confidence

## Related Code Files

**Đọc để hiểu context:**
- `blog-site/app/articles/[slug]/page.tsx` — article detail page
- `blog-site/lib/articles-api.ts` — Article type + API layer
- `blog-site/components/article-card.tsx` — card component để reuse
- `go-services/cmd/api-gateway/main.go` — route registration
- `go-services/internal/handlers/chart.go` — pattern tham khảo cho Go handler

**Tạo mới:**
- `blog-site/components/related-articles.tsx`
- `blog-site/components/stock-data-widget.tsx`
- `go-services/internal/handlers/articles.go` — thêm `/api/v1/articles/:slug/related`

**Sửa:**
- `blog-site/lib/articles-api.ts` — thêm `getRelatedArticles(slug, symbol)`
- `blog-site/app/articles/[slug]/page.tsx` — layout 2-cột trên desktop
- `go-services/cmd/api-gateway/main.go` — register route mới

## Implementation Steps

### Backend — Go (articles handler)

1. Tạo `go-services/internal/handlers/articles.go`:
   ```go
   // GET /api/v1/articles/:slug/related?limit=3
   // Query: SELECT * FROM articles WHERE symbol = $1 AND slug != $2 
   //        AND status = 'published' ORDER BY published_at DESC LIMIT $3
   func RelatedArticles() gin.HandlerFunc { ... }
   ```

2. Register trong `main.go`:
   ```go
   v1.GET("/articles/:slug/related", handlers.RelatedArticles())
   ```

3. Response format:
   ```json
   { "articles": [{ "id", "slug", "title", "symbol", "image_url", "published_at", "forecast_data" }] }
   ```

> **Lưu ý:** Nếu Go service chưa có DB access cho articles (chỉ có vnstock client), kiểm tra xem articles được lưu ở đâu. Nếu là PostgreSQL qua crawl-agent, cần inject DB connection vào Go handler. Fallback: call Python crawl-agent API.

### Frontend — Related Articles component

4. Tạo `blog-site/components/related-articles.tsx`:
   ```tsx
   // Server component, gọi getRelatedArticles(slug, symbol) 
   // Render 3 ArticleCard compact (no chart) trong 1 row grid
   // Header: "Phân tích liên quan" với brutalist style
   ```

5. `blog-site/lib/articles-api.ts` — thêm:
   ```ts
   export async function getRelatedArticles(slug: string, symbol: string): Promise<Article[]>
   ```

6. Insert vào `app/articles/[slug]/page.tsx` sau phần Source links:
   ```tsx
   <RelatedArticles slug={article.slug} symbol={article.symbol} />
   ```

### Frontend — Quick Data Widget

7. Tạo `blog-site/components/stock-data-widget.tsx`:
   ```tsx
   // Client component — fetch /api/v1/chart/:symbol?days=5 để lấy giá gần nhất
   // Hiển thị:
   //   Symbol badge (lớn)
   //   Giá đóng cửa gần nhất
   //   Thay đổi % hôm nay (▲/▼)
   //   Recommendation badge (MUA/BÁN/GIỮ)
   //   Confidence score (progress bar)
   //   RSI (nếu có trong forecast_data)
   // Style: neo-brutalism card, sticky top-24 trên desktop
   ```

8. Layout article detail — chuyển từ `max-w-3xl mx-auto` sang 2-cột trên `lg`:
   ```tsx
   <div className="grid grid-cols-1 lg:grid-cols-[1fr_280px] gap-8 max-w-5xl mx-auto">
     <div>{/* main content */}</div>
     <aside className="hidden lg:block">
       <div className="sticky top-24">
         <StockDataWidget symbol={article.symbol} forecast={forecast} />
       </div>
     </aside>
   </div>
   ```

## Success Criteria

- [ ] Related articles hiển thị 3 bài cùng symbol (hoặc empty state nếu không có)
- [ ] Quick data widget sticky bên phải trên desktop, ẩn trên mobile
- [ ] Widget load không block main content (Suspense / loading skeleton)
- [ ] Build `npm run build` pass 0 errors

## Risk Assessment

| Rủi ro | Xác suất | Mitigation |
|--------|----------|------------|
| Go service chưa có DB access cho articles | Cao | Kiểm tra architecture trước; fallback: fetch từ crawl-agent Python API |
| forecast_data không có RSI field | Trung bình | Chỉ hiển thị recommendation + confidence nếu không có RSI |
| Layout 2-cột phá vỡ mobile | Thấp | `hidden lg:block` cho sidebar — safe |
