package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
	pgtypeconv "github.com/winspire/winspire-core/libs/go/pgtype"
)

// MatchRepository handles match persistence
type MatchRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Match, error)
	GetByRoundID(ctx context.Context, roundID uuid.UUID) ([]domain.Match, error)
	GetForPlayer(ctx context.Context, playerID uuid.UUID) ([]domain.Match, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.MatchStatus) error
	UpdateReady(ctx context.Context, id uuid.UUID, playerID uuid.UUID, ready bool) error
	UpdateResult(ctx context.Context, id uuid.UUID, winnerID uuid.UUID, scorePlayer1, scorePlayer2 int, source domain.ResultSource) error
	UpdateDisconnect(ctx context.Context, id uuid.UUID, playerID uuid.UUID, disconnectedAt *time.Time) error
	UpdateScore(ctx context.Context, id uuid.UUID, scorePlayer1, scorePlayer2 int) error
	UpdateDisconnectedPlayer(ctx context.Context, id uuid.UUID, playerID uuid.UUID, disconnectedAt time.Time) error
	ClearDisconnectedPlayer(ctx context.Context, id uuid.UUID) error
}

type matchRepository struct {
	queries *sqlc.Queries
}

// NewMatchRepository creates a new match repository
func NewMatchRepository(queries *sqlc.Queries) MatchRepository {
	return &matchRepository{
		queries: queries,
	}
}

// GetByID retrieves a match by its ID
func (r *matchRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Match, error) {
	row, err := r.queries.GetMatchByID(ctx, pgtypeconv.UUIDToPgtype(id))
	if err != nil {
		return nil, fmt.Errorf("get match by ID: %w", err)
	}

	return r.rowToMatch(row), nil
}

// GetByRoundID retrieves all matches for a round
func (r *matchRepository) GetByRoundID(ctx context.Context, roundID uuid.UUID) ([]domain.Match, error) {
	rows, err := r.queries.GetMatchesByRoundID(ctx, pgtypeconv.UUIDToPgtype(roundID))
	if err != nil {
		return nil, fmt.Errorf("get matches by round ID: %w", err)
	}

	matches := make([]domain.Match, len(rows))
	for i, row := range rows {
		matches[i] = *r.rowToMatch(row)
	}

	return matches, nil
}

// GetForPlayer retrieves active matches for a player
func (r *matchRepository) GetForPlayer(ctx context.Context, playerID uuid.UUID) ([]domain.Match, error) {
	rows, err := r.queries.GetMatchesForPlayer(ctx, pgtypeconv.UUIDToPgtype(playerID))
	if err != nil {
		return nil, fmt.Errorf("get matches for player: %w", err)
	}

	matches := make([]domain.Match, len(rows))
	for i, row := range rows {
		matches[i] = *r.rowToMatch(row)
	}

	return matches, nil
}

// UpdateStatus updates a match's status
func (r *matchRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.MatchStatus) error {
	err := r.queries.UpdateMatchStatus(ctx, sqlc.UpdateMatchStatusParams{
		ID:     pgtypeconv.UUIDToPgtype(id),
		Status: string(status), // Status is string in SQLC (NOT NULL field)
	})
	if err != nil {
		return fmt.Errorf("update match status: %w", err)
	}

	return nil
}

// UpdateReady updates a player's ready status
func (r *matchRepository) UpdateReady(ctx context.Context, id uuid.UUID, playerID uuid.UUID, ready bool) error {
	err := r.queries.UpdateMatchReady(ctx, sqlc.UpdateMatchReadyParams{
		ID:                pgtypeconv.UUIDToPgtype(id),
		Participant1ID:    pgtypeconv.UUIDToPgtype(playerID),
		Participant1Ready: ready,
	})
	if err != nil {
		return fmt.Errorf("update match ready: %w", err)
	}

	return nil
}

// UpdateResult updates a match's result
func (r *matchRepository) UpdateResult(ctx context.Context, id uuid.UUID, winnerID uuid.UUID, scorePlayer1, scorePlayer2 int, source domain.ResultSource) error {
	err := r.queries.UpdateMatchResult(ctx, sqlc.UpdateMatchResultParams{
		ID:           pgtypeconv.UUIDToPgtype(id),
		WinnerID:     pgtypeconv.UUIDToPgtype(winnerID),
		ScorePlayer1: pgtypeconv.IntToPgtype(scorePlayer1),
		ScorePlayer2: pgtypeconv.IntToPgtype(scorePlayer2),
		ResultSource: pgtypeconv.StringToPgtype(string(source)),
	})
	if err != nil {
		return fmt.Errorf("update match result: %w", err)
	}

	return nil
}

// UpdateDisconnect tracks player disconnection
func (r *matchRepository) UpdateDisconnect(ctx context.Context, id uuid.UUID, playerID uuid.UUID, disconnectedAt *time.Time) error {
	if disconnectedAt == nil {
		// Clear disconnect
		err := r.queries.ClearMatchDisconnect(ctx, pgtypeconv.UUIDToPgtype(id))
		if err != nil {
			return fmt.Errorf("clear match disconnect: %w", err)
		}
	} else {
		// Set disconnect
		err := r.queries.UpdateMatchDisconnect(ctx, sqlc.UpdateMatchDisconnectParams{
			ID:                   pgtypeconv.UUIDToPgtype(id),
			DisconnectedPlayerID: pgtypeconv.UUIDPtrToPgtype(&playerID),
			DisconnectedAt:       pgtypeconv.TimePtrToPgtype(disconnectedAt),
		})
		if err != nil {
			return fmt.Errorf("update match disconnect: %w", err)
		}
	}

	return nil
}

// UpdateScore updates match scores
func (r *matchRepository) UpdateScore(ctx context.Context, id uuid.UUID, scorePlayer1, scorePlayer2 int) error {
	err := r.queries.UpdateMatchScore(ctx, sqlc.UpdateMatchScoreParams{
		ID:           pgtypeconv.UUIDToPgtype(id),
		ScorePlayer1: pgtypeconv.IntToPgtype(scorePlayer1),
		ScorePlayer2: pgtypeconv.IntToPgtype(scorePlayer2),
	})
	if err != nil {
		return fmt.Errorf("update match score: %w", err)
	}

	return nil
}

// rowToMatch converts a database row to a domain Match
func (r *matchRepository) rowToMatch(row sqlc.TournamentMatch) *domain.Match {
	match := &domain.Match{
		ID:                pgtypeconv.PgtypeToUUID(row.ID),
		RoundID:           pgtypeconv.PgtypeToUUID(row.RoundID),
		MatchNumber:       int(row.MatchNumber), // int32 -> int
		NextMatchID:       pgtypeconv.PgtypeToUUIDPtr(row.NextMatchID),
		Participant1ID:    pgtypeconv.PgtypeToUUID(row.Participant1ID),
		Participant2ID:    pgtypeconv.PgtypeToUUIDPtr(row.Participant2ID),
		Status:            domain.MatchStatus(row.Status), // string directly
		Participant1Ready: row.Participant1Ready,          // bool directly
		Participant2Ready: row.Participant2Ready,          // bool directly
		WinnerID:          pgtypeconv.PgtypeToUUIDPtr(row.WinnerID),
		ScorePlayer1:      pgtypeconv.PgtypeToIntPtr(row.ScorePlayer1),
		ScorePlayer2:      pgtypeconv.PgtypeToIntPtr(row.ScorePlayer2),
		CreatedAt:         row.CreatedAt, // time.Time directly
		StartedAt:         pgtypeconv.PgtypeToTimePtr(row.StartedAt),
		CompletedAt:       pgtypeconv.PgtypeToTimePtr(row.CompletedAt),
		UpdatedAt:         row.UpdatedAt, // time.Time directly
	}

	// Handle result source
	if pgtypeconv.PgtypeToStringPtr(row.ResultSource) != nil {
		source := domain.ResultSource(*pgtypeconv.PgtypeToStringPtr(row.ResultSource))
		match.ResultSource = &source
	}

	// Handle disconnect tracking
	match.DisconnectedPlayerID = pgtypeconv.PgtypeToUUIDPtr(row.DisconnectedPlayerID)
	match.DisconnectedAt = pgtypeconv.PgtypeToTimePtr(row.DisconnectedAt)

	// Handle game API tracking
	match.GameAPIMatchID = pgtypeconv.PgtypeToStringPtr(row.GameApiMatchID)
	match.GameAPIPollAttempts = int(row.GameApiPollAttempts.Int32) // pgtype.Int4 -> int
	match.GameAPILastPoll = pgtypeconv.PgtypeToTimePtr(row.GameApiLastPoll)

	return match
}

// UpdateDisconnectedPlayer updates disconnect tracking info (T114, T121)
func (r *matchRepository) UpdateDisconnectedPlayer(ctx context.Context, id uuid.UUID, playerID uuid.UUID, disconnectedAt time.Time) error {
	return r.queries.UpdateDisconnectedPlayerOnly(ctx, sqlc.UpdateDisconnectedPlayerOnlyParams{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		DisconnectedPlayerID: pgtype.UUID{
			Bytes: playerID,
			Valid: true,
		},
		DisconnectedAt: pgtype.Timestamp{
			Time:  disconnectedAt,
			Valid: true,
		},
	})
}

// ClearDisconnectedPlayer clears disconnect tracking info (T116)
func (r *matchRepository) ClearDisconnectedPlayer(ctx context.Context, id uuid.UUID) error {
	return r.queries.ClearDisconnectedPlayerInfo(ctx, pgtype.UUID{
		Bytes: id,
		Valid: true,
	})
}
