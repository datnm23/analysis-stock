#!/bin/bash
# PostgreSQL Restore Script
#
# Usage:
#   ./restore.sh /backups/vnstock_20260330_020000.sql.gz
#   DB_NAME=vnstock_staging ./restore.sh backup.sql.gz

set -euo pipefail

DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-vnstock}"
DB_NAME="${DB_NAME:-vnstock}"

BACKUP_FILE="${1:-}"

if [ -z "${BACKUP_FILE}" ]; then
    echo "Usage: $0 <backup_file.sql.gz>"
    echo ""
    echo "Available backups:"
    ls -lh /backups/*.sql.gz 2>/dev/null || echo "  (none found in /backups/)"
    exit 1
fi

if [ ! -f "${BACKUP_FILE}" ]; then
    echo "ERROR: Backup file not found: ${BACKUP_FILE}"
    exit 1
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Restoring ${BACKUP_FILE} → ${DB_NAME}@${DB_HOST}:${DB_PORT}"
echo ""
echo "⚠️  WARNING: This will DROP and recreate all tables in ${DB_NAME}!"
echo "Press Ctrl+C within 5 seconds to cancel..."
sleep 5

echo "Restoring..."
gunzip -c "${BACKUP_FILE}" | psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --single-transaction \
    --set ON_ERROR_STOP=on \
    2>&1

echo "[$(date '+%Y-%m-%d %H:%M:%S')] ✅ Restore complete"
