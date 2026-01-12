package games

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Game represents a game entity for business logic.
type Game struct {
	ID                uuid.UUID
	GameIntegrationID *uuid.UUID
	Slug              string
	Name              string
	Description       *string
	LogoURL           *string
	StoragePath       string
	Version           string
	IsActive          bool
	CreatedAt         string
	UpdatedAt         string
}

// CreateGameInput contains the data needed to create a new game.
type CreateGameInput struct {
	GameIntegrationID *uuid.UUID
	Slug              string
	Name              string
	Description       *string
	LogoURL           *string
	StoragePath       string
	Version           string
}

// UpdateGameInput contains the data that can be updated for a game.
type UpdateGameInput struct {
	ID                uuid.UUID
	GameIntegrationID *uuid.UUID
	Slug              *string
	Name              *string
	Description       *string
	LogoURL           *string
	StoragePath       *string
	Version           *string
	IsActive          *bool
}

// GameRepository handles database operations for games.
// Uses direct pgx queries against Supabase PostgreSQL.
type GameRepository struct {
	pool *pgxpool.Pool
}

// NewGameRepository creates a new GameRepository with a Supabase connection pool.
func NewGameRepository(pool *pgxpool.Pool) *GameRepository {
	return &GameRepository{
		pool: pool,
	}
}

// Create inserts a new game into the database.
func (r *GameRepository) Create(ctx context.Context, input CreateGameInput) (*Game, error) {
	query := `
		INSERT INTO games (game_integration_id, slug, name, description, logo_url, storage_path, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, game_integration_id, slug, name, description, logo_url, storage_path, version, is_active, created_at, updated_at
	`

	var game Game
	var createdAt, updatedAt time.Time

	err := r.pool.QueryRow(ctx, query,
		input.GameIntegrationID,
		input.Slug,
		input.Name,
		input.Description,
		input.LogoURL,
		input.StoragePath,
		input.Version,
	).Scan(
		&game.ID,
		&game.GameIntegrationID,
		&game.Slug,
		&game.Name,
		&game.Description,
		&game.LogoURL,
		&game.StoragePath,
		&game.Version,
		&game.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create game: %w", err)
	}

	game.CreatedAt = timeToString(createdAt)
	game.UpdatedAt = timeToString(updatedAt)

	return &game, nil
}

// Update modifies an existing game in the database.
func (r *GameRepository) Update(ctx context.Context, input UpdateGameInput) error {
	query := `
		UPDATE games
		SET
			game_integration_id = COALESCE($2, game_integration_id),
			slug = COALESCE($3, slug),
			name = COALESCE($4, name),
			description = COALESCE($5, description),
			logo_url = COALESCE($6, logo_url),
			storage_path = COALESCE($7, storage_path),
			version = COALESCE($8, version),
			is_active = COALESCE($9, is_active),
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		input.ID,
		input.GameIntegrationID,
		input.Slug,
		input.Name,
		input.Description,
		input.LogoURL,
		input.StoragePath,
		input.Version,
		input.IsActive,
	)
	if err != nil {
		return fmt.Errorf("update game: %w", err)
	}

	return nil
}

// Delete performs a soft delete on a game (sets is_active to false).
func (r *GameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE games SET is_active = false, updated_at = NOW() WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete game: %w", err)
	}

	return nil
}

// GetByID retrieves a game by its ID.
func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (*Game, error) {
	query := `
		SELECT id, game_integration_id, slug, name, description, logo_url, storage_path, version, is_active, created_at, updated_at
		FROM games
		WHERE id = $1
	`

	game, err := r.scanGame(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("get game by id: %w", err)
	}

	return game, nil
}

// GetBySlug retrieves an active game by its slug.
func (r *GameRepository) GetBySlug(ctx context.Context, slug string) (*Game, error) {
	query := `
		SELECT id, game_integration_id, slug, name, description, logo_url, storage_path, version, is_active, created_at, updated_at
		FROM games
		WHERE slug = $1 AND is_active = true
	`

	game, err := r.scanGame(r.pool.QueryRow(ctx, query, slug))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("get game by slug: %w", err)
	}

	return game, nil
}

// GetByIntegrationID retrieves a game by its game integration ID.
func (r *GameRepository) GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) (*Game, error) {
	query := `
		SELECT id, game_integration_id, slug, name, description, logo_url, storage_path, version, is_active, created_at, updated_at
		FROM games
		WHERE game_integration_id = $1 AND is_active = true
	`

	game, err := r.scanGame(r.pool.QueryRow(ctx, query, integrationID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("get game by integration id: %w", err)
	}

	return game, nil
}

// List retrieves all active games.
func (r *GameRepository) List(ctx context.Context) ([]Game, error) {
	query := `
		SELECT id, game_integration_id, slug, name, description, logo_url, storage_path, version, is_active, created_at, updated_at
		FROM games
		WHERE is_active = true
		ORDER BY name
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list games: %w", err)
	}
	defer rows.Close()

	return r.scanGames(rows)
}

// ListAll retrieves all games (including inactive).
func (r *GameRepository) ListAll(ctx context.Context) ([]Game, error) {
	query := `
		SELECT id, game_integration_id, slug, name, description, logo_url, storage_path, version, is_active, created_at, updated_at
		FROM games
		ORDER BY name
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list all games: %w", err)
	}
	defer rows.Close()

	return r.scanGames(rows)
}

// ============================================================================
// Helper functions
// ============================================================================

// scanGame scans a single row into a Game struct.
func (r *GameRepository) scanGame(row pgx.Row) (*Game, error) {
	var game Game
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&game.ID,
		&game.GameIntegrationID,
		&game.Slug,
		&game.Name,
		&game.Description,
		&game.LogoURL,
		&game.StoragePath,
		&game.Version,
		&game.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	game.CreatedAt = timeToString(createdAt)
	game.UpdatedAt = timeToString(updatedAt)

	return &game, nil
}

// scanGames scans multiple rows into a slice of Game structs.
func (r *GameRepository) scanGames(rows pgx.Rows) ([]Game, error) {
	var games []Game

	for rows.Next() {
		var game Game
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&game.ID,
			&game.GameIntegrationID,
			&game.Slug,
			&game.Name,
			&game.Description,
			&game.LogoURL,
			&game.StoragePath,
			&game.Version,
			&game.IsActive,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan game row: %w", err)
		}

		game.CreatedAt = timeToString(createdAt)
		game.UpdatedAt = timeToString(updatedAt)

		games = append(games, game)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate games: %w", err)
	}

	return games, nil
}

// timeToString converts time.Time to ISO 8601 string
func timeToString(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z07:00")
}
