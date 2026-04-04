-- 002_add_n8n_database.down.sql
DROP TRIGGER IF EXISTS update_stocks_updated_at ON stocks;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_stocks_exchange;
DROP INDEX IF EXISTS idx_stocks_industry;
DROP INDEX IF EXISTS idx_forecasts_recommendation;
DROP INDEX IF EXISTS idx_daily_reports_date;
