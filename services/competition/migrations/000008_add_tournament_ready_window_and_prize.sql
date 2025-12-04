-- Migration: Add ready_window and prize JSONB columns to tournaments
-- Purpose: Align database schema with repository and sqlc models
-- Date: 2025-12-04

ALTER TABLE tournaments
    ADD COLUMN IF NOT EXISTS ready_window JSONB,
    ADD COLUMN IF NOT EXISTS prize JSONB;

COMMENT ON COLUMN tournaments.ready_window IS 'JSONB configuration for tournament ready/check-in window';
COMMENT ON COLUMN tournaments.prize IS 'JSONB configuration for tournament prize structure';


