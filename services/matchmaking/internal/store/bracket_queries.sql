-- ============================================================================
-- Bracket Queries
-- ============================================================================

-- name: CreateBracket :one
INSERT INTO tournament_brackets (
    tournament_id,
    total_rounds,
    total_matches,
    byes_count,
    generated_at
) VALUES (
    $1, $2, $3, $4, NOW()
) RETURNING *;

-- name: GetBracketByID :one
SELECT * FROM tournament_brackets
WHERE id = $1;

-- name: GetBracketByTournamentID :one
SELECT * FROM tournament_brackets
WHERE tournament_id = $1;

-- name: GetBracketWithRoundsAndMatches :many
-- Complex join query for bracket visualization
SELECT 
    b.id as bracket_id,
    b.tournament_id,
    b.total_rounds as bracket_total_rounds,
    b.total_matches as bracket_total_matches,
    b.byes_count,
    b.generated_at as bracket_generated_at,
    r.id as round_id,
    r.round_number,
    r.round_name,
    r.matches_count as round_matches_count,
    r.status as round_status,
    r.started_at as round_started_at,
    r.completed_at as round_completed_at,
    m.id as match_id,
    m.match_number,
    m.participant1_id,
    m.participant2_id,
    m.status as match_status,
    m.participant1_ready,
    m.participant2_ready,
    m.winner_id,
    m.score_player1,
    m.score_player2,
    m.result_source,
    m.next_match_id,
    m.created_at as match_created_at,
    m.started_at as match_started_at,
    m.completed_at as match_completed_at
FROM tournament_brackets b
LEFT JOIN tournament_rounds r ON r.bracket_id = b.id
LEFT JOIN tournament_matches m ON m.round_id = r.id
WHERE b.tournament_id = $1
ORDER BY r.round_number ASC, m.match_number ASC;

-- name: UpdateBracketCompletedAt :exec
UPDATE tournament_brackets
SET completed_at = $2
WHERE id = $1;

-- name: DeleteBracket :exec
DELETE FROM tournament_brackets
WHERE id = $1;

