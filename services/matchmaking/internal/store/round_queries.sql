-- ============================================================================
-- Round Queries
-- ============================================================================

-- name: CreateRound :one
INSERT INTO tournament_rounds (
    bracket_id,
    round_number,
    round_name,
    matches_count,
    status
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetRoundByID :one
SELECT * FROM tournament_rounds
WHERE id = $1;

-- name: GetRoundsByBracketID :many
SELECT * FROM tournament_rounds
WHERE bracket_id = $1
ORDER BY round_number ASC;

-- name: GetRoundByBracketAndNumber :one
SELECT * FROM tournament_rounds
WHERE bracket_id = $1 AND round_number = $2;

-- name: UpdateRoundStatus :exec
UPDATE tournament_rounds
SET 
    status = $2,
    started_at = CASE 
        WHEN $2 = 'in_progress' AND started_at IS NULL THEN NOW() 
        ELSE started_at 
    END,
    completed_at = CASE 
        WHEN $2 = 'completed' THEN NOW() 
        ELSE completed_at 
    END
WHERE id = $1;

-- name: GetActiveRounds :many
-- Get all active rounds (in_progress status)
SELECT r.*, b.tournament_id
FROM tournament_rounds r
JOIN tournament_brackets b ON b.id = r.bracket_id
WHERE r.status = 'in_progress'
ORDER BY b.tournament_id, r.round_number;

