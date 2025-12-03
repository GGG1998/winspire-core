-- Migration: Create tournament_registrations table
-- Purpose: Store participant registrations for tournaments (scalable design for 10K+ participants)
-- Date: 2024-11-29

-- ============================================================================
-- TOURNAMENT REGISTRATIONS TABLE
-- ============================================================================
-- Represents participant registrations for tournaments.
-- Designed for high scalability (10,000+ concurrent participants per tournament).
-- Uses separate table instead of JSON for:
--   - Efficient indexing and querying
--   - Atomic concurrent updates
--   - Individual participant validation
--   - Pagination support

CREATE TABLE IF NOT EXISTS tournament_registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    team_id UUID,
    
    -- Registration status
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (
        status IN ('pending', 'confirmed', 'checked_in', 'withdrawn', 'disqualified')
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

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Primary lookup: Get all participants for a tournament (with pagination)
CREATE INDEX idx_tournament_registrations_tournament_id 
    ON tournament_registrations(tournament_id, registered_at DESC);

-- Find all tournaments a user is registered in
CREATE INDEX idx_tournament_registrations_user_id 
    ON tournament_registrations(user_id, registered_at DESC);

-- Filter by status within a tournament
CREATE INDEX idx_tournament_registrations_tournament_status 
    ON tournament_registrations(tournament_id, status);

-- Find ready participants
CREATE INDEX idx_tournament_registrations_tournament_ready 
    ON tournament_registrations(tournament_id, is_ready) 
    WHERE is_ready = true;

-- Team-based lookups
CREATE INDEX idx_tournament_registrations_team_id 
    ON tournament_registrations(team_id) 
    WHERE team_id IS NOT NULL;

-- ============================================================================
-- TRIGGER: Update updated_at timestamp
-- ============================================================================

CREATE TRIGGER update_tournament_registrations_updated_at
    BEFORE UPDATE ON tournament_registrations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE tournament_registrations IS 'Participant registrations for tournaments';
COMMENT ON COLUMN tournament_registrations.id IS 'Unique identifier for the registration';
COMMENT ON COLUMN tournament_registrations.tournament_id IS 'Reference to the tournament';
COMMENT ON COLUMN tournament_registrations.user_id IS 'Reference to the registered user';
COMMENT ON COLUMN tournament_registrations.team_id IS 'Optional reference to a team (for team tournaments)';
COMMENT ON COLUMN tournament_registrations.status IS 'Registration status: pending, confirmed, checked_in, withdrawn, disqualified';
COMMENT ON COLUMN tournament_registrations.display_name IS 'Participant display name (cached from user profile)';
COMMENT ON COLUMN tournament_registrations.avatar_url IS 'Participant avatar URL (cached from user profile)';
COMMENT ON COLUMN tournament_registrations.checked_in_at IS 'When the participant checked in';
COMMENT ON COLUMN tournament_registrations.is_ready IS 'Whether the participant has confirmed ready status';
COMMENT ON COLUMN tournament_registrations.registered_at IS 'When the participant registered';





