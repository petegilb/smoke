#!/bin/bash
# Resets the smoke database to a clean state by truncating all data tables
# and resetting the last_scraped_at timestamp to epoch.

DB_URL="${DATABASE_URL:-postgres://smoke:smoke@localhost:5432/smoke?sslmode=disable}"

echo "Resetting smoke database..."
psql "$DB_URL" <<SQL
TRUNCATE daily_snapshots, games CASCADE;
UPDATE scrape_metadata SET value = '1970-01-01T00:00:00Z' WHERE key = 'last_scraped_at';
SQL

echo "Database reset complete."
