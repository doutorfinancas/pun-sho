-- Partial indexes to speed up the dashboard and link analytics queries
-- on large data sets (millions of shorties / accesses).
--
-- IMPORTANT: CREATE INDEX CONCURRENTLY cannot run inside a transaction, and
-- golang-migrate wraps each migration in a transaction by default. Apply this
-- migration with the x-no-tx-wrap option on the database URL, e.g.:
--
--   migrate -database "${POSTGRES_URL}&x-no-tx-wrap=true" -path db_migrations up
--
-- Or apply manually via psql and then mark the migration as applied:
--
--   psql "$POSTGRES_URL" -f db_migrations/000011_add_perf_indexes.up.sql
--   psql "$POSTGRES_URL" -c "INSERT INTO schema_migrations (version, dirty) VALUES (11, false);"

-- Active links (deleted_at IS NULL) — used by COUNT, recent list, list page filters.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shorties_active_created_at
    ON shorties (created_at DESC)
    WHERE deleted_at IS NULL;

-- Active links filtered by ttl — used by the active/expired COUNT(*) on the dashboard.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shorties_active_ttl
    ON shorties (ttl)
    WHERE deleted_at IS NULL;

-- Redirected accesses by created_at — used by every analytics range query.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shorty_accesses_redirected_created_at
    ON shorty_accesses (created_at)
    WHERE status = 'redirected';

-- Redirected accesses per link — used by per-link analytics queries.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shorty_accesses_shorty_redirected
    ON shorty_accesses (shorty_id, created_at)
    WHERE status = 'redirected';
