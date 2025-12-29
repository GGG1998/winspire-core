-- Migration: Add 'starting' status to tournaments
-- Reason: Tournament start saga uses transient "starting" state before confirming with matchmaking
-- Date: 2025-12-06

-- Drop the existing CHECK constraint
ALTER TABLE tournaments
    DROP CONSTRAINT IF EXISTS tournaments_status_check;

-- Add new CHECK constraint with 'starting' status
ALTER TABLE tournaments
    ADD CONSTRAINT tournaments_status_check CHECK (
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
    );

-- Update comment to reflect new status
COMMENT ON CONSTRAINT tournaments_status_check ON tournaments IS 
'Allowed tournament statuses: draft (unpublished), scheduled (published but registration not open), registration_open (accepting signups), registration_closed (no more signups), starting (transient state during start saga), started (in progress), completed (finished), cancelled (terminated early)';

