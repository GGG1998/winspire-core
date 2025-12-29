//go:build integration

package application

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/winspire-core/services/matchmaking/internal/observability"
)

// competitionTestEnv holds resources for integration testing
type competitionTestEnv struct {
	pgContainer          *postgres.PostgresContainer
	redisContainer       tc.Container
	competitionContainer tc.Container
	pool                 *pgxpool.Pool
	competitionURL       string
	network              *tc.DockerNetwork
}

// setupCompetitionTestEnv creates the test environment with PostgreSQL and Competition service
func setupCompetitionTestEnv(ctx context.Context, t *testing.T) *competitionTestEnv {
	t.Helper()

	// Create a Docker network for container communication
	net, err := network.New(ctx, network.WithCheckDuplicate())
	require.NoError(t, err, "failed to create docker network")

	// Start PostgreSQL container
	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("tournament_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		tc.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
		network.WithNetwork([]string{"postgres"}, net),
	)
	require.NoError(t, err, "failed to start postgres container")

	// Start Redis container
	redisReq := tc.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		Networks:     []string{net.Name},
		NetworkAliases: map[string][]string{
			net.Name: {"redis"},
		},
		WaitingFor: wait.ForLog("Ready to accept connections").
			WithStartupTimeout(30 * time.Second),
	}
	redisContainer, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	require.NoError(t, err, "failed to start redis container")

	// Get connection string for migrations
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get postgres connection string")

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err, "failed to create postgres pool")

	// Run tournament service migrations
	err = runCompetitionMigrations(ctx, pool)
	require.NoError(t, err, "failed to run tournament migrations")

	// Get the internal network connection string for the tournament service
	networkConnStr := "postgresql://postgres:postgres@postgres:5432/tournament_test?sslmode=disable"

	// Build and start the tournament service container
	// Path: application -> internal -> matchmaking -> services -> winspire-core (4 levels up)
	_, thisFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(thisFile), "../../../..")

	competitionReq := tc.ContainerRequest{
		FromDockerfile: tc.FromDockerfile{
			Context:    projectRoot,
			Dockerfile: "services/tournament/cmd/tournament/Dockerfile.test",
		},
		ExposedPorts: []string{"8089/tcp"},
		Env: map[string]string{
			"APP_ENV":      "test",
			"SERVICE_PORT": "8089",
			"POSTGRES_DSN": networkConnStr,
			"REDIS_ADDR":   "redis:6379",
		},
		Networks: []string{net.Name},
		WaitingFor: wait.ForHTTP("/healthz").
			WithPort("8089/tcp").
			WithStartupTimeout(120 * time.Second),
	}

	competitionContainer, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: competitionReq,
		Started:          true,
	})
	require.NoError(t, err, "failed to start tournament service container")

	// Get the tournament service URL
	host, err := competitionContainer.Host(ctx)
	require.NoError(t, err, "failed to get tournament container host")

	port, err := competitionContainer.MappedPort(ctx, "8089")
	require.NoError(t, err, "failed to get tournament container port")

	competitionURL := fmt.Sprintf("http://%s:%s", host, port.Port())

	return &competitionTestEnv{
		pgContainer:          pgContainer,
		redisContainer:       redisContainer,
		competitionContainer: competitionContainer,
		pool:                 pool,
		competitionURL:       competitionURL,
		network:              net,
	}
}

// cleanup releases all test resources
func (e *competitionTestEnv) cleanup(ctx context.Context) {
	if e.pool != nil {
		e.pool.Close()
	}
	if e.competitionContainer != nil {
		_ = e.competitionContainer.Terminate(ctx)
	}
	if e.redisContainer != nil {
		_ = e.redisContainer.Terminate(ctx)
	}
	if e.pgContainer != nil {
		_ = e.pgContainer.Terminate(ctx)
	}
	if e.network != nil {
		_ = e.network.Remove(ctx)
	}
}

// runCompetitionMigrations executes tournament service migrations
func runCompetitionMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Path: application -> internal -> matchmaking -> services -> winspire-core (4 levels up)
	// Then: services/tournament/migrations
	_, thisFile, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "../../../..", "services/tournament/migrations")

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations directory %s: %w", migrationsDir, err)
	}

	// Sort migration files by name
	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Execute each migration
	for _, fileName := range sqlFiles {
		filePath := filepath.Join(migrationsDir, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read migration file %s: %w", fileName, err)
		}

		if _, err := pool.Exec(ctx, string(content)); err != nil {
			// Skip certain errors
			if strings.Contains(err.Error(), "already exists") {
				continue
			}
			return fmt.Errorf("execute migration %s: %w", fileName, err)
		}
	}

	return nil
}

// insertTestTournament creates a test tournament in the database
func insertTestTournament(ctx context.Context, pool *pgxpool.Pool, tournament testTournament) error {
	query := `
		INSERT INTO tournaments (
			id, host_id, name, description, status,
			scheduled_start_time_at, minimum_team_count, game_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := pool.Exec(ctx, query,
		tournament.ID,
		tournament.HostID,
		tournament.Name,
		tournament.Description,
		tournament.Status,
		tournament.ScheduledStartTime,
		tournament.MinParticipants,
		tournament.GameID,
	)
	return err
}

type testTournament struct {
	ID                 uuid.UUID
	HostID             uuid.UUID
	Name               string
	Description        *string
	Status             string
	ScheduledStartTime *time.Time
	MinParticipants    int
	GameID             *uuid.UUID
}

func TestCompetitionClient_GetTournamentInfo_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup test environment
	env := setupCompetitionTestEnv(ctx, t)
	defer env.cleanup(ctx)

	// Create competition client
	logger := observability.NewLogger("info", "text")
	client := NewCompetitionClient(env.competitionURL, logger)

	// Default host ID from migrations
	defaultHostID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Test cases
	tests := []struct {
		name            string
		setupTournament *testTournament
		queryID         uuid.UUID
		expectedResult  *TournamentInfo
		expectNil       bool
		expectError     bool
	}{
		{
			name: "tournament exists with all fields",
			setupTournament: &testTournament{
				ID:                 uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				HostID:             defaultHostID,
				Name:               "Test Tournament",
				Description:        strPtr("A test tournament"),
				Status:             "registration_open",
				ScheduledStartTime: timePtr(time.Now().Add(24 * time.Hour)),
				MinParticipants:    8,
				GameID:             uuidPtr(uuid.MustParse("22222222-2222-2222-2222-222222222222")),
			},
			queryID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			expectedResult: &TournamentInfo{
				ID:              uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				Name:            "Test Tournament",
				HostID:          defaultHostID,
				GameID:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				MinParticipants: 8,
				Status:          "registration_open",
			},
			expectNil:   false,
			expectError: false,
		},
		{
			name: "tournament exists without game ID",
			setupTournament: &testTournament{
				ID:                 uuid.MustParse("33333333-3333-3333-3333-333333333333"),
				HostID:             defaultHostID,
				Name:               "No Game Tournament",
				Description:        nil,
				Status:             "draft",
				ScheduledStartTime: nil,
				MinParticipants:    4,
				GameID:             nil,
			},
			queryID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			expectedResult: &TournamentInfo{
				ID:              uuid.MustParse("33333333-3333-3333-3333-333333333333"),
				Name:            "No Game Tournament",
				HostID:          defaultHostID,
				GameID:          uuid.UUID{}, // Empty when not set
				MinParticipants: 4,
				Status:          "draft",
			},
			expectNil:   false,
			expectError: false,
		},
		{
			name:            "tournament not found",
			setupTournament: nil,
			queryID:         uuid.MustParse("99999999-9999-9999-9999-999999999999"),
			expectedResult:  nil,
			expectNil:       true,
			expectError:     false,
		},
		{
			name: "tournament with minimum participants less than 2 defaults to 2",
			setupTournament: &testTournament{
				ID:                 uuid.MustParse("44444444-4444-4444-4444-444444444444"),
				HostID:             defaultHostID,
				Name:               "Small Tournament",
				Description:        nil,
				Status:             "started",
				ScheduledStartTime: nil,
				MinParticipants:    1, // Less than 2
				GameID:             nil,
			},
			queryID: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
			expectedResult: &TournamentInfo{
				ID:              uuid.MustParse("44444444-4444-4444-4444-444444444444"),
				Name:            "Small Tournament",
				HostID:          defaultHostID,
				GameID:          uuid.UUID{},
				MinParticipants: 2, // Should default to 2
				Status:          "started",
			},
			expectNil:   false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test data if provided
			if tt.setupTournament != nil {
				err := insertTestTournament(ctx, env.pool, *tt.setupTournament)
				require.NoError(t, err, "failed to insert test tournament")
			}

			// Execute
			result, err := client.GetTournamentInfo(ctx, tt.queryID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.expectNil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedResult.ID, result.ID)
			assert.Equal(t, tt.expectedResult.Name, result.Name)
			assert.Equal(t, tt.expectedResult.HostID, result.HostID)
			assert.Equal(t, tt.expectedResult.GameID, result.GameID)
			assert.Equal(t, tt.expectedResult.MinParticipants, result.MinParticipants)
			assert.Equal(t, tt.expectedResult.Status, result.Status)
		})
	}
}

func TestCompetitionClient_IsUserRegistered_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup test environment
	env := setupCompetitionTestEnv(ctx, t)
	defer env.cleanup(ctx)

	// Create competition client
	logger := observability.NewLogger("info", "text")
	client := NewCompetitionClient(env.competitionURL, logger)

	// Default host ID from migrations
	defaultHostID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tournamentID := uuid.New()
	registeredUserID := uuid.New()
	unregisteredUserID := uuid.New()

	// Create tournament
	err := insertTestTournament(ctx, env.pool, testTournament{
		ID:              tournamentID,
		HostID:          defaultHostID,
		Name:            "Registration Test Tournament",
		Status:          "registration_open",
		MinParticipants: 8,
	})
	require.NoError(t, err, "failed to insert test tournament")

	// Create registration for one user
	_, err = env.pool.Exec(ctx, `
		INSERT INTO tournament_registrations (tournament_id, user_id, status, display_name)
		VALUES ($1, $2, 'confirmed', 'Test Player')
	`, tournamentID, registeredUserID)
	require.NoError(t, err, "failed to insert test registration")

	tests := []struct {
		name         string
		tournamentID uuid.UUID
		userID       uuid.UUID
		expected     bool
	}{
		{
			name:         "user is registered",
			tournamentID: tournamentID,
			userID:       registeredUserID,
			expected:     true,
		},
		{
			name:         "user is not registered",
			tournamentID: tournamentID,
			userID:       unregisteredUserID,
			expected:     false,
		},
		{
			name:         "tournament does not exist",
			tournamentID: uuid.New(),
			userID:       registeredUserID,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache before each test
			client.ClearCache()

			result, err := client.IsUserRegistered(ctx, tt.tournamentID, tt.userID)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}
