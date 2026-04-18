"use client";

import { useState } from "react";
import { Article, updateArticleStatus } from "@/lib/api";

interface Props {
  article: Article;
  onUpdate: (id: number) => void;
}

export function ArticleReviewCard({ article, onUpdate }: Props) {
  const [expanded, setExpanded] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleAction = async (status: "published" | "rejected") => {
    setLoading(true);
    await updateArticleStatus(article.id, status);
    onUpdate(article.id);
    setLoading(false);
  };

  const date = new Date(article.created_at).toLocaleString("vi-VN");

  return (
    <div className="border rounded-lg p-4 mb-4 bg-white shadow-sm">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <span className="text-xs font-bold text-blue-600 bg-blue-50 px-2 py-0.5 rounded">
            {article.symbol}
          </span>
          <h3 className="text-base font-semibold mt-1 truncate">{article.title}</h3>
          <p className="text-sm text-gray-500 mt-1 line-clamp-2">{article.summary}</p>
          <p className="text-xs text-gray-400 mt-1">{date}</p>
        </div>
        <div className="flex gap-2 shrink-0">
          <button
            onClick={() => setExpanded(!expanded)}
            className="text-sm px-3 py-1 border rounded hover:bg-gray-50 transition-colors"
          >
            {expanded ? "Ẩn" : "Xem"}
          </button>
          <button
            disabled={loading}
            onClick={() => handleAction("published")}
            className="text-sm px-3 py-1 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50 transition-colors"
          >
            Duyệt
          </button>
          <button
            disabled={loading}
            onClick={() => handleAction("rejected")}
            className="text-sm px-3 py-1 bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50 transition-colors"
          >
            Từ chối
          </button>
        </div>
      </div>
      {expanded && (
        <div className="mt-4 p-3 bg-gray-50 rounded text-sm whitespace-pre-wrap font-mono max-h-96 overflow-y-auto">
          {article.content}
        </div>
      )}
    </div>
  );
}
