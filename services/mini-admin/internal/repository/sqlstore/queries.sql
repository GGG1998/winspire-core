-- name: CreateGame :one
-- Creates a new game record
INSERT INTO games (
    slug,
    name,
    description,
    logo_url,
    s3_path,
    version,
    versioning_enabled
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, slug, name, description, logo_url, s3_path, version, versioning_enabled, is_active, created_at, updated_at;

-- name: GetGameByID :one
-- Retrieves a game by its ID
SELECT id, slug, name, description, logo_url, s3_path, version, versioning_enabled, is_active, created_at, updated_at
FROM games
WHERE id = $1;

-- name: GetGameBySlug :one
-- Retrieves an active game by its slug
SELECT id, slug, name, description, logo_url, s3_path, version, versioning_enabled, is_active, created_at, updated_at
FROM games
WHERE slug = $1 AND is_active = true;

-- name: ListGames :many
-- Lists all active games
SELECT id, slug, name, description, logo_url, s3_path, version, versioning_enabled, is_active, created_at, updated_at
FROM games
WHERE is_active = true
ORDER BY created_at DESC;

-- name: ListAllGames :many
-- Lists all games (including inactive)
SELECT id, slug, name, description, logo_url, s3_path, version, versioning_enabled, is_active, created_at, updated_at
FROM games
ORDER BY created_at DESC;

-- name: UpdateGame :exec
-- Updates a game record
UPDATE games
SET
    slug = COALESCE(sqlc.narg('slug'), slug),
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    logo_url = COALESCE(sqlc.narg('logo_url'), logo_url),
    s3_path = COALESCE(sqlc.narg('s3_path'), s3_path),
    version = COALESCE(sqlc.narg('version'), version),
    versioning_enabled = COALESCE(sqlc.narg('versioning_enabled'), versioning_enabled),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- name: DeleteGame :exec
-- Soft deletes a game (sets is_active to false)
UPDATE games
SET is_active = false, updated_at = NOW()
WHERE id = $1;

-- name: HardDeleteGame :exec
-- Permanently deletes a game
DELETE FROM games
WHERE id = $1;

