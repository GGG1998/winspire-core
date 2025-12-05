-- Create tournament_matches table
CREATE TABLE tournament_matches (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Foreign key to round (same database)
    round_id UUID NOT NULL REFERENCES tournament_rounds(id) ON DELETE CASCADE,
    
    -- Match metadata
    match_number INTEGER NOT NULL CHECK (match_number > 0),
    next_match_id UUID REFERENCES tournament_matches(id), -- NULL for final
    
    -- Participants (UUIDs only, no FK - users table in different DB)
    -- participant2_id NULL = bye for participant1
    participant1_id UUID NOT NULL,
    participant2_id UUID, -- Nullable for byes
    
    -- Match state
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending',     -- Created, waiting for players
        'ready',       -- Both players ready
        'started',     -- Match in progress
        'paused',      -- Disconnection pause (CS:GO style)
        'completed',   -- Match finished
        'cancelled'    -- Host cancelled
    )),
    
    -- Ready state (per player)
    participant1_ready BOOLEAN NOT NULL DEFAULT false,
    participant2_ready BOOLEAN NOT NULL DEFAULT false,
    
    -- Results
    winner_id UUID, -- Must be participant1_id or participant2_id (checked in app layer)
    score_player1 INTEGER CHECK (score_player1 >= 0),
    score_player2 INTEGER CHECK (score_player2 >= 0),
    result_source VARCHAR(20) CHECK (result_source IN ('game_api', 'host_manual', 'walkover')),
    
    -- Disconnect tracking (CS:GO style)
    disconnected_player_id UUID, -- Must be participant1_id or participant2_id
    disconnected_at TIMESTAMP,
    
    -- Game API integration
    game_api_match_id VARCHAR(255),  -- External game match ID
    game_api_poll_attempts INTEGER DEFAULT 0,
    game_api_last_poll TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT check_completed_has_winner CHECK (
        (status = 'completed' AND winner_id IS NOT NULL) OR status != 'completed'
    ),
    CONSTRAINT unique_round_match_number UNIQUE (round_id, match_number)
);

-- Note: No foreign keys on participant1_id, participant2_id, winner_id, 
-- or disconnected_player_id because users table exists in competition_db 
-- (separate database). Application layer validates user existence.

-- Indexes
CREATE INDEX idx_matches_round_id ON tournament_matches(round_id);
CREATE INDEX idx_matches_next_match_id ON tournament_matches(next_match_id) 
    WHERE next_match_id IS NOT NULL;
CREATE INDEX idx_matches_participants ON tournament_matches(participant1_id, participant2_id)
    WHERE status IN ('pending', 'ready', 'started');
CREATE INDEX idx_matches_game_api_pending ON tournament_matches(game_api_match_id, game_api_last_poll)
    WHERE status = 'started' AND result_source IS NULL;
CREATE INDEX idx_matches_status ON tournament_matches(status);

COMMENT ON TABLE tournament_matches IS 'Individual 1v1 matches between participants. Core entity for match state management.';
COMMENT ON COLUMN tournament_matches.participant2_id IS 'NULL for bye matches (participant1 auto-advances)';
COMMENT ON COLUMN tournament_matches.next_match_id IS 'Winner advances to this match. NULL for final match.';
COMMENT ON COLUMN tournament_matches.status IS 'State machine: pending → ready → started → [paused →] completed';
COMMENT ON COLUMN tournament_matches.disconnected_at IS 'When disconnect occurred. Used for 30s reconnection window enforcement.';
COMMENT ON COLUMN tournament_matches.result_source IS 'How result was determined: game_api (polled), host_manual (override), walkover (no-show)';
