/**
 * VN Stock Analysis API Client
 *
 * Typed client for the Go API Gateway.
 */

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export interface TechnicalAnalysis {
  symbol: string;
  timestamp: string;
  price: {
    open: number;
    high: number;
    low: number;
    close: number;
    volume: number;
    change_percent: number;
  };
  rsi: number;
  macd: {
    macd: number;
    signal: number;
    histogram: number;
  };
  bollinger: {
    upper: number;
    middle: number;
    lower: number;
  };
  sma_20: number;
  sma_50: number;
  atr: number;
  signal: string;
  confidence: number;
  score: number;
  reasons: string[];
}

export interface ForecastResult {
  symbol: string;
  timestamp: string;
  technical_score: number;
  sentiment_score: number;
  market_score: number;
  combined_score: number;
  recommendation: string;
  confidence: number;
  support_price?: number;
  resistance_price?: number;
  reasoning: string[];
  anomaly_detected?: boolean;
  weights_used: {
    technical: number;
    sentiment: number;
    market: number;
  };
}

export interface DailyReport {
  date: string;
  symbols: string[];
  analyses: ForecastResult[];
  summary: string;
}

export interface SentimentResult {
  id: string;
  sentiment: string;
  confidence: number;
  source_weight: number;
  is_duplicate: boolean;
  symbols: string[];
  keywords: string[];
}

export interface RumorResult {
  rumors: Array<{
    symbol: string;
    social_mentions: number;
    official_mentions: number;
    risk_level: string;
    unique_sources: number;
    warning: string;
  }>;
  verified_symbols: string[];
  stats: {
    total_items: number;
    rumors_detected: number;
  };
}

async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!res.ok) {
    throw new Error(`API Error: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

// ---- Technical Analysis ----

export async function getTechnicalAnalysis(symbol: string): Promise<TechnicalAnalysis> {
  return fetchAPI<TechnicalAnalysis>(`/technical/${symbol}`);
}

export async function batchTechnicalAnalysis(symbols: string[]): Promise<TechnicalAnalysis[]> {
  return fetchAPI<TechnicalAnalysis[]>("/technical/batch", {
    method: "POST",
    body: JSON.stringify({ symbols }),
  });
}

// ---- Forecast ----

export async function getForecast(symbol: string): Promise<ForecastResult> {
  return fetchAPI<ForecastResult>(`/forecast/${symbol}`);
}

// ---- Reports ----

export async function getDailyReport(): Promise<DailyReport> {
  return fetchAPI<DailyReport>("/reports/daily");
}

export async function generateDailyReport(): Promise<DailyReport> {
  return fetchAPI<DailyReport>("/reports/generate", { method: "POST" });
}

// ---- Sentiment (proxied through Go gateway) ----

export async function analyzeSentiment(
  texts: Array<{ id: string; content: string; source?: string }>
): Promise<{ results: SentimentResult[]; processing_time_ms: number }> {
  return fetchAPI("/sentiment", {
    method: "POST",
    body: JSON.stringify({ texts }),
  });
}

// ---- Queue (async analysis) ----

export async function enqueueAnalysis(symbol: string): Promise<{ job_id: string }> {
  return fetchAPI("/analysis/queue", {
    method: "POST",
    body: JSON.stringify({ symbol }),
  });
}

export async function getAnalysisStatus(
  jobId: string
): Promise<{ status: string; result?: ForecastResult }> {
  return fetchAPI(`/analysis/status/${jobId}`);
}
