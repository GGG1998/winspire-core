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
        'pending', 'ready', 'loading', 'started', 'paused', 'completed', 'cancelled'
    )),
    participant1_ready BOOLEAN NOT NULL DEFAULT false,
    participant2_ready BOOLEAN NOT NULL DEFAULT false,
    participant1_game_loaded BOOLEAN NOT NULL DEFAULT false,
    participant2_game_loaded BOOLEAN NOT NULL DEFAULT false,
    winner_id UUID,
    score_player1 INTEGER CHECK (score_player1 >= 0),
    score_player2 INTEGER CHECK (score_player2 >= 0),
    result_source VARCHAR(20) CHECK (result_source IN ('game_api', 'host_manual', 'walkover')),
    disconnected_player_id UUID,
    disconnected_at TIMESTAMP,
    game_api_match_id VARCHAR(255),
    game_api_poll_attempts INTEGER DEFAULT 0,
    game_api_last_poll TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
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

-- ============================================================================
-- Games Table (BC: GameLibrary - merged from game-management service)
-- ============================================================================
CREATE TABLE games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_integration_id UUID UNIQUE,
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    logo_url TEXT,
    storage_path TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '1.0.0',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- Tournament Service Tables (Consolidated)
-- ============================================================================

-- Hosts representing organizations or individuals
CREATE TABLE hosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    logo_url VARCHAR(512),
    website_url VARCHAR(512),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- User permissions for hosts
CREATE TABLE host_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id UUID NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(host_id, user_id)
);

-- Tournaments hosted by organizations
CREATE TABLE tournaments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id UUID NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    external_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (
        status IN ('draft', 'scheduled', 'registration_open', 'registration_closed', 'started', 'completed', 'cancelled')
    ),
    scheduled_start_time_at TIMESTAMP,
    registration_window_open_at TIMESTAMP,
    actual_start_time_at TIMESTAMP,
    completed_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    minimum_team_count INTEGER DEFAULT 8,
    maximum_team_count INTEGER,
    team_size INTEGER NOT NULL DEFAULT 1,
    auto_force_ready BOOLEAN DEFAULT true,
    game_id UUID,
    space_id UUID,
    template_id UUID,
    ready_window JSONB,
    prize JSONB,
    game_snapshot JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Participant registrations for tournaments
CREATE TABLE tournament_registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    team_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (
        status IN ('pending', 'registered', 'confirmed', 'checked_in', 'withdrawn', 'disqualified')
    ),
    display_name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    checked_in_at TIMESTAMP,
    is_ready BOOLEAN DEFAULT false,
    registered_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tournament_id, user_id)
);
