-- Migration: Create games table in Supabase
-- Purpose: Move games from matchmaking DB to Supabase for frontend direct access

-- ============================================================================
-- GAMES TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_integration_id UUID UNIQUE,
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    logo_url TEXT,
    storage_path TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '1.0.0',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_games_slug ON public.games(slug);
CREATE INDEX idx_games_is_active ON public.games(is_active);
CREATE INDEX idx_games_name ON public.games(name);

-- Trigger for updated_at (reuses existing function from migration 005)
DROP TRIGGER IF EXISTS update_games_updated_at ON public.games;
CREATE TRIGGER update_games_updated_at
    BEFORE UPDATE ON public.games
    FOR EACH ROW
    EXECUTE FUNCTION public.update_updated_at();

-- ============================================================================
-- ROW LEVEL SECURITY
-- ============================================================================
ALTER TABLE public.games ENABLE ROW LEVEL SECURITY;

-- Public read access for active games (anyone can view games, even unauthenticated)
CREATE POLICY "Public read access for active games"
    ON public.games
    FOR SELECT
    USING (is_active = true);

-- Service role can do everything (for Go backend admin operations)
CREATE POLICY "Service role full access"
    ON public.games
    TO service_role
    USING (true)
    WITH CHECK (true);

-- ============================================================================
-- COMMENTS
-- ============================================================================
COMMENT ON TABLE public.games IS 'Game catalog for the tournament platform';
COMMENT ON COLUMN public.games.id IS 'Unique identifier for the game';
COMMENT ON COLUMN public.games.game_integration_id IS 'External game integration reference';
COMMENT ON COLUMN public.games.slug IS 'URL-friendly unique identifier';
COMMENT ON COLUMN public.games.name IS 'Display name of the game';
COMMENT ON COLUMN public.games.description IS 'Game description';
COMMENT ON COLUMN public.games.logo_url IS 'URL to game logo/thumbnail';
COMMENT ON COLUMN public.games.storage_path IS 'S3 path prefix for game bundle';
COMMENT ON COLUMN public.games.version IS 'Current game version';
COMMENT ON COLUMN public.games.is_active IS 'Whether game is available for tournaments';
