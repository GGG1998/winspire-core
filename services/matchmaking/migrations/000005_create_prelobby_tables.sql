-- Migration: 000005_create_prelobby_tables.sql
-- Purpose: Create tables for tournament pre-lobby functionality
-- Date: 2025-12-05

-- ============================================================================
-- PRE-LOBBY TABLE
-- ============================================================================
-- Represents the waiting room state for a tournament before it starts.
-- Only one pre-lobby per tournament (tournament_id is UNIQUE).

CREATE TABLE IF NOT EXISTS prelobbies (
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

-- Index for status queries (find active pre-lobbies for recovery on startup)
CREATE INDEX IF NOT EXISTS idx_prelobbies_status 
    ON prelobbies(status) 
    WHERE status IN ('waiting', 'grace_period');

-- Index for tournament lookup
CREATE INDEX IF NOT EXISTS idx_prelobbies_tournament_id 
    ON prelobbies(tournament_id);

-- ============================================================================
-- PARTICIPANT SNAPSHOT TABLE
-- ============================================================================
-- Immutable record of participants present when grace period ended.
-- Used as the source of truth for bracket generation.

CREATE TABLE IF NOT EXISTS prelobby_participant_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL UNIQUE REFERENCES prelobbies(tournament_id) ON DELETE CASCADE,
    participants JSONB NOT NULL,
    participant_count INTEGER NOT NULL CHECK (participant_count >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- ACTIVITY FEED TABLE
-- ============================================================================
-- Chronological log of events in the pre-lobby (joins, leaves, system messages).
-- API returns only 20 most recent events per tournament.

CREATE TABLE IF NOT EXISTS prelobby_activity_feed (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES prelobbies(tournament_id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL CHECK (
        event_type IN ('participant_joined', 'participant_left', 'grace_period_started', 
                       'grace_period_ended', 'tournament_cancelled', 'system_message')
    ),
    message TEXT NOT NULL,
    participant_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for efficient retrieval of recent events (ordered by time DESC)
CREATE INDEX IF NOT EXISTS idx_activity_feed_tournament_time 
    ON prelobby_activity_feed(tournament_id, created_at DESC);

