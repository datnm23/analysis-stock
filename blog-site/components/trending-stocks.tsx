import Link from "next/link";

const API_URL = process.env.API_URL || "http://localhost:8080";

interface TrendingItem {
  symbol: string;
  count: number;
  recommendation: string;
  last_published: string;
}

const REC_BADGE: Record<string, string> = {
  BUY: "badge-buy", SELL: "badge-sell", HOLD: "badge-hold",
};
const REC_LABEL: Record<string, string> = {
  BUY: "MUA", SELL: "BÁN", HOLD: "GIỮ",
};

async function getTrending(days = 7, limit = 5): Promise<TrendingItem[]> {
  try {
    const res = await fetch(
      `${API_URL}/api/v1/trending?days=${days}&limit=${limit}`,
      { next: { revalidate: 300 } }
    );
    if (!res.ok) return [];
    const json = await res.json();
    return json.trending ?? [];
  } catch {
    return [];
  }
}

export async function TrendingStocks({ days = 7, limit = 5 }: { days?: number; limit?: number }) {
  const items = await getTrending(days, limit);
  if (!items.length) return null;

  return (
    <div className="border-3 border-ink shadow-brutal">
      <div className="bg-ink px-4 py-2 flex items-center justify-between">
        <h2 className="text-yellow font-black text-sm uppercase tracking-widest">
          🔥 Hot trong {days} ngày
        </h2>
        <Link href="/screener" className="text-white/50 text-[10px] font-mono uppercase hover:text-yellow transition-colors">
          Xem tất cả →
        </Link>
      </div>

      <div className="divide-y-3 divide-ink bg-cream">
        {items.map((item, idx) => (
          <Link
            key={item.symbol}
            href={`/symbols/${item.symbol}`}
            className="flex items-center gap-3 px-4 py-3 hover:bg-yellow/20 transition-colors group"
          >
            <span className="w-6 text-center font-black text-ink/30 text-sm tabular-nums">
              {idx + 1}
            </span>
            <span className="badge-symbol flex-shrink-0">{item.symbol}</span>
            <span className="flex-1 text-xs text-ink/50 font-mono">
              {item.count} bài
            </span>
            {item.recommendation && (
              <span className={`${REC_BADGE[item.recommendation] ?? "badge-hold"} text-[10px]`}>
                {REC_LABEL[item.recommendation] ?? item.recommendation}
              </span>
            )}
            <span className="text-ink/30 group-hover:text-ink text-xs">→</span>
          </Link>
        ))}
      </div>
    </div>
  );
}
