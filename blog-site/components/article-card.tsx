import Link from "next/link";
import { Article } from "@/lib/articles-api";

export function ArticleCard({ article }: { article: Article }) {
  const date = article.published_at
    ? new Date(article.published_at).toLocaleDateString("vi-VN")
    : "";

  return (
    <Link href={`/articles/${article.slug}`} className="block group">
      <div className="bg-white rounded-lg border p-5 hover:shadow-md transition-shadow h-full">
        <span className="text-xs font-bold text-blue-600 bg-blue-50 px-2 py-0.5 rounded">
          {article.symbol}
        </span>
        <h2 className="text-base font-semibold mt-2 group-hover:text-blue-700 line-clamp-2">
          {article.title}
        </h2>
        <p className="text-sm text-gray-500 mt-1 line-clamp-2">{article.summary}</p>
        <p className="text-xs text-gray-400 mt-3">{date}</p>
      </div>
    </Link>
  );
}
