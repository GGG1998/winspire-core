package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/winspire/game-management/internal/repository/sqlstore"
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

// GameRepository handles database operations for games using sqlc-generated queries.
type GameRepository struct {
	pool    *pgxpool.Pool
	queries *sqlstore.Queries
}

// NewGameRepository creates a new GameRepository.
func NewGameRepository(pool *pgxpool.Pool) *GameRepository {
	return &GameRepository{
		pool:    pool,
		queries: sqlstore.New(pool),
	}
}

// Create inserts a new game into the database using sqlc-generated query.
func (r *GameRepository) Create(ctx context.Context, input CreateGameInput) (*Game, error) {
	sqlcGame, err := r.queries.CreateGame(ctx, sqlstore.CreateGameParams{
		GameIntegrationID: uuidToPgtypeNullable(input.GameIntegrationID),
		Slug:              input.Slug,
		Name:              input.Name,
		Description:       stringToPgtypeText(input.Description),
		LogoUrl:           stringToPgtypeText(input.LogoURL),
		StoragePath:       input.StoragePath,
		Version:           input.Version,
	})
	if err != nil {
		return nil, fmt.Errorf("create game: %w", err)
	}

	return sqlcGameToGame(sqlcGame), nil
}

// Update modifies an existing game in the database using sqlc-generated query.
func (r *GameRepository) Update(ctx context.Context, input UpdateGameInput) error {
	err := r.queries.UpdateGame(ctx, sqlstore.UpdateGameParams{
		ID:                uuidToPgtype(input.ID),
		GameIntegrationID: optionalUUIDToPgtype(input.GameIntegrationID),
		Slug:              optionalStringToPgtypeText(input.Slug),
		Name:              optionalStringToPgtypeText(input.Name),
		Description:       optionalStringToPgtypeText(input.Description),
		LogoUrl:           optionalStringToPgtypeText(input.LogoURL),
		StoragePath:       optionalStringToPgtypeText(input.StoragePath),
		Version:           optionalStringToPgtypeText(input.Version),
		IsActive:          optionalBoolToPgtype(input.IsActive),
	})
	if err != nil {
		return fmt.Errorf("update game: %w", err)
	}

	return nil
}

// Delete performs a soft delete on a game (sets is_active to false).
func (r *GameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteGame(ctx, uuidToPgtype(id))
	if err != nil {
		return fmt.Errorf("delete game: %w", err)
	}

	return nil
}

// GetByID retrieves a game by its ID using sqlc-generated query.
func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (*Game, error) {
	sqlcGame, err := r.queries.GetGameByID(ctx, uuidToPgtype(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("get game by id: %w", err)
	}

	return sqlcGameToGame(sqlcGame), nil
}

// GetBySlug retrieves an active game by its slug using sqlc-generated query.
func (r *GameRepository) GetBySlug(ctx context.Context, slug string) (*Game, error) {
	sqlcGame, err := r.queries.GetGameBySlug(ctx, slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("get game by slug: %w", err)
	}

	return sqlcGameToGame(sqlcGame), nil
}

// GetByIntegrationID retrieves a game by its game integration ID using sqlc-generated query.
func (r *GameRepository) GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) (*Game, error) {
	sqlcGame, err := r.queries.GetGameByIntegrationID(ctx, uuidToPgtype(integrationID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("get game by integration id: %w", err)
	}

	return sqlcGameToGame(sqlcGame), nil
}

// List retrieves all active games using sqlc-generated query.
func (r *GameRepository) List(ctx context.Context) ([]Game, error) {
	sqlcGames, err := r.queries.ListGames(ctx)
	if err != nil {
		return nil, fmt.Errorf("list games: %w", err)
	}

	games := make([]Game, 0, len(sqlcGames))
	for _, sg := range sqlcGames {
		games = append(games, *sqlcGameToGame(sg))
	}

	return games, nil
}

// ListAll retrieves all games (including inactive) using sqlc-generated query.
func (r *GameRepository) ListAll(ctx context.Context) ([]Game, error) {
	sqlcGames, err := r.queries.ListAllGames(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all games: %w", err)
	}
	games := make([]Game, 0, len(sqlcGames))
	for _, sg := range sqlcGames {
		games = append(games, *sqlcGameToGame(sg))
	}

	return games, nil
}

// ============================================================================
// Helper functions for type conversion between sqlc types and business types
// ============================================================================

// uuidToPgtype converts standard UUID to pgtype.UUID
func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

// uuidToPgtypeNullable converts *uuid.UUID to pgtype.UUID (nullable)
func uuidToPgtypeNullable(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{
		Bytes: *id,
		Valid: true,
	}
}

// pgtypeToUUID converts pgtype.UUID to standard UUID
func pgtypeToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.UUID{}
	}
	return id.Bytes
}

// pgtypeToUUIDNullable converts pgtype.UUID to *uuid.UUID
func pgtypeToUUIDNullable(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	result := uuid.UUID(id.Bytes)
	return &result
}

// stringToPgtypeText creates a pgtype.Text from *string
func stringToPgtypeText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: *s,
		Valid:  true,
	}
}

// optionalStringToPgtypeText creates a pgtype.Text from *string for optional updates
func optionalStringToPgtypeText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: *s,
		Valid:  true,
	}
}

// optionalUUIDToPgtype creates a pgtype.UUID from *uuid.UUID for optional updates
func optionalUUIDToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{
		Bytes: *id,
		Valid: true,
	}
}

// optionalBoolToPgtype creates a pgtype.Bool from *bool for optional updates
func optionalBoolToPgtype(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{
		Bool:  *b,
		Valid: true,
	}
}

// pgtypeTextToString converts pgtype.Text to *string (nil if not valid)
func pgtypeTextToString(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// pgtypeTimestampToString converts pgtype.Timestamptz to ISO 8601 string
func pgtypeTimestampToString(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("2006-01-02T15:04:05Z07:00")
}

// sqlcGameToGame converts sqlstore.Game to repository.Game
func sqlcGameToGame(g sqlstore.Game) *Game {
	return &Game{
		ID:                pgtypeToUUID(g.ID),
		GameIntegrationID: pgtypeToUUIDNullable(g.GameIntegrationID),
		Slug:              g.Slug,
		Name:              g.Name,
		Description:       pgtypeTextToString(g.Description),
		LogoURL:           pgtypeTextToString(g.LogoUrl),
		StoragePath:       g.StoragePath,
		Version:           g.Version,
		IsActive:          g.IsActive,
		CreatedAt:         pgtypeTimestampToString(g.CreatedAt),
		UpdatedAt:         pgtypeTimestampToString(g.UpdatedAt),
	}
}
