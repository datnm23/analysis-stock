---
title: Phase 04 – Homepage redesign
status: completed
progress: 100%
completed: 2026-04-20
---

# Phase 04 – Homepage redesign: Market Overview Dashboard

## Overview

Thay thế blog-hero bằng market dashboard. Layout:
```
┌──────────────────────────────────────────────────────────┐
│  VNINDEX 1817 ▲0.29%  │  VN30 2028 ▼0.1%  │  HNX...    │  ← IndexBar (client, poll 60s)
└──────────────────────────────────────────────────────────┘
┌──────────────────────────────────────────────────────────┐
│  [Tất cả] [Ngân hàng] [BĐS] [Thép] [CK] [Dầu khí] ...  │
│  Mã  │ Giá      │ +/-    │ % KL   │ AI Rec │ TC  │ Sàn  │
│  VCB │ 85,000   │ ▲200   │ +0.24% │ MUA    │ ...  │ ...  │
└──────────────────────────────────────────────────────────┘
┌──────────────────────┐  ┌──────────────────────────────┐
│  Hot 7 ngày          │  │  Bài phân tích mới nhất      │
└──────────────────────┘  └──────────────────────────────┘
```

## Files

- **Sửa**: `blog-site/app/page.tsx` — rewrite hoàn toàn
- **Tạo mới**: `blog-site/components/market-index-bar.tsx` — client component
- **Tạo mới**: `blog-site/components/market-board-table.tsx` — client component
- **Đọc tham khảo**: `blog-site/lib/sector-mapping.ts` (Phase 03), `blog-site/app/globals.css`

## Component 1: `market-index-bar.tsx`

```tsx
"use client";
import { useEffect, useState } from "react";

interface IndexSnap {
  symbol: string;
  close: number;
  change: number;
  change_pct: number;
  volume: number;
}

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const LABELS: Record<string, string> = {
  VNINDEX: "VN-Index", VN30: "VN30",
  HNXINDEX: "HNX-Index", UPCOMINDEX: "UPCOM",
};

export function MarketIndexBar() {
  const [indices, setIndices] = useState<IndexSnap[]>([]);

  async function load() {
    try {
      const res = await fetch(`${API}/api/v1/market/indices`);
      if (!res.ok) return;
      const data = await res.json();
      setIndices(data.indices ?? []);
    } catch { /* silent fail */ }
  }

  useEffect(() => {
    load();
    const t = setInterval(load, 60_000); // poll every 60s
    return () => clearInterval(t);
  }, []);

  if (!indices.length) return null;

  return (
    <div className="flex flex-wrap gap-px bg-ink border-3 border-ink mb-6 shadow-brutal">
      {indices.map((idx) => {
        const up = idx.change >= 0;
        return (
          <div key={idx.symbol}
            className="flex items-center gap-3 bg-cream px-4 py-2.5 flex-1 min-w-[160px]">
            <span className="font-black text-xs uppercase tracking-widest text-ink/50">
              {LABELS[idx.symbol] ?? idx.symbol}
            </span>
            <span className="font-black text-base tabular-nums">
              {idx.close.toLocaleString("vi-VN")}
            </span>
            <span className={`font-bold text-xs tabular-nums ${up ? "text-[#05D98F]" : "text-[#FF4A4A]"}`}>
              {up ? "▲" : "▼"} {Math.abs(idx.change_pct).toFixed(2)}%
            </span>
          </div>
        );
      })}
    </div>
  );
}
```

## Component 2: `market-board-table.tsx`

```tsx
"use client";
import { useEffect, useState, useMemo } from "react";
import Link from "next/link";
import { SECTORS, getSector } from "@/lib/sector-mapping";

interface StockRow {
  symbol: string;
  price: number;
  change: number;
  change_pct: number;
  volume: number;
  ceiling: number;
  floor: number;
  // AI data (merged client-side from screener cache)
  recommendation?: string;
  confidence?: number;
  article_slug?: string;
}

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const REC_BADGE: Record<string, string> = {
  BUY: "badge-buy", SELL: "badge-sell", HOLD: "badge-hold"
};
const REC_LABEL: Record<string, string> = {
  BUY: "MUA", SELL: "BÁN", HOLD: "GIỮ"
};

export function MarketBoardTable() {
  const [rows, setRows] = useState<StockRow[]>([]);
  const [aiMap, setAiMap] = useState<Record<string, StockRow>>({});
  const [sector, setSector] = useState("all");
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");

  // Fetch price board
  useEffect(() => {
    async function loadBoard() {
      setLoading(true);
      try {
        const res = await fetch(`${API}/api/v1/market/board?exchange=ALL`);
        if (!res.ok) return;
        const data = await res.json();
        setRows(data.symbols ?? []);
      } finally {
        setLoading(false);
      }
    }
    loadBoard();
  }, []);

  // Fetch AI screener (merge overlay)
  useEffect(() => {
    async function loadAI() {
      try {
        const res = await fetch(`${API}/api/v1/screener?limit=200`);
        if (!res.ok) return;
        const data = await res.json();
        const map: Record<string, StockRow> = {};
        for (const item of (data.results ?? [])) {
          map[item.symbol] = item;
        }
        setAiMap(map);
      } catch { /* silent */ }
    }
    loadAI();
  }, []);

  // Filter + merge
  const displayed = useMemo(() => {
    let list = rows.map((r) => ({
      ...r,
      ...(aiMap[r.symbol] ? {
        recommendation: aiMap[r.symbol].recommendation,
        confidence: aiMap[r.symbol].confidence,
        article_slug: aiMap[r.symbol].article_slug,
      } : {}),
    }));

    if (search) {
      list = list.filter((r) => r.symbol.includes(search.toUpperCase()));
    }
    if (sector !== "all") {
      list = list.filter((r) => getSector(r.symbol) === sector);
    }
    return list;
  }, [rows, aiMap, sector, search]);

  const fmtPrice = (v: number) => v >= 1000 ? v.toLocaleString("vi-VN") : v.toFixed(2);

  return (
    <div>
      {/* Sector tabs + search */}
      <div className="flex flex-wrap gap-2 mb-4 items-center">
        {SECTORS.map((s) => (
          <button key={s.key} onClick={() => setSector(s.key)}
            className={`font-black text-xs px-3 py-1.5 border-3 border-ink uppercase tracking-wide transition-none
              ${sector === s.key ? "bg-ink text-yellow" : "bg-white text-ink hover:bg-yellow"}`}>
            {s.label}
          </button>
        ))}
        <input
          type="text" placeholder="Tìm mã..." value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="ml-auto border-3 border-ink px-3 py-1.5 text-sm font-mono w-28 focus:outline-none focus:bg-yellow/20"
        />
      </div>

      {loading ? (
        <div className="border-3 border-ink p-8 text-center">
          <span className="text-ink/40 font-mono text-sm uppercase animate-pulse">Đang tải bảng giá...</span>
        </div>
      ) : (
        <div className="border-3 border-ink overflow-x-auto shadow-brutal">
          <table className="w-full text-sm border-collapse">
            <thead>
              <tr className="bg-ink text-yellow">
                {["Mã", "Giá", "±", "%", "KL (triệu)", "AI", ""].map((h) => (
                  <th key={h} className="px-3 py-2.5 text-left font-black text-xs uppercase tracking-widest whitespace-nowrap">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y-2 divide-ink/10 bg-cream">
              {displayed.slice(0, 300).map((row) => {
                const up = row.change >= 0;
                const isAI = !!row.recommendation;
                return (
                  <tr key={row.symbol}
                    className={`hover:bg-yellow/10 transition-colors ${isAI ? "bg-yellow/5" : ""}`}>
                    <td className="px-3 py-2 font-black">
                      <Link href={`/symbols/${row.symbol}`} className="badge-symbol hover:opacity-80">
                        {row.symbol}
                      </Link>
                    </td>
                    <td className="px-3 py-2 tabular-nums font-bold">{fmtPrice(row.price)}</td>
                    <td className={`px-3 py-2 tabular-nums font-bold ${up ? "text-[#05D98F]" : "text-[#FF4A4A]"}`}>
                      {up ? "+" : ""}{row.change.toFixed(2)}
                    </td>
                    <td className={`px-3 py-2 tabular-nums font-bold ${up ? "text-[#05D98F]" : "text-[#FF4A4A]"}`}>
                      {up ? "+" : ""}{row.change_pct.toFixed(2)}%
                    </td>
                    <td className="px-3 py-2 tabular-nums text-ink/60">
                      {(row.volume / 1_000_000).toFixed(2)}
                    </td>
                    <td className="px-3 py-2">
                      {row.recommendation ? (
                        <span className={`${REC_BADGE[row.recommendation] ?? "badge-hold"} text-[10px]`}>
                          {REC_LABEL[row.recommendation] ?? row.recommendation}
                        </span>
                      ) : null}
                    </td>
                    <td className="px-3 py-2">
                      {row.article_slug && (
                        <Link href={`/articles/${row.article_slug}`}
                          className="text-[10px] font-black border-2 border-ink px-1.5 py-0.5 hover:bg-yellow transition-colors">
                          AI →
                        </Link>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
          {displayed.length > 300 && (
            <p className="text-center text-xs text-ink/40 py-2 font-mono">
              Hiển thị 300/{displayed.length} — dùng bộ lọc để thu hẹp
            </p>
          )}
        </div>
      )}
    </div>
  );
}
```

## `app/page.tsx` Redesign

```tsx
import { Suspense } from "react";
import Link from "next/link";
import { getLatestArticles } from "@/lib/articles-api";
import { ArticleCard } from "@/components/article-card";
import { TrendingStocks } from "@/components/trending-stocks";
import { MarketIndexBar } from "@/components/market-index-bar";
import { MarketBoardTable } from "@/components/market-board-table";

export const revalidate = 60;

export default async function HomePage() {
  const articles = await getLatestArticles(4);

  return (
    <div className="space-y-10">
      {/* Index bar */}
      <MarketIndexBar />

      {/* Market board */}
      <section>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-black uppercase tracking-tight text-ink border-b-3 border-yellow pb-1">
            Bảng giá thị trường
          </h2>
          <Link href="/market" className="btn-brutal bg-ink text-yellow text-xs font-black px-4 py-2 uppercase tracking-widest">
            Xem biểu đồ →
          </Link>
        </div>
        <MarketBoardTable />
      </section>

      {/* Trending + Latest */}
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_280px] gap-8 items-start">
        <section>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-black uppercase tracking-tight text-ink border-b-3 border-yellow pb-1">
              Phân tích mới nhất
            </h2>
            <Link href="/articles" className="btn-brutal bg-ink text-yellow text-xs font-black px-4 py-2 uppercase tracking-widest">
              Xem tất cả →
            </Link>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {articles.map((a) => <ArticleCard key={a.id} article={a} />)}
          </div>
        </section>
        <aside>
          <Suspense fallback={null}>
            <TrendingStocks days={7} limit={8} />
          </Suspense>
        </aside>
      </div>
    </div>
  );
}
```

## Success Criteria

- `http://localhost:3001/` hiển thị IndexBar + bảng giá (hoặc empty state khi crawl-agent chưa chạy)
- Sector filter hoạt động client-side, không gọi API lại
- AI badge hiện trên mã có data screener
- Không lỗi TypeScript (`npx tsc --noEmit`)
- Graceful empty state khi price board 502 (crawl-agent offline)
