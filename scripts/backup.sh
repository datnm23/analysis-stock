#!/bin/bash
# PostgreSQL Automated Backup Script
# Runs daily via cron or Docker healthcheck-style scheduling.
# Retains backups for 7 days by default.
#
# Usage:
#   ./backup.sh                      # Backup all databases
#   RETENTION_DAYS=14 ./backup.sh    # Custom retention
#   S3_BUCKET=my-bucket ./backup.sh  # Upload to S3/MinIO

set -euo pipefail

# ---- Configuration ----
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-vnstock}"
DB_NAME="${DB_NAME:-vnstock}"
BACKUP_DIR="${BACKUP_DIR:-/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
S3_BUCKET="${S3_BUCKET:-}"
S3_ENDPOINT="${S3_ENDPOINT:-}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql.gz"

# ---- Functions ----
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# ---- Pre-checks ----
mkdir -p "${BACKUP_DIR}"

if ! command -v pg_dump &> /dev/null; then
    log "ERROR: pg_dump not found. Install postgresql-client."
    exit 1
fi

# ---- Backup ----
log "Starting backup: ${DB_NAME}@${DB_HOST}:${DB_PORT}"

pg_dump \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --format=plain \
    --no-owner \
    --no-privileges \
    --verbose \
    2>/dev/null | gzip > "${BACKUP_FILE}"

BACKUP_SIZE=$(du -sh "${BACKUP_FILE}" | cut -f1)
log "Backup complete: ${BACKUP_FILE} (${BACKUP_SIZE})"

# ---- Upload to S3/MinIO (optional) ----
if [ -n "${S3_BUCKET}" ]; then
    S3_PATH="s3://${S3_BUCKET}/db-backups/$(basename ${BACKUP_FILE})"
    log "Uploading to ${S3_PATH} ..."

    AWS_ARGS=""
    if [ -n "${S3_ENDPOINT}" ]; then
        AWS_ARGS="--endpoint-url ${S3_ENDPOINT}"
    fi

    if command -v aws &> /dev/null; then
        aws s3 cp ${AWS_ARGS} "${BACKUP_FILE}" "${S3_PATH}"
        log "Upload complete: ${S3_PATH}"
    else
        log "WARNING: aws CLI not found, skipping S3 upload"
    fi
fi

# ---- Cleanup old backups ----
log "Cleaning backups older than ${RETENTION_DAYS} days ..."
DELETED=$(find "${BACKUP_DIR}" -name "*.sql.gz" -mtime +${RETENTION_DAYS} -delete -print | wc -l)
log "Deleted ${DELETED} old backup(s)"

# ---- Summary ----
REMAINING=$(find "${BACKUP_DIR}" -name "*.sql.gz" | wc -l)
TOTAL_SIZE=$(du -sh "${BACKUP_DIR}" 2>/dev/null | cut -f1)
log "Backup summary: ${REMAINING} files, ${TOTAL_SIZE} total"
