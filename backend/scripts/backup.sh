#!/usr/bin/env bash
#
# Daily Postgres backup for smoke.
# Dumps the database via the running postgres container and prunes dumps
# older than a week. Intended to be run from a host-level cron job or
# systemd timer.
#
set -euo pipefail

# Backups are written to <repo>/backups (gitignored) unless BACKUP_DIR is set.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="${BACKUP_DIR:-$(cd "$SCRIPT_DIR/../.." && pwd)/backups}"
CONTAINER="${DB_CONTAINER:-smoke-db}"
DB_USER="${POSTGRES_USER:-smoke}"
DB_NAME="${POSTGRES_DB:-smoke}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"

mkdir -p "$BACKUP_DIR"

timestamp="$(date +%Y%m%d_%H%M%S)"
outfile="$BACKUP_DIR/smoke_${timestamp}.dump"

echo "[$(date -u '+%Y-%m-%dT%H:%M:%SZ')] dumping $DB_NAME -> $outfile"

# Dump to a temp file first so a failed/partial dump never replaces a good one.
tmpfile="${outfile}.tmp"
if docker exec "$CONTAINER" pg_dump -U "$DB_USER" -d "$DB_NAME" -F c > "$tmpfile"; then
  mv "$tmpfile" "$outfile"
  echo "[$(date -u '+%Y-%m-%dT%H:%M:%SZ')] backup complete ($(du -h "$outfile" | cut -f1))"
else
  rm -f "$tmpfile"
  echo "[$(date -u '+%Y-%m-%dT%H:%M:%SZ')] backup FAILED" >&2
  exit 1
fi

# Prune dumps older than RETENTION_DAYS.
deleted="$(find "$BACKUP_DIR" -name 'smoke_*.dump' -mtime "+${RETENTION_DAYS}" -print -delete)"
if [ -n "$deleted" ]; then
  echo "[$(date -u '+%Y-%m-%dT%H:%M:%SZ')] pruned:"
  echo "$deleted"
fi
