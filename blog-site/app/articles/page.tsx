import { getArticles } from "@/lib/articles-api";
import { ArticleCard } from "@/components/article-card";

export const revalidate = 3600;

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
      <h1 className="text-2xl font-bold mb-2">Tất cả bài viết</h1>
      <p className="text-sm text-gray-500 mb-6">{total} bài viết</p>
      {articles.length === 0 ? (
        <p className="text-gray-400 py-12 text-center">Chưa có bài viết nào.</p>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {articles.map((a) => (
              <ArticleCard key={a.id} article={a} />
            ))}
          </div>
          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-8">
              {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                <a
                  key={p}
                  href={`/articles?page=${p}`}
                  className={`px-3 py-1 rounded border text-sm transition-colors ${
                    p === page
                      ? "bg-blue-600 text-white border-blue-600"
                      : "hover:bg-gray-100 border-gray-200"
                  }`}
                >
                  {p}
                </a>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}
