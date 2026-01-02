-- Migration: Create tournaments table
-- Purpose: Store tournament information for tournament management
-- Date: 2024-11-28

-- ============================================================================
-- TOURNAMENTS TABLE
-- ============================================================================
-- Represents tournaments hosted by organizations

CREATE TABLE IF NOT EXISTS tournaments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id UUID NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    
    -- Basic information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    external_id VARCHAR(255),
    
    -- Status and lifecycle
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (
        status IN ('draft', 'registration_open', 'registration_closed', 'started', 'completed', 'cancelled')
    ),
    
    -- Timing
    scheduled_start_time_at TIMESTAMPTZ,
    registration_window_open_at TIMESTAMPTZ,
    actual_start_time_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    
    -- Configuration
    minimum_team_count INTEGER DEFAULT 8,
    maximum_team_count INTEGER,
    team_size INTEGER NOT NULL DEFAULT 1,
    auto_force_ready BOOLEAN DEFAULT true,
    
    -- Optional references
    game_id UUID,
    space_id UUID,
    template_id UUID,
    
    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Primary lookup indexes
CREATE INDEX IF NOT EXISTS idx_tournaments_host_id ON tournaments(host_id);
CREATE INDEX IF NOT EXISTS idx_tournaments_status ON tournaments(status);
CREATE INDEX IF NOT EXISTS idx_tournaments_external_id ON tournaments(external_id) WHERE external_id IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_tournaments_host_status ON tournaments(host_id, status);
CREATE INDEX IF NOT EXISTS idx_tournaments_host_created ON tournaments(host_id, created_at DESC);

-- Timing indexes
CREATE INDEX IF NOT EXISTS idx_tournaments_scheduled_start ON tournaments(scheduled_start_time_at) WHERE scheduled_start_time_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tournaments_registration_open ON tournaments(registration_window_open_at) WHERE registration_window_open_at IS NOT NULL;

-- Optional references
CREATE INDEX IF NOT EXISTS idx_tournaments_game_id ON tournaments(game_id) WHERE game_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tournaments_space_id ON tournaments(space_id) WHERE space_id IS NOT NULL;

-- ============================================================================
-- TRIGGER: Update updated_at timestamp
-- ============================================================================

DROP TRIGGER IF EXISTS update_tournaments_updated_at ON tournaments;
CREATE TRIGGER update_tournaments_updated_at
    BEFORE UPDATE ON tournaments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE tournaments IS 'Tournaments hosted by organizations';
COMMENT ON COLUMN tournaments.id IS 'Unique identifier for the tournament';
COMMENT ON COLUMN tournaments.host_id IS 'Reference to the host organization';
COMMENT ON COLUMN tournaments.name IS 'Display name of the tournament';
COMMENT ON COLUMN tournaments.description IS 'Optional description';
COMMENT ON COLUMN tournaments.external_id IS 'External identifier for idempotency';
COMMENT ON COLUMN tournaments.status IS 'Current tournament status';
COMMENT ON COLUMN tournaments.scheduled_start_time_at IS 'When the tournament is scheduled to start';
COMMENT ON COLUMN tournaments.registration_window_open_at IS 'When registration opens';
COMMENT ON COLUMN tournaments.actual_start_time_at IS 'When the tournament actually started';
COMMENT ON COLUMN tournaments.completed_at IS 'When the tournament was completed';
COMMENT ON COLUMN tournaments.cancelled_at IS 'When the tournament was cancelled';
COMMENT ON COLUMN tournaments.minimum_team_count IS 'Minimum number of teams required to start';
COMMENT ON COLUMN tournaments.maximum_team_count IS 'Maximum number of teams allowed';
COMMENT ON COLUMN tournaments.team_size IS 'Number of players per team';
COMMENT ON COLUMN tournaments.auto_force_ready IS 'Whether to automatically force ready status';
COMMENT ON COLUMN tournaments.game_id IS 'Optional reference to the game';
COMMENT ON COLUMN tournaments.space_id IS 'Optional reference to the space';
COMMENT ON COLUMN tournaments.template_id IS 'Optional reference to a tournament template';

