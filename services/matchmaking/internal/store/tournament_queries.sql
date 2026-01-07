-- ============================================================================
-- TOURNAMENT SCHEMA QUERIES
-- Generated code for tournament bounded context (merged from tournament service)
-- ============================================================================

-- ============================================================================
-- HOST QUERIES
-- ============================================================================

-- name: GetTournamentHostByID :one
SELECT id, name, description, logo_url, website_url, created_at, updated_at
FROM tournament.hosts
WHERE id = $1;

-- name: CreateTournamentHost :one
INSERT INTO tournament.hosts (name, description, logo_url, website_url)
VALUES ($1, $2, $3, $4)
RETURNING id, name, description, logo_url, website_url, created_at, updated_at;

-- name: UpdateTournamentHost :one
UPDATE tournament.hosts
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    logo_url = COALESCE(sqlc.narg('logo_url'), logo_url),
    website_url = COALESCE(sqlc.narg('website_url'), website_url),
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, description, logo_url, website_url, created_at, updated_at;

-- name: ListTournamentHosts :many
SELECT id, name, description, logo_url, website_url, created_at, updated_at
FROM tournament.hosts
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- ============================================================================
-- HOST MEMBERS QUERIES
-- ============================================================================

-- name: IsTournamentUserHostAdmin :one
SELECT EXISTS(
    SELECT 1 FROM tournament.host_members
    WHERE host_id = $1
    AND user_id = $2
    AND role IN ('admin', 'owner')
) AS is_admin;

-- name: GetTournamentUserHostRole :one
SELECT role
FROM tournament.host_members
WHERE host_id = $1 AND user_id = $2;

-- name: AddTournamentHostMember :one
INSERT INTO tournament.host_members (host_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (host_id, user_id) DO UPDATE
SET role = EXCLUDED.role, updated_at = NOW()
RETURNING id, host_id, user_id, role, created_at, updated_at;

-- name: RemoveTournamentHostMember :exec
DELETE FROM tournament.host_members
WHERE host_id = $1 AND user_id = $2;

-- name: GetTournamentHostMembers :many
SELECT id, host_id, user_id, role, created_at, updated_at
FROM tournament.host_members
WHERE host_id = $1
ORDER BY created_at DESC;

-- name: GetTournamentUserHosts :many
SELECT h.id, h.name, h.description, h.logo_url, h.website_url, h.created_at, h.updated_at
FROM tournament.hosts h
INNER JOIN tournament.host_members hm ON h.id = hm.host_id
WHERE hm.user_id = $1
ORDER BY h.created_at DESC;

-- name: GetTournamentUserHostsWithRole :many
SELECT
    h.id,
    h.name,
    h.description,
    h.logo_url,
    h.website_url,
    h.created_at,
    h.updated_at,
    hm.role as member_role
FROM tournament.hosts h
INNER JOIN tournament.host_members hm ON h.id = hm.host_id
WHERE hm.user_id = $1
ORDER BY h.created_at DESC;

-- name: CreateTournamentHostWithOwner :one
WITH new_host AS (
    INSERT INTO tournament.hosts (name, description, logo_url, website_url)
    VALUES ($1, $2, $3, $4)
    RETURNING id, name, description, logo_url, website_url, created_at, updated_at
),
new_member AS (
    INSERT INTO tournament.host_members (host_id, user_id, role)
    SELECT id, $5, 'owner'
    FROM new_host
    RETURNING id
)
SELECT * FROM new_host;

-- name: CreateTournamentHostWithIDAndOwner :one
-- Creates a host with a specific ID (used for personal hosts where host_id = user_id)
WITH new_host AS (
    INSERT INTO tournament.hosts (id, name, description, logo_url, website_url)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id, name, description, logo_url, website_url, created_at, updated_at
),
new_member AS (
    INSERT INTO tournament.host_members (host_id, user_id, role)
    SELECT id, $6, 'owner'
    FROM new_host
    RETURNING id
)
SELECT * FROM new_host;

-- ============================================================================
-- TOURNAMENT QUERIES
-- ============================================================================

-- name: GetTournamentByID :one
SELECT
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize, game_snapshot,
    created_at, updated_at
FROM tournament.tournaments
WHERE id = $1;

-- name: GetTournamentByHostAndID :one
SELECT
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize, game_snapshot,
    created_at, updated_at
FROM tournament.tournaments
WHERE host_id = $1 AND id = $2;

-- name: ListTournamentsByHostID :many
SELECT
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize, game_snapshot,
    created_at, updated_at
FROM tournament.tournaments
WHERE host_id = $1
ORDER BY created_at DESC;

-- name: ListTournamentsByHostIDWithStatus :many
SELECT
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize, game_snapshot,
    created_at, updated_at
FROM tournament.tournaments
WHERE host_id = $1 AND status = $2
ORDER BY created_at DESC;

-- name: CreateTournament :one
INSERT INTO tournament.tournaments (
    host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id, game_snapshot
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize, game_snapshot,
    created_at, updated_at;

-- name: SaveTournament :one
UPDATE tournament.tournaments
SET
    name = $2,
    description = $3,
    status = $4,
    scheduled_start_time_at = $5,
    actual_start_time_at = $6,
    completed_at = $7,
    cancelled_at = $8,
    minimum_team_count = $9,
    maximum_team_count = $10,
    ready_window = $11,
    prize = $12,
    game_snapshot = $13,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize, game_snapshot,
    created_at, updated_at;

-- name: UpdateTournamentStatus :exec
UPDATE tournament.tournaments
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: StartTournament :exec
UPDATE tournament.tournaments
SET
    status = 'started',
    actual_start_time_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: CompleteTournament :exec
UPDATE tournament.tournaments
SET
    status = 'completed',
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: CancelTournament :exec
UPDATE tournament.tournaments
SET
    status = 'cancelled',
    cancelled_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteTournament :exec
DELETE FROM tournament.tournaments
WHERE id = $1;

-- name: TournamentExists :one
SELECT EXISTS(
    SELECT 1 FROM tournament.tournaments WHERE id = $1
) AS exists;

-- name: GetTournamentByExternalID :one
SELECT
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize, game_snapshot,
    created_at, updated_at
FROM tournament.tournaments
WHERE external_id = $1 AND host_id = $2;

-- name: ListTournamentsByStatus :many
SELECT
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize, game_snapshot,
    created_at, updated_at
FROM tournament.tournaments
WHERE status = ANY($1::text[])
ORDER BY scheduled_start_time_at ASC NULLS LAST, created_at ASC;

-- ============================================================================
-- TOURNAMENT REGISTRATION QUERIES
-- ============================================================================

-- name: CreateTournamentRegistration :one
INSERT INTO tournament.registrations (
    tournament_id, user_id, team_id,
    status, display_name, avatar_url
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (tournament_id, user_id) DO UPDATE
SET
    status = EXCLUDED.status,
    display_name = EXCLUDED.display_name,
    avatar_url = EXCLUDED.avatar_url,
    updated_at = NOW()
RETURNING
    id, tournament_id, user_id, team_id,
    status, display_name, avatar_url,
    checked_in_at, is_ready,
    registered_at, updated_at;

-- name: GetTournamentRegistration :one
SELECT
    id, tournament_id, user_id, team_id,
    status, display_name, avatar_url,
    checked_in_at, is_ready,
    registered_at, updated_at
FROM tournament.registrations
WHERE tournament_id = $1 AND user_id = $2;

-- name: ListTournamentRegistrations :many
SELECT
    id, tournament_id, user_id, team_id,
    status, display_name, avatar_url,
    checked_in_at, is_ready,
    registered_at, updated_at
FROM tournament.registrations
WHERE tournament_id = $1
ORDER BY registered_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllTournamentRegistrations :many
SELECT
    id, tournament_id, user_id, team_id,
    status, display_name, avatar_url,
    checked_in_at, is_ready,
    registered_at, updated_at
FROM tournament.registrations
WHERE tournament_id = $1
ORDER BY registered_at ASC;

-- name: ListConfirmedTournamentRegistrations :many
SELECT
    id, tournament_id, user_id, team_id,
    status, display_name, avatar_url,
    checked_in_at, is_ready,
    registered_at, updated_at
FROM tournament.registrations
WHERE tournament_id = $1 AND status IN ('confirmed', 'checked_in')
ORDER BY registered_at ASC;

-- name: GetConfirmedTournamentUserIDs :many
SELECT user_id
FROM tournament.registrations
WHERE tournament_id = $1 AND status IN ('confirmed', 'checked_in')
ORDER BY registered_at ASC;

-- name: CountTournamentRegistrations :one
SELECT COUNT(*) AS count
FROM tournament.registrations
WHERE tournament_id = $1;

-- name: CountTournamentRegistrationsByStatus :one
SELECT COUNT(*) AS count
FROM tournament.registrations
WHERE tournament_id = $1 AND status = $2;

-- name: CountConfirmedTournamentRegistrations :one
SELECT COUNT(*) AS count
FROM tournament.registrations
WHERE tournament_id = $1 AND status IN ('confirmed', 'checked_in');

-- name: CountReadyTournamentParticipants :one
SELECT COUNT(*) AS count
FROM tournament.registrations
WHERE tournament_id = $1 AND is_ready = true;

-- name: UpdateTournamentRegistrationStatus :exec
UPDATE tournament.registrations
SET status = $3, updated_at = NOW()
WHERE tournament_id = $1 AND user_id = $2;

-- name: UpdateTournamentRegistrationReady :exec
UPDATE tournament.registrations
SET is_ready = $3, updated_at = NOW()
WHERE tournament_id = $1 AND user_id = $2;

-- name: CheckInTournamentParticipant :exec
UPDATE tournament.registrations
SET
    status = 'checked_in',
    checked_in_at = NOW(),
    is_ready = true,
    updated_at = NOW()
WHERE tournament_id = $1 AND user_id = $2;

-- name: WithdrawTournamentRegistration :exec
UPDATE tournament.registrations
SET
    status = 'withdrawn',
    is_ready = false,
    updated_at = NOW()
WHERE tournament_id = $1 AND user_id = $2;

-- name: DeleteTournamentRegistration :exec
DELETE FROM tournament.registrations
WHERE tournament_id = $1 AND user_id = $2;

-- name: GetUserTournamentRegistrations :many
SELECT
    id, tournament_id, user_id, team_id,
    status, display_name, avatar_url,
    checked_in_at, is_ready,
    registered_at, updated_at
FROM tournament.registrations
WHERE user_id = $1
ORDER BY registered_at DESC
LIMIT $2 OFFSET $3;
