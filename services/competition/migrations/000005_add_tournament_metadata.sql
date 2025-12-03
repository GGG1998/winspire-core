-- Migration: Add tournament metadata fields
-- Purpose: Add ready_window and prize JSONB fields to tournaments table
-- Date: 2024-12-01

-- ============================================================================
-- ADD METADATA COLUMNS
-- ============================================================================

-- Add ready_window as JSONB to store check-in window
ALTER TABLE tournaments
ADD COLUMN ready_window JSONB;

-- Add prize as JSONB to store prize configuration
ALTER TABLE tournaments
ADD COLUMN prize JSONB;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON COLUMN tournaments.ready_window IS 'Check-in window as JSONB: { "startsAt": "ISO8601", "endsAt": "ISO8601" }';
COMMENT ON COLUMN tournaments.prize IS 'Prize configuration as JSONB: { "type": "custom|cash|points", "description": "...", "value": 0, "currency": "USD" }';



