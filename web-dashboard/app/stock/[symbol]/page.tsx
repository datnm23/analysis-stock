'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { getTechnicalAnalysis, getForecast, type TechnicalAnalysis, type ForecastResult } from '@/lib/api';

function signalColor(rec: string): string {
  if (rec?.includes('BUY')) return 'var(--green)';
  if (rec?.includes('SELL')) return 'var(--red)';
  return 'var(--yellow)';
}

function Indicator({ label, value, unit }: { label: string; value?: number; unit?: string }) {
  return (
    <div className="stat-card">
      <div className="stat-label">{label}</div>
      <div className="stat-value">{value != null ? value.toFixed(2) : '—'}</div>
      {unit && <small style={{ color: 'var(--text-secondary)' }}>{unit}</small>}
    </div>
  );
}

function ScoreBar({ value, label, maxValue = 100 }: { value: number; label: string; maxValue?: number }) {
  const pct = Math.min((value / maxValue) * 100, 100);
  const color = pct >= 60 ? 'var(--green)' : pct >= 40 ? 'var(--yellow)' : 'var(--red)';
  return (
    <div style={{ marginBottom: 16 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13, marginBottom: 6 }}>
        <span style={{ color: 'var(--text-secondary)' }}>{label}</span>
        <span style={{ fontWeight: 700, color: 'var(--text-primary)' }}>{value.toFixed(1)}</span>
      </div>
      <div style={{ background: 'var(--bg-tertiary)', borderRadius: 6, height: 8, overflow: 'hidden' }}>
        <div style={{ width: `${pct}%`, background: color, height: '100%', borderRadius: 6, transition: 'width 0.8s cubic-bezier(0.4, 0, 0.2, 1)' }} />
      </div>
    </div>
  );
}

export default function StockDetailPage({ params }: { params: { symbol: string } }) {
  const router = useRouter();
  const symbol = params.symbol?.toUpperCase() || '';
  const [technical, setTechnical] = useState<TechnicalAnalysis | null>(null);
  const [forecast, setForecast] = useState<ForecastResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const fetchData = () => {
    if (!symbol) return;
    
    Promise.allSettled([
      getTechnicalAnalysis(symbol),
      getForecast(symbol),
    ]).then(([techRes, forecastRes]) => {
      let isSuccess = false;
      if (techRes.status === 'fulfilled') {
        setTechnical(techRes.value);
        isSuccess = true;
      }
      if (forecastRes.status === 'fulfilled') {
        setForecast(forecastRes.value);
        isSuccess = true;
      }
      if (techRes.status === 'rejected' && forecastRes.status === 'rejected') {
        setError('Không thể kết nối API. Vui lòng kiểm tra hệ thống.');
      } else {
        setError(null);
      }
      
      if (isSuccess) setLastUpdated(new Date());
    }).finally(() => {
      setLoading(false);
    });
  };

  useEffect(() => {
    setLoading(true);
    fetchData();
    const interval = setInterval(fetchData, 30000); // Poll every 30s
    return () => clearInterval(interval);
  }, [symbol]);

  if (!symbol) {
    return (
      <div className="fade-in">
        <div className="card" style={{ textAlign: 'center', padding: 40 }}>
          <p style={{ color: 'var(--text-secondary)' }}>Vui lòng chọn mã chứng khoán.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="fade-in">
      <div style={{ marginBottom: 20 }}>
        <button 
          onClick={() => router.push('/')}
          style={{ 
            background: 'none', 
            border: 'none', 
            color: 'var(--accent)', 
            cursor: 'pointer',
            fontSize: 14,
            fontWeight: 600,
            display: 'flex',
            alignItems: 'center',
            gap: 6
          }}
        >
          <span>←</span> Quay lại Tổng quan
        </button>
      </div>

      <div className="page-header" style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap' }}>
        <div>
          <h2>📈 {symbol}</h2>
          <p>Phân tích chi tiết kỹ thuật và dự báo</p>
        </div>
        
        {lastUpdated ? (
          <div className="live-indicator">
            <div className="live-dot" />
            Cập nhật: {lastUpdated.toLocaleTimeString('vi-VN')}
          </div>
        ) : loading ? (
          <div className="live-indicator" style={{ opacity: 0.5 }}>
            <span style={{ fontSize: 12 }}>Đang kết nối API...</span>
          </div>
        ) : null}
      </div>

      {loading && !forecast && !technical && (
        <div className="card" style={{ textAlign: 'center', padding: 40 }}>
          <div className="loading" style={{ height: 'auto', background: 'transparent' }}>
            <span style={{ fontSize: 16 }}>⏳ Đang tải dữ liệu...</span>
          </div>
        </div>
      )}

      {error && (
        <div className="card" style={{ borderLeft: '4px solid var(--red)', marginBottom: 20 }}>
          <p style={{ color: 'var(--red)' }}>⚠️ {error}</p>
        </div>
      )}

      {/* Forecast Recommendation */}
      {forecast && (
        <div className="card" style={{ textAlign: 'center', marginBottom: 24, position: 'relative', overflow: 'hidden' }}>
          {forecast.anomaly_detected && (
           <div style={{ position: 'absolute', top: 0, right: 0, padding: '4px 20px', background: 'var(--red)', color: '#fff', fontSize: 12, fontWeight: 700, letterSpacing: 1, transform: 'rotate(45deg) translate(30%, -150%)', boxShadow: '0 2px 8px rgba(0,0,0,0.3)' }}>
             ANOMALY
           </div>
          )}
          <div
            style={{
              display: 'inline-block',
              padding: '12px 48px',
              borderRadius: 32,
              background: signalColor(forecast.recommendation) + '20', // Add slight transparency for modern look
              border: `2px solid ${signalColor(forecast.recommendation)}`,
              color: signalColor(forecast.recommendation),
              fontWeight: 800,
              fontSize: 24,
              letterSpacing: 2,
              boxShadow: `0 4px 24px ${signalColor(forecast.recommendation)}40`
            }}
          >
            {forecast.recommendation || 'HOLD'}
          </div>
          <p style={{ marginTop: 16, color: 'var(--text-secondary)', fontSize: 15 }}>
            Điểm tổng: <strong style={{color: 'var(--text-primary)'}}>{forecast.combined_score.toFixed(1)}</strong> / 100 
            <span style={{opacity: 0.5, margin: '0 12px'}}>|</span>
            Độ tin cậy: <strong style={{color: 'var(--text-primary)'}}>{forecast.confidence.toFixed(1)}%</strong>
          </p>
        </div>
      )}

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '24px', marginBottom: '24px' }}>
        {/* Score Breakdown */}
        {forecast && (
          <div className="card">
            <div className="card-header"><h3 className="card-title">📊 Score Breakdown</h3></div>
            <ScoreBar value={forecast.technical_score} label={`Technical (${((forecast.weights_used?.technical || 0.6) * 100).toFixed(0)}%)`} />
            <ScoreBar value={forecast.sentiment_score} label={`Sentiment (${((forecast.weights_used?.sentiment || 0.3) * 100).toFixed(0)}%)`} />
            <ScoreBar value={forecast.market_score} label={`Market (${((forecast.weights_used?.market || 0.1) * 100).toFixed(0)}%)`} />
            <ScoreBar value={forecast.combined_score} label="Combined Score" />
          </div>
        )}

        {/* Support/Resistance & Quick Stats */}
        {forecast && (forecast.support_price || forecast.resistance_price) && (
          <div className="card">
            <div className="card-header"><h3 className="card-title">🎯 Các Mốc Quan Trọng</h3></div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              <div style={{ padding: 16, background: 'var(--green-soft)', borderLeft: '4px solid var(--green)', borderRadius: 8 }}>
                <span style={{ fontSize: 13, color: 'var(--text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Kháng cự mục tiêu</span>
                <div style={{ fontSize: 24, fontWeight: 800, color: 'var(--green)', marginTop: 4 }}>
                  {forecast.resistance_price?.toLocaleString() ?? '—'}
                </div>
              </div>
              <div style={{ padding: 16, background: 'var(--red-soft)', borderLeft: '4px solid var(--red)', borderRadius: 8 }}>
                <span style={{ fontSize: 13, color: 'var(--text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Hỗ trợ rủi ro</span>
                <div style={{ fontSize: 24, fontWeight: 800, color: 'var(--red)', marginTop: 4 }}>
                  {forecast.support_price?.toLocaleString() ?? '—'}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Technical Indicators */}
      {technical && (
        <>
          <div className="card" style={{ marginBottom: 24 }}>
            <div className="card-header"><h3 className="card-title">🔧 Chỉ Báo Kỹ Thuật Chính</h3></div>
            
            {/* Price Info Grid inside card */}
            <div className="stats-grid" style={{ marginBottom: 24, padding: 16, background: 'rgba(0,0,0,0.1)', borderRadius: 12 }}>
              <Indicator label="Giá Mở" value={technical.price?.open} />
              <Indicator label="Giá Cao" value={technical.price?.high} />
              <Indicator label="Giá Thấp" value={technical.price?.low} />
              <Indicator label="Giá Đóng" value={technical.price?.close} />
            </div>

            <div className="stats-grid">
              <Indicator label="RSI (14)" value={technical.rsi} />
              <Indicator label="MACD" value={technical.macd?.macd} />
              <Indicator label="MACD Signal" value={technical.macd?.signal} />
              <Indicator label="SMA 20" value={technical.sma_20} />
              <Indicator label="SMA 50" value={technical.sma_50} />
              <Indicator label="ATR" value={technical.atr} />
            </div>
          </div>
        </>
      )}

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '24px' }}>
        {/* Reasoning */}
        {forecast && forecast.reasoning && forecast.reasoning.length > 0 && (
          <div className="card">
            <div className="card-header"><h3 className="card-title">📝 Tổng hợp Phân Tích</h3></div>
            <ul style={{ color: 'var(--text-secondary)', lineHeight: 1.9, paddingLeft: 20 }}>
              {forecast.reasoning.map((r, i) => <li key={i}>{r}</li>)}
            </ul>
          </div>
        )}

        {/* Technical Reasons */}
        {technical && technical.reasons && technical.reasons.length > 0 && (
          <div className="card">
            <div className="card-header"><h3 className="card-title">🔍 Chi Tiết Kỹ Thuật</h3></div>
            <ul style={{ color: 'var(--text-secondary)', lineHeight: 1.9, paddingLeft: 20 }}>
              {technical.reasons.map((r, i) => <li key={i}>{r}</li>)}
            </ul>
          </div>
        )}
      </div>
    </div>
  );
}
