import Link from "next/link";
import { getLatestArticles } from "@/lib/articles-api";
import { ArticleCard } from "@/components/article-card";

export const revalidate = 3600;

export default async function HomePage() {
  const articles = await getLatestArticles(6);

  return (
    <div>
      <h1 className="text-2xl font-bold mb-2">Phân tích chứng khoán mới nhất</h1>
      <p className="text-sm text-gray-500 mb-6">
        Bài viết tổng hợp từ tin tức và phân tích kỹ thuật tự động
      </p>
      {articles.length === 0 ? (
        <p className="text-gray-400 py-12 text-center">Chưa có bài viết nào.</p>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {articles.map((a) => (
            <ArticleCard key={a.id} article={a} />
          ))}
        </div>
      )}
      <div className="mt-8 text-center">
        <Link href="/articles" className="text-blue-600 hover:underline text-sm">
          Xem tất cả bài viết →
        </Link>
      </div>
    </div>
  );
}
