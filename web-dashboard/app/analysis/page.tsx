'use client';

import { useEffect, useState } from 'react';

interface AnalysisResult {
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
  weights_used: { technical: number; sentiment: number; market: number };
}

const SYMBOLS = ['VNM', 'FPT', 'VIC', 'HPG', 'VHM', 'VCB', 'TCB', 'MBB', 'MWG', 'SSI'];
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

function signalColor(rec: string): string {
  if (rec?.includes('BUY')) return 'var(--green)';
  if (rec?.includes('SELL')) return 'var(--red)';
  return 'var(--yellow)';
}

function ScoreBar({ value, label }: { value: number; label: string }) {
  const color = value >= 60 ? 'var(--green)' : value >= 40 ? 'var(--yellow)' : 'var(--red)';
  return (
    <div style={{ marginBottom: 8 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, marginBottom: 2 }}>
        <span style={{ color: 'var(--text-secondary)' }}>{label}</span>
        <span style={{ fontWeight: 600 }}>{value.toFixed(1)}</span>
      </div>
      <div style={{ background: 'var(--bg-tertiary)', borderRadius: 4, height: 6, overflow: 'hidden' }}>
        <div style={{ width: `${value}%`, background: color, height: '100%', borderRadius: 4, transition: 'width 0.5s' }} />
      </div>
    </div>
  );
}

export default function AnalysisPage() {
  const [symbol, setSymbol] = useState('FPT');
  const [result, setResult] = useState<AnalysisResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function analyze() {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/forecast/${symbol}`);
      if (!res.ok) throw new Error(`API Error: ${res.status}`);
      const data = await res.json();
      setResult(data);
    } catch (err: any) {
      setError(err.message || 'Không thể kết nối API');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="fade-in">
      <div className="page-header">
        <h2>🔍 Phân Tích Mã Chứng Khoán</h2>
        <p>Chọn mã CK để xem phân tích kỹ thuật + sentiment + dự báo</p>
      </div>

      {/* Symbol Selector */}
      <div className="card" style={{ marginBottom: 20 }}>
        <div style={{ display: 'flex', gap: 12, alignItems: 'center', flexWrap: 'wrap' }}>
          <select
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
            style={{
              padding: '8px 16px',
              borderRadius: 8,
              border: '1px solid var(--border)',
              background: 'var(--bg-secondary)',
              color: 'var(--text-primary)',
              fontSize: 14,
            }}
          >
            {SYMBOLS.map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
          <input
            type="text"
            placeholder="Hoặc nhập mã CK..."
            value={symbol}
            onChange={(e) => setSymbol(e.target.value.toUpperCase())}
            style={{
              padding: '8px 16px',
              borderRadius: 8,
              border: '1px solid var(--border)',
              background: 'var(--bg-secondary)',
              color: 'var(--text-primary)',
              fontSize: 14,
              width: 180,
            }}
          />
          <button
            onClick={analyze}
            disabled={loading || !symbol}
            style={{
              padding: '8px 24px',
              borderRadius: 8,
              border: 'none',
              background: 'var(--accent)',
              color: '#fff',
              fontWeight: 600,
              cursor: loading ? 'wait' : 'pointer',
              opacity: loading ? 0.7 : 1,
            }}
          >
            {loading ? '⏳ Đang phân tích...' : '🔍 Phân tích'}
          </button>
        </div>
      </div>

      {error && (
        <div className="card" style={{ borderLeft: '4px solid var(--red)', marginBottom: 20 }}>
          <p style={{ color: 'var(--red)' }}>⚠️ {error}</p>
        </div>
      )}

      {result && (
        <>
          {/* Recommendation Header */}
          <div className="card" style={{ textAlign: 'center', marginBottom: 20 }}>
            <h3 style={{ fontSize: 28, marginBottom: 8 }}>{result.symbol}</h3>
            <div
              style={{
                display: 'inline-block',
                padding: '8px 32px',
                borderRadius: 20,
                background: signalColor(result.recommendation),
                color: '#fff',
                fontWeight: 700,
                fontSize: 20,
              }}
            >
              {result.recommendation}
            </div>
            <p style={{ marginTop: 12, color: 'var(--text-secondary)' }}>
              Điểm tổng: <strong>{result.combined_score.toFixed(1)}</strong> / 100 |
              Độ tin cậy: <strong>{result.confidence.toFixed(1)}%</strong>
            </p>
          </div>

          {/* Score Breakdown */}
          <div className="stats-grid" style={{ marginBottom: 20 }}>
            <div className="stat-card">
              <div className="stat-label">📊 Technical</div>
              <div className="stat-value">{result.technical_score.toFixed(1)}</div>
              <small style={{ color: 'var(--text-secondary)' }}>
                Weight: {(result.weights_used.technical * 100).toFixed(0)}%
              </small>
            </div>
            <div className="stat-card">
              <div className="stat-label">💬 Sentiment</div>
              <div className="stat-value">{result.sentiment_score.toFixed(1)}</div>
              <small style={{ color: 'var(--text-secondary)' }}>
                Weight: {(result.weights_used.sentiment * 100).toFixed(0)}%
              </small>
            </div>
            <div className="stat-card">
              <div className="stat-label">🌐 Market</div>
              <div className="stat-value">{result.market_score.toFixed(1)}</div>
              <small style={{ color: 'var(--text-secondary)' }}>
                Weight: {(result.weights_used.market * 100).toFixed(0)}%
              </small>
            </div>
            <div className="stat-card">
              <div className="stat-label">🎯 Combined</div>
              <div className="stat-value" style={{ color: signalColor(result.recommendation) }}>
                {result.combined_score.toFixed(1)}
              </div>
            </div>
          </div>

          {/* Score Bars */}
          <div className="card" style={{ marginBottom: 20 }}>
            <div className="card-header"><h3 className="card-title">Score Breakdown</h3></div>
            <ScoreBar value={result.technical_score} label="Technical Score" />
            <ScoreBar value={result.sentiment_score} label="Sentiment Score" />
            <ScoreBar value={result.market_score} label="Market Context" />
            <ScoreBar value={result.combined_score} label="Combined Score" />
          </div>

          {/* Support/Resistance */}
          {(result.support_price || result.resistance_price) && (
            <div className="stats-grid" style={{ marginBottom: 20 }}>
              <div className="stat-card">
                <div className="stat-label">📉 Hỗ trợ</div>
                <div className="stat-value green">
                  {result.support_price?.toLocaleString() ?? '—'}
                </div>
              </div>
              <div className="stat-card">
                <div className="stat-label">📈 Kháng cự</div>
                <div className="stat-value red">
                  {result.resistance_price?.toLocaleString() ?? '—'}
                </div>
              </div>
            </div>
          )}

          {/* Reasoning */}
          <div className="card">
            <div className="card-header"><h3 className="card-title">📝 Lý do phân tích</h3></div>
            <ul style={{ color: 'var(--text-secondary)', lineHeight: 1.8, paddingLeft: 20 }}>
              {result.reasoning.map((r, i) => <li key={i}>{r}</li>)}
            </ul>
          </div>
        </>
      )}
    </div>
  );
}
