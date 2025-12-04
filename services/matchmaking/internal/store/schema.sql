-- Schema reference for SQLC code generation
-- This file aggregates all table definitions for SQLC to understand the database structure

-- ============================================================================
-- Tournament Brackets Table
-- ============================================================================
CREATE TABLE tournament_brackets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL UNIQUE,
    total_rounds INTEGER NOT NULL CHECK (total_rounds > 0),
    total_matches INTEGER NOT NULL CHECK (total_matches > 0),
    byes_count INTEGER NOT NULL DEFAULT 0 CHECK (byes_count >= 0),
    generated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- Tournament Rounds Table
-- ============================================================================
CREATE TABLE tournament_rounds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bracket_id UUID NOT NULL REFERENCES tournament_brackets(id) ON DELETE CASCADE,
    round_number INTEGER NOT NULL CHECK (round_number > 0),
    round_name VARCHAR(100) NOT NULL,
    matches_count INTEGER NOT NULL CHECK (matches_count > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'in_progress', 'completed')),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    CONSTRAINT unique_bracket_round_number UNIQUE (bracket_id, round_number)
);

-- ============================================================================
-- Tournament Matches Table
-- ============================================================================
CREATE TABLE tournament_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    round_id UUID NOT NULL REFERENCES tournament_rounds(id) ON DELETE CASCADE,
    match_number INTEGER NOT NULL CHECK (match_number > 0),
    next_match_id UUID REFERENCES tournament_matches(id),
    participant1_id UUID NOT NULL,
    participant2_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending', 'ready', 'started', 'paused', 'completed', 'cancelled'
    )),
    participant1_ready BOOLEAN NOT NULL DEFAULT false,
    participant2_ready BOOLEAN NOT NULL DEFAULT false,
    winner_id UUID,
    score_player1 INTEGER CHECK (score_player1 >= 0),
    score_player2 INTEGER CHECK (score_player2 >= 0),
    result_source VARCHAR(20) CHECK (result_source IN ('game_api', 'host_manual', 'walkover')),
    disconnected_player_id UUID,
    disconnected_at TIMESTAMP,
    game_api_match_id VARCHAR(255),
    game_api_poll_attempts INTEGER DEFAULT 0,
    game_api_last_poll TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT check_completed_has_winner CHECK (
        (status = 'completed' AND winner_id IS NOT NULL) OR status != 'completed'
    ),
    CONSTRAINT unique_round_match_number UNIQUE (round_id, match_number)
);

