DROP INDEX IF EXISTS idx_shorty_accesses_country;
ALTER TABLE shorty_accesses DROP COLUMN IF EXISTS city;
ALTER TABLE shorty_accesses DROP COLUMN IF EXISTS country;
