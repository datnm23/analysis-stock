import { notFound } from "next/navigation";
import Link from "next/link";
import type { Metadata } from "next";
import { getArticlesBySymbol, getArticles } from "@/lib/articles-api";
import { ArticleCard } from "@/components/article-card";
import { StockChart } from "@/components/stock-chart";

export const revalidate = 60;
export const dynamicParams = true;

export async function generateStaticParams() {
  try {
    const { articles } = await getArticles(100, 0);
    const seen = new Set<string>();
    const symbols = articles.map((a) => a.symbol).filter((s) => seen.has(s) ? false : seen.add(s) && true);
    return symbols.map((symbol) => ({ symbol }));
  } catch {
    return [];
  }
}

export async function generateMetadata({
  params,
}: {
  params: { symbol: string };
}): Promise<Metadata> {
  const symbol = params.symbol.toUpperCase();
  return {
    title: `Phân tích ${symbol} — VietStock AI`,
    description: `Tổng hợp phân tích kỹ thuật và AI cho mã cổ phiếu ${symbol} (HSX/HNX/UPCOM)`,
  };
}

function parseForecast(raw?: string) {
  if (!raw) return null;
  try { return JSON.parse(raw); } catch { return null; }
}

const REC_CONFIG: Record<string, { bg: string; label: string; badge: string }> = {
  BUY:  { bg: "bg-[#05D98F]", label: "KHUYẾN NGHỊ MUA", badge: "badge-buy" },
  SELL: { bg: "bg-[#FF4A4A]", label: "KHUYẾN NGHỊ BÁN", badge: "badge-sell" },
  HOLD: { bg: "bg-yellow",    label: "KHUYẾN NGHỊ GIỮ", badge: "badge-hold" },
};

const PAGE_SIZE = 12;

export default async function SymbolPage({
  params,
  searchParams,
}: {
  params: { symbol: string };
  searchParams: { page?: string };
}) {
  const symbol = params.symbol.toUpperCase();
  const page = Math.max(1, parseInt(searchParams.page ?? "1"));
  const offset = (page - 1) * PAGE_SIZE;

  const { articles, total } = await getArticlesBySymbol(symbol, PAGE_SIZE, offset);

  if (total === 0 && page === 1) notFound();

  const totalPages = Math.ceil(total / PAGE_SIZE);
  const latest = articles[0];
  const forecast = parseForecast(latest?.forecast_data);
  const rec = forecast?.recommendation as string | undefined;
  const recCfg = rec ? REC_CONFIG[rec] ?? null : null;

  return (
    <div className="max-w-5xl mx-auto">
      {/* Back */}
      <Link
        href="/articles"
        className="inline-flex items-center gap-2 font-bold text-sm text-ink uppercase tracking-wide border-b-2 border-ink hover:border-yellow hover:text-yellow transition-colors duration-100 mb-8"
      >
        ← Tất cả bài viết
      </Link>

      {/* Header */}
      <div className="border-3 border-ink bg-cream shadow-brutal mb-6 p-6">
        <div className="flex flex-wrap items-center gap-3 mb-3">
          <span className="badge-symbol text-xl px-4 py-2">{symbol}</span>
          {rec && recCfg && (
            <span className={`${recCfg.badge} text-base px-3 py-1`}>{recCfg.label}</span>
          )}
        </div>
        <h1 className="text-3xl font-black text-ink uppercase tracking-tight">
          Phân tích AI — {symbol}
        </h1>
        <p className="text-ink/50 text-sm mt-2">{total} bài phân tích</p>
      </div>

      {/* Forecast bar */}
      {forecast && rec && recCfg && (
        <div className={`border-3 border-ink ${recCfg.bg} shadow-brutal-sm mb-8 grid grid-cols-3 divide-x-3 divide-ink`}>
          {[
            { label: "Độ tin cậy", value: `${Math.round((forecast.confidence ?? 0) * 100)}%` },
            { label: "Kỹ thuật",   value: `${((forecast.technical_score ?? 0) * 10).toFixed(1)}/10` },
            { label: "Sentiment",  value: `${((forecast.sentiment_score ?? 0) * 10).toFixed(1)}/10` },
          ].map(({ label, value }) => (
            <div key={label} className="px-4 py-4 text-center">
              <p className="text-xs font-black text-ink/60 uppercase tracking-widest mb-1">{label}</p>
              <p className="text-2xl font-black text-ink tabular-nums">{value}</p>
            </div>
          ))}
        </div>
      )}

      {/* Chart — 6 months */}
      <div className="border-3 border-ink shadow-brutal mb-8 overflow-hidden">
        <StockChart symbol={symbol} days={180} />
      </div>

      {/* Article grid */}
      <div className="mb-4 flex items-center gap-3">
        <h2 className="text-xl font-black text-ink uppercase tracking-tight">Lịch sử phân tích</h2>
        <span className="bg-yellow border-2 border-ink text-ink font-black text-xs px-2 py-1">
          {total} bài
        </span>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6 mb-10">
        {articles.map((a) => (
          <ArticleCard key={a.id} article={a} />
        ))}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex justify-center items-center gap-2 mt-4">
          {page > 1 && (
            <a href={`/symbols/${symbol}?page=${page - 1}`}
              className="btn-brutal bg-white text-ink font-black text-sm px-4 py-2 uppercase tracking-wide">
              ← Trước
            </a>
          )}
          {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
            <a key={p} href={`/symbols/${symbol}?page=${p}`}
              className={`font-black text-sm px-4 py-2 border-3 border-ink transition-none ${p === page ? "bg-ink text-yellow" : "bg-white text-ink hover:bg-yellow"}`}>
              {p}
            </a>
          ))}
          {page < totalPages && (
            <a href={`/symbols/${symbol}?page=${page + 1}`}
              className="btn-brutal bg-ink text-yellow font-black text-sm px-4 py-2 uppercase tracking-wide">
              Tiếp →
            </a>
          )}
        </div>
      )}
    </div>
  );
}
