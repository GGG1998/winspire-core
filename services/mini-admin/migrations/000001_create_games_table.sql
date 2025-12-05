-- +migrate Up
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    logo_url TEXT,
    s3_path TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '1.0.0',
    versioning_enabled BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_games_slug ON games(slug) WHERE is_active = true;
CREATE INDEX idx_games_active ON games(is_active);
CREATE INDEX idx_games_created_at ON games(created_at DESC);

-- +migrate Down
DROP INDEX IF EXISTS idx_games_created_at;
DROP INDEX IF EXISTS idx_games_active;
DROP INDEX IF EXISTS idx_games_slug;
DROP TABLE IF EXISTS games;

