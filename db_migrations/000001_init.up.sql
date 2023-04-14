CREATE TABLE IF NOT EXISTS pun_sho.shorties (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    public_id TEXT,
    link TEXT,
    ttl TIMESTAMP DEFAULT NULL,
    created_at TIMESTAMP DEFAULT now(),
    deleted_at TIMESTAMP DEFAULT NULL,
    INDEX (ttl),
    INDEX (deleted_at),
    INDEX (created_at),
    UNIQUE (public_id)
);

CREATE TABLE IF NOT EXISTS pun_sho.shorty_accesses (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    shorty_id UUID,
    meta JSONB,
    user_agent TEXT,
    ip_address TEXT,
    extra TEXT,
    operating_system TEXT,
    browser TEXT,
    created_at TIMESTAMP DEFAULT now(),
    INDEX (shorty_id),
    INDEX (ip_address),
    INDEX (created_at),
    INDEX (extra)
);

CREATE INDEX shorty_accesses_meta ON pun_sho.shorty_accesses USING GIN (meta);
