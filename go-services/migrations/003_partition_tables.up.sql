-- 003_partition_tables.up.sql
-- Convert technical_analysis and sentiment_analysis to RANGE-partitioned tables
-- Partitioned by month for 10x query performance on data > 1 year
--
-- Strategy:
--   1. Rename existing tables to *_old (preserves data during migration)
--   2. Create new partitioned table with identical schema
--   3. Copy existing data into the partitioned table
--   4. Drop the old tables
--
-- After applying:
--   - Monthly partitions exist from 2024-01 through 2027-12
--   - A DEFAULT partition catches any rows outside those bounds

-- ===========================================================================
-- technical_analysis
-- ===========================================================================

ALTER TABLE technical_analysis RENAME TO technical_analysis_old;
DROP INDEX IF EXISTS idx_technical_symbol_time;

CREATE TABLE technical_analysis (
    id          BIGSERIAL,
    symbol      VARCHAR(10) NOT NULL,
    timestamp   TIMESTAMPTZ NOT NULL,

    -- Price data
    open_price  DECIMAL(12, 2),
    high_price  DECIMAL(12, 2),
    low_price   DECIMAL(12, 2),
    close_price DECIMAL(12, 2),
    volume      BIGINT,

    -- Indicators
    rsi_14          DECIMAL(5, 2),
    macd_line       DECIMAL(10, 4),
    macd_signal     DECIMAL(10, 4),
    macd_histogram  DECIMAL(10, 4),
    bb_upper        DECIMAL(12, 2),
    bb_middle       DECIMAL(12, 2),
    bb_lower        DECIMAL(12, 2),
    sma_20          DECIMAL(12, 2),
    sma_50          DECIMAL(12, 2),
    ema_12          DECIMAL(12, 2),
    ema_26          DECIMAL(12, 2),
    adx             DECIMAL(5, 2),
    atr             DECIMAL(12, 2),
    stoch_k         DECIMAL(5, 2),
    stoch_d         DECIMAL(5, 2),

    -- Signal
    signal      VARCHAR(15) CHECK (signal IN ('STRONG_BUY', 'BUY', 'HOLD', 'SELL', 'STRONG_SELL')),
    confidence  DECIMAL(5, 2),
    score       DECIMAL(5, 2),

    created_at  TIMESTAMPTZ DEFAULT NOW(),

    -- UNIQUE must include partition key (timestamp) — PostgreSQL requirement
    UNIQUE (symbol, timestamp)
) PARTITION BY RANGE (timestamp);

-- Monthly partitions 2024
CREATE TABLE technical_analysis_2024_01 PARTITION OF technical_analysis FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE technical_analysis_2024_02 PARTITION OF technical_analysis FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
CREATE TABLE technical_analysis_2024_03 PARTITION OF technical_analysis FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');
CREATE TABLE technical_analysis_2024_04 PARTITION OF technical_analysis FOR VALUES FROM ('2024-04-01') TO ('2024-05-01');
CREATE TABLE technical_analysis_2024_05 PARTITION OF technical_analysis FOR VALUES FROM ('2024-05-01') TO ('2024-06-01');
CREATE TABLE technical_analysis_2024_06 PARTITION OF technical_analysis FOR VALUES FROM ('2024-06-01') TO ('2024-07-01');
CREATE TABLE technical_analysis_2024_07 PARTITION OF technical_analysis FOR VALUES FROM ('2024-07-01') TO ('2024-08-01');
CREATE TABLE technical_analysis_2024_08 PARTITION OF technical_analysis FOR VALUES FROM ('2024-08-01') TO ('2024-09-01');
CREATE TABLE technical_analysis_2024_09 PARTITION OF technical_analysis FOR VALUES FROM ('2024-09-01') TO ('2024-10-01');
CREATE TABLE technical_analysis_2024_10 PARTITION OF technical_analysis FOR VALUES FROM ('2024-10-01') TO ('2024-11-01');
CREATE TABLE technical_analysis_2024_11 PARTITION OF technical_analysis FOR VALUES FROM ('2024-11-01') TO ('2024-12-01');
CREATE TABLE technical_analysis_2024_12 PARTITION OF technical_analysis FOR VALUES FROM ('2024-12-01') TO ('2025-01-01');

-- Monthly partitions 2025
CREATE TABLE technical_analysis_2025_01 PARTITION OF technical_analysis FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE technical_analysis_2025_02 PARTITION OF technical_analysis FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE technical_analysis_2025_03 PARTITION OF technical_analysis FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
CREATE TABLE technical_analysis_2025_04 PARTITION OF technical_analysis FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');
CREATE TABLE technical_analysis_2025_05 PARTITION OF technical_analysis FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');
CREATE TABLE technical_analysis_2025_06 PARTITION OF technical_analysis FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');
CREATE TABLE technical_analysis_2025_07 PARTITION OF technical_analysis FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE technical_analysis_2025_08 PARTITION OF technical_analysis FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE technical_analysis_2025_09 PARTITION OF technical_analysis FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE technical_analysis_2025_10 PARTITION OF technical_analysis FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE technical_analysis_2025_11 PARTITION OF technical_analysis FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE technical_analysis_2025_12 PARTITION OF technical_analysis FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

-- Monthly partitions 2026
CREATE TABLE technical_analysis_2026_01 PARTITION OF technical_analysis FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE technical_analysis_2026_02 PARTITION OF technical_analysis FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE technical_analysis_2026_03 PARTITION OF technical_analysis FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE technical_analysis_2026_04 PARTITION OF technical_analysis FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE technical_analysis_2026_05 PARTITION OF technical_analysis FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE technical_analysis_2026_06 PARTITION OF technical_analysis FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE technical_analysis_2026_07 PARTITION OF technical_analysis FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE technical_analysis_2026_08 PARTITION OF technical_analysis FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE technical_analysis_2026_09 PARTITION OF technical_analysis FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE technical_analysis_2026_10 PARTITION OF technical_analysis FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE technical_analysis_2026_11 PARTITION OF technical_analysis FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE technical_analysis_2026_12 PARTITION OF technical_analysis FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

-- Monthly partitions 2027
CREATE TABLE technical_analysis_2027_01 PARTITION OF technical_analysis FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');
CREATE TABLE technical_analysis_2027_02 PARTITION OF technical_analysis FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');
CREATE TABLE technical_analysis_2027_03 PARTITION OF technical_analysis FOR VALUES FROM ('2027-03-01') TO ('2027-04-01');
CREATE TABLE technical_analysis_2027_04 PARTITION OF technical_analysis FOR VALUES FROM ('2027-04-01') TO ('2027-05-01');
CREATE TABLE technical_analysis_2027_05 PARTITION OF technical_analysis FOR VALUES FROM ('2027-05-01') TO ('2027-06-01');
CREATE TABLE technical_analysis_2027_06 PARTITION OF technical_analysis FOR VALUES FROM ('2027-06-01') TO ('2027-07-01');
CREATE TABLE technical_analysis_2027_07 PARTITION OF technical_analysis FOR VALUES FROM ('2027-07-01') TO ('2027-08-01');
CREATE TABLE technical_analysis_2027_08 PARTITION OF technical_analysis FOR VALUES FROM ('2027-08-01') TO ('2027-09-01');
CREATE TABLE technical_analysis_2027_09 PARTITION OF technical_analysis FOR VALUES FROM ('2027-09-01') TO ('2027-10-01');
CREATE TABLE technical_analysis_2027_10 PARTITION OF technical_analysis FOR VALUES FROM ('2027-10-01') TO ('2027-11-01');
CREATE TABLE technical_analysis_2027_11 PARTITION OF technical_analysis FOR VALUES FROM ('2027-11-01') TO ('2027-12-01');
CREATE TABLE technical_analysis_2027_12 PARTITION OF technical_analysis FOR VALUES FROM ('2027-12-01') TO ('2028-01-01');

-- Catch-all for dates outside defined range
CREATE TABLE technical_analysis_default PARTITION OF technical_analysis DEFAULT;

-- Indexes on partitioned table (automatically propagated to all partitions)
CREATE INDEX idx_technical_symbol_time ON technical_analysis (symbol, timestamp DESC);
CREATE INDEX idx_technical_created_at  ON technical_analysis (created_at DESC);

-- Migrate existing data
INSERT INTO technical_analysis
    SELECT * FROM technical_analysis_old;

DROP TABLE technical_analysis_old;

-- ===========================================================================
-- sentiment_analysis
-- ===========================================================================

ALTER TABLE sentiment_analysis RENAME TO sentiment_analysis_old;
DROP INDEX IF EXISTS idx_sentiment_symbol;
DROP INDEX IF EXISTS idx_sentiment_analyzed;

CREATE TABLE sentiment_analysis (
    id           BIGSERIAL,
    symbol       VARCHAR(10),
    source_url   TEXT,
    source_type  VARCHAR(50),

    -- Sentiment data
    text_content TEXT,
    sentiment    VARCHAR(20) CHECK (sentiment IN ('positive', 'negative', 'neutral')),
    confidence   DECIMAL(5, 2),
    keywords     JSONB,

    -- Metadata
    published_at  TIMESTAMPTZ,
    analyzed_at   TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    model_version VARCHAR(50) DEFAULT 'phobert-base'
) PARTITION BY RANGE (analyzed_at);

-- Monthly partitions 2024
CREATE TABLE sentiment_analysis_2024_01 PARTITION OF sentiment_analysis FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE sentiment_analysis_2024_02 PARTITION OF sentiment_analysis FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
CREATE TABLE sentiment_analysis_2024_03 PARTITION OF sentiment_analysis FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');
CREATE TABLE sentiment_analysis_2024_04 PARTITION OF sentiment_analysis FOR VALUES FROM ('2024-04-01') TO ('2024-05-01');
CREATE TABLE sentiment_analysis_2024_05 PARTITION OF sentiment_analysis FOR VALUES FROM ('2024-05-01') TO ('2024-06-01');
CREATE TABLE sentiment_analysis_2024_06 PARTITION OF sentiment_analysis FOR VALUES FROM ('2024-06-01') TO ('2024-07-01');
CREATE TABLE sentiment_analysis_2024_07 PARTITION OF sentiment_analysis FOR VALUES FROM ('2024-07-01') TO ('2024-08-01');
CREATE TABLE sentiment_analysis_2024_08 PARTITION OF sentiment_analysis FOR VALUES FROM ('2024-08-01') TO ('2024-09-01');
CREATE TABLE sentiment_analysis_2024_09 PARTITION OF sentiment_analysis FOR VALUES FROM ('2024-09-01') TO ('2024-10-01');
CREATE TABLE sentiment_analysis_2024_10 PARTITION OF sentiment_analysis FOR VALUES FROM ('2024-10-01') TO ('2024-11-01');
CREATE TABLE sentiment_analysis_2024_11 PARTITION OF sentiment_analysis FOR VALUES FROM ('2024-11-01') TO ('2024-12-01');
CREATE TABLE sentiment_analysis_2024_12 PARTITION OF sentiment_analysis FOR VALUES FROM ('2024-12-01') TO ('2025-01-01');

-- Monthly partitions 2025
CREATE TABLE sentiment_analysis_2025_01 PARTITION OF sentiment_analysis FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE sentiment_analysis_2025_02 PARTITION OF sentiment_analysis FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE sentiment_analysis_2025_03 PARTITION OF sentiment_analysis FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
CREATE TABLE sentiment_analysis_2025_04 PARTITION OF sentiment_analysis FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');
CREATE TABLE sentiment_analysis_2025_05 PARTITION OF sentiment_analysis FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');
CREATE TABLE sentiment_analysis_2025_06 PARTITION OF sentiment_analysis FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');
CREATE TABLE sentiment_analysis_2025_07 PARTITION OF sentiment_analysis FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE sentiment_analysis_2025_08 PARTITION OF sentiment_analysis FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE sentiment_analysis_2025_09 PARTITION OF sentiment_analysis FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE sentiment_analysis_2025_10 PARTITION OF sentiment_analysis FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE sentiment_analysis_2025_11 PARTITION OF sentiment_analysis FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE sentiment_analysis_2025_12 PARTITION OF sentiment_analysis FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

-- Monthly partitions 2026
CREATE TABLE sentiment_analysis_2026_01 PARTITION OF sentiment_analysis FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE sentiment_analysis_2026_02 PARTITION OF sentiment_analysis FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE sentiment_analysis_2026_03 PARTITION OF sentiment_analysis FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE sentiment_analysis_2026_04 PARTITION OF sentiment_analysis FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE sentiment_analysis_2026_05 PARTITION OF sentiment_analysis FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE sentiment_analysis_2026_06 PARTITION OF sentiment_analysis FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE sentiment_analysis_2026_07 PARTITION OF sentiment_analysis FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE sentiment_analysis_2026_08 PARTITION OF sentiment_analysis FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE sentiment_analysis_2026_09 PARTITION OF sentiment_analysis FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE sentiment_analysis_2026_10 PARTITION OF sentiment_analysis FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE sentiment_analysis_2026_11 PARTITION OF sentiment_analysis FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE sentiment_analysis_2026_12 PARTITION OF sentiment_analysis FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

-- Monthly partitions 2027
CREATE TABLE sentiment_analysis_2027_01 PARTITION OF sentiment_analysis FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');
CREATE TABLE sentiment_analysis_2027_02 PARTITION OF sentiment_analysis FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');
CREATE TABLE sentiment_analysis_2027_03 PARTITION OF sentiment_analysis FOR VALUES FROM ('2027-03-01') TO ('2027-04-01');
CREATE TABLE sentiment_analysis_2027_04 PARTITION OF sentiment_analysis FOR VALUES FROM ('2027-04-01') TO ('2027-05-01');
CREATE TABLE sentiment_analysis_2027_05 PARTITION OF sentiment_analysis FOR VALUES FROM ('2027-05-01') TO ('2027-06-01');
CREATE TABLE sentiment_analysis_2027_06 PARTITION OF sentiment_analysis FOR VALUES FROM ('2027-06-01') TO ('2027-07-01');
CREATE TABLE sentiment_analysis_2027_07 PARTITION OF sentiment_analysis FOR VALUES FROM ('2027-07-01') TO ('2027-08-01');
CREATE TABLE sentiment_analysis_2027_08 PARTITION OF sentiment_analysis FOR VALUES FROM ('2027-08-01') TO ('2027-09-01');
CREATE TABLE sentiment_analysis_2027_09 PARTITION OF sentiment_analysis FOR VALUES FROM ('2027-09-01') TO ('2027-10-01');
CREATE TABLE sentiment_analysis_2027_10 PARTITION OF sentiment_analysis FOR VALUES FROM ('2027-10-01') TO ('2027-11-01');
CREATE TABLE sentiment_analysis_2027_11 PARTITION OF sentiment_analysis FOR VALUES FROM ('2027-11-01') TO ('2027-12-01');
CREATE TABLE sentiment_analysis_2027_12 PARTITION OF sentiment_analysis FOR VALUES FROM ('2027-12-01') TO ('2028-01-01');

-- Default partition
CREATE TABLE sentiment_analysis_default PARTITION OF sentiment_analysis DEFAULT;

-- Indexes
CREATE INDEX idx_sentiment_symbol   ON sentiment_analysis (symbol);
CREATE INDEX idx_sentiment_analyzed ON sentiment_analysis (analyzed_at DESC);

-- Migrate existing data
INSERT INTO sentiment_analysis
    SELECT * FROM sentiment_analysis_old;

DROP TABLE sentiment_analysis_old;
