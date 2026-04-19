import { Suspense } from "react";
import Link from "next/link";
import { getLatestArticles } from "@/lib/articles-api";
import { ArticleCard } from "@/components/article-card";
import { TrendingStocks } from "@/components/trending-stocks";
import { MarketIndexBar } from "@/components/market-index-bar";
import { MarketBoardTable } from "@/components/market-board-table";

export const revalidate = 60;

export default async function HomePage() {
  const articles = await getLatestArticles(4);

  return (
    <div className="space-y-8">

      {/* ─── Index bar (client, polls every 60s) ─── */}
      <MarketIndexBar />

      {/* ─── Market price board ─── */}
      <section>
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-xl font-black uppercase tracking-tight text-ink border-b-3 border-yellow pb-1">
            Bảng giá thị trường
          </h1>
          <div className="flex items-center gap-2">
            <Link
              href="/market"
              className="btn-brutal bg-ink text-yellow text-xs font-black px-3 py-1.5 uppercase tracking-widest"
            >
              Biểu đồ →
            </Link>
            <Link
              href="/screener"
              className="btn-brutal bg-yellow text-ink text-xs font-black px-3 py-1.5 uppercase tracking-widest"
            >
              AI Screener →
            </Link>
          </div>
        </div>
        <MarketBoardTable />
      </section>

      {/* ─── Trending + Latest articles ─── */}
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_260px] gap-8 items-start">

        {/* Latest articles */}
        <section>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-black uppercase tracking-tight text-ink border-b-3 border-yellow pb-1">
              Phân tích mới nhất
            </h2>
            <Link
              href="/articles"
              className="btn-brutal bg-ink text-yellow text-xs font-black px-3 py-1.5 uppercase tracking-widest"
            >
              Xem tất cả →
            </Link>
          </div>

          {articles.length === 0 ? (
            <div className="border-3 border-ink bg-white py-12 text-center shadow-brutal">
              <p className="font-black text-ink/30 text-lg uppercase tracking-widest mb-1">
                Chưa có bài nào.
              </p>
              <p className="text-ink/40 text-sm">AI đang xử lý — quay lại sau.</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
              {articles.map((a) => (
                <ArticleCard key={a.id} article={a} />
              ))}
            </div>
          )}
        </section>

        {/* Trending sidebar */}
        <aside className="space-y-4">
          <Suspense fallback={null}>
            <TrendingStocks days={7} limit={8} />
          </Suspense>

          {/* Quick links */}
          <div className="border-3 border-ink bg-ink p-4 shadow-brutal">
            <p className="text-yellow font-black text-xs uppercase tracking-widest mb-3">Công cụ phân tích</p>
            <div className="flex flex-col gap-2">
              {[
                { href: "/market",   label: "📊 Biểu đồ chỉ số" },
                { href: "/screener", label: "🤖 AI Screener" },
                { href: "/articles", label: "📰 Bài phân tích" },
              ].map(({ href, label }) => (
                <Link
                  key={href}
                  href={href}
                  className="text-white/70 hover:text-yellow text-sm font-semibold transition-colors py-0.5"
                >
                  {label}
                </Link>
              ))}
            </div>
          </div>
        </aside>
      </div>

    </div>
  );
}
