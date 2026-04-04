-- 004_auto_partition_function.down.sql
-- Drop the auto-partition function

DROP FUNCTION IF EXISTS ensure_partitions(TEXT, TEXT, INT);
