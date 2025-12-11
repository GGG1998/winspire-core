-- Add game loading tracking and optimistic locking support

-- Add 'loading' status to valid status values
ALTER TABLE tournament_matches DROP CONSTRAINT IF EXISTS tournament_matches_status_check;
ALTER TABLE tournament_matches ADD CONSTRAINT tournament_matches_status_check 
    CHECK (status IN (
        'pending',     -- Created, waiting for players
        'ready',       -- Both players ready
        'loading',     -- Game being loaded by players
        'started',     -- Match in progress
        'paused',      -- Disconnection pause
        'completed',   -- Match finished
        'cancelled'    -- Host cancelled
    ));

-- Add game loaded tracking columns
ALTER TABLE tournament_matches 
    ADD COLUMN participant1_game_loaded BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN participant2_game_loaded BOOLEAN NOT NULL DEFAULT false;

-- Add version column for optimistic locking
ALTER TABLE tournament_matches 
    ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

-- Update existing 'started', 'completed', and 'paused' matches to set game_loaded = true
-- This ensures backward compatibility for matches that started before this migration
UPDATE tournament_matches
SET 
    participant1_game_loaded = true,
    participant2_game_loaded = CASE 
        WHEN participant2_id IS NOT NULL THEN true 
        ELSE false 
    END
WHERE status IN ('started', 'completed', 'paused');

-- Add constraints to ensure valid state transitions
ALTER TABLE tournament_matches 
    ADD CONSTRAINT check_loading_requires_both_ready 
    CHECK (
        status != 'loading' OR 
        (participant1_ready = true AND participant2_ready = true)
    );

-- This constraint now works because we've updated existing matches above
ALTER TABLE tournament_matches 
    ADD CONSTRAINT check_started_requires_both_loaded
    CHECK (
        status NOT IN ('started', 'paused') OR
        (participant1_game_loaded = true AND (participant2_id IS NULL OR participant2_game_loaded = true))
    );

-- Add index for timeout monitoring queries
CREATE INDEX idx_matches_loading_timeout ON tournament_matches (status, updated_at) 
    WHERE status = 'loading';

-- Update comments
COMMENT ON COLUMN tournament_matches.status IS 'State machine: pending → ready → loading → started → [paused →] completed';
COMMENT ON COLUMN tournament_matches.participant1_game_loaded IS 'Whether participant1 has loaded the game (confirmed via callback)';
COMMENT ON COLUMN tournament_matches.participant2_game_loaded IS 'Whether participant2 has loaded the game (confirmed via callback)';
COMMENT ON COLUMN tournament_matches.version IS 'Optimistic locking version number, incremented on each update';

