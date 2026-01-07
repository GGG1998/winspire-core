-- Migration: Create tournament schema (merged from tournament service)
-- Purpose: Support tournament lifecycle management within matchmaking monolith
-- Date: 2025-01-07

-- ============================================================================
-- CREATE TOURNAMENT SCHEMA
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS tournament;

-- ============================================================================
-- HOSTS TABLE
-- ============================================================================
-- Represents organizations or individuals that can host tournaments

CREATE TABLE IF NOT EXISTS tournament.hosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    logo_url VARCHAR(512),
    website_url VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tournament_hosts_name ON tournament.hosts(name);

-- ============================================================================
-- HOST MEMBERS TABLE
-- ============================================================================
-- Manages user permissions for hosts (authorization)
-- Roles: 'owner', 'admin', 'member'

CREATE TABLE IF NOT EXISTS tournament.host_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id UUID NOT NULL REFERENCES tournament.hosts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(host_id, user_id)
);

CREATE INDEX idx_tournament_host_members_host_id ON tournament.host_members(host_id);
CREATE INDEX idx_tournament_host_members_user_id ON tournament.host_members(user_id);
CREATE INDEX idx_tournament_host_members_role ON tournament.host_members(role);
CREATE INDEX idx_tournament_host_members_host_user ON tournament.host_members(host_id, user_id);

-- ============================================================================
-- TOURNAMENTS TABLE
-- ============================================================================
-- Represents tournaments hosted by organizations

CREATE TABLE IF NOT EXISTS tournament.tournaments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id UUID NOT NULL REFERENCES tournament.hosts(id) ON DELETE CASCADE,

    -- Basic information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    external_id VARCHAR(255),

    -- Status and lifecycle
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (
        status IN (
            'draft',
            'scheduled',
            'registration_open',
            'registration_closed',
            'starting',
            'started',
            'completed',
            'cancelled'
        )
    ),

    -- Timing
    scheduled_start_time_at TIMESTAMPTZ,
    registration_window_open_at TIMESTAMPTZ,
    actual_start_time_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,

    -- Configuration
    minimum_team_count INTEGER DEFAULT 2,
    maximum_team_count INTEGER,
    team_size INTEGER NOT NULL DEFAULT 1,
    auto_force_ready BOOLEAN DEFAULT true,

    -- Optional references
    game_id UUID,
    space_id UUID,
    template_id UUID,

    -- JSONB fields for flexible configuration
    ready_window JSONB,
    prize JSONB,
    game_snapshot JSONB,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Primary lookup indexes
CREATE INDEX idx_tournament_tournaments_host_id ON tournament.tournaments(host_id);
CREATE INDEX idx_tournament_tournaments_status ON tournament.tournaments(status);
CREATE INDEX idx_tournament_tournaments_external_id ON tournament.tournaments(external_id) WHERE external_id IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX idx_tournament_tournaments_host_status ON tournament.tournaments(host_id, status);
CREATE INDEX idx_tournament_tournaments_host_created ON tournament.tournaments(host_id, created_at DESC);

-- Timing indexes
CREATE INDEX idx_tournament_tournaments_scheduled_start ON tournament.tournaments(scheduled_start_time_at) WHERE scheduled_start_time_at IS NOT NULL;
CREATE INDEX idx_tournament_tournaments_registration_open ON tournament.tournaments(registration_window_open_at) WHERE registration_window_open_at IS NOT NULL;

-- Optional references
CREATE INDEX idx_tournament_tournaments_game_id ON tournament.tournaments(game_id) WHERE game_id IS NOT NULL;
CREATE INDEX idx_tournament_tournaments_space_id ON tournament.tournaments(space_id) WHERE space_id IS NOT NULL;

-- Game snapshot lookup
CREATE INDEX idx_tournament_tournaments_game_snapshot_slug ON tournament.tournaments ((game_snapshot->>'slug'));

-- ============================================================================
-- TOURNAMENT REGISTRATIONS TABLE
-- ============================================================================
-- Represents participant registrations for tournaments
-- Designed for high scalability (10,000+ concurrent participants per tournament)

CREATE TABLE IF NOT EXISTS tournament.registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournament.tournaments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    team_id UUID,

    -- Registration status
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (
        status IN ('pending', 'registered', 'confirmed', 'checked_in', 'withdrawn', 'disqualified')
    ),

    -- Participant display info (denormalized for performance)
    display_name VARCHAR(255) NOT NULL,
    avatar_url TEXT,

    -- Check-in tracking
    checked_in_at TIMESTAMPTZ,
    is_ready BOOLEAN DEFAULT false,

    -- Metadata
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure unique registration per user per tournament
    UNIQUE(tournament_id, user_id)
);

-- Primary lookup: Get all participants for a tournament (with pagination)
CREATE INDEX idx_tournament_registrations_tournament_id ON tournament.registrations(tournament_id, registered_at DESC);

-- Find all tournaments a user is registered in
CREATE INDEX idx_tournament_registrations_user_id ON tournament.registrations(user_id, registered_at DESC);

-- Filter by status within a tournament
CREATE INDEX idx_tournament_registrations_tournament_status ON tournament.registrations(tournament_id, status);

-- Find ready participants
CREATE INDEX idx_tournament_registrations_tournament_ready ON tournament.registrations(tournament_id, is_ready) WHERE is_ready = true;

-- Team-based lookups
CREATE INDEX idx_tournament_registrations_team_id ON tournament.registrations(team_id) WHERE team_id IS NOT NULL;

-- ============================================================================
-- TRIGGERS: Update updated_at timestamp
-- ============================================================================

-- Reuse existing update_updated_at_column function from public schema
-- (created in matchmaking migrations 000001)

CREATE TRIGGER update_tournament_hosts_updated_at
    BEFORE UPDATE ON tournament.hosts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tournament_host_members_updated_at
    BEFORE UPDATE ON tournament.host_members
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tournament_tournaments_updated_at
    BEFORE UPDATE ON tournament.tournaments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tournament_registrations_updated_at
    BEFORE UPDATE ON tournament.registrations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON SCHEMA tournament IS 'Tournament management bounded context (merged from tournament service)';

COMMENT ON TABLE tournament.hosts IS 'Organizations or individuals that can host tournaments';
COMMENT ON COLUMN tournament.hosts.id IS 'Unique identifier for the host';
COMMENT ON COLUMN tournament.hosts.name IS 'Display name of the host organization';
COMMENT ON COLUMN tournament.hosts.description IS 'Optional description of the host';
COMMENT ON COLUMN tournament.hosts.logo_url IS 'URL to host logo/avatar';
COMMENT ON COLUMN tournament.hosts.website_url IS 'Host website URL';

COMMENT ON TABLE tournament.host_members IS 'User permissions and roles for hosts';
COMMENT ON COLUMN tournament.host_members.host_id IS 'Reference to the host';
COMMENT ON COLUMN tournament.host_members.user_id IS 'Reference to user (from auth system)';
COMMENT ON COLUMN tournament.host_members.role IS 'User role: owner, admin, or member';

COMMENT ON TABLE tournament.tournaments IS 'Tournaments hosted by organizations';
COMMENT ON COLUMN tournament.tournaments.id IS 'Unique identifier for the tournament';
COMMENT ON COLUMN tournament.tournaments.host_id IS 'Reference to the host organization';
COMMENT ON COLUMN tournament.tournaments.name IS 'Display name of the tournament';
COMMENT ON COLUMN tournament.tournaments.description IS 'Optional description';
COMMENT ON COLUMN tournament.tournaments.external_id IS 'External identifier for idempotency';
COMMENT ON COLUMN tournament.tournaments.status IS 'Current tournament status: draft, scheduled, registration_open, registration_closed, starting, started, completed, cancelled';
COMMENT ON COLUMN tournament.tournaments.scheduled_start_time_at IS 'When the tournament is scheduled to start';
COMMENT ON COLUMN tournament.tournaments.registration_window_open_at IS 'When registration opens';
COMMENT ON COLUMN tournament.tournaments.actual_start_time_at IS 'When the tournament actually started';
COMMENT ON COLUMN tournament.tournaments.completed_at IS 'When the tournament was completed';
COMMENT ON COLUMN tournament.tournaments.cancelled_at IS 'When the tournament was cancelled';
COMMENT ON COLUMN tournament.tournaments.minimum_team_count IS 'Minimum number of teams required to start';
COMMENT ON COLUMN tournament.tournaments.maximum_team_count IS 'Maximum number of teams allowed';
COMMENT ON COLUMN tournament.tournaments.team_size IS 'Number of players per team';
COMMENT ON COLUMN tournament.tournaments.auto_force_ready IS 'Whether to automatically force ready status';
COMMENT ON COLUMN tournament.tournaments.game_id IS 'Optional reference to the game';
COMMENT ON COLUMN tournament.tournaments.space_id IS 'Optional reference to the space';
COMMENT ON COLUMN tournament.tournaments.template_id IS 'Optional reference to a tournament template';
COMMENT ON COLUMN tournament.tournaments.ready_window IS 'JSONB: Check-in window { startsAt, endsAt }';
COMMENT ON COLUMN tournament.tournaments.prize IS 'JSONB: Prize config { type, description, value, currency }';
COMMENT ON COLUMN tournament.tournaments.game_snapshot IS 'JSONB: Denormalized game data snapshot (slug, name, version, logo)';

COMMENT ON TABLE tournament.registrations IS 'Participant registrations for tournaments';
COMMENT ON COLUMN tournament.registrations.id IS 'Unique identifier for the registration';
COMMENT ON COLUMN tournament.registrations.tournament_id IS 'Reference to the tournament';
COMMENT ON COLUMN tournament.registrations.user_id IS 'Reference to the registered user';
COMMENT ON COLUMN tournament.registrations.team_id IS 'Optional reference to a team (for team tournaments)';
COMMENT ON COLUMN tournament.registrations.status IS 'Registration status: pending, registered, confirmed, checked_in, withdrawn, disqualified';
COMMENT ON COLUMN tournament.registrations.display_name IS 'Participant display name (cached from user profile)';
COMMENT ON COLUMN tournament.registrations.avatar_url IS 'Participant avatar URL (cached from user profile)';
COMMENT ON COLUMN tournament.registrations.checked_in_at IS 'When the participant checked in';
COMMENT ON COLUMN tournament.registrations.is_ready IS 'Whether the participant has confirmed ready status';
COMMENT ON COLUMN tournament.registrations.registered_at IS 'When the participant registered';
