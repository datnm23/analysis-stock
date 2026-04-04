-- 004_auto_partition_function.up.sql
-- Create a PL/pgSQL function that automatically creates monthly partitions
-- for the next N months. Designed to be called periodically (at app startup,
-- or via pg_cron) so that INSERT never hits the DEFAULT partition.

CREATE OR REPLACE FUNCTION ensure_partitions(
    p_table_name   TEXT,    -- e.g. 'technical_analysis'
    p_column_name  TEXT,    -- e.g. 'timestamp' or 'analyzed_at'
    p_months_ahead INT DEFAULT 3  -- how many months into the future to pre-create
) RETURNS INT AS $$
DECLARE
    v_start     DATE;
    v_end       DATE;
    v_partition TEXT;
    v_count     INT := 0;
BEGIN
    FOR i IN 0..p_months_ahead LOOP
        v_start := date_trunc('month', CURRENT_DATE + (i || ' months')::INTERVAL)::DATE;
        v_end   := (v_start + INTERVAL '1 month')::DATE;
        v_partition := p_table_name || '_' || to_char(v_start, 'YYYY_MM');

        -- Skip if partition already exists
        IF NOT EXISTS (
            SELECT 1 FROM pg_class
            WHERE relname = v_partition
              AND relkind = 'r'  -- regular table (partition)
        ) THEN
            EXECUTE format(
                'CREATE TABLE %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
                v_partition, p_table_name, v_start, v_end
            );
            RAISE NOTICE 'Created partition: %', v_partition;
            v_count := v_count + 1;
        END IF;
    END LOOP;

    RETURN v_count;
END;
$$ LANGUAGE plpgsql;

-- Call it now to ensure the next 6 months exist
SELECT ensure_partitions('technical_analysis', 'timestamp', 6);
SELECT ensure_partitions('sentiment_analysis', 'analyzed_at', 6);

COMMENT ON FUNCTION ensure_partitions IS
    'Idempotently creates monthly range partitions for the next N months. '
    'Call at application startup or via pg_cron schedule.';
