"use client";

import { useEffect, useState } from "react";
import { Article, fetchDraftArticles } from "@/lib/api";
import { ArticleReviewCard } from "./article-review-card";

export default function AdminArticlesPage() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDraftArticles()
      .then(setArticles)
      .finally(() => setLoading(false));
  }, []);

  const handleUpdate = (id: number) => {
    setArticles((prev) => prev.filter((a) => a.id !== id));
  };

  if (loading) {
    return <div className="p-8 text-gray-500">Đang tải...</div>;
  }

  return (
    <div className="max-w-4xl mx-auto p-8">
      <h1 className="text-2xl font-bold mb-2">Duyệt bài viết</h1>
      <p className="text-sm text-gray-500 mb-6">
        {articles.length} bài đang chờ duyệt
      </p>
      {articles.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <p className="text-4xl mb-3">✓</p>
          <p>Không có bài viết nào cần duyệt.</p>
        </div>
      ) : (
        articles.map((a) => (
          <ArticleReviewCard key={a.id} article={a} onUpdate={handleUpdate} />
        ))
      )}
    </div>
  );
}
