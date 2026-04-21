ALTER TABLE shorties ADD CONSTRAINT chk_public_id_chars
    CHECK (public_id ~ '^[a-zA-Z0-9_-]+$') NOT VALID;

ALTER TABLE shorties VALIDATE CONSTRAINT chk_public_id_chars;
