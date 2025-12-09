-- Schema reference for SQLC code generation
-- This file aggregates all table definitions for SQLC to understand the database structure

-- ============================================================================
-- Tournament Brackets Table
-- ============================================================================
CREATE TABLE tournament_brackets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL UNIQUE,
    game_snapshot JSONB,
    total_rounds INTEGER NOT NULL CHECK (total_rounds > 0),
    total_matches INTEGER NOT NULL CHECK (total_matches > 0),
    byes_count INTEGER NOT NULL DEFAULT 0 CHECK (byes_count >= 0),
    generated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP NULL
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

-- ============================================================================
-- Pre-Lobby Tables (Tournament waiting room before start)
-- ============================================================================

-- Pre-lobby state for tournaments
CREATE TABLE prelobbies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL UNIQUE,
    status VARCHAR(30) NOT NULL DEFAULT 'waiting' CHECK (
        status IN ('waiting', 'grace_period', 'generating_bracket', 'started', 'cancelled')
    ),
    grace_period_start TIMESTAMP,
    grace_period_end TIMESTAMP,
    min_participants INTEGER NOT NULL DEFAULT 2 CHECK (min_participants >= 2),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Participant snapshot when grace period ends (immutable, used for bracket generation)
CREATE TABLE prelobby_participant_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL UNIQUE REFERENCES prelobbies(tournament_id) ON DELETE CASCADE,
    participants JSONB NOT NULL,
    participant_count INTEGER NOT NULL CHECK (participant_count >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Activity feed for pre-lobby events
CREATE TABLE prelobby_activity_feed (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES prelobbies(tournament_id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL CHECK (
        event_type IN ('participant_joined', 'participant_left', 'grace_period_started', 
                       'grace_period_ended', 'bracket_generation', 'tournament_cancelled', 'system_message')
    ),
    message TEXT NOT NULL,
    participant_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

