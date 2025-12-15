-- Migration: Add game_snapshot JSONB column to tournaments
-- Reason: Store denormalized game data snapshot to avoid HTTP calls in event-driven flows
-- Date: 2025-12-09

-- Add game_snapshot column to tournaments table
ALTER TABLE tournaments 
ADD COLUMN game_snapshot JSONB;

-- Add comment explaining the column purpose
COMMENT ON COLUMN tournaments.game_snapshot IS 
'Snapshot of game data at tournament creation (slug, name, version, logo, etc.). Used for event-driven flows to avoid synchronous HTTP calls to game-management service.';

-- Create GIN index for JSONB queries (optional, for future query optimization)
CREATE INDEX idx_tournaments_game_snapshot_slug 
ON tournaments ((game_snapshot->>'slug'));

COMMENT ON INDEX idx_tournaments_game_snapshot_slug IS 
'Index on game slug within snapshot for efficient queries by game slug';









