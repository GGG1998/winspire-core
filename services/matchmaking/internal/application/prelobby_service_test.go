package application

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/repository"
	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
	"github.com/winspire-core/services/matchmaking/internal/testutil"
)

var testDB *testutil.TestDB

// TestMain sets up and tears down the test database
func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	testDB, err = testutil.SetupTestDB(ctx)
	if err != nil {
		panic("failed to setup test database: " + err.Error())
	}

	code := m.Run()

	testDB.Cleanup(ctx)
	os.Exit(code)
}

func TestCanAcceptParticipants(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Truncate tables for clean state
	err := testutil.TruncateTables(ctx, testDB.Pool)
	require.NoError(t, err)

	// Create repositories and service
	queries := sqlc.New(testDB.Pool)
	preLobbyRepo := repository.NewPreLobbyRepository(queries, testDB.Pool, testDB.Pool)
	matchRepo := repository.NewMatchRepository(queries)
	roundRepo := repository.NewRoundRepository(queries)
	bracketRepo := repository.NewBracketRepository(queries, testDB.Pool, testDB.Pool)
	logger := observability.NewLogger("info", "text")
	participantStore := NewInMemoryPreLobbyParticipantStore()
	service := NewPreLobbyService(preLobbyRepo, matchRepo, roundRepo, nil, nil, nil, logger, participantStore)

	// Generate 10 test users
	users := testutil.GenerateTestUsers(10)

	// Test cases for status-based acceptance
	statusTests := []struct {
		name           string
		setupStatus    func(ctx context.Context, tournamentID uuid.UUID) error
		expectedAccept bool
		expectedReason string
	}{
		{
			name:           "no pre-lobby exists - should accept",
			setupStatus:    nil, // no setup needed
			expectedAccept: true,
			expectedReason: "",
		},
		{
			name: "waiting status - should accept",
			setupStatus: func(ctx context.Context, tournamentID uuid.UUID) error {
				preLobby, err := domain.NewPreLobby(tournamentID, 2)
				if err != nil {
					return err
				}
				return preLobbyRepo.Create(ctx, preLobby)
			},
			expectedAccept: true,
			expectedReason: "",
		},
		{
			name: "grace_period status - should accept",
			setupStatus: func(ctx context.Context, tournamentID uuid.UUID) error {
				preLobby, err := domain.NewPreLobby(tournamentID, 2)
				if err != nil {
					return err
				}
				if err := preLobbyRepo.Create(ctx, preLobby); err != nil {
					return err
				}
				_, err = preLobbyRepo.StartGracePeriod(ctx, tournamentID)
				return err
			},
			expectedAccept: true,
			expectedReason: "",
		},
		{
			name: "generating_bracket status - should NOT accept",
			setupStatus: func(ctx context.Context, tournamentID uuid.UUID) error {
				preLobby, err := domain.NewPreLobby(tournamentID, 2)
				if err != nil {
					return err
				}
				if err := preLobbyRepo.Create(ctx, preLobby); err != nil {
					return err
				}
				if _, err := preLobbyRepo.StartGracePeriod(ctx, tournamentID); err != nil {
					return err
				}
				_, err = preLobbyRepo.UpdateStatus(ctx, tournamentID, domain.PreLobbyStatusGeneratingBracket)
				return err
			},
			expectedAccept: false,
			expectedReason: "Pre-lobby is not accepting new participants",
		},
		{
			name: "started status - should NOT accept",
			setupStatus: func(ctx context.Context, tournamentID uuid.UUID) error {
				preLobby, err := domain.NewPreLobby(tournamentID, 2)
				if err != nil {
					return err
				}
				if err := preLobbyRepo.Create(ctx, preLobby); err != nil {
					return err
				}
				if _, err := preLobbyRepo.StartGracePeriod(ctx, tournamentID); err != nil {
					return err
				}
				if _, err := preLobbyRepo.UpdateStatus(ctx, tournamentID, domain.PreLobbyStatusGeneratingBracket); err != nil {
					return err
				}
				_, err = preLobbyRepo.UpdateStatus(ctx, tournamentID, domain.PreLobbyStatusStarted)
				return err
			},
			expectedAccept: false,
			expectedReason: "Pre-lobby is not accepting new participants",
		},
		{
			name: "cancelled status - should NOT accept",
			setupStatus: func(ctx context.Context, tournamentID uuid.UUID) error {
				preLobby, err := domain.NewPreLobby(tournamentID, 2)
				if err != nil {
					return err
				}
				if err := preLobbyRepo.Create(ctx, preLobby); err != nil {
					return err
				}
				_, err = preLobbyRepo.UpdateStatus(ctx, tournamentID, domain.PreLobbyStatusCancelled)
				return err
			},
			expectedAccept: false,
			expectedReason: "Pre-lobby is not accepting new participants",
		},
	}

	for _, tt := range statusTests {
		t.Run(tt.name, func(t *testing.T) {
			tournamentID := uuid.New()
			participantID := users[0]

			if tt.setupStatus != nil {
				err := tt.setupStatus(ctx, tournamentID)
				require.NoError(t, err)
			}

			canAccept, reason, err := service.CanAcceptParticipants(ctx, tournamentID, &participantID)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedAccept, canAccept)
			assert.Equal(t, tt.expectedReason, reason)
		})
	}

	t.Run("without participant ID - only checks pre-lobby status", func(t *testing.T) {
		tournamentID := uuid.New()
		preLobby, err := domain.NewPreLobby(tournamentID, 2)
		require.NoError(t, err)
		err = preLobbyRepo.Create(ctx, preLobby)
		require.NoError(t, err)

		canAccept, reason, err := service.CanAcceptParticipants(ctx, tournamentID, nil)

		require.NoError(t, err)
		assert.True(t, canAccept)
		assert.Empty(t, reason)
	})

	t.Run("full tournament with 10 participants in grace_period - all accepted", func(t *testing.T) {
		tournamentID := uuid.New()

		preLobby, err := domain.NewPreLobby(tournamentID, 8)
		require.NoError(t, err)
		err = preLobbyRepo.Create(ctx, preLobby)
		require.NoError(t, err)
		_, err = preLobbyRepo.StartGracePeriod(ctx, tournamentID)
		require.NoError(t, err)

		for i, user := range users {
			canAccept, reason, err := service.CanAcceptParticipants(ctx, tournamentID, &user)
			require.NoError(t, err)
			assert.True(t, canAccept, "user[%d] should be accepted during grace_period", i)
			assert.Empty(t, reason)
		}
	})

	t.Run("tournament with brackets and matches - started status rejects all", func(t *testing.T) {
		tournamentID := uuid.New()

		// Setup pre-lobby in started status
		preLobby, err := domain.NewPreLobby(tournamentID, 8)
		require.NoError(t, err)
		err = preLobbyRepo.Create(ctx, preLobby)
		require.NoError(t, err)
		_, err = preLobbyRepo.StartGracePeriod(ctx, tournamentID)
		require.NoError(t, err)
		_, err = preLobbyRepo.UpdateStatus(ctx, tournamentID, domain.PreLobbyStatusGeneratingBracket)
		require.NoError(t, err)
		_, err = preLobbyRepo.UpdateStatus(ctx, tournamentID, domain.PreLobbyStatusStarted)
		require.NoError(t, err)

		// Create bracket with 10 participants
		bracket, err := domain.NewBracket(tournamentID, users)
		require.NoError(t, err)

		rounds, err := bracket.GenerateRounds(users)
		require.NoError(t, err)

		// Generate matches for round 1
		matches := make([]domain.Match, 0)
		for i := 0; i < len(users)/2; i++ {
			match := domain.Match{
				ID:             uuid.New(),
				RoundID:        rounds[0].ID,
				MatchNumber:    i + 1,
				Participant1ID: users[i*2],
				Participant2ID: &users[i*2+1],
				Status:         domain.MatchStatusPending,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			matches = append(matches, match)
		}

		err = bracketRepo.Create(ctx, bracket, rounds, matches)
		require.NoError(t, err)

		// Complete some matches
		matchResults := []struct {
			matchIdx int
			winnerID uuid.UUID
		}{
			{0, users[1]}, // user[0] loses to user[1]
			{1, users[3]}, // user[2] loses to user[3]
			{2, users[4]}, // user[4] wins against user[5]
		}

		for _, mr := range matchResults {
			err := queries.UpdateMatchResult(ctx, sqlc.UpdateMatchResultParams{
				ID:           testutil.UUIDToPgtype(matches[mr.matchIdx].ID),
				WinnerID:     testutil.UUIDToPgtype(mr.winnerID),
				ScorePlayer1: testutil.IntToPgtype(1),
				ScorePlayer2: testutil.IntToPgtype(0),
				ResultSource: testutil.StringToPgtype("host_manual"),
			})
			require.NoError(t, err)
		}

		// All participants should be rejected - started status doesn't accept
		// Note: The elimination check code is unreachable because started status
		// fails CanAcceptParticipants() domain check first
		for i, user := range users {
			canAccept, reason, err := service.CanAcceptParticipants(ctx, tournamentID, &user)
			require.NoError(t, err)
			assert.False(t, canAccept, "user[%d] should be rejected - started status", i)
			assert.Equal(t, "Pre-lobby is not accepting new participants", reason)
		}
	})
}
