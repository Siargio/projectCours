CREATE TABLE IF NOT EXISTS links (
    id         SERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL UNIQUE,
    long_url   TEXT NOT NULL,
    clicks     INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_short_code ON links(short_code);