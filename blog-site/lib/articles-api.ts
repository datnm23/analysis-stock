const API_URL = process.env.API_URL || "http://localhost:8080";

export interface Article {
  id: number;
  symbol: string;
  title: string;
  slug: string;
  summary: string;
  content: string;
  source_urls: string;
  image_url?: string;
  published_at?: string;
  created_at: string;
  forecast_data?: string;
}

export interface ArticleListResult {
  articles: Article[];
  total: number;
  limit: number;
  offset: number;
}

export async function getArticles(
  limit = 20,
  offset = 0,
  filters?: { q?: string; recommendation?: string }
): Promise<ArticleListResult> {
  try {
    const params = new URLSearchParams({
      status: "published",
      limit: String(limit),
      offset: String(offset),
    });
    if (filters?.q) params.set("q", filters.q);
    if (filters?.recommendation) params.set("recommendation", filters.recommendation);
    const res = await fetch(`${API_URL}/api/v1/articles?${params}`, {
      next: { revalidate: 60 },
    });
    if (!res.ok) return { articles: [], total: 0, limit, offset };
    return res.json();
  } catch {
    return { articles: [], total: 0, limit, offset };
  }
}

export async function getArticleBySlug(slug: string): Promise<Article | null> {
  try {
    const res = await fetch(`${API_URL}/api/v1/articles/${slug}`, {
      next: { revalidate: 60 },
    });
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

export async function getLatestArticles(limit = 6): Promise<Article[]> {
  const { articles } = await getArticles(limit, 0);
  return articles;
}

export async function getArticlesBySymbol(
  symbol: string,
  limit = 20,
  offset = 0
): Promise<ArticleListResult> {
  try {
    const params = new URLSearchParams({
      status: "published",
      symbol: symbol.toUpperCase(),
      limit: String(limit),
      offset: String(offset),
    });
    const res = await fetch(`${API_URL}/api/v1/articles?${params}`, {
      next: { revalidate: 60 },
    });
    if (!res.ok) return { articles: [], total: 0, limit, offset };
    return res.json();
  } catch {
    return { articles: [], total: 0, limit, offset };
  }
}

export async function getRelatedArticles(slug: string, symbol: string, limit = 3): Promise<Article[]> {
  try {
    const res = await fetch(
      `${API_URL}/api/v1/articles/${slug}/related?symbol=${symbol}&limit=${limit}`,
      { next: { revalidate: 60 } }
    );
    if (!res.ok) return [];
    const json = await res.json();
    return json.articles ?? [];
  } catch {
    return [];
  }
}
