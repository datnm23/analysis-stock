"use client";

import { useEffect, useState } from "react";

interface WidgetProps {
  symbol: string;
  forecast?: {
    recommendation?: string;
    confidence?: number;
    technical_score?: number;
    sentiment_score?: number;
  } | null;
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

const REC_STYLE: Record<string, string> = {
  BUY: "bg-[#05D98F] text-ink",
  SELL: "bg-[#FF4A4A] text-white",
  HOLD: "bg-yellow text-ink",
};
const REC_LABEL: Record<string, string> = {
  BUY: "KHUYẾN NGHỊ MUA", SELL: "KHUYẾN NGHỊ BÁN", HOLD: "KHUYẾN NGHỊ GIỮ",
};

function fmtPrice(v: number) {
  return v >= 1000 ? v.toLocaleString("vi-VN") : v.toFixed(2);
}

export function StockDataWidget({ symbol, forecast }: WidgetProps) {
  const [price, setPrice] = useState<number | null>(null);
  const [change, setChange] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchPrice() {
      try {
        const res = await fetch(`${API_URL}/api/v1/chart/${symbol}?days=5`);
        if (!res.ok) return;
        const json = await res.json();
        const bars = json.bars ?? [];
        if (bars.length >= 2) {
          const last = bars[bars.length - 1];
          const prev = bars[bars.length - 2];
          setPrice(last.close);
          setChange(((last.close - prev.close) / prev.close) * 100);
        } else if (bars.length === 1) {
          setPrice(bars[0].close);
          setChange(0);
        }
      } finally {
        setLoading(false);
      }
    }
    fetchPrice();
  }, [symbol]);

  const rec = forecast?.recommendation;
  const confidence = forecast?.confidence ?? 0;
  const techScore = (forecast?.technical_score ?? 0) * 10;
  const sentScore = (forecast?.sentiment_score ?? 0) * 10;

  return (
    <div className="border-3 border-ink bg-cream shadow-brutal">
      {/* Symbol header */}
      <div className="bg-ink px-4 py-3 flex items-center justify-between">
        <span className="badge-symbol text-base px-3 py-1">{symbol}</span>
        <span className="text-white/40 text-[10px] font-mono uppercase tracking-widest">Live Data</span>
      </div>

      {/* Price */}
      <div className="px-4 py-4 border-b-3 border-ink">
        {loading ? (
          <div className="h-8 bg-ink/10 animate-pulse rounded" />
        ) : price !== null ? (
          <div className="flex items-baseline gap-3">
            <span className="text-2xl font-black text-ink tabular-nums">{fmtPrice(price)}</span>
            {change !== null && (
              <span className={`text-sm font-bold ${change >= 0 ? "text-[#05D98F]" : "text-[#FF4A4A]"}`}>
                {change >= 0 ? "▲" : "▼"} {Math.abs(change).toFixed(2)}%
              </span>
            )}
          </div>
        ) : (
          <span className="text-ink/30 text-xs font-mono">Không có dữ liệu</span>
        )}
        <p className="text-[10px] text-ink/40 font-mono uppercase mt-1">Giá đóng cửa gần nhất (VNĐ)</p>
      </div>

      {/* Recommendation */}
      {rec && (
        <div className={`px-4 py-3 border-b-3 border-ink ${REC_STYLE[rec] ?? "bg-yellow text-ink"}`}>
          <p className="text-xs font-black uppercase tracking-widest">{REC_LABEL[rec] ?? rec}</p>
          <div className="mt-2">
            <div className="flex items-center justify-between text-[10px] font-bold uppercase mb-1">
              <span>Độ tin cậy</span>
              <span>{Math.round(confidence * 100)}%</span>
            </div>
            <div className="h-2 bg-ink/20 border border-ink/30">
              <div
                className="h-full bg-ink/60"
                style={{ width: `${Math.round(confidence * 100)}%` }}
              />
            </div>
          </div>
        </div>
      )}

      {/* Scores */}
      {(forecast?.technical_score !== undefined || forecast?.sentiment_score !== undefined) && (
        <div className="grid grid-cols-2 divide-x-3 divide-ink border-b-3 border-ink">
          {[
            { label: "Kỹ thuật", value: techScore },
            { label: "Sentiment", value: sentScore },
          ].map(({ label, value }) => (
            <div key={label} className="px-3 py-3 text-center">
              <p className="text-[10px] text-ink/50 font-black uppercase tracking-wider">{label}</p>
              <p className="text-lg font-black text-ink tabular-nums">{value.toFixed(1)}</p>
              <p className="text-[10px] text-ink/40">/10</p>
            </div>
          ))}
        </div>
      )}

      <div className="px-4 py-2 text-[10px] text-ink/40 font-mono">
        Nguồn: KB Securities · Tự động cập nhật
      </div>
    </div>
  );
}
