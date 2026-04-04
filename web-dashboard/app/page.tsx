'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { getDailyReport, type DailyReport } from '@/lib/api';

const WATCHLIST = ['VNM', 'FPT', 'VIC', 'HPG', 'VHM', 'VCB', 'TCB', 'MBB', 'MWG', 'SSI'];

function signalLabel(rec: string | undefined): string {
  if (!rec) return 'GIỮ';
  if (rec.includes('STRONG_BUY')) return 'MUA MẠNH';
  if (rec.includes('BUY')) return 'MUA';
  if (rec.includes('STRONG_SELL')) return 'BÁN MẠNH';
  if (rec.includes('SELL')) return 'BÁN';
  return 'GIỮ';
}

export default function DashboardPage() {
  const router = useRouter();
  const [report, setReport] = useState<DailyReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const fetchData = async () => {
    try {
      const data = await getDailyReport();
      setReport(data);
      setLastUpdated(new Date());
    } catch (err) {
      console.error('Failed to fetch overview data via API', err);
    } finally {
      if (loading) setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 30000); // 30s polling
    return () => clearInterval(interval);
  }, []);

  const buy_signals = report?.analyses?.filter((a) => a.recommendation?.includes('BUY')).length || 0;
  const sell_signals = report?.analyses?.filter((a) => a.recommendation?.includes('SELL')).length || 0;
  const hold_signals = report?.analyses?.filter((a) => a.recommendation?.includes('HOLD') || !a.recommendation).length || 0;
  const total_symbols_analyzed = report?.analyses?.length || (report as any)?.total_symbols_analyzed || 0;

  // Use the original object.entries structure if the API returns results instead of analyses
  let analysisMap = new Map();
  if (report?.analyses) {
    for (const a of report.analyses) {
      analysisMap.set(a.symbol, a);
    }
  } else if ((report as any)?.results) {
    const results = (report as any).results;
    for (const key of Object.keys(results)) {
      analysisMap.set(key, results[key]);
    }
  }

  // Display top watchlist and any other analyzed stocks
  const displaySymbols = [...new Set([...WATCHLIST, ...Array.from(analysisMap.keys())])].slice(0, 15);

  return (
    <div className="fade-in">
      <div className="page-header" style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap' }}>
        <div>
          <h2>📊 Tổng Quan Thị Trường</h2>
          <p>Phân tích tự động chứng khoán HSX / HNX / UPCOM</p>
        </div>
        
        {lastUpdated ? (
          <div className="live-indicator">
            <div className="live-dot" />
            Cập nhật: {lastUpdated.toLocaleTimeString('vi-VN')}
          </div>
        ) : (
          <div className="live-indicator" style={{ opacity: 0.5 }}>
            <span style={{ fontSize: 12 }}>Đang kết nối API...</span>
          </div>
        )}
      </div>

      {loading && !report ? (
        <div className="loading">Đang tải dữ liệu...</div>
      ) : (
        <>
          {/* Stats Overview */}
          <div className="stats-grid">
            <div className="stat-card">
              <div className="stat-label">Tổng mã phân tích</div>
              <div className="stat-value">{total_symbols_analyzed || '—'}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">🟢 Tín hiệu Mua</div>
              <div className="stat-value green">{buy_signals || ((report as any)?.buy_signals) || '—'}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">🔴 Tín hiệu Bán</div>
              <div className="stat-value red">{sell_signals || ((report as any)?.sell_signals) || '—'}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">⚪ Giữ</div>
              <div className="stat-value yellow">{hold_signals || ((report as any)?.hold_signals) || '—'}</div>
            </div>
          </div>

          {/* Watchlist Table */}
          <div className="card" style={{ marginBottom: 32 }}>
            <div className="card-header">
              <h3 className="card-title">Danh sách theo dõi</h3>
            </div>
            <div className="table-wrapper">
              <table>
                <thead>
                  <tr>
                    <th>Mã CK</th>
                    <th>Điểm KT / KQ</th>
                    <th>Tín hiệu</th>
                    <th>Điểm Tổng</th>
                    <th className="hidden-mobile">Độ tin cậy</th>
                  </tr>
                </thead>
                <tbody>
                  {displaySymbols.map((symbol) => {
                    const data = analysisMap.get(symbol);
                    if (!data) {
                      return (
                        <tr key={symbol}>
                          <td><strong>{symbol}</strong></td>
                          <td>—</td>
                          <td><span className="signal-badge hold">Chưa có</span></td>
                          <td>—</td>
                          <td className="hidden-mobile">—</td>
                        </tr>
                      );
                    }
                    return (
                      <tr key={symbol} onClick={() => router.push(`/stock/${symbol}`)}>
                        <td><strong style={{ fontSize: 16 }}>{symbol}</strong></td>
                        <td>
                          <div style={{ fontSize: 13, color: 'var(--text-secondary)' }}>
                            KT: {data.technical_score?.toFixed(0) ?? 0} | 
                            TT: {data.sentiment_score?.toFixed(0) ?? 0}
                          </div>
                        </td>
                        <td>
                          <span
                            className={`signal-badge ${
                              data.recommendation?.includes('BUY') ? 'buy' :
                              data.recommendation?.includes('SELL') ? 'sell' : 'hold'
                            }`}
                          >
                            {signalLabel(data.recommendation)}
                          </span>
                        </td>
                        <td>
                          <strong style={{ fontSize: 16 }}>{data.combined_score?.toFixed(1) ?? '—'}</strong>
                        </td>
                        <td className="hidden-mobile">{data.confidence?.toFixed(0) ?? '—'}%</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>

          {/* Market Summary */}
          {((report as any)?.market_summary || report?.summary) && (
            <div className="card fade-in">
              <div className="card-header">
                <h3 className="card-title">📝 Nhận Xét Thị Trường</h3>
              </div>
              <p style={{ color: 'var(--text-secondary)', lineHeight: 1.8 }}>
                {(report as any)?.market_summary || report?.summary}
              </p>
            </div>
          )}
        </>
      )}
    </div>
  );
}
