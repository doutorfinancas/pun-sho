CREATE TABLE IF NOT EXISTS shorties (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    public_id TEXT,
    link TEXT,
    ttl TIMESTAMP DEFAULT NULL,
    created_at TIMESTAMP DEFAULT now(),
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_shorties_link ON shorties (ttl);
CREATE INDEX idx_shorties_deleted_at ON shorties (deleted_at);
CREATE INDEX idx_shorties_created_at ON shorties (created_at);
CREATE UNIQUE INDEX idx_shorties_public_id ON shorties (public_id);

CREATE TABLE IF NOT EXISTS shorty_accesses (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    shorty_id UUID,
    meta JSONB,
    user_agent TEXT,
    ip_address TEXT,
    extra TEXT,
    operating_system TEXT,
    browser TEXT,
    created_at TIMESTAMP DEFAULT now()
);

CREATE INDEX shorty_accesses_meta ON shorty_accesses USING GIN (meta);
CREATE INDEX shorty_accesses_extra ON shorty_accesses (extra);
CREATE INDEX shorty_accesses_created_at ON shorty_accesses (created_at);
CREATE INDEX shorty_accesses_ip_address ON shorty_accesses (ip_address);
CREATE INDEX shorty_accesses_shorty_id ON shorty_accesses (shorty_id);
