import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import { getArticleBySlug, getArticles } from "@/lib/articles-api";
import { MarkdownContent } from "@/components/markdown-content";
import { StockChart } from "@/components/stock-chart";

export const revalidate = 60;
export const dynamicParams = true;

export async function generateStaticParams() {
  try {
    const { articles } = await getArticles(100, 0);
    return articles.map((a) => ({ slug: a.slug }));
  } catch {
    return [];
  }
}

const REC_CONFIG: Record<string, { bg: string; label: string; badge: string }> = {
  BUY:  { bg: "bg-[#05D98F]", label: "KHUYẾN NGHỊ MUA", badge: "badge-buy" },
  SELL: { bg: "bg-[#FF4A4A]", label: "KHUYẾN NGHỊ BÁN", badge: "badge-sell" },
  HOLD: { bg: "bg-yellow",    label: "KHUYẾN NGHỊ GIỮ", badge: "badge-hold" },
};

function parseForecast(raw?: string) {
  if (!raw) return null;
  try { return JSON.parse(raw); } catch { return null; }
}

function parseSourceUrls(raw?: string): string[] {
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed.filter(Boolean) : [];
  } catch { return []; }
}

export default async function ArticleDetailPage({
  params,
}: {
  params: { slug: string };
}) {
  const article = await getArticleBySlug(params.slug);
  if (!article) notFound();

  const date = article.published_at
    ? new Date(article.published_at).toLocaleDateString("vi-VN", {
        year: "numeric", month: "long", day: "numeric",
      })
    : "";

  const forecast = parseForecast(article.forecast_data);
  const rec = forecast?.recommendation as string | undefined;
  const recCfg = rec ? REC_CONFIG[rec] ?? REC_CONFIG.HOLD : null;
  const sourceUrls = parseSourceUrls(article.source_urls);

  return (
    <div className="max-w-3xl mx-auto">
      {/* ── Back ── */}
      <Link
        href="/articles"
        className="inline-flex items-center gap-2 font-bold text-sm text-ink uppercase tracking-wide border-b-2 border-ink hover:border-yellow hover:text-yellow transition-colors duration-100 mb-8"
      >
        ← Tất cả bài viết
      </Link>

      {/* ── Hero image ── */}
      {article.image_url ? (
        <div className="relative w-full h-64 border-3 border-ink shadow-brutal-lg mb-8 overflow-hidden">
          <Image
            src={article.image_url}
            alt={article.title}
            fill
            className="object-cover"
            sizes="(max-width: 768px) 100vw, 768px"
            priority
          />
          {/* Symbol watermark */}
          <div className="absolute bottom-0 left-0 bg-ink/80 px-3 py-1">
            <span className="badge-symbol">{article.symbol}</span>
          </div>
        </div>
      ) : (
        <div className="w-full h-36 border-3 border-ink bg-ink shadow-brutal-lg mb-8 flex items-center justify-center">
          <span className="text-6xl font-black text-yellow/20 tracking-widest">{article.symbol}</span>
        </div>
      )}

      {/* ── Header block ── */}
      <div className="border-3 border-ink bg-cream shadow-brutal mb-6 p-6">
        {/* Badges */}
        <div className="flex flex-wrap items-center gap-2 mb-4">
          <span className="badge-symbol text-base px-3 py-1">{article.symbol}</span>
          {rec && recCfg && (
            <span className={`${recCfg.badge} text-base px-3 py-1`}>{recCfg.label}</span>
          )}
        </div>

        <h1 className="text-2xl md:text-3xl font-black text-ink leading-tight uppercase tracking-tight mb-3">
          {article.title}
        </h1>

        {date && (
          <p className="text-xs font-semibold text-ink/40 uppercase tracking-widest">
            {date}
          </p>
        )}
      </div>

      {/* ── Forecast metrics bar ── */}
      {forecast && rec && recCfg && (
        <div className={`border-3 border-ink ${recCfg.bg} shadow-brutal-sm mb-8 grid grid-cols-3 divide-x-3 divide-ink`}>
          {[
            { label: "Độ tin cậy", value: `${Math.round((forecast.confidence ?? 0) * 100)}%` },
            { label: "Kỹ thuật", value: `${((forecast.technical_score ?? 0) * 10).toFixed(1)}/10` },
            { label: "Sentiment", value: `${((forecast.sentiment_score ?? 0) * 10).toFixed(1)}/10` },
          ].map(({ label, value }) => (
            <div key={label} className="px-4 py-4 text-center">
              <p className="text-xs font-black text-ink/60 uppercase tracking-widest mb-1">{label}</p>
              <p className="text-2xl font-black text-ink tabular-nums">{value}</p>
            </div>
          ))}
        </div>
      )}

      {/* ── Interactive chart ── */}
      <div className="border-3 border-ink shadow-brutal mb-8 overflow-hidden">
        <StockChart symbol={article.symbol} days={90} />
      </div>

      {/* ── Article content ── */}
      <div className="border-3 border-ink bg-white shadow-brutal p-6 md:p-8 mb-8">
        <MarkdownContent content={article.content} />
      </div>

      {/* ── Disclaimer ── */}
      <div className="border-3 border-ink bg-yellow shadow-brutal-sm p-4 mb-8">
        <p className="text-xs font-black text-ink uppercase tracking-wide mb-1">
          ⚠ Lưu ý quan trọng
        </p>
        <p className="text-xs text-ink/70 leading-relaxed">
          Bài viết do AI tổng hợp, chỉ mang tính tham khảo — không phải lời khuyên đầu tư.
          Thị trường luôn có rủi ro. Quyết định cuối cùng là của bạn.
        </p>
      </div>

      {/* ── Source links ── */}
      {sourceUrls.length > 0 && (
        <div className="border-3 border-ink bg-cream shadow-brutal-sm p-4 mb-8">
          <p className="text-xs font-black text-ink uppercase tracking-widest mb-3">
            Nguồn tham khảo
          </p>
          <ul className="space-y-1">
            {sourceUrls.map((url, i) => (
              <li key={i}>
                <a
                  href={url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-xs font-semibold text-ink underline underline-offset-2 decoration-yellow hover:decoration-ink break-all"
                >
                  {url}
                </a>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
