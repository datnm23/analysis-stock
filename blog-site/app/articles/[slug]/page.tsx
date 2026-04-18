import { notFound } from "next/navigation";
import { getArticleBySlug, getArticles } from "@/lib/articles-api";

export const revalidate = 3600;

// dynamicParams=true allows on-demand generation for slugs not pre-rendered
export const dynamicParams = true;

export async function generateStaticParams() {
  try {
    const { articles } = await getArticles(100, 0);
    return articles.map((a) => ({ slug: a.slug }));
  } catch {
    // API unavailable at build time — pages generated on first request
    return [];
  }
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
        year: "numeric",
        month: "long",
        day: "numeric",
      })
    : "";

  return (
    <article className="max-w-2xl">
      <div className="mb-6">
        <span className="text-sm font-bold text-blue-600 bg-blue-50 px-2 py-0.5 rounded">
          {article.symbol}
        </span>
        <h1 className="text-2xl font-bold mt-3 leading-snug">{article.title}</h1>
        {date && <p className="text-sm text-gray-400 mt-1">{date}</p>}
      </div>
      <div className="prose prose-sm max-w-none text-gray-800 leading-relaxed whitespace-pre-wrap">
        {article.content}
      </div>
    </article>
  );
}
