-- Migration: allow bracket_generation activity event
-- Reason: application emits bracket_generation to activity feed during bracket generation completion
-- Date: 2025-12-06

-- Drop existing check constraint
ALTER TABLE prelobby_activity_feed
    DROP CONSTRAINT IF EXISTS prelobby_activity_feed_event_type_check;

-- Add updated check constraint including bracket_generation
ALTER TABLE prelobby_activity_feed
    ADD CONSTRAINT prelobby_activity_feed_event_type_check CHECK (
        event_type IN (
            'participant_joined',
            'participant_left',
            'grace_period_started',
            'grace_period_ended',
            'bracket_generation',
            'tournament_cancelled',
            'system_message'
        )
    );


