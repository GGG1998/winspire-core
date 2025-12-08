-- Migration: allow bracket_generation activity event (rebase for shared migration history)
-- Reason: ensure event_type check includes bracket_generation even if previous version numbering differs
-- Date: 2025-12-06

ALTER TABLE prelobby_activity_feed
    DROP CONSTRAINT IF EXISTS prelobby_activity_feed_event_type_check;

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



