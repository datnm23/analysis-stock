import { getArticles } from "@/lib/articles-api";
import { ArticleCard } from "@/components/article-card";

export const revalidate = 60;

const REC_FILTERS = [
  { value: "", label: "TẤT CẢ" },
  { value: "BUY", label: "MUA" },
  { value: "HOLD", label: "GIỮ" },
  { value: "SELL", label: "BÁN" },
];

function buildHref(params: { page?: number; q?: string; rec?: string }) {
  const p = new URLSearchParams();
  if (params.page && params.page > 1) p.set("page", String(params.page));
  if (params.q) p.set("q", params.q);
  if (params.rec) p.set("rec", params.rec);
  const s = p.toString();
  return `/articles${s ? `?${s}` : ""}`;
}

export default async function ArticlesPage({
  searchParams,
}: {
  searchParams: { page?: string; q?: string; rec?: string };
}) {
  const page = Math.max(1, parseInt(searchParams.page ?? "1"));
  const q = searchParams.q?.trim() ?? "";
  const rec = searchParams.rec ?? "";
  const limit = 12;
  const offset = (page - 1) * limit;

  const { articles, total } = await getArticles(limit, offset, {
    q: q || undefined,
    recommendation: rec || undefined,
  });
  const totalPages = Math.ceil(total / limit);

  const hasFilter = q || rec;

  return (
    <div>
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-end gap-4 mb-1">
          <h1 className="text-4xl font-black text-ink uppercase tracking-tight">Kho phân tích</h1>
          <span className="bg-yellow border-2 border-ink text-ink font-black text-sm px-3 py-1 mb-1">
            {total} bài
          </span>
        </div>
        <p className="text-ink/50 text-sm mt-2 mb-1">
          Phân tích kỹ thuật · Sentiment · Dự báo AI — mỗi ngày giao dịch.
        </p>
        <div className="h-1 w-24 bg-ink" />
      </div>

      {/* Search + filter */}
      <form method="GET" action="/articles" className="mb-6 flex flex-col sm:flex-row gap-3">
        <input
          type="search"
          name="q"
          defaultValue={q}
          placeholder="Tìm mã hoặc tiêu đề..."
          className="flex-1 px-4 py-2 border-3 border-ink font-mono text-sm bg-white text-ink placeholder:text-ink/40 outline-none focus:ring-0"
        />
        <button
          type="submit"
          className="btn-brutal bg-ink text-yellow font-black text-sm px-5 py-2 uppercase tracking-wide border-3 border-ink"
        >
          Tìm →
        </button>
      </form>

      {/* Recommendation filter pills */}
      <div className="flex flex-wrap gap-2 mb-6">
        {REC_FILTERS.map((f) => (
          <a
            key={f.value || "all"}
            href={buildHref({ q: q || undefined, rec: f.value || undefined })}
            className={`font-black text-sm px-3 py-1.5 border-3 border-ink transition-none
              ${rec === f.value ? "bg-ink text-yellow" : "bg-white text-ink hover:bg-yellow"}`}
          >
            {f.label}
          </a>
        ))}
      </div>

      {/* Active filter label */}
      {hasFilter && (
        <p className="text-sm font-bold text-ink/60 mb-4">
          {q && `Kết quả cho "${q}"`}{q && rec && " · "}{rec && `Lọc: ${rec}`}
          {" — "}{total} bài
          {" · "}
          <a href="/articles" className="underline decoration-yellow hover:decoration-ink">
            Xóa bộ lọc
          </a>
        </p>
      )}

      {articles.length === 0 ? (
        <div className="border-3 border-ink bg-white py-20 text-center shadow-brutal">
          <p className="font-black text-ink/30 text-xl uppercase tracking-widest mb-2">
            Không tìm thấy bài nào.
          </p>
          <p className="text-ink/30 text-sm">
            {hasFilter ? "Thử tìm kiếm khác." : "AI đang quét dữ liệu — quay lại sau."}
          </p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
            {articles.map((a) => <ArticleCard key={a.id} article={a} />)}
          </div>

          {totalPages > 1 && (
            <div className="flex justify-center items-center gap-2 mt-12">
              {page > 1 && (
                <a href={buildHref({ page: page - 1, q: q || undefined, rec: rec || undefined })}
                  className="btn-brutal bg-white text-ink font-black text-sm px-4 py-2 uppercase tracking-wide">
                  ← Trước
                </a>
              )}
              {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                <a key={p} href={buildHref({ page: p, q: q || undefined, rec: rec || undefined })}
                  className={`font-black text-sm px-4 py-2 border-3 border-ink transition-none ${
                    p === page ? "bg-ink text-yellow" : "bg-white text-ink hover:bg-yellow"
                  }`}>
                  {p}
                </a>
              ))}
              {page < totalPages && (
                <a href={buildHref({ page: page + 1, q: q || undefined, rec: rec || undefined })}
                  className="btn-brutal bg-ink text-yellow font-black text-sm px-4 py-2 uppercase tracking-wide">
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
