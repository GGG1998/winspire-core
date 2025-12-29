-- Migration: Add 'scheduled' status to tournaments
-- Purpose: Enable tournament scheduling phase before registration opens
-- Date: 2024-12-01

-- ============================================================================
-- UPDATE STATUS CHECK CONSTRAINT
-- ============================================================================

-- Drop the existing CHECK constraint
ALTER TABLE tournaments
DROP CONSTRAINT IF EXISTS tournaments_status_check;

-- Add new CHECK constraint with 'scheduled' status
ALTER TABLE tournaments
ADD CONSTRAINT tournaments_status_check 
CHECK (status IN ('draft', 'scheduled', 'registration_open', 'registration_closed', 'started', 'completed', 'cancelled'));

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON CONSTRAINT tournaments_status_check ON tournaments IS 
'Allowed tournament statuses: draft (unpublished), scheduled (published but registration not open), registration_open (accepting signups), registration_closed (no more signups), started (in progress), completed (finished), cancelled (terminated early)';



