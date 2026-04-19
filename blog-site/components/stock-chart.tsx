"use client";

import { useEffect, useRef, useState } from "react";

interface Bar {
  time: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

interface TooltipData {
  date: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  change: number;
  changePercent: number;
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const VALID_SYMBOL = /^[A-Z0-9]{2,10}$/;

function fmtPrice(v: number) {
  return v >= 1000 ? v.toLocaleString("vi-VN") : v.toFixed(2);
}
function fmtVol(v: number) {
  if (v >= 1_000_000) return (v / 1_000_000).toFixed(2) + "M";
  if (v >= 1_000) return (v / 1_000).toFixed(0) + "K";
  return String(v);
}

function buildSmaData(bars: Bar[], period: number): { time: string; value: number }[] {
  const closes = bars.map((b) => b.close);
  const result: { time: string; value: number }[] = [];
  for (let i = period - 1; i < bars.length; i++) {
    const avg = closes.slice(i - period + 1, i + 1).reduce((a, b) => a + b, 0) / period;
    result.push({ time: bars[i].time, value: avg });
  }
  return result;
}

export function StockChart({ symbol, days = 90 }: { symbol: string; days?: number }) {
  const containerRef = useRef<HTMLDivElement>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const chartRef = useRef<any>(null);
  const [tooltip, setTooltip] = useState<TooltipData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dataWarning, setDataWarning] = useState<string | null>(null);

  const normalizedSymbol = (symbol ?? "").trim().toUpperCase();
  const isValid = VALID_SYMBOL.test(normalizedSymbol);

  useEffect(() => {
    if (!isValid || !containerRef.current) return;
    setDataWarning(null);

    let destroyed = false;

    async function init() {
      try {
        const res = await fetch(`${API_URL}/api/v1/chart/${normalizedSymbol}?days=${days + 30}`);
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const json = await res.json();
        const bars: Bar[] = json.bars || [];
        if (!bars.length) throw new Error("No data");

        if (!destroyed) {
          if (bars.length < 20) setDataWarning("Không đủ dữ liệu lịch sử để tính SMA");
          else if (bars.length < 50) setDataWarning("Chưa đủ 50 phiên — đường SMA50 chưa hiển thị");
        }

        const { createChart, CandlestickSeries, HistogramSeries, LineSeries } = await import("lightweight-charts");
        if (destroyed || !containerRef.current) return;

        const chart = createChart(containerRef.current, {
          layout: { background: { color: "#0f0f1a" }, textColor: "#c8c8d4" },
          grid: { vertLines: { color: "#1e1e2e" }, horzLines: { color: "#1e1e2e" } },
          crosshair: { mode: 1 },
          rightPriceScale: { borderColor: "#2a2a3e" },
          timeScale: { borderColor: "#2a2a3e", timeVisible: true, secondsVisible: false },
          height: 420,
          localization: { priceFormatter: (p: number) => fmtPrice(p) },
        });
        chartRef.current = chart;

        const candleSeries = chart.addSeries(CandlestickSeries, {
          upColor: "#00d4aa", downColor: "#ff4444",
          borderUpColor: "#00d4aa", borderDownColor: "#ff4444",
          wickUpColor: "#00d4aa", wickDownColor: "#ff4444",
          priceScaleId: "right",
        });
        candleSeries.setData(bars.map((b) => ({ time: b.time, open: b.open, high: b.high, low: b.low, close: b.close })));

        const volSeries = chart.addSeries(HistogramSeries, {
          priceScaleId: "volume",
          priceFormat: { type: "volume" },
        });
        volSeries.setData(bars.map((b) => ({
          time: b.time,
          value: b.volume,
          color: b.close >= b.open ? "#00d4aa44" : "#ff444444",
        })));
        chart.priceScale("volume").applyOptions({ scaleMargins: { top: 0.75, bottom: 0 } });
        chart.priceScale("right").applyOptions({ scaleMargins: { top: 0, bottom: 0.28 } });

        const sma20Series = chart.addSeries(LineSeries, {
          color: "#4fc3f7", lineWidth: 1, priceScaleId: "right",
          crosshairMarkerVisible: false, lastValueVisible: false, priceLineVisible: false,
        });
        sma20Series.setData(buildSmaData(bars, 20));

        const sma50Series = chart.addSeries(LineSeries, {
          color: "#ffb74d", lineWidth: 1, priceScaleId: "right",
          crosshairMarkerVisible: false, lastValueVisible: false, priceLineVisible: false,
        });
        sma50Series.setData(buildSmaData(bars, 50));

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        chart.subscribeCrosshairMove((param: any) => {
          if (!param?.time) { setTooltip(null); return; }
          const cd = param.seriesData.get(candleSeries);
          const vd = param.seriesData.get(volSeries);
          if (!cd) return;
          const idx = bars.findIndex((b) => b.time === param.time);
          const prev = idx > 0 ? bars[idx - 1].close : cd.close;
          const change = cd.close - prev;
          setTooltip({
            date: String(param.time),
            open: cd.open, high: cd.high, low: cd.low, close: cd.close,
            volume: vd?.value ?? 0,
            change,
            changePercent: prev ? (change / prev) * 100 : 0,
          });
        });

        chart.timeScale().fitContent();
        setLoading(false);
      } catch (e: unknown) {
        if (!destroyed) setError(e instanceof Error ? e.message : "Unknown error");
        setLoading(false);
      }
    }

    init();
    return () => {
      destroyed = true;
      chartRef.current?.remove();
      chartRef.current = null;
    };
  }, [isValid, normalizedSymbol, days]);

  useEffect(() => {
    if (!containerRef.current || !chartRef.current) return;
    const ro = new ResizeObserver(() => {
      if (containerRef.current && chartRef.current)
        chartRef.current.applyOptions({ width: containerRef.current.clientWidth });
    });
    ro.observe(containerRef.current);
    return () => ro.disconnect();
  }, [loading]);

  // Early return AFTER all hooks — React hooks rules compliance
  if (!isValid) return null;

  return (
    <div className="overflow-hidden bg-[#0f0f1a]">
      <div className="flex items-center justify-between px-4 py-2 border-b border-[#2a2a3e]">
        <span className="text-sm font-semibold text-slate-200">
          {normalizedSymbol} — Biểu đồ giá & khối lượng ({days} ngày)
        </span>
        <div className="flex gap-3 text-xs text-slate-400">
          <span className="flex items-center gap-1"><span className="inline-block w-6 h-0.5 bg-[#4fc3f7]" />SMA20</span>
          <span className="flex items-center gap-1"><span className="inline-block w-6 h-0.5 bg-[#ffb74d]" />SMA50</span>
        </div>
      </div>

      <div className="px-4 py-1 min-h-[28px] text-xs font-mono flex flex-wrap gap-x-4 gap-y-0.5 bg-[#0f0f1a] text-slate-300">
        {tooltip ? (
          <>
            <span className="text-slate-400">{tooltip.date}</span>
            <span>O <b>{fmtPrice(tooltip.open)}</b></span>
            <span>H <b className="text-[#00d4aa]">{fmtPrice(tooltip.high)}</b></span>
            <span>L <b className="text-[#ff4444]">{fmtPrice(tooltip.low)}</b></span>
            <span>C <b>{fmtPrice(tooltip.close)}</b></span>
            <span className={tooltip.change >= 0 ? "text-[#00d4aa]" : "text-[#ff4444]"}>
              {tooltip.change >= 0 ? "▲" : "▼"} {Math.abs(tooltip.change).toLocaleString("vi-VN")} ({tooltip.changePercent.toFixed(2)}%)
            </span>
            <span className="text-slate-400">KL <b className="text-slate-200">{fmtVol(tooltip.volume)}</b></span>
          </>
        ) : (
          <span className="text-slate-600 font-mono uppercase tracking-widest text-[10px]">
            Di chuột / chạm → xem giá & KL từng ngày
          </span>
        )}
      </div>

      <div className="relative">
        {loading && (
          <div className="absolute inset-0 flex items-center justify-center bg-[#0f0f1a] z-10">
            <span className="text-slate-500 text-xs font-mono uppercase tracking-widest animate-pulse">
              Đang tải biểu đồ {normalizedSymbol}...
            </span>
          </div>
        )}
        {error && (
          <div className="absolute inset-0 flex items-center justify-center bg-[#0f0f1a] z-10 px-6">
            <span className="text-[#FF4A4A] text-xs font-mono uppercase tracking-wide text-center">
              Không lấy được dữ liệu — thử tải lại trang.
            </span>
          </div>
        )}
        <div ref={containerRef} style={{ height: 420 }} />
      </div>

      {dataWarning && (
        <div className="px-4 py-1 text-[10px] text-amber-400 font-mono bg-[#0f0f1a] border-t border-[#2a2a3e]">
          ⚠ {dataWarning}
        </div>
      )}
      <div className="px-4 py-1.5 text-[11px] text-slate-500 border-t border-[#2a2a3e] font-mono">
        Nguồn: KB Securities · Giá VNĐ · {days} ngày gần nhất
      </div>
    </div>
  );
}
