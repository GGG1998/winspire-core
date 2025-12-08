package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
)

// TestPrelobbyActivityBracketGeneration ensures bracket_generation event_type passes the constraint.
func TestPrelobbyActivityBracketGeneration(t *testing.T) {
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
		_ = postgresContainer.Terminate(ctx)
	}()

	// Get connection string
	pgConnStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Apply minimal migrations for prelobby tables
	err = runPrelobbyMigrations(ctx, pgConnStr)
	require.NoError(t, err)

	// Setup pool and queries
	pool, err := pgxpool.New(ctx, pgConnStr)
	require.NoError(t, err)
	defer pool.Close()

	queries := sqlc.New(pool)

	// Create prelobby
	tournamentID := uuid.New()
	_, err = queries.CreatePreLobby(ctx, sqlc.CreatePreLobbyParams{
		TournamentID:    uuidToPgtype(tournamentID),
		MinParticipants: 2,
	})
	require.NoError(t, err)

	// Insert activity with bracket_generation event_type
	participantName := pgtypeText("system")
	_, err = queries.AddActivityFeedEvent(ctx, sqlc.AddActivityFeedEventParams{
		TournamentID:    uuidToPgtype(tournamentID),
		EventType:       "bracket_generation",
		Message:         "Bracket generated",
		ParticipantName: participantName,
	})
	require.NoError(t, err, "bracket_generation should be allowed by constraint")
}

// runPrelobbyMigrations applies only the prelobby-related tables with updated constraint.
func runPrelobbyMigrations(ctx context.Context, connStr string) error {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return err
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS prelobbies (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tournament_id UUID NOT NULL UNIQUE,
			status VARCHAR(30) NOT NULL DEFAULT 'waiting' CHECK (
				status IN ('waiting', 'grace_period', 'generating_bracket', 'started', 'cancelled')
			),
			grace_period_start TIMESTAMP,
			grace_period_end TIMESTAMP,
			min_participants INTEGER NOT NULL DEFAULT 2 CHECK (min_participants >= 2),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS prelobby_activity_feed (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tournament_id UUID NOT NULL REFERENCES prelobbies(tournament_id) ON DELETE CASCADE,
			event_type VARCHAR(50) NOT NULL CHECK (
				event_type IN ('participant_joined', 'participant_left', 'grace_period_started',
							   'grace_period_ended', 'bracket_generation', 'tournament_cancelled', 'system_message')
			),
			message TEXT NOT NULL,
			participant_name VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

// Helpers for pgtype conversions (local minimal helpers)
func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	var v pgtype.UUID
	v.Bytes = id
	v.Valid = true
	return v
}

func pgtypeText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}
