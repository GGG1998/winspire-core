-- Migration: Add 'registered' status to tournament_registrations
-- Reason: Align with participant flow: pending → registered → confirmed → checked_in
-- Date: 2025-12-05

-- Drop the existing CHECK constraint
ALTER TABLE tournament_registrations
    DROP CONSTRAINT IF EXISTS tournament_registrations_status_check;

-- Add new CHECK constraint with 'registered' status
ALTER TABLE tournament_registrations
    ADD CONSTRAINT tournament_registrations_status_check 
    CHECK (status IN ('pending', 'registered', 'confirmed', 'checked_in', 'withdrawn', 'disqualified'));

-- Update comment to reflect new status
COMMENT ON COLUMN tournament_registrations.status IS 
'Registration status: pending (initial), registered (took a spot), confirmed (manually confirmed), checked_in (checked in before tournament), withdrawn (left tournament), disqualified (removed by host)';



