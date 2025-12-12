-- Migration: Add game_snapshot JSONB column to tournament_brackets
-- Reason: Store denormalized game data snapshot to avoid HTTP calls during match lifecycle
-- Date: 2025-12-09

-- Add game_snapshot column to tournament_brackets table
ALTER TABLE tournament_brackets 
ADD COLUMN game_snapshot JSONB;

-- Add comment explaining the column purpose
COMMENT ON COLUMN tournament_brackets.game_snapshot IS 
'Denormalized game data snapshot from tournament - used for match URLs and display. Contains slug, name, version, etc.';

-- Create GIN index for JSONB queries on game slug (optional, for future query optimization)
CREATE INDEX idx_brackets_game_snapshot_slug 
ON tournament_brackets ((game_snapshot->>'slug'));

COMMENT ON INDEX idx_brackets_game_snapshot_slug IS 
'Index on game slug within snapshot for efficient queries by game slug';







