'use client';

import { useEffect, useState } from 'react';

interface Report {
  date: string;
  total_symbols_analyzed: number;
  report_json: any;
  top_picks: Array<{
    symbol: string;
    recommendation: string;
    combined_score: number;
    reasoning: string[];
  }>;
  market_summary: string;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export default function ReportsPage() {
  const [report, setReport] = useState<Report | null>(null);
  const [loading, setLoading] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function fetchReport() {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/reports/daily`);
      if (!res.ok) throw new Error(`API Error: ${res.status}`);
      setReport(await res.json());
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function generateReport() {
    setGenerating(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/reports/generate`, { method: 'POST' });
      if (!res.ok) throw new Error(`API Error: ${res.status}`);
      setReport(await res.json());
    } catch (err: any) {
      setError(err.message);
    } finally {
      setGenerating(false);
    }
  }

  useEffect(() => { fetchReport(); }, []);

  return (
    <div className="fade-in">
      <div className="page-header">
        <h2>📋 Báo Cáo Hàng Ngày</h2>
        <p>Báo cáo tổng hợp phân tích chứng khoán tự động</p>
      </div>

      {/* Actions */}
      <div style={{ display: 'flex', gap: 12, marginBottom: 20 }}>
        <button
          onClick={fetchReport}
          disabled={loading}
          style={{
            padding: '8px 20px', borderRadius: 8, border: '1px solid var(--border)',
            background: 'var(--bg-secondary)', color: 'var(--text-primary)', cursor: 'pointer',
          }}
        >
          {loading ? '⏳ Đang tải...' : '🔄 Tải báo cáo mới nhất'}
        </button>
        <button
          onClick={generateReport}
          disabled={generating}
          style={{
            padding: '8px 20px', borderRadius: 8, border: 'none',
            background: 'var(--accent)', color: '#fff', fontWeight: 600, cursor: 'pointer',
            opacity: generating ? 0.7 : 1,
          }}
        >
          {generating ? '⏳ Đang tạo...' : '🚀 Tạo báo cáo mới'}
        </button>
      </div>

      {error && (
        <div className="card" style={{ borderLeft: '4px solid var(--red)', marginBottom: 20 }}>
          <p style={{ color: 'var(--red)' }}>⚠️ {error}</p>
        </div>
      )}

      {report ? (
        <>
          {/* Report Header */}
          <div className="card" style={{ marginBottom: 20 }}>
            <div className="card-header">
              <h3 className="card-title">📊 Báo cáo ngày {report.date || 'Hôm nay'}</h3>
            </div>
            <p style={{ color: 'var(--text-secondary)' }}>
              Tổng số mã phân tích: <strong>{report.total_symbols_analyzed}</strong>
            </p>
          </div>

          {/* Top Picks */}
          {report.top_picks && report.top_picks.length > 0 && (
            <div className="card" style={{ marginBottom: 20 }}>
              <div className="card-header">
                <h3 className="card-title">⭐ Top Picks</h3>
              </div>
              <div className="table-wrapper">
                <table>
                  <thead>
                    <tr>
                      <th>Mã CK</th>
                      <th>Khuyến nghị</th>
                      <th>Điểm</th>
                      <th>Lý do</th>
                    </tr>
                  </thead>
                  <tbody>
                    {report.top_picks.map((pick) => (
                      <tr key={pick.symbol}>
                        <td><strong>{pick.symbol}</strong></td>
                        <td>
                          <span className={`signal-badge ${
                            pick.recommendation?.includes('BUY') ? 'buy' :
                            pick.recommendation?.includes('SELL') ? 'sell' : 'hold'
                          }`}>
                            {pick.recommendation}
                          </span>
                        </td>
                        <td>{pick.combined_score?.toFixed(1)}</td>
                        <td style={{ fontSize: 12, color: 'var(--text-secondary)' }}>
                          {pick.reasoning?.slice(0, 2).join('; ') || '—'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Market Summary */}
          {report.market_summary && (
            <div className="card">
              <div className="card-header">
                <h3 className="card-title">📝 Nhận Xét Thị Trường</h3>
              </div>
              <p style={{ color: 'var(--text-secondary)', lineHeight: 1.8 }}>
                {report.market_summary}
              </p>
            </div>
          )}
        </>
      ) : !loading && !error ? (
        <div className="card" style={{ textAlign: 'center', padding: 40 }}>
          <p style={{ color: 'var(--text-secondary)', fontSize: 16 }}>
            📭 Chưa có báo cáo. Nhấn "Tạo báo cáo mới" để bắt đầu.
          </p>
        </div>
      ) : null}
    </div>
  );
}
