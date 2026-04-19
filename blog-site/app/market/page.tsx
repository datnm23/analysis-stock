import { Suspense } from "react";
import Link from "next/link";
import type { Metadata } from "next";
import { StockChart } from "@/components/stock-chart";
import { ScreenerTable } from "@/components/screener-table";
import { MarketTabSwitcher } from "./market-tab-switcher";

export const metadata: Metadata = {
  title: "Thị trường – VietStock AI",
  description: "Biểu đồ VN-Index, VN30, HNX và các khuyến nghị AI cập nhật",
};

const TABS = [
  { label: "VN-Index",  symbol: "VNINDEX" },
  { label: "VN30",      symbol: "VN30" },
  { label: "HNX",       symbol: "HNXINDEX" },
  { label: "UPCOM",     symbol: "UPCOMINDEX" },
];

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const REC_BADGE: Record<string, string> = { BUY: "badge-buy", SELL: "badge-sell" };
const REC_LABEL: Record<string, string> = { BUY: "MUA", SELL: "BÁN" };

interface ScreenerItem {
  symbol: string;
  recommendation: string;
  confidence: number;
  article_slug: string;
}

async function fetchTopRec(rec: "BUY" | "SELL"): Promise<ScreenerItem[]> {
  try {
    const res = await fetch(`${API}/api/v1/screener?recommendation=${rec}&limit=3`, {
      next: { revalidate: 300 },
    });
    if (!res.ok) return [];
    const data = await res.json();
    return data.results ?? [];
  } catch {
    return [];
  }
}

function TopRecList({ title, items, rec }: { title: string; items: ScreenerItem[]; rec: "BUY" | "SELL" }) {
  return (
    <div className="border-3 border-ink bg-white shadow-brutal">
      <div className="bg-ink px-4 py-2 flex items-center gap-2">
        <span className={`${REC_BADGE[rec]} text-xs`}>{REC_LABEL[rec]}</span>
        <h2 className="text-yellow font-black text-sm uppercase tracking-widest">{title}</h2>
      </div>
      {items.length === 0 ? (
        <p className="px-4 py-3 text-ink/40 text-sm font-mono">Chưa có dữ liệu</p>
      ) : (
        <ul className="divide-y-3 divide-ink/10">
          {items.map((item) => (
            <li key={item.symbol} className="flex items-center justify-between px-4 py-3 hover:bg-yellow/10">
              <Link href={`/symbols/${item.symbol}`} className="badge-symbol hover:opacity-80">
                {item.symbol}
              </Link>
              <span className="text-xs font-bold text-ink/50 tabular-nums">
                {Math.round(item.confidence * 100)}%
              </span>
              <Link
                href={`/articles/${item.article_slug}`}
                className="text-xs font-black uppercase border-2 border-ink px-2 py-0.5 hover:bg-yellow transition-colors"
              >
                Xem →
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

export default async function MarketPage({
  searchParams,
}: {
  searchParams: Promise<{ tab?: string }>;
}) {
  const params = await searchParams;
  const activeSymbol = TABS.find((t) => t.symbol === params.tab)?.symbol ?? "VNINDEX";
  const [buyTop, sellTop] = await Promise.all([fetchTopRec("BUY"), fetchTopRec("SELL")]);

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-black uppercase tracking-tight text-ink border-b-3 border-ink pb-3">
        Thị trường
      </h1>

      {/* Index tabs */}
      <Suspense fallback={
        <div className="flex gap-2">
          {TABS.map((t) => (
            <div key={t.symbol} className="h-10 w-20 bg-ink/10 animate-pulse border-3 border-ink" />
          ))}
        </div>
      }>
        <MarketTabSwitcher tabs={TABS} activeSymbol={activeSymbol} />
      </Suspense>

      {/* Chart */}
      <StockChart symbol={activeSymbol} days={90} />

      {/* Top recommendations */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
        <TopRecList title="Top khuyến nghị MUA" items={buyTop} rec="BUY" />
        <TopRecList title="Top khuyến nghị BÁN" items={sellTop} rec="SELL" />
      </div>

      {/* Full screener */}
      <div>
        <h2 className="text-lg font-black uppercase tracking-tight text-ink border-b-3 border-ink pb-2 mb-6">
          Screener AI
        </h2>
        <ScreenerTable />
      </div>
    </div>
  );
}
