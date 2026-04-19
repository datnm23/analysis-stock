"use client";

import Image from "next/image";
import Link from "next/link";
import { Article } from "@/lib/articles-api";

const REC_BADGE: Record<string, string> = {
  BUY: "badge-buy",
  SELL: "badge-sell",
  HOLD: "badge-hold",
};

const REC_LABEL: Record<string, string> = {
  BUY: "MUA",
  SELL: "BÁN",
  HOLD: "GIỮ",
};

function getRecommendation(forecastData?: string): string | null {
  if (!forecastData) return null;
  try {
    return JSON.parse(forecastData).recommendation ?? null;
  } catch {
    return null;
  }
}

export function ArticleCard({ article }: { article: Article }) {
  const date = new Date(article.published_at ?? article.created_at).toLocaleDateString("vi-VN", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });

  const rec = getRecommendation(article.forecast_data);

  return (
    <Link href={`/articles/${article.slug}`} className="block group">
      <div className="card-brutal flex flex-col h-full">
        {/* ── Thumbnail ── */}
        {article.image_url ? (
          <div className="relative w-full h-44 flex-shrink-0 border-b-3 border-ink overflow-hidden">
            <Image
              src={article.image_url}
              alt={article.title}
              fill
              className="object-cover group-hover:scale-105 transition-transform duration-300"
              sizes="(max-width: 768px) 100vw, 50vw"
            />
          </div>
        ) : (
          <div className="w-full h-44 flex-shrink-0 border-b-3 border-ink bg-ink flex items-center justify-center">
            <span className="text-4xl font-black text-yellow/30 tracking-widest uppercase">
              {article.symbol}
            </span>
          </div>
        )}

        {/* ── Body ── */}
        <div className="p-5 flex flex-col flex-1 gap-3">
          {/* Badges row */}
          <div className="flex items-center gap-2 flex-wrap">
            <Link href={`/symbols/${article.symbol}`} onClick={(e) => e.stopPropagation()} className="badge-symbol hover:opacity-70 transition-opacity">{article.symbol}</Link>
            {rec && (
              <span className={REC_BADGE[rec] ?? "badge-hold"}>
                {REC_LABEL[rec] ?? rec}
              </span>
            )}
          </div>

          {/* Title */}
          <h2 className="text-sm font-bold text-ink leading-snug line-clamp-3 flex-1 group-hover:underline underline-offset-2 decoration-2 decoration-yellow">
            {article.title}
          </h2>

          {/* Summary */}
          {article.summary && article.summary !== "Tóm tắt" && (
            <p className="text-xs text-ink/60 line-clamp-2 leading-relaxed">
              {article.summary}
            </p>
          )}

          {/* Date */}
          <p className="text-xs font-semibold text-ink/40 uppercase tracking-wide mt-auto pt-2 border-t-2 border-ink/10">
            {date}
          </p>
        </div>
      </div>
    </Link>
  );
}
