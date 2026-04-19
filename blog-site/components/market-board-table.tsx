"use client";

import { useEffect, useMemo, useState } from "react";
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
  ref: number;
  // merged from AI screener
  recommendation?: string;
  confidence?: number;
  article_slug?: string;
}

interface AiItem {
  symbol: string;
  recommendation: string;
  confidence: number;
  article_slug: string;
}

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const REC_BADGE: Record<string, string> = {
  BUY: "badge-buy", SELL: "badge-sell", HOLD: "badge-hold",
};
const REC_LABEL: Record<string, string> = {
  BUY: "MUA", SELL: "BÁN", HOLD: "GIỮ",
};

type SortKey = "symbol" | "price" | "change_pct" | "volume";

function fmtPrice(v: number): string {
  if (!v) return "—";
  return v >= 1000 ? v.toLocaleString("vi-VN") : v.toFixed(2);
}

function fmtVol(v: number): string {
  if (!v) return "—";
  if (v >= 1_000_000) return (v / 1_000_000).toFixed(2) + "M";
  if (v >= 1_000) return (v / 1_000).toFixed(0) + "K";
  return String(v);
}

export function MarketBoardTable() {
  const [rows, setRows]       = useState<StockRow[]>([]);
  const [aiMap, setAiMap]     = useState<Record<string, AiItem>>({});
  const [sector, setSector]   = useState("all");
  const [search, setSearch]   = useState("");
  const [loading, setLoading] = useState(true);
  const [sortKey, setSortKey] = useState<SortKey>("volume");
  const [sortAsc, setSortAsc] = useState(false);
  const [error, setError]     = useState(false);

  // Load price board
  useEffect(() => {
    setLoading(true);
    fetch(`${API}/api/v1/market/board?exchange=ALL`)
      .then((r) => r.ok ? r.json() : Promise.reject(r.status))
      .then((data) => { setRows(data.symbols ?? []); setError(false); })
      .catch(() => setError(true))
      .finally(() => setLoading(false));
  }, []);

  // Load AI screener overlay (non-blocking)
  useEffect(() => {
    fetch(`${API}/api/v1/screener?limit=200`)
      .then((r) => r.ok ? r.json() : null)
      .then((data) => {
        if (!data) return;
        const map: Record<string, AiItem> = {};
        for (const item of (data.results ?? [])) map[item.symbol] = item;
        setAiMap(map);
      })
      .catch(() => {});
  }, []);

  function handleSort(key: SortKey) {
    if (sortKey === key) setSortAsc((p) => !p);
    else { setSortKey(key); setSortAsc(false); }
  }

  const displayed = useMemo(() => {
    let list = rows.map((r) => ({
      ...r,
      ...(aiMap[r.symbol] ? {
        recommendation: aiMap[r.symbol].recommendation,
        confidence:     aiMap[r.symbol].confidence,
        article_slug:   aiMap[r.symbol].article_slug,
      } : {}),
    }));

    if (search) list = list.filter((r) => r.symbol.includes(search.toUpperCase()));
    if (sector !== "all") list = list.filter((r) => getSector(r.symbol) === sector);

    list.sort((a, b) => {
      const av = a[sortKey] as number | string;
      const bv = b[sortKey] as number | string;
      const cmp = av < bv ? -1 : av > bv ? 1 : 0;
      return sortAsc ? cmp : -cmp;
    });

    return list;
  }, [rows, aiMap, sector, search, sortKey, sortAsc]);

  function SortArrow({ k }: { k: SortKey }) {
    if (sortKey !== k) return <span className="text-yellow/30">↕</span>;
    return <span>{sortAsc ? "↑" : "↓"}</span>;
  }

  if (loading) {
    return (
      <div className="border-3 border-ink bg-white p-10 text-center shadow-brutal">
        <p className="font-mono text-sm text-ink/40 uppercase tracking-widest animate-pulse">
          Đang tải bảng giá...
        </p>
      </div>
    );
  }

  if (error || rows.length === 0) {
    return (
      <div className="border-3 border-ink bg-white p-10 text-center shadow-brutal">
        <p className="font-black text-ink/30 uppercase text-lg">Bảng giá chưa sẵn sàng</p>
        <p className="text-xs text-ink/40 mt-2">Crawl-agent đang khởi động — thử lại sau vài phút</p>
      </div>
    );
  }

  return (
    <div>
      {/* Sector tabs + search bar */}
      <div className="flex flex-wrap gap-1.5 mb-4 items-center">
        {SECTORS.map((s) => (
          <button
            key={s.key}
            onClick={() => setSector(s.key)}
            className={`font-black text-[11px] px-2.5 py-1.5 border-3 border-ink uppercase tracking-wide transition-none
              ${sector === s.key ? "bg-ink text-yellow" : "bg-white text-ink hover:bg-yellow"}`}
          >
            {s.label}
            {s.key === "all" && rows.length > 0 && (
              <span className="ml-1 opacity-50">({rows.length})</span>
            )}
          </button>
        ))}
        <input
          type="text"
          placeholder="Tìm mã..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="ml-auto border-3 border-ink px-3 py-1.5 text-sm font-mono w-24 focus:outline-none focus:bg-yellow/20 uppercase"
        />
      </div>

      {/* Table */}
      <div className="border-3 border-ink overflow-x-auto shadow-brutal">
        <table className="w-full text-sm border-collapse">
          <thead>
            <tr className="bg-ink text-yellow">
              {(
                [
                  ["symbol",     "Mã",        true],
                  [null,         "Ngành",      false],
                  ["price",      "Giá",        true],
                  ["change_pct", "+/- %",      true],
                  ["volume",     "KL",         true],
                  [null,         "AI",         false],
                  [null,         "",           false],
                ] as [SortKey | null, string, boolean][]
              ).map(([key, label]) => (
                <th
                  key={label || Math.random()}
                  onClick={() => key && handleSort(key)}
                  className={`px-3 py-2.5 text-left font-black text-[11px] uppercase tracking-widest whitespace-nowrap
                    ${key ? "cursor-pointer hover:text-white" : ""}`}
                >
                  {label} {key && <SortArrow k={key} />}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y-2 divide-ink/10 bg-cream">
            {displayed.slice(0, 300).map((row) => {
              const up  = row.change_pct >= 0;
              const atC = row.ceiling > 0 && row.price >= row.ceiling;
              const atF = row.floor > 0 && row.price <= row.floor;
              const priceColor = atC
                ? "text-[#00E5FF]"
                : atF
                ? "text-[#800000]"
                : up
                ? "text-[#05D98F]"
                : "text-[#FF4A4A]";
              const sectorLabel = getSector(row.symbol);

              return (
                <tr
                  key={row.symbol}
                  className={`hover:bg-yellow/10 transition-colors ${row.recommendation ? "bg-yellow/5" : ""}`}
                >
                  <td className="px-3 py-2 font-black">
                    <Link href={`/symbols/${row.symbol}`} className="badge-symbol hover:opacity-80">
                      {row.symbol}
                    </Link>
                  </td>
                  <td className="px-3 py-2 text-[10px] text-ink/40 font-mono uppercase">
                    {sectorLabel !== "other" ? sectorLabel : ""}
                  </td>
                  <td className={`px-3 py-2 tabular-nums font-bold ${priceColor}`}>
                    {fmtPrice(row.price)}
                  </td>
                  <td className={`px-3 py-2 tabular-nums font-bold ${up ? "text-[#05D98F]" : "text-[#FF4A4A]"}`}>
                    {row.change_pct > 0 ? "+" : ""}{row.change_pct.toFixed(2)}%
                  </td>
                  <td className="px-3 py-2 tabular-nums text-ink/60 text-xs">
                    {fmtVol(row.volume)}
                  </td>
                  <td className="px-3 py-2">
                    {row.recommendation && (
                      <span className={`${REC_BADGE[row.recommendation] ?? "badge-hold"} text-[10px]`}>
                        {REC_LABEL[row.recommendation] ?? row.recommendation}
                      </span>
                    )}
                  </td>
                  <td className="px-3 py-2">
                    {row.article_slug && (
                      <Link
                        href={`/articles/${row.article_slug}`}
                        className="text-[10px] font-black border-2 border-ink px-1.5 py-0.5 hover:bg-yellow transition-colors whitespace-nowrap"
                      >
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
          <p className="text-center text-xs text-ink/40 py-2 font-mono border-t-2 border-ink/10">
            Hiển thị 300/{displayed.length} — dùng bộ lọc ngành hoặc tìm kiếm
          </p>
        )}
      </div>
    </div>
  );
}
