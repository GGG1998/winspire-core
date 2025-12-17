package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
	pgtypeconv "github.com/winspire/winspire-core/libs/go/pgtype"
)

// RoundRepository handles round persistence
type RoundRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Round, error)
	GetByBracketID(ctx context.Context, bracketID uuid.UUID) ([]domain.Round, error)
	GetLatestByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*domain.Round, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.RoundStatus) error
}

type roundRepository struct {
	queries *sqlc.Queries
}

// NewRoundRepository creates a new round repository
func NewRoundRepository(queries *sqlc.Queries) RoundRepository {
	return &roundRepository{
		queries: queries,
	}
}

// GetByID retrieves a round by its ID
func (r *roundRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Round, error) {
	row, err := r.queries.GetRoundByID(ctx, pgtypeconv.UUIDToPgtype(id))
	if err != nil {
		return nil, fmt.Errorf("get round by ID: %w", err)
	}

	return &domain.Round{
		ID:           pgtypeconv.PgtypeToUUID(row.ID),
		BracketID:    pgtypeconv.PgtypeToUUID(row.BracketID),
		RoundNumber:  int(row.RoundNumber),           // int32 -> int
		RoundName:    row.RoundName,                  // string directly
		MatchesCount: int(row.MatchesCount),          // int32 -> int
		Status:       domain.RoundStatus(row.Status), // string directly
		StartedAt:    pgtypeconv.PgtypeToTimePtr(row.StartedAt),
		CompletedAt:  pgtypeconv.PgtypeToTimePtr(row.CompletedAt),
	}, nil
}

// GetByBracketID retrieves all rounds for a bracket
func (r *roundRepository) GetByBracketID(ctx context.Context, bracketID uuid.UUID) ([]domain.Round, error) {
	rows, err := r.queries.GetRoundsByBracketID(ctx, pgtypeconv.UUIDToPgtype(bracketID))
	if err != nil {
		return nil, fmt.Errorf("get rounds by bracket ID: %w", err)
	}

	rounds := make([]domain.Round, len(rows))
	for i, row := range rows {
		rounds[i] = domain.Round{
			ID:           pgtypeconv.PgtypeToUUID(row.ID),
			BracketID:    pgtypeconv.PgtypeToUUID(row.BracketID),
			RoundNumber:  int(row.RoundNumber),           // int32 -> int
			RoundName:    row.RoundName,                  // string directly
			MatchesCount: int(row.MatchesCount),          // int32 -> int
			Status:       domain.RoundStatus(row.Status), // string directly
			StartedAt:    pgtypeconv.PgtypeToTimePtr(row.StartedAt),
			CompletedAt:  pgtypeconv.PgtypeToTimePtr(row.CompletedAt),
		}
	}

	return rounds, nil
}

// GetLatestByTournamentID retrieves the latest (highest round_number) round for a tournament
func (r *roundRepository) GetLatestByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*domain.Round, error) {
	row, err := r.queries.GetLatestRoundByTournamentID(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return nil, fmt.Errorf("get latest round by tournament ID: %w", err)
	}

	return &domain.Round{
		ID:           pgtypeconv.PgtypeToUUID(row.ID),
		BracketID:    pgtypeconv.PgtypeToUUID(row.BracketID),
		RoundNumber:  int(row.RoundNumber),
		RoundName:    row.RoundName,
		MatchesCount: int(row.MatchesCount),
		Status:       domain.RoundStatus(row.Status),
		StartedAt:    pgtypeconv.PgtypeToTimePtr(row.StartedAt),
		CompletedAt:  pgtypeconv.PgtypeToTimePtr(row.CompletedAt),
	}, nil
}

// UpdateStatus updates a round's status
func (r *roundRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.RoundStatus) error {
	err := r.queries.UpdateRoundStatus(ctx, sqlc.UpdateRoundStatusParams{
		ID:     pgtypeconv.UUIDToPgtype(id),
		Status: string(status), // Status is string in SQLC (NOT NULL field)
	})
	if err != nil {
		return fmt.Errorf("update round status: %w", err)
	}

	return nil
}
