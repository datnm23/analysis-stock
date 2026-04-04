-- 003_partition_tables.down.sql
-- Revert partitioned tables back to plain tables

-- ===========================================================================
-- technical_analysis
-- ===========================================================================

CREATE TABLE technical_analysis_plain (LIKE technical_analysis INCLUDING ALL);
INSERT INTO technical_analysis_plain SELECT * FROM technical_analysis;
DROP TABLE technical_analysis;
ALTER TABLE technical_analysis_plain RENAME TO technical_analysis;

CREATE INDEX IF NOT EXISTS idx_technical_symbol_time ON technical_analysis (symbol, timestamp DESC);

-- ===========================================================================
-- sentiment_analysis
-- ===========================================================================

CREATE TABLE sentiment_analysis_plain (LIKE sentiment_analysis INCLUDING ALL);
INSERT INTO sentiment_analysis_plain SELECT * FROM sentiment_analysis;
DROP TABLE sentiment_analysis;
ALTER TABLE sentiment_analysis_plain RENAME TO sentiment_analysis;

CREATE INDEX IF NOT EXISTS idx_sentiment_symbol   ON sentiment_analysis (symbol);
CREATE INDEX IF NOT EXISTS idx_sentiment_analyzed ON sentiment_analysis (analyzed_at DESC);
