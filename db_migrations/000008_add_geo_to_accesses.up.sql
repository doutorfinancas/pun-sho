ALTER TABLE shorty_accesses ADD COLUMN country VARCHAR(100);
ALTER TABLE shorty_accesses ADD COLUMN city VARCHAR(200);
CREATE INDEX idx_shorty_accesses_country ON shorty_accesses(country);
