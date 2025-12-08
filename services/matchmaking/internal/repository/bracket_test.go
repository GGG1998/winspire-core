package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
)

// TestBracketRepository_Create_Integration tests bracket creation with real database
func TestBracketRepository_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	postgresContainer, err := postgres.RunContainer(ctx,
		tc.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("matchmaking_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		tc.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	}()

	// Get connection string
	pgConnStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Run database migrations
	err = runMigrations(ctx, pgConnStr)
	require.NoError(t, err, "failed to run migrations")

	// Setup database connection pool
	dbConfig, err := pgxpool.ParseConfig(pgConnStr)
	require.NoError(t, err)
	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	require.NoError(t, err)
	defer pool.Close()

	// Create repository
	queries := sqlc.New(pool)
	bracketRepo := NewBracketRepository(queries, pool, pool)

	t.Run("Create bracket with 2 participants", func(t *testing.T) {
		// Setup
		tournamentID := uuid.New()
		participants := []uuid.UUID{uuid.New(), uuid.New()}

		// Create bracket domain model
		bracket, err := domain.NewBracket(tournamentID, participants)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, bracket.ID, "bracket should have a valid ID")

		// Generate rounds
		rounds, err := bracket.GenerateRounds(participants)
		require.NoError(t, err)
		require.NotEmpty(t, rounds, "should have at least one round")

		// Generate matches (simplified - just create empty matches for test)
		matches := []domain.Match{}
		for _, round := range rounds {
			for i := 0; i < round.MatchesCount; i++ {
				match := domain.Match{
					ID:             uuid.New(),
					RoundID:        round.ID,
					MatchNumber:    i + 1,
					Participant1ID: participants[0],
					Participant2ID: &participants[1],
					Status:         domain.MatchStatusPending,
				}
				matches = append(matches, match)
			}
		}

		// Execute - Create bracket with rounds and matches
		err = bracketRepo.Create(ctx, bracket, rounds, matches)
		require.NoError(t, err, "bracket creation should succeed without FK violations")

		// Verify - Retrieve bracket and check it exists
		retrievedBracket, err := bracketRepo.GetByTournamentID(ctx, tournamentID)
		require.NoError(t, err)
		assert.Equal(t, bracket.ID, retrievedBracket.ID, "bracket ID should match")
		assert.Equal(t, bracket.TournamentID, retrievedBracket.TournamentID)
		assert.Equal(t, bracket.TotalRounds, retrievedBracket.TotalRounds)
		assert.Equal(t, bracket.TotalMatches, retrievedBracket.TotalMatches)

		// Verify rounds were created with correct bracket_id
		bracket2, rounds2, matches2, err := bracketRepo.GetWithRoundsAndMatches(ctx, tournamentID)
		require.NoError(t, err)
		assert.Equal(t, bracket.ID, bracket2.ID)
		assert.Equal(t, len(rounds), len(rounds2), "should have same number of rounds")
		assert.Equal(t, len(matches), len(matches2), "should have same number of matches")

		// Verify round bracket_id FK is correct
		for _, round := range rounds2 {
			assert.Equal(t, bracket.ID, round.BracketID, "round should reference correct bracket_id")
		}
	})

	t.Run("Create bracket with 4 participants", func(t *testing.T) {
		// Setup
		tournamentID := uuid.New()
		participants := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}

		// Create bracket domain model
		bracket, err := domain.NewBracket(tournamentID, participants)
		require.NoError(t, err)

		// Generate rounds
		rounds, err := bracket.GenerateRounds(participants)
		require.NoError(t, err)

		// Generate matches (simplified)
		matches := []domain.Match{}
		for _, round := range rounds {
			for i := 0; i < round.MatchesCount; i++ {
				match := domain.Match{
					ID:             uuid.New(),
					RoundID:        round.ID,
					MatchNumber:    i + 1,
					Participant1ID: participants[i%len(participants)],
					Status:         domain.MatchStatusPending,
				}
				if i+1 < len(participants) {
					match.Participant2ID = &participants[(i+1)%len(participants)]
				}
				matches = append(matches, match)
			}
		}

		// Execute
		err = bracketRepo.Create(ctx, bracket, rounds, matches)
		require.NoError(t, err, "bracket creation with 4 participants should succeed")

		// Verify
		retrievedBracket, err := bracketRepo.GetByTournamentID(ctx, tournamentID)
		require.NoError(t, err)
		assert.Equal(t, bracket.ID, retrievedBracket.ID)
		assert.Equal(t, 2, retrievedBracket.TotalRounds, "4 participants should need 2 rounds")
	})
}

// runMigrations runs database migrations using atlas
func runMigrations(ctx context.Context, connStr string) error {
	// This is a simplified version - in real tests you'd use atlas or another migration tool
	// For now, we'll just create the tables directly
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return err
	}
	defer pool.Close()

	// Create brackets table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS tournament_brackets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tournament_id UUID NOT NULL UNIQUE,
			total_rounds INTEGER NOT NULL CHECK (total_rounds > 0),
			total_matches INTEGER NOT NULL CHECK (total_matches > 0),
			byes_count INTEGER NOT NULL DEFAULT 0 CHECK (byes_count >= 0),
			generated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMP NULL
		);
	`)
	if err != nil {
		return err
	}

	// Create rounds table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS tournament_rounds (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			bracket_id UUID NOT NULL REFERENCES tournament_brackets(id) ON DELETE CASCADE,
			round_number INTEGER NOT NULL CHECK (round_number > 0),
			round_name VARCHAR(100) NOT NULL,
			matches_count INTEGER NOT NULL CHECK (matches_count > 0),
			status VARCHAR(20) NOT NULL DEFAULT 'pending' 
				CHECK (status IN ('pending', 'in_progress', 'completed')),
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			CONSTRAINT unique_bracket_round_number UNIQUE (bracket_id, round_number)
		);
	`)
	if err != nil {
		return err
	}

	// Create matches table (full schema with all columns)
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS tournament_matches (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			round_id UUID NOT NULL REFERENCES tournament_rounds(id) ON DELETE CASCADE,
			match_number INTEGER NOT NULL CHECK (match_number > 0),
			next_match_id UUID REFERENCES tournament_matches(id),
			participant1_id UUID NOT NULL,
			participant2_id UUID,
			status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN (
				'pending', 'ready', 'started', 'paused', 'completed', 'cancelled'
			)),
			participant1_ready BOOLEAN NOT NULL DEFAULT false,
			participant2_ready BOOLEAN NOT NULL DEFAULT false,
			winner_id UUID,
			score_player1 INTEGER CHECK (score_player1 >= 0),
			score_player2 INTEGER CHECK (score_player2 >= 0),
			result_source VARCHAR(20) CHECK (result_source IN ('game_api', 'host_manual', 'walkover')),
			disconnected_player_id UUID,
			disconnected_at TIMESTAMP,
			game_api_match_id VARCHAR(255),
			game_api_poll_attempts INTEGER DEFAULT 0,
			game_api_last_poll TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			CONSTRAINT check_completed_has_winner CHECK (
				(status = 'completed' AND winner_id IS NOT NULL) OR status != 'completed'
			),
			CONSTRAINT unique_round_match_number UNIQUE (round_id, match_number)
		);
	`)
	if err != nil {
		return err
	}

	return nil
}

