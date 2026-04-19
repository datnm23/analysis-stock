# Phase 3: Symbol Page `/symbols/[symbol]`

**Priority:** High  
**Effort:** ~1 ngày  
**Status:** pending  
**Depends on:** Phase 1 (getRelatedArticles pattern), Phase 2 (article filtering)

## Overview

Trang hub cho từng mã cổ phiếu — tập trung toàn bộ phân tích AI của 1 symbol.  
URL: `/symbols/VNM`, `/symbols/HPG`, v.v.

**Value:** SEO goldmine — "phân tích VNM", "VNM cổ phiếu" low-competition keywords.

## Related Code Files

**Tạo mới:**
- `blog-site/app/symbols/[symbol]/page.tsx` — symbol hub page
- `blog-site/lib/articles-api.ts` — thêm `getArticlesBySymbol(symbol, limit, offset)`

**Tham khảo:**
- `blog-site/app/articles/[slug]/page.tsx` — pattern StockChart + layout
- `blog-site/components/stock-chart.tsx` — reuse trực tiếp
- `blog-site/components/article-card.tsx` — reuse cho article list

## Implementation Steps

### API

1. Thêm `getArticlesBySymbol()` vào `articles-api.ts`:
   ```ts
   export async function getArticlesBySymbol(
     symbol: string, limit = 20, offset = 0
   ): Promise<ArticleListResult>
   // Calls: GET /api/v1/articles?symbol=VNM&status=published&limit=...
   ```

2. Go handler cập nhật `/api/v1/articles` hỗ trợ `?symbol=VNM` filter (exact match).

### Page Structure

3. Tạo `blog-site/app/symbols/[symbol]/page.tsx`:

```tsx
// generateStaticParams: lấy distinct symbols từ articles
// revalidate: 60

export default async function SymbolPage({ params }: { params: { symbol: string } }) {
  const symbol = params.symbol.toUpperCase()
  const { articles, total } = await getArticlesBySymbol(symbol, 20, 0)
  
  // notFound nếu total === 0
  
  // Lấy bài mới nhất để hiển thị recommendation hiện tại
  const latest = articles[0]
  const forecast = parseForecast(latest?.forecast_data)
}
```

4. Layout page:
```
┌──────────────────────────────────────────────┐
│  ← Tất cả mã          [Symbol badge large]   │
│  VNM — VINAMILK                              │
│  [MUA badge] [Confidence: 87%]               │
├──────────────────────────────────────────────┤
│  [StockChart symbol={symbol} days={180}]      │  ← 6 tháng
├──────────────────────────────────────────────┤
│  Lịch sử phân tích AI ({total} bài)          │
│  [ArticleCard][ArticleCard][ArticleCard]      │
│  [ArticleCard][ArticleCard][ArticleCard]      │
│  [Pagination nếu > 12 bài]                   │
└──────────────────────────────────────────────┘
```

5. SEO metadata động:
```tsx
export async function generateMetadata({ params }) {
  return {
    title: `Phân tích ${params.symbol.toUpperCase()} — VietStock AI`,
    description: `Tổng hợp ${total} bài phân tích kỹ thuật và AI cho mã ${symbol}`
  }
}
```

6. Thêm link từ `article-card.tsx` — symbol badge click → `/symbols/VNM`:
```tsx
<Link href={`/symbols/${article.symbol}`} className="badge-symbol">
  {article.symbol}
</Link>
```

7. Thêm `/symbols` vào navbar (optional — trang index liệt kê tất cả symbols):
```tsx
// /symbols/page.tsx: Grid tất cả symbols đã có bài, sort by article count
```

## generateStaticParams

```ts
export async function generateStaticParams() {
  // Fetch distinct symbols từ articles API
  // Fallback: return [] nếu lỗi (dynamicParams = true)
}
export const dynamicParams = true
```

## Success Criteria

- [ ] `/symbols/VNM` hiển thị chart 6 tháng + tất cả bài về VNM
- [ ] Symbol badge trên article card link đến symbol page
- [ ] SEO metadata đúng cho từng symbol
- [ ] `notFound()` cho symbol không có bài
- [ ] Pagination nếu > 12 bài
