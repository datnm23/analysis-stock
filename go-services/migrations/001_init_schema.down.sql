-- 001_init_schema.down.sql
DROP INDEX IF EXISTS idx_forecast_symbol_time;
DROP INDEX IF EXISTS idx_sentiment_analyzed;
DROP INDEX IF EXISTS idx_sentiment_symbol;
DROP INDEX IF EXISTS idx_technical_symbol_time;

DROP TABLE IF EXISTS daily_reports;
DROP TABLE IF EXISTS forecasts;
DROP TABLE IF EXISTS sentiment_analysis;
DROP TABLE IF EXISTS technical_analysis;
DROP TABLE IF EXISTS stocks;
