// Package testutil provides test utilities for integration testing
package testutil

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDB holds test database resources
type TestDB struct {
	Pool      *pgxpool.Pool
	Container *postgres.PostgresContainer
	ConnStr   string
}

// SetupTestDB creates a PostgreSQL testcontainer and runs migrations
func SetupTestDB(ctx context.Context) (*TestDB, error) {
	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("matchmaking_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		tc.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, err
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = postgresContainer.Terminate(ctx)
		return nil, err
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		_ = postgresContainer.Terminate(ctx)
		return nil, err
	}

	// Run migrations
	if err := RunMigrations(ctx, pool); err != nil {
		pool.Close()
		_ = postgresContainer.Terminate(ctx)
		return nil, err
	}

	return &TestDB{
		Pool:      pool,
		Container: postgresContainer,
		ConnStr:   connStr,
	}, nil
}

// Cleanup closes the pool and terminates the container
func (db *TestDB) Cleanup(ctx context.Context) {
	if db.Pool != nil {
		db.Pool.Close()
	}
	if db.Container != nil {
		_ = db.Container.Terminate(ctx)
	}
}

// RunMigrations executes all SQL migration files from the migrations directory
// It skips duplicate migrations (e.g., 000007 and 000014 both add game_snapshot)
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, b, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(b), "../../migrations")

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	// Sort migration files by name (they are numbered)
	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Track which migrations have been applied (by content hash to detect duplicates)
	appliedPatterns := make(map[string]bool)

	// Execute each migration
	for _, fileName := range sqlFiles {
		// Skip known duplicate migrations
		if isDuplicateMigration(fileName, appliedPatterns) {
			continue
		}

		filePath := filepath.Join(migrationsDir, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		if _, err := pool.Exec(ctx, string(content)); err != nil {
			// Log but continue for certain errors (like duplicate column)
			if strings.Contains(err.Error(), "already exists") {
				continue
			}
			return err
		}
	}

	return nil
}

// isDuplicateMigration checks if a migration is a known duplicate
func isDuplicateMigration(fileName string, applied map[string]bool) bool {
	// Map of known duplicate migrations
	duplicates := map[string]string{
		"000014_add_game_snapshot_to_brackets.sql": "000007_add_game_snapshot_to_brackets.sql",
		"000012_allow_bracket_generation_activity.sql": "000006_allow_bracket_generation_activity.sql",
	}

	if original, isDupe := duplicates[fileName]; isDupe {
		if applied[original] {
			return true
		}
	}

	// Mark this migration as applied
	applied[fileName] = true
	return false
}

// TruncateTables truncates all test tables for clean state between tests
func TruncateTables(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE
			tournament_matches,
			tournament_rounds,
			tournament_brackets,
			prelobby_activity_feed,
			prelobby_participant_snapshots,
			prelobbies
		CASCADE;
	`)
	return err
}

// GenerateTestUsers generates n test user UUIDs
func GenerateTestUsers(n int) []uuid.UUID {
	users := make([]uuid.UUID, n)
	for i := range n {
		users[i] = uuid.New()
	}
	return users
}

// UUIDToPgtype converts uuid.UUID to pgtype.UUID
func UUIDToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// IntToPgtype converts int to pgtype.Int4
func IntToPgtype(i int) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(i), Valid: true}
}

// StringToPgtype converts string to pgtype.Text
func StringToPgtype(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}
