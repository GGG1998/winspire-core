-- ============================================================================
-- HOST QUERIES
-- Generated code will be used by internal/repository/host.go
-- ============================================================================

-- name: GetHostByID :one
-- Retrieves a host by its ID
SELECT id, name, description, logo_url, website_url, created_at, updated_at
FROM hosts
WHERE id = $1;

-- name: CreateHost :one
-- Creates a new host and returns the created record
INSERT INTO hosts (id, name, description, logo_url, website_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, description, logo_url, website_url, created_at, updated_at;

-- name: UpdateHost :one
-- Updates a host's information
UPDATE hosts
SET 
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    logo_url = COALESCE(sqlc.narg('logo_url'), logo_url),
    website_url = COALESCE(sqlc.narg('website_url'), website_url),
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, description, logo_url, website_url, created_at, updated_at;

-- name: DeleteHost :exec
-- Deletes a host (cascades to host_members due to FK constraint)
DELETE FROM hosts
WHERE id = $1;

-- name: ListHosts :many
-- Lists all hosts (optionally paginated)
SELECT id, name, description, logo_url, website_url, created_at, updated_at
FROM hosts
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- ============================================================================
-- HOST MEMBERS QUERIES
-- ============================================================================

-- name: IsUserHostAdmin :one
-- Checks if a user is an admin/owner of the specified host
SELECT EXISTS(
    SELECT 1 FROM host_members
    WHERE host_id = $1
    AND user_id = $2
    AND role IN ('admin', 'owner')
) AS is_admin;

-- name: GetUserHostRole :one
-- Gets the user's role for a specific host
SELECT role
FROM host_members
WHERE host_id = $1 AND user_id = $2;

-- name: AddHostMember :one
-- Adds a user to a host with a specified role
INSERT INTO host_members (host_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (host_id, user_id) DO UPDATE
SET role = EXCLUDED.role, updated_at = NOW()
RETURNING id, host_id, user_id, role, created_at, updated_at;

-- name: RemoveHostMember :exec
-- Removes a user from a host
DELETE FROM host_members
WHERE host_id = $1 AND user_id = $2;

-- name: GetHostMembers :many
-- Lists all members of a host
SELECT hm.id, hm.host_id, hm.user_id, hm.role, hm.created_at, hm.updated_at
FROM host_members hm
WHERE hm.host_id = $1
ORDER BY hm.created_at DESC;

-- name: GetUserHosts :many
-- Returns all hosts where the user is a member
SELECT h.id, h.name, h.description, h.logo_url, h.website_url, h.created_at, h.updated_at
FROM hosts h
INNER JOIN host_members hm ON h.id = hm.host_id
WHERE hm.user_id = $1
ORDER BY h.created_at DESC;

-- name: GetUserHostsWithRole :many
-- Returns all hosts where the user is a member, including their role
SELECT 
    h.id, 
    h.name, 
    h.description, 
    h.logo_url, 
    h.website_url, 
    h.created_at, 
    h.updated_at,
    hm.role as member_role
FROM hosts h
INNER JOIN host_members hm ON h.id = hm.host_id
WHERE hm.user_id = $1
ORDER BY h.created_at DESC;

-- name: HostExists :one
-- Checks if a host exists
SELECT EXISTS(
    SELECT 1 FROM hosts WHERE id = $1
) AS exists;

-- name: CreateHostWithOwner :one
-- Creates a host and adds the creator as owner in a single transaction
-- Note: This must be called within a transaction (pgx.Tx)
WITH new_host AS (
    INSERT INTO hosts (id, name, description, logo_url, website_url)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id, name, description, logo_url, website_url, created_at, updated_at
),
new_member AS (
    INSERT INTO host_members (host_id, user_id, role)
    SELECT id, $6, 'owner'
    FROM new_host
    RETURNING id
)
SELECT * FROM new_host;

-- ============================================================================
-- TOURNAMENT QUERIES
-- Generated code will be used by internal/repository/tournament.go
-- ============================================================================

-- name: GetTournamentByID :one
-- Retrieves a tournament by its ID
SELECT 
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize,
    created_at, updated_at
FROM tournaments
WHERE id = $1;

-- name: GetTournamentByHostAndID :one
-- Retrieves a tournament scoped to a host
SELECT 
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize,
    created_at, updated_at
FROM tournaments
WHERE host_id = $1 AND id = $2;

-- name: ListTournamentsByHostID :many
-- Lists all tournaments for a specific host
SELECT 
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    created_at, updated_at
FROM tournaments
WHERE host_id = $1
ORDER BY created_at DESC;

-- name: ListTournamentsByHostIDWithStatus :many
-- Lists tournaments for a host filtered by status
SELECT 
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    created_at, updated_at
FROM tournaments
WHERE host_id = $1 AND status = $2
ORDER BY created_at DESC;

-- name: CreateTournament :one
-- Creates a new tournament
INSERT INTO tournaments (
    host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING 
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize,
    created_at, updated_at;

-- name: UpdateTournament :one
-- Updates tournament information with optional fields
-- Deprecated: Use SaveTournament for DDD aggregate persistence
UPDATE tournaments
SET 
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    scheduled_start_time_at = COALESCE(sqlc.narg('scheduled_start_time_at'), scheduled_start_time_at),
    minimum_team_count = COALESCE(sqlc.narg('minimum_team_count'), minimum_team_count),
    maximum_team_count = COALESCE(sqlc.narg('maximum_team_count'), maximum_team_count),
    updated_at = NOW()
WHERE id = $1
RETURNING 
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    created_at, updated_at;

-- name: SaveTournament :one
-- Saves the full tournament aggregate state (DDD pattern)
-- Used after applying aggregate business methods
UPDATE tournaments
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
    updated_at = NOW()
WHERE id = $1
RETURNING 
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize,
    created_at, updated_at;

-- name: UpdateTournamentStatus :exec
-- Updates tournament status
UPDATE tournaments
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: StartTournament :exec
-- Marks tournament as started
UPDATE tournaments
SET 
    status = 'started',
    actual_start_time_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: CompleteTournament :exec
-- Marks tournament as completed
UPDATE tournaments
SET 
    status = 'completed',
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: CancelTournament :exec
-- Marks tournament as cancelled
UPDATE tournaments
SET 
    status = 'cancelled',
    cancelled_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteTournament :exec
-- Deletes a tournament
DELETE FROM tournaments
WHERE id = $1;

-- name: TournamentExists :one
-- Checks if a tournament exists
SELECT EXISTS(
    SELECT 1 FROM tournaments WHERE id = $1
) AS exists;

-- name: GetTournamentByExternalID :one
-- Retrieves a tournament by external ID for idempotency
SELECT 
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    created_at, updated_at
FROM tournaments
WHERE external_id = $1 AND host_id = $2;

3-- name: ListTournamentsByStatus :many
-- Lists tournaments by status for scheduler
SELECT 
    id, host_id, name, description, external_id,
    status, scheduled_start_time_at, registration_window_open_at,
    actual_start_time_at, completed_at, cancelled_at,
    minimum_team_count, maximum_team_count, team_size, auto_force_ready,
    game_id, space_id, template_id,
    ready_window, prize,
    created_at, updated_at
FROM tournaments
WHERE status = ANY($1::text[])
ORDER BY scheduled_start_time_at ASC NULLS LAST, created_at ASC;

-- ============================================================================
-- TOURNAMENT REGISTRATION QUERIES
-- Scalable participant management for tournaments (10K+ participants)
-- ============================================================================

-- name: CreateTournamentRegistration :one
-- Registers a user for a tournament
INSERT INTO tournament_registrations (
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
-- Gets a specific registration
SELECT 
    id, tournament_id, user_id, team_id,
    status, display_name, avatar_url,
    checked_in_at, is_ready,
    registered_at, updated_at
FROM tournament_registrations
WHERE tournament_id = $1 AND user_id = $2;

-- name: ListTournamentRegistrations :many
-- Lists registrations for a tournament (paginated, newest first)
SELECT 
    id, tournament_id, user_id, team_id,
    status, display_name, avatar_url,
    checked_in_at, is_ready,
    registered_at, updated_at
FROM tournament_registrations
WHERE tournament_id = $1
ORDER BY registered_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTournamentRegistrations :one
-- Counts total registrations for a tournament
SELECT COUNT(*) AS count
FROM tournament_registrations
WHERE tournament_id = $1;

-- name: CountTournamentRegistrationsByStatus :one
-- Counts registrations by status
SELECT COUNT(*) AS count
FROM tournament_registrations
WHERE tournament_id = $1 AND status = $2;

-- name: CountReadyParticipants :one
-- Counts participants who are ready
SELECT COUNT(*) AS count
FROM tournament_registrations
WHERE tournament_id = $1 AND is_ready = true;

-- name: UpdateRegistrationStatus :exec
-- Updates registration status
UPDATE tournament_registrations
SET status = $2, updated_at = NOW()
WHERE tournament_id = $1 AND user_id = $3;

-- name: UpdateRegistrationReady :exec
-- Updates ready status
UPDATE tournament_registrations
SET is_ready = $2, updated_at = NOW()
WHERE tournament_id = $1 AND user_id = $3;

-- name: CheckInParticipant :exec
-- Marks participant as checked in
UPDATE tournament_registrations
SET 
    status = 'checked_in',
    checked_in_at = NOW(),
    is_ready = true,
    updated_at = NOW()
WHERE tournament_id = $1 AND user_id = $2;

-- name: WithdrawRegistration :exec
-- Withdraws a registration
UPDATE tournament_registrations
SET 
    status = 'withdrawn',
    is_ready = false,
    updated_at = NOW()
WHERE tournament_id = $1 AND user_id = $2;

-- name: DeleteRegistration :exec
-- Deletes a registration (hard delete)
DELETE FROM tournament_registrations
WHERE tournament_id = $1 AND user_id = $2;

-- name: GetUserTournamentRegistrations :many
-- Gets all tournaments a user is registered in
SELECT 
    id, tournament_id, user_id, team_id,
    status, display_name, avatar_url,
    checked_in_at, is_ready,
    registered_at, updated_at
FROM tournament_registrations
WHERE user_id = $1
ORDER BY registered_at DESC
LIMIT $2 OFFSET $3;

