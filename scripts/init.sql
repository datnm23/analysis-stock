-- Initialize database for vnstock hybrid architecture
-- Create additional database for n8n
CREATE DATABASE n8n;
CREATE USER n8n WITH ENCRYPTED PASSWORD 'n8n_password';
GRANT ALL PRIVILEGES ON DATABASE n8n TO n8n;

-- Create tables for vnstock
\c vnstock;

-- Stock symbols
CREATE TABLE IF NOT EXISTS stocks (
    symbol VARCHAR(10) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    exchange VARCHAR(10) NOT NULL CHECK (exchange IN ('HSX', 'HNX', 'UPCOM')),
    industry VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Technical analysis results
CREATE TABLE IF NOT EXISTS technical_analysis (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,

    -- Price data
    open_price DECIMAL(12, 2),
    high_price DECIMAL(12, 2),
    low_price DECIMAL(12, 2),
    close_price DECIMAL(12, 2),
    volume BIGINT,

    -- Indicators
    rsi_14 DECIMAL(5, 2),
    macd_line DECIMAL(10, 4),
    macd_signal DECIMAL(10, 4),
    macd_histogram DECIMAL(10, 4),
    bb_upper DECIMAL(12, 2),
    bb_middle DECIMAL(12, 2),
    bb_lower DECIMAL(12, 2),
    sma_20 DECIMAL(12, 2),
    sma_50 DECIMAL(12, 2),
    ema_12 DECIMAL(12, 2),
    ema_26 DECIMAL(12, 2),
    adx DECIMAL(5, 2),
    atr DECIMAL(12, 2),
    stoch_k DECIMAL(5, 2),
    stoch_d DECIMAL(5, 2),

    -- Signal
    signal VARCHAR(15) CHECK (signal IN ('STRONG_BUY', 'BUY', 'HOLD', 'SELL', 'STRONG_SELL')),
    confidence DECIMAL(5, 2),
    score DECIMAL(5, 2),

    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(symbol, timestamp)
);

-- Sentiment analysis results
CREATE TABLE IF NOT EXISTS sentiment_analysis (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10),
    source_url TEXT,
    source_type VARCHAR(50),

    -- Sentiment data
    text_content TEXT,
    sentiment VARCHAR(20) CHECK (sentiment IN ('positive', 'negative', 'neutral')),
    confidence DECIMAL(5, 2),
    keywords TEXT[],

    -- Metadata
    published_at TIMESTAMPTZ,
    analyzed_at TIMESTAMPTZ DEFAULT NOW(),
    model_version VARCHAR(50) DEFAULT 'phobert-base'
);

-- Forecast results
CREATE TABLE IF NOT EXISTS forecasts (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,

    -- Component scores
    technical_score DECIMAL(5, 2),
    sentiment_score DECIMAL(5, 2),
    market_score DECIMAL(5, 2),

    -- Final recommendation
    recommendation VARCHAR(20) CHECK (recommendation IN
        ('STRONG_BUY', 'BUY', 'HOLD', 'SELL', 'STRONG_SELL')),
    confidence DECIMAL(5, 2),

    -- Price targets
    support_price DECIMAL(12, 2),
    resistance_price DECIMAL(12, 2),

    reasoning TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Daily reports
CREATE TABLE IF NOT EXISTS daily_reports (
    id BIGSERIAL PRIMARY KEY,
    report_date DATE UNIQUE NOT NULL,

    -- Summary
    total_symbols_analyzed INT,
    buy_signals INT,
    sell_signals INT,
    hold_signals INT,

    -- Top picks
    top_picks JSONB,
    market_summary TEXT,

    -- Report content
    report_json JSONB,
    report_url TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_technical_symbol_time ON technical_analysis(symbol, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_sentiment_symbol ON sentiment_analysis(symbol);
CREATE INDEX IF NOT EXISTS idx_sentiment_analyzed ON sentiment_analysis(analyzed_at DESC);
CREATE INDEX IF NOT EXISTS idx_forecast_symbol_time ON forecasts(symbol, timestamp DESC);

-- Insert some sample Vietnamese stocks
INSERT INTO stocks (symbol, name, exchange, industry) VALUES
('VNM', 'Công ty CP Sữa Việt Nam', 'HSX', 'Thực phẩm'),
('FPT', 'Công ty CP FPT', 'HSX', 'Công nghệ'),
('VIC', 'Tập đoàn Vingroup', 'HSX', 'Bất động sản'),
('HPG', 'Tập đoàn Hòa Phát', 'HSX', 'Thép'),
('VHM', 'Vinhomes', 'HSX', 'Bất động sản'),
('VCB', 'Vietcombank', 'HSX', 'Ngân hàng'),
('BID', 'BIDV', 'HSX', 'Ngân hàng'),
('TCB', 'Techcombank', 'HSX', 'Ngân hàng'),
('MBB', 'MB Bank', 'HSX', 'Ngân hàng'),
('VPB', 'VPBank', 'HSX', 'Ngân hàng'),
('MWG', 'Thế Giới Di Động', 'HSX', 'Bán lẻ'),
('VNR', 'Tổng Công ty Cổ phần Tái bảo hiểm Quốc gia', 'HSX', 'Bảo hiểm'),
('SSI', 'SSI Securities', 'HSX', 'Chứng khoán'),
('VND', 'VNDirect Securities', 'HSX', 'Chứng khoán'),
('PNJ', 'Phú Nhuận Jewelry', 'HSX', 'Bán lẻ')
ON CONFLICT (symbol) DO NOTHING;
