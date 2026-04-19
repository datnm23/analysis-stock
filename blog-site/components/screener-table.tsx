"use client";

import Link from "next/link";
import { useState, useEffect } from "react";

interface ScreenerItem {
  symbol: string;
  recommendation: string;
  confidence: number;
  technical_score: number;
  sentiment_score: number;
  last_analyzed: string;
  article_slug: string;
}

type SortKey = "confidence" | "technical_score" | "sentiment_score" | "last_analyzed" | "symbol";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

const REC_FILTERS = ["", "BUY", "HOLD", "SELL"] as const;
const REC_LABEL: Record<string, string> = { BUY: "MUA", SELL: "BÁN", HOLD: "GIỮ" };
const REC_BADGE: Record<string, string> = { BUY: "badge-buy", SELL: "badge-sell", HOLD: "badge-hold" };

function ConfidenceBar({ value }: { value: number }) {
  const pct = Math.round(value * 100);
  return (
    <div className="flex items-center gap-2">
      <div className="w-16 h-2 bg-ink/10 border border-ink/20 flex-shrink-0">
        <div className="h-full bg-ink/50" style={{ width: `${pct}%` }} />
      </div>
      <span className="tabular-nums text-xs font-bold">{pct}%</span>
    </div>
  );
}

export function ScreenerTable() {
  const [items, setItems] = useState<ScreenerItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [recFilter, setRecFilter] = useState<string>("");
  const [minConf, setMinConf] = useState(false);
  const [sortKey, setSortKey] = useState<SortKey>("confidence");
  const [sortAsc, setSortAsc] = useState(false);

  useEffect(() => {
    async function load() {
      setLoading(true);
      try {
        const params = new URLSearchParams({ limit: "200" });
        if (recFilter) params.set("recommendation", recFilter);
        if (minConf) params.set("min_confidence", "0.7");
        const res = await fetch(`${API_URL}/api/v1/screener?${params}`);
        if (!res.ok) return;
        const json = await res.json();
        setItems(json.results ?? []);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [recFilter, minConf]);

  const sorted = [...items].sort((a, b) => {
    const av = a[sortKey] as string | number;
    const bv = b[sortKey] as string | number;
    const cmp = av < bv ? -1 : av > bv ? 1 : 0;
    return sortAsc ? cmp : -cmp;
  });

  function handleSort(key: SortKey) {
    if (sortKey === key) setSortAsc((p) => !p);
    else { setSortKey(key); setSortAsc(false); }
  }

  function SortIndicator({ k }: { k: SortKey }) {
    if (sortKey !== k) return <span className="text-ink/20">↕</span>;
    return <span>{sortAsc ? "↑" : "↓"}</span>;
  }

  const formatDate = (iso: string) => {
    try {
      return new Date(iso).toLocaleDateString("vi-VN", { day: "2-digit", month: "2-digit" });
    } catch { return iso.slice(0, 10); }
  };

  return (
    <div>
      {/* Filter bar */}
      <div className="flex flex-wrap items-center gap-2 mb-6">
        {REC_FILTERS.map((r) => (
          <button
            key={r || "all"}
            onClick={() => setRecFilter(r)}
            className={`font-black text-sm px-3 py-1.5 border-3 border-ink transition-none
              ${recFilter === r ? "bg-ink text-yellow" : "bg-white text-ink hover:bg-yellow"}`}
          >
            {r ? REC_LABEL[r] : "TẤT CẢ"}
            {!r && items.length > 0 && ` (${items.length})`}
          </button>
        ))}
        <button
          onClick={() => setMinConf((p) => !p)}
          className={`font-bold text-sm px-3 py-1.5 border-3 border-ink transition-none
            ${minConf ? "bg-ink text-yellow" : "bg-white text-ink hover:bg-yellow"}`}
        >
          {minConf ? "✓" : ""} Tin cậy &gt; 70%
        </button>
      </div>

      {loading ? (
        <div className="border-3 border-ink p-8 text-center">
          <span className="text-ink/40 font-mono text-sm uppercase tracking-widest animate-pulse">
            Đang tải...
          </span>
        </div>
      ) : sorted.length === 0 ? (
        <div className="border-3 border-ink bg-white p-8 text-center">
          <p className="font-black text-ink/30 uppercase">Không có kết quả</p>
        </div>
      ) : (
        <div className="border-3 border-ink overflow-x-auto shadow-brutal">
          <table className="w-full text-sm border-collapse">
            <thead>
              <tr className="bg-ink text-yellow">
                {([
                  ["symbol", "Mã"],
                  [null, "Khuyến nghị"],
                  ["confidence", "Tin cậy"],
                  ["technical_score", "Kỹ thuật"],
                  ["sentiment_score", "Sentiment"],
                  ["last_analyzed", "Cập nhật"],
                ] as [SortKey | null, string][]).map(([key, label]) => (
                  <th
                    key={label}
                    onClick={() => key && handleSort(key)}
                    className={`px-4 py-3 text-left font-black text-xs uppercase tracking-widest whitespace-nowrap
                      ${key ? "cursor-pointer hover:text-white" : ""}`}
                  >
                    {label} {key && <SortIndicator k={key} />}
                  </th>
                ))}
                <th className="px-4 py-3 text-left font-black text-xs uppercase tracking-widest">Xem</th>
              </tr>
            </thead>
            <tbody className="divide-y-3 divide-ink bg-cream">
              {sorted.map((item) => (
                <tr key={item.symbol} className="hover:bg-yellow/10 transition-colors">
                  <td className="px-4 py-3">
                    <Link href={`/symbols/${item.symbol}`} className="badge-symbol hover:opacity-80">
                      {item.symbol}
                    </Link>
                  </td>
                  <td className="px-4 py-3">
                    {item.recommendation ? (
                      <span className={`${REC_BADGE[item.recommendation] ?? "badge-hold"} text-xs`}>
                        {REC_LABEL[item.recommendation] ?? item.recommendation}
                      </span>
                    ) : "—"}
                  </td>
                  <td className="px-4 py-3"><ConfidenceBar value={item.confidence} /></td>
                  <td className="px-4 py-3 font-bold tabular-nums">{(item.technical_score * 10).toFixed(1)}</td>
                  <td className="px-4 py-3 font-bold tabular-nums">{(item.sentiment_score * 10).toFixed(1)}</td>
                  <td className="px-4 py-3 text-ink/50 font-mono text-xs">{formatDate(item.last_analyzed)}</td>
                  <td className="px-4 py-3">
                    <Link
                      href={`/articles/${item.article_slug}`}
                      className="font-black text-xs uppercase tracking-wide border-2 border-ink px-2 py-1 hover:bg-yellow transition-colors"
                    >
                      Xem →
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
