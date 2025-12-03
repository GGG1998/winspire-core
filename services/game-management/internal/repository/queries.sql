-- ============================================================================
-- GAME QUERIES
-- Generated code will be used by internal/repository/game.go
-- ============================================================================

-- name: CreateGame :one
-- Creates a new game and returns the created record
INSERT INTO games (
    game_integration_id,
    slug,
    name,
    description,
    logo_url,
    storage_path,
    version
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, game_integration_id, slug, name, description, logo_url, storage_path, version, is_active, created_at, updated_at;

-- name: UpdateGame :exec
-- Updates a game's information using optional parameters
UPDATE games
SET
    game_integration_id = COALESCE(sqlc.narg('game_integration_id'), game_integration_id),
    slug = COALESCE(sqlc.narg('slug'), slug),
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    logo_url = COALESCE(sqlc.narg('logo_url'), logo_url),
    storage_path = COALESCE(sqlc.narg('storage_path'), storage_path),
    version = COALESCE(sqlc.narg('version'), version),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteGame :exec
-- Soft delete - sets is_active to false
UPDATE games
SET
    is_active = false,
    updated_at = NOW()
WHERE id = $1;

-- name: HardDeleteGame :exec
-- Hard delete - permanently removes game (use with caution)
DELETE FROM games WHERE id = $1;

-- name: GetGameByID :one
-- Retrieves a game by its ID
SELECT id, game_integration_id, slug, name, description, logo_url, storage_path, version, is_active, created_at, updated_at
FROM games
WHERE id = $1;

-- name: GetGameBySlug :one
-- Retrieves an active game by its slug
SELECT id, game_integration_id, slug, name, description, logo_url, storage_path, version, is_active, created_at, updated_at
FROM games
WHERE slug = $1 AND is_active = true;

-- name: GetGameByIntegrationID :one
-- Retrieves an active game by its game integration ID
SELECT id, game_integration_id, slug, name, description, logo_url, storage_path, version, is_active, created_at, updated_at
FROM games
WHERE game_integration_id = $1 AND is_active = true;

-- name: ListGames :many
-- Lists all active games ordered by name
SELECT id, game_integration_id, slug, name, description, logo_url, storage_path, version, is_active, created_at, updated_at
FROM games
WHERE is_active = true
ORDER BY name ASC;

-- name: ListAllGames :many
-- Lists all games including inactive ones (admin use)
SELECT id, game_integration_id, slug, name, description, logo_url, storage_path, version, is_active, created_at, updated_at
FROM games
ORDER BY name ASC;

-- name: GameExists :one
-- Checks if a game exists
SELECT EXISTS(
    SELECT 1 FROM games WHERE id = $1
) AS exists;

-- name: GameSlugExists :one
-- Checks if a slug is already taken
SELECT EXISTS(
    SELECT 1 FROM games WHERE slug = $1 AND is_active = true
) AS exists;

-- name: CountGames :one
-- Counts total active games
SELECT COUNT(*) FROM games WHERE is_active = true;

-- name: CountAllGames :one
-- Counts all games including inactive
SELECT COUNT(*) FROM games;

