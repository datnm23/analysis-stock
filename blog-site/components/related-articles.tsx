import Link from "next/link";
import Image from "next/image";
import { getRelatedArticles } from "@/lib/articles-api";

function parseForecastRec(raw?: string): string | null {
  if (!raw) return null;
  try { return JSON.parse(raw).recommendation ?? null; } catch { return null; }
}

const REC_BADGE: Record<string, string> = {
  BUY: "badge-buy", SELL: "badge-sell", HOLD: "badge-hold",
};
const REC_LABEL: Record<string, string> = {
  BUY: "MUA", SELL: "BÁN", HOLD: "GIỮ",
};

export async function RelatedArticles({ slug, symbol }: { slug: string; symbol: string }) {
  const articles = await getRelatedArticles(slug, symbol, 3);
  if (!articles.length) return null;

  return (
    <div className="border-3 border-ink shadow-brutal mb-8">
      <div className="bg-ink px-4 py-2">
        <h3 className="text-yellow font-black text-sm uppercase tracking-widest">
          Phân tích liên quan — {symbol}
        </h3>
      </div>
      <div className="divide-y-3 divide-ink bg-cream">
        {articles.map((a) => {
          const rec = parseForecastRec(a.forecast_data);
          return (
            <Link
              key={a.id}
              href={`/articles/${a.slug}`}
              className="flex gap-3 p-4 hover:bg-yellow/20 transition-colors group"
            >
              {a.image_url && (
                <div className="relative w-16 h-16 flex-shrink-0 border-2 border-ink overflow-hidden">
                  <Image src={a.image_url} alt={a.title} fill className="object-cover" sizes="64px" />
                </div>
              )}
              <div className="flex-1 min-w-0">
                <p className="text-xs font-black text-ink line-clamp-2 group-hover:underline decoration-yellow">
                  {a.title}
                </p>
                {rec && (
                  <span className={`${REC_BADGE[rec] ?? "badge-hold"} text-[10px] mt-1 inline-block`}>
                    {REC_LABEL[rec] ?? rec}
                  </span>
                )}
              </div>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
