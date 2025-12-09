// Package repository provides data access layer for matchmaking entities
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
	pgtypeconv "github.com/winspire/winspire-core/libs/go/pgtype"
)

// BracketRepository handles bracket persistence
type BracketRepository interface {
	Create(ctx context.Context, bracket *domain.Bracket, rounds []domain.Round, matches []domain.Match) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Bracket, error)
	GetByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*domain.Bracket, error)
	GetWithRoundsAndMatches(ctx context.Context, tournamentID uuid.UUID) (*domain.Bracket, []domain.Round, []domain.Match, error)
	UpdateCompletedAt(ctx context.Context, id uuid.UUID, completedAt time.Time) error
}

type bracketRepository struct {
	queries *sqlc.Queries
	db      sqlc.DBTX
	pool    *pgxpool.Pool // Needed for transactions
}

// NewBracketRepository creates a new bracket repository
func NewBracketRepository(queries *sqlc.Queries, db sqlc.DBTX, pool *pgxpool.Pool) BracketRepository {
	return &bracketRepository{
		queries: queries,
		db:      db,
		pool:    pool,
	}
}

// Create creates a bracket with all its rounds and matches in a single transaction
func (r *bracketRepository) Create(ctx context.Context, bracket *domain.Bracket, rounds []domain.Round, matches []domain.Match) error {
	// Start transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback if commit is not called

	qtx := r.queries.WithTx(tx)

	// Marshal game snapshot to JSONB if present
	var gameSnapshotJSON []byte
	if bracket.GameSnapshot != nil {
		gameSnapshotJSON, _ = json.Marshal(bracket.GameSnapshot)
	}

	// Create bracket (use domain-generated ID to ensure FK consistency)
	_, err = qtx.CreateBracket(ctx, sqlc.CreateBracketParams{
		ID:           pgtypeconv.UUIDToPgtype(bracket.ID),
		TournamentID: pgtypeconv.UUIDToPgtype(bracket.TournamentID),
		GameSnapshot: gameSnapshotJSON,
		TotalRounds:  int32(bracket.TotalRounds),
		TotalMatches: int32(bracket.TotalMatches),
		ByesCount:    int32(bracket.ByesCount),
	})
	if err != nil {
		return fmt.Errorf("create bracket: %w", err)
	}

	// Create rounds (use domain-generated IDs to ensure FK consistency)
	for _, round := range rounds {
		_, err = qtx.CreateRound(ctx, sqlc.CreateRoundParams{
			ID:           pgtypeconv.UUIDToPgtype(round.ID),
			BracketID:    pgtypeconv.UUIDToPgtype(bracket.ID),
			RoundNumber:  int32(round.RoundNumber),
			RoundName:    round.RoundName,
			MatchesCount: int32(round.MatchesCount),
			Status:       string(round.Status),
		})
		if err != nil {
			return fmt.Errorf("create round %d: %w", round.RoundNumber, err)
		}
	}

	// Create matches (use domain-generated IDs to ensure FK consistency)
	for _, match := range matches {
		_, err = qtx.CreateMatch(ctx, sqlc.CreateMatchParams{
			ID:             pgtypeconv.UUIDToPgtype(match.ID),
			RoundID:        pgtypeconv.UUIDToPgtype(match.RoundID),
			MatchNumber:    int32(match.MatchNumber),
			NextMatchID:    pgtypeconv.UUIDPtrToPgtype(match.NextMatchID),
			Participant1ID: pgtypeconv.UUIDToPgtype(match.Participant1ID),
			Participant2ID: pgtypeconv.UUIDPtrToPgtype(match.Participant2ID),
			Status:         string(match.Status),
		})
		if err != nil {
			return fmt.Errorf("create match %d in round %s: %w", match.MatchNumber, match.RoundID, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a bracket by its ID
func (r *bracketRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bracket, error) {
	row, err := r.queries.GetBracketByID(ctx, pgtypeconv.UUIDToPgtype(id))
	if err != nil {
		return nil, fmt.Errorf("get bracket by ID: %w", err)
	}

	// Unmarshal game snapshot if present
	var gameSnapshot *domain.GameSnapshot
	if len(row.GameSnapshot) > 0 {
		gameSnapshot = &domain.GameSnapshot{}
		if err := json.Unmarshal(row.GameSnapshot, gameSnapshot); err != nil {
			// Log error but don't fail - continue without snapshot
			fmt.Printf("WARN: Failed to unmarshal game snapshot for bracket %s: %v\n", id, err)
		}
	}

	return &domain.Bracket{
		ID:           pgtypeconv.PgtypeToUUID(row.ID),
		TournamentID: pgtypeconv.PgtypeToUUID(row.TournamentID),
		GameSnapshot: gameSnapshot,
		TotalRounds:  int(row.TotalRounds),
		TotalMatches: int(row.TotalMatches),
		ByesCount:    int(row.ByesCount),
		GeneratedAt:  row.GeneratedAt,
	}, nil
}

// GetByTournamentID retrieves a bracket by tournament ID
func (r *bracketRepository) GetByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*domain.Bracket, error) {
	row, err := r.queries.GetBracketByTournamentID(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return nil, fmt.Errorf("get bracket by tournament ID: %w", err)
	}

	// Unmarshal game snapshot if present
	var gameSnapshot *domain.GameSnapshot
	if len(row.GameSnapshot) > 0 {
		gameSnapshot = &domain.GameSnapshot{}
		if err := json.Unmarshal(row.GameSnapshot, gameSnapshot); err != nil {
			// Log error but don't fail - continue without snapshot
			fmt.Printf("WARN: Failed to unmarshal game snapshot for bracket tournament %s: %v\n", tournamentID, err)
		}
	}

	return &domain.Bracket{
		ID:           pgtypeconv.PgtypeToUUID(row.ID),
		TournamentID: pgtypeconv.PgtypeToUUID(row.TournamentID),
		GameSnapshot: gameSnapshot,
		TotalRounds:  int(row.TotalRounds),
		TotalMatches: int(row.TotalMatches),
		ByesCount:    int(row.ByesCount),
		GeneratedAt:  row.GeneratedAt,
	}, nil
}

// GetWithRoundsAndMatches retrieves a complete bracket hierarchy
func (r *bracketRepository) GetWithRoundsAndMatches(ctx context.Context, tournamentID uuid.UUID) (*domain.Bracket, []domain.Round, []domain.Match, error) {
	rows, err := r.queries.GetBracketWithRoundsAndMatches(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get bracket with rounds and matches: %w", err)
	}

	if len(rows) == 0 {
		return nil, nil, nil, fmt.Errorf("bracket not found for tournament %s", tournamentID)
	}

	// Unmarshal game snapshot if present
	var gameSnapshot *domain.GameSnapshot
	if len(rows[0].GameSnapshot) > 0 {
		gameSnapshot = &domain.GameSnapshot{}
		if err := json.Unmarshal(rows[0].GameSnapshot, gameSnapshot); err != nil {
			// Log error but don't fail - continue without snapshot
			fmt.Printf("WARN: Failed to unmarshal game snapshot: %v\n", err)
		}
	}

	// Extract bracket from first row
	bracket := &domain.Bracket{
		ID:           pgtypeconv.PgtypeToUUID(rows[0].BracketID),
		TournamentID: pgtypeconv.PgtypeToUUID(rows[0].TournamentID),
		GameSnapshot: gameSnapshot,
		TotalRounds:  int(rows[0].BracketTotalRounds),
		TotalMatches: int(rows[0].BracketTotalMatches),
		ByesCount:    int(rows[0].ByesCount),
		GeneratedAt:  rows[0].BracketGeneratedAt,
	}

	// Group rounds and matches
	roundMap := make(map[uuid.UUID]*domain.Round)
	var rounds []domain.Round
	var matches []domain.Match

	for _, row := range rows {
		roundID := pgtypeconv.PgtypeToUUID(row.RoundID)
		// Add round if not seen
		if _, exists := roundMap[roundID]; !exists {
			round := domain.Round{
				ID:           roundID,
				BracketID:    pgtypeconv.PgtypeToUUID(row.BracketID),
				RoundNumber:  pgtypeconv.PgtypeToInt(row.RoundNumber),
				RoundName:    pgtypeconv.PgtypeToString(row.RoundName),
				MatchesCount: pgtypeconv.PgtypeToInt(row.RoundMatchesCount),
				Status:       domain.RoundStatus(pgtypeconv.PgtypeToString(row.RoundStatus)),
				StartedAt:    pgtypeconv.PgtypeToTimePtr(row.RoundStartedAt),
				CompletedAt:  pgtypeconv.PgtypeToTimePtr(row.RoundCompletedAt),
			}
			roundMap[roundID] = &round
			rounds = append(rounds, round)
		}

		// Add match
		match := domain.Match{
			ID:                pgtypeconv.PgtypeToUUID(row.MatchID),
			RoundID:           roundID,
			MatchNumber:       pgtypeconv.PgtypeToInt(row.MatchNumber),
			NextMatchID:       pgtypeconv.PgtypeToUUIDPtr(row.NextMatchID),
			Participant1ID:    pgtypeconv.PgtypeToUUID(row.Participant1ID),
			Participant2ID:    pgtypeconv.PgtypeToUUIDPtr(row.Participant2ID),
			Status:            domain.MatchStatus(pgtypeconv.PgtypeToString(row.MatchStatus)),
			Participant1Ready: pgtypeconv.PgtypeToBool(row.Participant1Ready),
			Participant2Ready: pgtypeconv.PgtypeToBool(row.Participant2Ready),
			WinnerID:          pgtypeconv.PgtypeToUUIDPtr(row.WinnerID),
			ScorePlayer1:      pgtypeconv.PgtypeToIntPtr(row.ScorePlayer1),
			ScorePlayer2:      pgtypeconv.PgtypeToIntPtr(row.ScorePlayer2),
			CreatedAt:         row.MatchCreatedAt.Time,
			StartedAt:         pgtypeconv.PgtypeToTimePtr(row.MatchStartedAt),
			CompletedAt:       pgtypeconv.PgtypeToTimePtr(row.MatchCompletedAt),
		}

		// Handle result source
		if pgtypeconv.PgtypeToStringPtr(row.ResultSource) != nil {
			source := domain.ResultSource(*pgtypeconv.PgtypeToStringPtr(row.ResultSource))
			match.ResultSource = &source
		}

		matches = append(matches, match)
	}

	return bracket, rounds, matches, nil
}

// UpdateCompletedAt marks a bracket as completed (T109)
func (r *bracketRepository) UpdateCompletedAt(ctx context.Context, id uuid.UUID, completedAt time.Time) error {
	return r.queries.UpdateBracketCompletedAt(ctx, sqlc.UpdateBracketCompletedAtParams{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		CompletedAt: pgtype.Timestamp{
			Time:  completedAt,
			Valid: true,
		},
	})
}
