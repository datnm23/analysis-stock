import { getArticles } from "@/lib/articles-api";
import { ArticleCard } from "@/components/article-card";

export const revalidate = 60;

export default async function ArticlesPage({
  searchParams,
}: {
  searchParams: { page?: string };
}) {
  const page = Math.max(1, parseInt(searchParams.page ?? "1"));
  const limit = 12;
  const offset = (page - 1) * limit;
  const { articles, total } = await getArticles(limit, offset);
  const totalPages = Math.ceil(total / limit);

  return (
    <div>
      {/* ─── Page header ─── */}
      <div className="mb-8">
        <div className="flex items-end gap-4 mb-1">
          <h1 className="text-4xl font-black text-ink uppercase tracking-tight">
            Kho phân tích
          </h1>
          <span className="bg-yellow border-2 border-ink text-ink font-black text-sm px-3 py-1 mb-1">
            {total} bài
          </span>
        </div>
        <p className="text-ink/50 text-sm mt-2 mb-1">
          Phân tích kỹ thuật · Sentiment · Dự báo AI — mỗi ngày giao dịch.
        </p>
        <div className="h-1 w-24 bg-ink" />
      </div>

      {articles.length === 0 ? (
        <div className="border-3 border-ink bg-white py-20 text-center shadow-brutal">
          <p className="font-black text-ink/30 text-xl uppercase tracking-widest mb-2">
            Chưa có bài nào.
          </p>
          <p className="text-ink/30 text-sm">AI đang quét dữ liệu — quay lại sau.</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
            {articles.map((a) => (
              <ArticleCard key={a.id} article={a} />
            ))}
          </div>

          {/* ─── Pagination ─── */}
          {totalPages > 1 && (
            <div className="flex justify-center items-center gap-2 mt-12">
              {page > 1 && (
                <a
                  href={`/articles?page=${page - 1}`}
                  className="btn-brutal bg-white text-ink font-black text-sm px-4 py-2 uppercase tracking-wide"
                >
                  ← Trước
                </a>
              )}
              {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                <a
                  key={p}
                  href={`/articles?page=${p}`}
                  className={`font-black text-sm px-4 py-2 border-3 border-ink transition-none ${
                    p === page
                      ? "bg-ink text-yellow"
                      : "bg-white text-ink hover:bg-yellow"
                  }`}
                >
                  {p}
                </a>
              ))}
              {page < totalPages && (
                <a
                  href={`/articles?page=${page + 1}`}
                  className="btn-brutal bg-ink text-yellow font-black text-sm px-4 py-2 uppercase tracking-wide"
                >
                  Tiếp →
                </a>
              )}
            </div>
          )}
        </>
      )}
    </div>
  );
}
