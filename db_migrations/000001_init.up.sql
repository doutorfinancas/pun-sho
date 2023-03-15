CREATE TABLE IF NOT EXISTS pun_sho.shorties (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    public_id STRING,
    link STRING,
    ttl TIMESTAMP DEFAULT NULL,
    created_at TIMESTAMP DEFAULT now(),
    deleted_at TIMESTAMP DEFAULT NULL,
    INDEX (ttl),
    INDEX (deleted_at),
    INDEX (created_at),
    UNIQUE INDEX (public_id)
);

CREATE TABLE IF NOT EXISTS pun_sho.shorty_accesses (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    shorty_id UUID,
    meta jsonb,
    user_agent string,
    ip_address string,
    extra string,
    created_at TIMESTAMP DEFAULT now(),
    INDEX (shorty_id),
    INDEX (ip_address),
    INDEX (created_at),
    INDEX (extra)
);

CREATE INDEX pun_sho.shorty_accesses_meta ON shorty_accesses USING GIN (meta);
