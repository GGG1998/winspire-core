-- Migration: allow scheduled status in tournaments
-- Reason: application uses "scheduled" state when publishing tournaments
-- Date: 2025-12-04

ALTER TABLE tournaments
    DROP CONSTRAINT IF EXISTS tournaments_status_check;

ALTER TABLE tournaments
    ADD CONSTRAINT tournaments_status_check CHECK (
        status IN (
            'draft',
            'scheduled',
            'registration_open',
            'registration_closed',
            'started',
            'completed',
            'cancelled'
        )
    );



