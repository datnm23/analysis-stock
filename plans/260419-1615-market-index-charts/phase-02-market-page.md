---
title: Phase 02 – /market page
status: pending
file: blog-site/app/market/page.tsx (NEW)
---

# Phase 02 – Trang /market

## Layout

```
┌─────────────────────────────────────────┐
│  [VNINDEX] [VN30] [HNXINDEX] [UPCOM]   │  ← tab bar (client, useSearchParams)
├─────────────────────────────────────────┤
│  StockChart (candlestick, 90 ngày)      │  ← reuse component hiện có
├──────────────┬──────────────────────────┤
│  Top MUA (3) │  Top BÁN (3)            │  ← từ screener API
├──────────────┴──────────────────────────┤
│  ScreenerTable (tất cả symbols)         │  ← reuse component hiện có
└─────────────────────────────────────────┘
```

## Architecture

- `page.tsx` = Server Component: fetch top recs (BUY/SELL) + screener data server-side
- `MarketTabSwitcher` = Client Component riêng (cần `useSearchParams`), wrap trong `<Suspense>`
- Truyền `activeTab` qua URL param `?tab=VNINDEX` (default VNINDEX)

## Files

- **Tạo mới**: `blog-site/app/market/page.tsx`
- **Đọc tham khảo**: `blog-site/components/stock-chart.tsx`, `blog-site/components/screener-table.tsx` (nếu tồn tại), `blog-site/lib/articles-api.ts`

## Implementation Steps

1. Kiểm tra `screener-table.tsx` tồn tại chưa — nếu chưa thì chỉ render list đơn giản

2. Tạo `blog-site/app/market/page.tsx`:

```tsx
import { Suspense } from "react";
import { StockChart } from "@/components/stock-chart";
import { MarketTabSwitcher } from "./market-tab-switcher"; // client component

const TABS = [
  { label: "VN-Index", symbol: "VNINDEX" },
  { label: "VN30",     symbol: "VN30" },
  { label: "HNX",      symbol: "HNXINDEX" },
  { label: "UPCOM",    symbol: "UPCOMINDEX" },
];

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

async function fetchTopRec(rec: "BUY" | "SELL") {
  const res = await fetch(`${API}/api/v1/screener?recommendation=${rec}&limit=3`, {
    next: { revalidate: 300 }
  });
  if (!res.ok) return [];
  const data = await res.json();
  return data.items ?? data ?? [];
}

export default async function MarketPage({
  searchParams,
}: {
  searchParams: { tab?: string };
}) {
  const activeSymbol = TABS.find(t => t.symbol === searchParams.tab)?.symbol ?? "VNINDEX";
  const [buyTop, sellTop] = await Promise.all([fetchTopRec("BUY"), fetchTopRec("SELL")]);

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-black uppercase tracking-tight text-ink border-b-3 border-ink pb-3">
        Thị trường
      </h1>

      {/* Tab switcher (client, needs Suspense) */}
      <Suspense fallback={<div className="h-10 bg-ink/10 animate-pulse rounded" />}>
        <MarketTabSwitcher tabs={TABS} activeSymbol={activeSymbol} />
      </Suspense>

      {/* Chart */}
      <StockChart symbol={activeSymbol} days={90} />

      {/* Top recommendations */}
      <div className="grid grid-cols-2 gap-6">
        <TopRecList title="Top MUA" items={buyTop} badgeClass="badge-buy" />
        <TopRecList title="Top BÁN" items={sellTop} badgeClass="badge-sell" />
      </div>
    </div>
  );
}
```

3. Tạo `blog-site/app/market/market-tab-switcher.tsx` (client component):

```tsx
"use client";
import { useRouter, useSearchParams } from "next/navigation";

export function MarketTabSwitcher({ tabs, activeSymbol }) {
  const router = useRouter();
  return (
    <div className="flex gap-2 flex-wrap">
      {tabs.map(tab => (
        <button
          key={tab.symbol}
          onClick={() => router.push(`/market?tab=${tab.symbol}`)}
          className={tab.symbol === activeSymbol
            ? "btn-brutal-active"  // xác định class từ globals.css
            : "btn-brutal"}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
```

4. Tạo `TopRecList` component inline trong page (< 30 lines) hoặc tách ra nếu cần.

5. Check Tailwind classes: reuse `badge-buy`, `badge-sell`, `card-brutal` đã có trong `globals.css`.

## Success Criteria

- `http://localhost:3001/market` → render trang, không lỗi
- Tab click → URL đổi `?tab=VN30`, chart đổi symbol
- Top MUA/BÁN render (hoặc empty state nếu screener trả rỗng)
