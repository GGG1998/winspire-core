-- ============================================================================
-- Match Queries
-- ============================================================================

-- name: CreateMatch :one
INSERT INTO tournament_matches (
    id,
    round_id,
    match_number,
    next_match_id,
    participant1_id,
    participant2_id,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetMatchByID :one
SELECT * FROM tournament_matches
WHERE id = $1;

-- name: GetMatchesByRoundID :many
SELECT * FROM tournament_matches
WHERE round_id = $1
ORDER BY match_number ASC;

-- name: GetMatchesByTournamentAndRound :many
-- Get all matches for a specific tournament and round number
SELECT m.* FROM tournament_matches m
JOIN tournament_rounds r ON m.round_id = r.id
JOIN tournament_brackets b ON r.bracket_id = b.id
WHERE b.tournament_id = $1 AND r.round_number = $2
ORDER BY m.match_number ASC;

-- name: GetMatchesForPlayer :many
-- Get all matches for a specific player (participant1 or participant2)
SELECT * FROM tournament_matches
WHERE (participant1_id = $1 OR participant2_id = $1)
  AND status IN ('pending', 'ready', 'started', 'paused')
ORDER BY created_at DESC;

-- name: UpdateMatchStatus :exec
UPDATE tournament_matches
SET 
    status = $2,
    updated_at = NOW(),
    started_at = CASE 
        WHEN $2::VARCHAR = 'started' AND started_at IS NULL THEN NOW() 
        ELSE started_at 
    END,
    completed_at = CASE 
        WHEN $2::VARCHAR = 'completed' THEN NOW() 
        ELSE completed_at 
    END
WHERE id = $1;

-- name: UpdateMatchReady :exec
-- Update ready status for a specific participant
UPDATE tournament_matches
SET 
    participant1_ready = CASE 
        WHEN participant1_id = $2 THEN $3 
        ELSE participant1_ready 
    END,
    participant2_ready = CASE 
        WHEN participant2_id = $2 THEN $3 
        ELSE participant2_ready 
    END,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateMatchResult :exec
UPDATE tournament_matches
SET 
    winner_id = $2,
    score_player1 = $3,
    score_player2 = $4,
    result_source = $5,
    status = 'completed',
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateMatchDisconnect :exec
-- Track player disconnect for CS:GO-style handling
UPDATE tournament_matches
SET 
    disconnected_player_id = $2,
    disconnected_at = $3,
    status = 'paused',
    updated_at = NOW()
WHERE id = $1;

-- name: ClearMatchDisconnect :exec
-- Clear disconnect tracking when player reconnects
UPDATE tournament_matches
SET 
    disconnected_player_id = NULL,
    disconnected_at = NULL,
    status = 'started',
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateMatchScore :exec
-- Update scores during match (for disconnect point adjustments)
UPDATE tournament_matches
SET 
    score_player1 = $2,
    score_player2 = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateMatchGameAPIInfo :exec
-- Track Game API polling attempts
UPDATE tournament_matches
SET 
    game_api_match_id = $2,
    game_api_poll_attempts = $3,
    game_api_last_poll = $4,
    updated_at = NOW()
WHERE id = $1;

-- name: GetMatchesForGameAPIPolling :many
-- Get matches that need Game API polling
SELECT * FROM tournament_matches
WHERE status = 'started'
  AND game_api_match_id IS NOT NULL
  AND result_source IS NULL
  AND (game_api_last_poll IS NULL OR game_api_last_poll < NOW() - INTERVAL '5 seconds')
ORDER BY game_api_last_poll ASC NULLS FIRST
LIMIT $1;

-- name: AssignWinnerToNextMatch :exec
-- Assign winner to next match's participant slot
-- Handles both NULL and uuid.Nil (00000000-0000-0000-0000-000000000000) as empty slots
-- Later round matches are created with participant1_id = uuid.Nil (not NULL) as placeholder
UPDATE tournament_matches
SET
    participant1_id = CASE
        WHEN participant1_id IS NULL
             OR participant1_id = '00000000-0000-0000-0000-000000000000' THEN $2
        ELSE participant1_id
    END,
    participant2_id = CASE
        WHEN participant1_id IS NOT NULL
             AND participant1_id != '00000000-0000-0000-0000-000000000000'
             AND (participant2_id IS NULL OR participant2_id = '00000000-0000-0000-0000-000000000000')
             AND participant1_id != $2 THEN $2
        ELSE participant2_id
    END,
    updated_at = NOW()
WHERE id = $1;

-- name: GetFinalMatch :one
-- Get the final match (next_match_id IS NULL)
SELECT * FROM tournament_matches
WHERE round_id IN (
    SELECT id FROM tournament_rounds WHERE bracket_id = (
        SELECT id FROM tournament_brackets WHERE tournament_id = $1
    )
)
AND next_match_id IS NULL;

-- name: UpdateDisconnectedPlayerOnly :exec
-- Update disconnected player info without changing status
UPDATE tournament_matches
SET disconnected_player_id = $2, disconnected_at = $3, updated_at = NOW()
WHERE id = $1;

-- name: ClearDisconnectedPlayerInfo :exec
-- Clear disconnected player info
UPDATE tournament_matches
SET disconnected_player_id = NULL, disconnected_at = NULL, updated_at = NOW()
WHERE id = $1;

-- name: MarkGameLoadedAndCheckBoth :one
-- Atomically mark game loaded and transition to loading state if pending
-- New flow: pending → loading → ready → started
-- Uses idempotency check - only updates if player not already marked as loaded
UPDATE tournament_matches
SET
    status = 'loading',
    participant1_game_loaded = CASE
        WHEN participant1_id = $2 THEN true
        ELSE participant1_game_loaded
    END,
    participant2_game_loaded = CASE
        WHEN participant2_id = $2 THEN true
        ELSE participant2_game_loaded
    END,
    updated_at = NOW()
WHERE id = $1
  AND status IN ('pending', 'loading')
  AND (
    (participant1_id = $2 AND NOT participant1_game_loaded)
    OR
    (participant2_id = $2 AND NOT participant2_game_loaded)
  )
RETURNING 
    id,
    round_id,
    match_number,
    next_match_id,
    participant1_id,
    participant2_id,
    status,
    participant1_ready,
    participant2_ready,
    participant1_game_loaded,
    participant2_game_loaded,
    winner_id,
    score_player1,
    score_player2,
    result_source,
    disconnected_player_id,
    disconnected_at,
    game_api_match_id,
    game_api_poll_attempts,
    game_api_last_poll,
    version,
    created_at,
    started_at,
    completed_at,
    updated_at;

-- name: UpdateMatchStatusWithVersion :one
-- Update match status with optimistic locking
UPDATE tournament_matches
SET 
    status = $2,
    version = version + 1,
    updated_at = NOW(),
    started_at = CASE 
        WHEN $2::VARCHAR = 'started' AND started_at IS NULL THEN NOW() 
        ELSE started_at 
    END,
    completed_at = CASE 
        WHEN $2::VARCHAR = 'completed' THEN NOW() 
        ELSE completed_at 
    END
WHERE id = $1 AND version = $3
RETURNING *;

-- name: FindLoadingMatchesOlderThan :many
-- Find matches in 'loading' status older than specified timestamp (for timeout monitoring)
SELECT * FROM tournament_matches
WHERE status = 'loading'
  AND updated_at < $1
ORDER BY updated_at ASC;

-- name: HasParticipantLostInTournament :one
-- Check if a participant has lost any match in the specified tournament
SELECT EXISTS(
    SELECT 1 FROM tournament_matches m
    JOIN tournament_rounds r ON m.round_id = r.id
    JOIN tournament_brackets b ON r.bracket_id = b.id
    WHERE b.tournament_id = $1
      AND m.status = 'completed'
      AND (
        (m.participant1_id = $2 AND m.winner_id != $2) OR
        (m.participant2_id = $2 AND m.winner_id != $2)
      )
) AS has_lost;

