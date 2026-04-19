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

const INDEX_LABELS: Record<string, string> = {
  VNINDEX:    "VN-Index",
  VN30:       "VN30",
  HNXINDEX:   "HNX-Index",
  UPCOMINDEX: "UPCOM",
};

export function MarketIndexBar() {
  const [indices, setIndices] = useState<IndexSnap[]>([]);

  async function load() {
    try {
      const res = await fetch(`${API}/api/v1/market/indices`, { cache: "no-store" });
      if (!res.ok) return;
      const data = await res.json();
      setIndices(data.indices ?? []);
    } catch { /* silent — bar simply stays empty */ }
  }

  useEffect(() => {
    load();
    const t = setInterval(load, 60_000);
    return () => clearInterval(t);
  }, []);

  if (!indices.length) return null;

  return (
    <div className="flex flex-wrap gap-px bg-ink border-3 border-ink shadow-brutal overflow-hidden">
      {indices.map((idx) => {
        const up = idx.change >= 0;
        const color = up ? "text-[#05D98F]" : "text-[#FF4A4A]";
        const arrow = up ? "▲" : "▼";
        return (
          <div
            key={idx.symbol}
            className="flex items-center gap-3 bg-cream px-4 py-2.5 flex-1 min-w-[155px]"
          >
            <span className="font-black text-[11px] uppercase tracking-widest text-ink/50 whitespace-nowrap">
              {INDEX_LABELS[idx.symbol] ?? idx.symbol}
            </span>
            <span className="font-black text-sm tabular-nums text-ink">
              {idx.close.toLocaleString("vi-VN")}
            </span>
            <span className={`font-bold text-xs tabular-nums ${color}`}>
              {arrow} {Math.abs(idx.change_pct).toFixed(2)}%
            </span>
          </div>
        );
      })}
    </div>
  );
}
