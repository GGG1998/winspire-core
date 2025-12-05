-- SQLC queries for pre-lobby operations
-- Feature: Tournament Pre-Lobby Backend

-- ============================================================================
-- PRE-LOBBY QUERIES
-- ============================================================================

-- name: CreatePreLobby :one
-- Creates a new pre-lobby for a tournament
INSERT INTO prelobbies (tournament_id, min_participants)
VALUES ($1, $2)
RETURNING *;

-- name: GetPreLobbyByTournament :one
-- Gets pre-lobby state by tournament ID
SELECT * FROM prelobbies WHERE tournament_id = $1;

-- name: GetPreLobbyByID :one
-- Gets pre-lobby state by pre-lobby ID
SELECT * FROM prelobbies WHERE id = $1;

-- name: UpdatePreLobbyStatus :one
-- Updates pre-lobby status
UPDATE prelobbies 
SET status = $2, updated_at = NOW()
WHERE tournament_id = $1
RETURNING *;

-- name: StartGracePeriod :one
-- Transitions pre-lobby to grace_period status with 30-second window
UPDATE prelobbies 
SET status = 'grace_period', 
    grace_period_start = NOW(),
    grace_period_end = NOW() + INTERVAL '30 seconds',
    updated_at = NOW()
WHERE tournament_id = $1 AND status = 'waiting'
RETURNING *;

-- name: GetActiveGracePeriods :many
-- Gets all pre-lobbies currently in grace period (for recovery on startup)
SELECT * FROM prelobbies 
WHERE status = 'grace_period' AND grace_period_end > NOW();

-- name: GetPreLobbiesByStatus :many
-- Gets all pre-lobbies with a specific status
SELECT * FROM prelobbies WHERE status = $1;

-- name: DeletePreLobby :exec
-- Deletes a pre-lobby (cascades to snapshots and activity feed)
DELETE FROM prelobbies WHERE tournament_id = $1;

-- ============================================================================
-- PARTICIPANT SNAPSHOT QUERIES
-- ============================================================================

-- name: CreateParticipantSnapshot :one
-- Creates immutable participant snapshot when grace period ends
INSERT INTO prelobby_participant_snapshots (tournament_id, participants, participant_count)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetParticipantSnapshot :one
-- Gets participant snapshot for a tournament
SELECT * FROM prelobby_participant_snapshots WHERE tournament_id = $1;

-- name: ParticipantSnapshotExists :one
-- Checks if a snapshot already exists for a tournament
SELECT EXISTS(SELECT 1 FROM prelobby_participant_snapshots WHERE tournament_id = $1) AS exists;

-- ============================================================================
-- ACTIVITY FEED QUERIES
-- ============================================================================

-- name: AddActivityFeedEvent :one
-- Adds an event to the activity feed
INSERT INTO prelobby_activity_feed (tournament_id, event_type, message, participant_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRecentActivityFeed :many
-- Gets the 20 most recent activity feed events for a tournament
SELECT * FROM prelobby_activity_feed 
WHERE tournament_id = $1 
ORDER BY created_at DESC 
LIMIT 20;

-- name: GetActivityFeedCount :one
-- Gets total count of activity feed events for a tournament
SELECT COUNT(*) AS count FROM prelobby_activity_feed WHERE tournament_id = $1;

-- name: DeleteOldActivityFeedEvents :exec
-- Deletes activity feed events older than a certain threshold (cleanup)
DELETE FROM prelobby_activity_feed 
WHERE tournament_id = $1 
AND created_at < NOW() - INTERVAL '24 hours';

