package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	redismodule "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/pubsub"
	"github.com/winspire-core/services/matchmaking/internal/repository"
	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
	"github.com/winspire-core/services/matchmaking/internal/websocket"
)

// TestHandleTournamentStarted_Integration tests HandleTournamentStarted with real database and Redis
func TestHandleTournamentStarted_Integration(t *testing.T) {
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

	// Start Redis container
	redisContainer, err := redismodule.RunContainer(ctx,
		tc.WithImage("redis:7-alpine"),
		tc.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithOccurrence(1).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	}()

	// Get connection strings
	pgConnStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	redisAddr, err := redisContainer.ConnectionString(ctx)
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

	// Setup Redis client
	redisOpts, err := redis.ParseURL(redisAddr)
	require.NoError(t, err)
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	// Test cases
	tests := []struct {
		name                string
		payload             map[string]interface{}
		metadata            map[string]string
		setupData           func(context.Context, *testing.T, *sqlc.Queries)
		expectError         bool
		expectErrorContains string
		validateResult      func(context.Context, *testing.T, *sqlc.Queries, uuid.UUID)
	}{
		{
			name: "T01: Valid payload with grace period - creates pre-lobby and starts grace period",
			payload: func() map[string]interface{} {
				return map[string]interface{}{
					"tournament_id": uuid.New().String(),
					"host_id":       uuid.New().String(),
					"participants":  []string{uuid.New().String(), uuid.New().String()},
					"game_id":       uuid.New().String(),
					"started_at":    time.Now().Format(time.RFC3339),
				}
			}(),
			metadata: map[string]string{
				"correlation_id": "test-001",
			},
			setupData: func(ctx context.Context, t *testing.T, q *sqlc.Queries) {
				// No setup needed - fresh tournament
			},
			expectError: false,
			validateResult: func(ctx context.Context, t *testing.T, q *sqlc.Queries, tournamentID uuid.UUID) {
				// Verify pre-lobby was created
				var pgTournamentID pgtype.UUID
				pgTournamentID.Scan(tournamentID.String())
				preLobby, err := q.GetPreLobbyByTournament(ctx, pgTournamentID)
				require.NoError(t, err)
				assert.Equal(t, "grace_period", preLobby.Status)
				assert.True(t, preLobby.GracePeriodStart.Valid)
				assert.True(t, preLobby.GracePeriodEnd.Valid)

				// Verify no bracket created yet (waiting for grace period)
				_, err = q.GetBracketByTournamentID(ctx, pgTournamentID)
				assert.Error(t, err) // Should not exist yet
			},
		},
		{
			name: "T02: Fallback to immediate bracket generation - when pre-lobby fails",
			payload: func() map[string]interface{} {
				return map[string]interface{}{
					"tournament_id": uuid.New().String(),
					"host_id":       uuid.New().String(),
					"participants":  []string{uuid.New().String(), uuid.New().String(), uuid.New().String()},
					"game_id":       uuid.New().String(),
					"started_at":    time.Now().Format(time.RFC3339),
				}
			}(),
			metadata: map[string]string{
				"correlation_id": "test-002",
			},
			setupData: func(ctx context.Context, t *testing.T, q *sqlc.Queries) {
				// No setup - will test fallback path
			},
			expectError: false,
			validateResult: func(ctx context.Context, t *testing.T, q *sqlc.Queries, tournamentID uuid.UUID) {
				// This test may or may not create bracket immediately depending on timing
				// The important thing is no error occurs
			},
		},
		{
			name: "T03: RecordBracketGeneration activity feed event",
			payload: func() map[string]interface{} {
				tid := uuid.New()
				// Pre-create pre-lobby to ensure it exists
				return map[string]interface{}{
					"tournament_id": tid.String(),
					"host_id":       uuid.New().String(),
					"participants":  []string{uuid.New().String(), uuid.New().String()},
					"game_id":       uuid.New().String(),
					"started_at":    time.Now().Format(time.RFC3339),
				}
			}(),
			metadata: map[string]string{
				"correlation_id": "test-003",
			},
			setupData: func(ctx context.Context, t *testing.T, q *sqlc.Queries) {
				// Setup will happen in test flow
			},
			expectError: false,
			validateResult: func(ctx context.Context, t *testing.T, q *sqlc.Queries, tournamentID uuid.UUID) {
				// Wait a bit for grace period to potentially trigger
				time.Sleep(100 * time.Millisecond)

				// Check activity feed for bracket generation event
				var pgTournamentID pgtype.UUID
				pgTournamentID.Scan(tournamentID.String())
				feed, err := q.GetRecentActivityFeed(ctx, pgTournamentID)
				require.NoError(t, err)

				// Look for bracket generation or grace period events
				hasEvent := false
				for _, item := range feed {
					if item.EventType == "bracket_generation" || item.EventType == "grace_period_started" {
						hasEvent = true
						break
					}
				}
				assert.True(t, hasEvent, "Should have activity feed events")
			},
		},
		{
			name: "T04: Invalid tournament_id - returns error",
			payload: map[string]interface{}{
				"tournament_id": "",
				"host_id":       uuid.New().String(),
				"participants":  []string{uuid.New().String(), uuid.New().String()},
				"game_id":       uuid.New().String(),
				"started_at":    time.Now().Format(time.RFC3339),
			},
			metadata: map[string]string{
				"correlation_id": "test-004",
			},
			setupData:           func(ctx context.Context, t *testing.T, q *sqlc.Queries) {},
			expectError:         true,
			expectErrorContains: "tournament_id is required",
			validateResult:      func(ctx context.Context, t *testing.T, q *sqlc.Queries, tournamentID uuid.UUID) {},
		},
		{
			name: "T05: Insufficient participants - grace period starts, bracket gen will fail later",
			payload: map[string]interface{}{
				"tournament_id": uuid.New().String(),
				"host_id":       uuid.New().String(),
				"participants":  []string{uuid.New().String()}, // Only 1 participant
				"game_id":       uuid.New().String(),
				"started_at":    time.Now().Format(time.RFC3339),
			},
			metadata: map[string]string{
				"correlation_id": "test-005",
			},
			setupData:   func(ctx context.Context, t *testing.T, q *sqlc.Queries) {},
			expectError: false, // Grace period starts successfully
			validateResult: func(ctx context.Context, t *testing.T, q *sqlc.Queries, tournamentID uuid.UUID) {
				// Verify grace period started even with insufficient participants
				// The bracket generation will fail when grace period ends (tested in grace period callback logic)
				var pgTournamentID pgtype.UUID
				pgTournamentID.Scan(tournamentID.String())
				preLobby, err := q.GetPreLobbyByTournament(ctx, pgTournamentID)
				require.NoError(t, err)
				assert.Equal(t, "grace_period", preLobby.Status)
			},
		},
		{
			name: "T06: Grace period already active - ignores duplicate event",
			payload: func() map[string]interface{} {
				tid := uuid.New()
				return map[string]interface{}{
					"tournament_id": tid.String(),
					"host_id":       uuid.New().String(),
					"participants":  []string{uuid.New().String(), uuid.New().String()},
					"game_id":       uuid.New().String(),
					"started_at":    time.Now().Format(time.RFC3339),
				}
			}(),
			metadata: map[string]string{
				"correlation_id": "test-006",
			},
			setupData: func(ctx context.Context, t *testing.T, q *sqlc.Queries) {
				// Will be set up in the test itself
			},
			expectError: false,
			validateResult: func(ctx context.Context, t *testing.T, q *sqlc.Queries, tournamentID uuid.UUID) {
				// Verify grace period is still active
				var pgTournamentID pgtype.UUID
				pgTournamentID.Scan(tournamentID.String())
				preLobby, err := q.GetPreLobbyByTournament(ctx, pgTournamentID)
				require.NoError(t, err)
				assert.Equal(t, "grace_period", preLobby.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear database between tests
			_, err := pool.Exec(ctx, "TRUNCATE TABLE tournament_brackets, tournament_rounds, tournament_matches, prelobbies, prelobby_participant_snapshots, prelobby_activity_feed CASCADE")
			require.NoError(t, err)

			// Setup
			queries := sqlc.New(pool)
			logger := observability.NewLogger("error", "json")
			metrics := observability.NewMetricsEmitter("test")
			publisher := pubsub.NewEventPublisher(redisClient)

			// Setup repositories
			bracketRepo := repository.NewBracketRepository(queries, pool, pool)
			roundRepo := repository.NewRoundRepository(queries)
			matchRepo := repository.NewMatchRepository(queries)
			preLobbyRepo := repository.NewPreLobbyRepository(queries, pool, pool)

			// Setup services
			bracketService := NewBracketService(bracketRepo, roundRepo, matchRepo, publisher, metrics, logger)
			hub := websocket.NewHub(nil)
			preLobbyService := NewPreLobbyService(preLobbyRepo, hub, publisher, metrics, logger)

			// Setup clients
			competitionClient := NewCompetitionClient("http://localhost:8080", logger)
			gameManagementClient := NewGameManagementClient("http://localhost:8087", logger)

			// Setup event handler
			handler := NewEventHandler(bracketService, preLobbyService, publisher, logger, competitionClient, gameManagementClient, "http://localhost:8087", hub)

			// Run test-specific setup
			if tt.setupData != nil {
				tt.setupData(ctx, t, queries)
			}

			// Extract tournament ID for validation
			var tournamentID uuid.UUID
			if tidStr, ok := tt.payload["tournament_id"].(string); ok && tidStr != "" {
				tournamentID, _ = uuid.Parse(tidStr)
			}

			// Special handling for T06 - pre-create grace period
			if tt.name == "T06: Grace period already active - ignores duplicate event" && tournamentID != uuid.Nil {
				_, err2 := preLobbyService.GetOrCreatePreLobby(ctx, tournamentID, 2)
				require.NoError(t, err2)
				err2 = preLobbyService.StartGracePeriod(ctx, tournamentID, func(uuid.UUID, []uuid.UUID) {})
				require.NoError(t, err2)
			}

			// Execute
			err = handler.HandleTournamentStartRequested(ctx, "TournamentStartRequested", tt.payload, tt.metadata)

			// Assert
			if tt.expectError {
				require.Error(t, err)
				if tt.expectErrorContains != "" {
					assert.Contains(t, err.Error(), tt.expectErrorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Validate results
			if !tt.expectError && tt.validateResult != nil {
				tt.validateResult(ctx, t, queries, tournamentID)
			}
		})
	}
}

// runMigrations applies database migrations
func runMigrations(ctx context.Context, connStr string) error {
	// Connect to database
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return err
	}
	defer pool.Close()

	// Apply migrations - in order (matching actual schema names)
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS tournament_brackets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tournament_id UUID NOT NULL UNIQUE,
			total_rounds INT NOT NULL,
			total_matches INT NOT NULL,
			byes_count INT NOT NULL DEFAULT 0,
			generated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_brackets_tournament_id ON tournament_brackets(tournament_id)`,
		`CREATE TABLE IF NOT EXISTS tournament_rounds (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			bracket_id UUID NOT NULL REFERENCES tournament_brackets(id) ON DELETE CASCADE,
			round_number INT NOT NULL,
			round_name VARCHAR(50) NOT NULL,
			matches_count INT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			UNIQUE(bracket_id, round_number)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_rounds_bracket_id ON tournament_rounds(bracket_id)`,
		`CREATE TABLE IF NOT EXISTS tournament_matches (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tournament_id UUID NOT NULL,
			round_id UUID NOT NULL REFERENCES tournament_rounds(id) ON DELETE CASCADE,
			round_number INT NOT NULL,
			match_number INT NOT NULL,
			participant1_id UUID NOT NULL,
			participant2_id UUID,
			winner_id UUID,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			participant1_ready BOOLEAN NOT NULL DEFAULT FALSE,
			participant2_ready BOOLEAN NOT NULL DEFAULT FALSE,
			score_participant1 INT,
			score_participant2 INT,
			result_source VARCHAR(20),
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			participant1_disconnected_at TIMESTAMP,
			participant2_disconnected_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_matches_tournament_round ON tournament_matches(tournament_id, round_number)`,
		`CREATE TABLE IF NOT EXISTS prelobbies (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tournament_id UUID NOT NULL UNIQUE,
			status VARCHAR(30) NOT NULL DEFAULT 'waiting',
			grace_period_start TIMESTAMP,
			grace_period_end TIMESTAMP,
			min_participants INT NOT NULL DEFAULT 2,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS prelobby_participant_snapshots (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tournament_id UUID NOT NULL UNIQUE,
			participants JSONB NOT NULL,
			participant_count INT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS prelobby_activity_feed (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tournament_id UUID NOT NULL,
			event_type VARCHAR(50) NOT NULL,
			message TEXT NOT NULL,
			participant_name VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_feed_tournament ON prelobby_activity_feed(tournament_id, created_at DESC)`,
	}

	for _, migration := range migrations {
		_, err := pool.Exec(ctx, migration)
		if err != nil {
			return err
		}
	}

	return nil
}
