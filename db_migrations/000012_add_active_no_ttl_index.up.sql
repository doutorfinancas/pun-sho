-- Partial index for the "active links with no TTL" subset.
--
-- The previous ActiveLinks count used `(ttl IS NULL OR ttl > NOW())`, which
-- Postgres can't satisfy with a single B-tree scan because of the IS NULL
-- branch. The query has been split into two halves; this index makes the
-- `ttl IS NULL` half index-only and ~constant time even on multi-million row
-- tables (the majority of links have no TTL set).
--
-- IMPORTANT: CREATE INDEX CONCURRENTLY cannot run inside a transaction, and
-- golang-migrate wraps each migration in a transaction by default. Apply with
-- the x-no-tx-wrap option on the database URL, e.g.:
--
--   migrate -database "${POSTGRES_URL}&x-no-tx-wrap=true" -path db_migrations up
--
-- Or apply manually via psql and then mark the migration as applied:
--
--   psql "$POSTGRES_URL" -f db_migrations/000012_add_active_no_ttl_index.up.sql
--   psql "$POSTGRES_URL" -c "INSERT INTO schema_migrations (version, dirty) VALUES (12, false);"

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shorties_active_no_ttl
    ON shorties (id)
    WHERE deleted_at IS NULL AND ttl IS NULL;
