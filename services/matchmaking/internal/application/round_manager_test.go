package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
)

// ==================== Mock Repositories ====================

type MockMatchRepository struct {
	mock.Mock
}

func (m *MockMatchRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Match, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Match), args.Error(1)
}

func (m *MockMatchRepository) GetByRoundID(ctx context.Context, roundID uuid.UUID) ([]domain.Match, error) {
	args := m.Called(ctx, roundID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Match), args.Error(1)
}

func (m *MockMatchRepository) GetByTournamentAndRound(ctx context.Context, tournamentID uuid.UUID, roundNumber int) ([]domain.Match, error) {
	args := m.Called(ctx, tournamentID, roundNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Match), args.Error(1)
}

func (m *MockMatchRepository) GetForPlayer(ctx context.Context, playerID uuid.UUID) ([]domain.Match, error) {
	args := m.Called(ctx, playerID)
	return args.Get(0).([]domain.Match), args.Error(1)
}

func (m *MockMatchRepository) GetCurrentForUser(ctx context.Context, userID uuid.UUID) (*domain.Match, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Match), args.Error(1)
}

func (m *MockMatchRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.MatchStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockMatchRepository) UpdateReady(ctx context.Context, id uuid.UUID, playerID uuid.UUID, ready bool) error {
	args := m.Called(ctx, id, playerID, ready)
	return args.Error(0)
}

func (m *MockMatchRepository) UpdateResult(ctx context.Context, id uuid.UUID, winnerID uuid.UUID, scorePlayer1, scorePlayer2 int, source domain.ResultSource) error {
	args := m.Called(ctx, id, winnerID, scorePlayer1, scorePlayer2, source)
	return args.Error(0)
}

func (m *MockMatchRepository) UpdateDisconnect(ctx context.Context, id uuid.UUID, playerID uuid.UUID, disconnectedAt *time.Time) error {
	args := m.Called(ctx, id, playerID, disconnectedAt)
	return args.Error(0)
}

func (m *MockMatchRepository) UpdateScore(ctx context.Context, id uuid.UUID, scorePlayer1, scorePlayer2 int) error {
	args := m.Called(ctx, id, scorePlayer1, scorePlayer2)
	return args.Error(0)
}

func (m *MockMatchRepository) UpdateDisconnectedPlayer(ctx context.Context, id uuid.UUID, playerID uuid.UUID, disconnectedAt time.Time) error {
	args := m.Called(ctx, id, playerID, disconnectedAt)
	return args.Error(0)
}

func (m *MockMatchRepository) ClearDisconnectedPlayer(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMatchRepository) UpdateGameLoadedAtomic(ctx context.Context, id uuid.UUID, playerID uuid.UUID) (*domain.Match, error) {
	args := m.Called(ctx, id, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Match), args.Error(1)
}

func (m *MockMatchRepository) FindLoadingMatchesOlderThan(ctx context.Context, duration time.Duration) ([]domain.Match, error) {
	args := m.Called(ctx, duration)
	return args.Get(0).([]domain.Match), args.Error(1)
}

func (m *MockMatchRepository) HasParticipantLostInTournament(ctx context.Context, tournamentID, participantID uuid.UUID) (bool, error) {
	args := m.Called(ctx, tournamentID, participantID)
	return args.Bool(0), args.Error(1)
}

func (m *MockMatchRepository) AssignWinnerToNextMatch(ctx context.Context, nextMatchID uuid.UUID, winnerID uuid.UUID) error {
	args := m.Called(ctx, nextMatchID, winnerID)
	return args.Error(0)
}

type MockRoundRepository struct {
	mock.Mock
}

func (m *MockRoundRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Round, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Round), args.Error(1)
}

func (m *MockRoundRepository) GetByBracketID(ctx context.Context, bracketID uuid.UUID) ([]domain.Round, error) {
	args := m.Called(ctx, bracketID)
	return args.Get(0).([]domain.Round), args.Error(1)
}

func (m *MockRoundRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.RoundStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockBracketRepository struct {
	mock.Mock
}

func (m *MockBracketRepository) Create(ctx context.Context, bracket *domain.Bracket, rounds []domain.Round, matches []domain.Match) error {
	args := m.Called(ctx, bracket, rounds, matches)
	return args.Error(0)
}

func (m *MockBracketRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bracket, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Bracket), args.Error(1)
}

func (m *MockBracketRepository) GetByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*domain.Bracket, error) {
	args := m.Called(ctx, tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Bracket), args.Error(1)
}

func (m *MockBracketRepository) GetWithRoundsAndMatches(ctx context.Context, tournamentID uuid.UUID) (*domain.Bracket, []domain.Round, []domain.Match, error) {
	args := m.Called(ctx, tournamentID)
	return args.Get(0).(*domain.Bracket), args.Get(1).([]domain.Round), args.Get(2).([]domain.Match), args.Error(3)
}

func (m *MockBracketRepository) UpdateCompletedAt(ctx context.Context, id uuid.UUID, completedAt time.Time) error {
	args := m.Called(ctx, id, completedAt)
	return args.Error(0)
}

// ==================== Mock PreLobbyService ====================

type MockPreLobbyService struct {
	mock.Mock
}

func (m *MockPreLobbyService) PrepareForNextRound(ctx context.Context, tournamentID uuid.UUID, expectedParticipants int) error {
	args := m.Called(ctx, tournamentID, expectedParticipants)
	return args.Error(0)
}

func (m *MockPreLobbyService) StartGracePeriod(ctx context.Context, tournamentID uuid.UUID, onComplete func(uuid.UUID, []uuid.UUID)) error {
	args := m.Called(ctx, tournamentID, onComplete)
	return args.Error(0)
}

// ==================== Test Helpers ====================

// MockPreLobbyServiceForRoundManager is a minimal mock for PreLobbyService
type MockPreLobbyServiceForRoundManager struct{}

func (m *MockPreLobbyServiceForRoundManager) PrepareForNextRound(_ context.Context, _ uuid.UUID, _ int) error {
	return nil
}

func (m *MockPreLobbyServiceForRoundManager) StartGracePeriod(_ context.Context, _ uuid.UUID, _ func(uuid.UUID, []uuid.UUID)) error {
	return nil
}

// createTestRoundManagerWithMocks creates RoundManager with mock PreLobbyService that won't panic
func createTestRoundManager(
	matchRepo *MockMatchRepository,
	roundRepo *MockRoundRepository,
	bracketRepo *MockBracketRepository,
	transitionStore RoundTransitionStoreInterface,
) *RoundManager {
	logger := observability.NewLogger("info", "text")
	if transitionStore == nil {
		transitionStore = NewInMemoryRoundTransitionStore(logger)
	}
	// Note: We pass nil for preLobbyService and matchAssignmentService
	// The RoundManager handles nil preLobbyService gracefully
	return NewRoundManager(matchRepo, roundRepo, bracketRepo, nil, nil, transitionStore, logger)
}

func createTestMatch(roundID uuid.UUID, matchNumber int, p1, p2 uuid.UUID, status domain.MatchStatus, nextMatchID *uuid.UUID) *domain.Match {
	return &domain.Match{
		ID:             uuid.New(),
		RoundID:        roundID,
		MatchNumber:    matchNumber,
		Participant1ID: p1,
		Participant2ID: &p2,
		Status:         status,
		NextMatchID:    nextMatchID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// ==================== Tests for OnMatchCompleted ====================

func TestRoundManager_OnMatchCompleted(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                 string
		setupMocks           func(*MockMatchRepository, *MockRoundRepository, *MockBracketRepository) (matchID, winnerID, tournamentID uuid.UUID)
		expectedError        bool
		expectedWinnersCount int
		checkState           func(*testing.T, *RoundManager, uuid.UUID)
	}{
		{
			name: "tracks winner when match completed and round not complete",
			setupMocks: func(matchRepo *MockMatchRepository, roundRepo *MockRoundRepository, bracketRepo *MockBracketRepository) (uuid.UUID, uuid.UUID, uuid.UUID) {
				tournamentID := uuid.New()
				bracketID := uuid.New()
				roundID := uuid.New()
				nextMatchID := uuid.New()
				matchID := uuid.New()
				p1, p2 := uuid.New(), uuid.New()
				winnerID := p1

				match := &domain.Match{
					ID:             matchID,
					RoundID:        roundID,
					MatchNumber:    1,
					Participant1ID: p1,
					Participant2ID: &p2,
					Status:         domain.MatchStatusCompleted,
					NextMatchID:    &nextMatchID, // Not final round
				}

				round := &domain.Round{
					ID:          roundID,
					BracketID:   bracketID,
					RoundNumber: 1,
				}

				bracket := &domain.Bracket{
					ID:           bracketID,
					TournamentID: tournamentID,
				}

				// Two matches in round, only one completed
				allMatches := []domain.Match{
					*match,
					{
						ID:             uuid.New(),
						RoundID:        roundID,
						MatchNumber:    2,
						Participant1ID: uuid.New(),
						Participant2ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
						Status:         domain.MatchStatusPending, // Not completed
						NextMatchID:    &nextMatchID,
					},
				}

				matchRepo.On("GetByID", ctx, matchID).Return(match, nil)
				roundRepo.On("GetByID", ctx, roundID).Return(round, nil)
				bracketRepo.On("GetByID", ctx, bracketID).Return(bracket, nil)
				matchRepo.On("GetByRoundID", ctx, roundID).Return(allMatches, nil)

				return matchID, winnerID, tournamentID
			},
			expectedError:        false,
			expectedWinnersCount: 1,
			checkState: func(t *testing.T, rm *RoundManager, tournamentID uuid.UUID) {
				state := rm.GetTransitionState(ctx, tournamentID)
				require.NotNil(t, state, "transition state should exist")
				assert.Len(t, state.ExpectedWinners, 1, "should have 1 expected winner")
				assert.False(t, state.GraceStarted, "grace should not start yet")
			},
		},
		{
			name: "skips final round match (NextMatchID is nil)",
			setupMocks: func(matchRepo *MockMatchRepository, roundRepo *MockRoundRepository, bracketRepo *MockBracketRepository) (uuid.UUID, uuid.UUID, uuid.UUID) {
				roundID := uuid.New()
				matchID := uuid.New()
				p1, p2 := uuid.New(), uuid.New()
				winnerID := p1
				tournamentID := uuid.New() // Won't be used since we skip final round

				match := &domain.Match{
					ID:             matchID,
					RoundID:        roundID,
					MatchNumber:    1,
					Participant1ID: p1,
					Participant2ID: &p2,
					Status:         domain.MatchStatusCompleted,
					NextMatchID:    nil, // Final round!
				}

				matchRepo.On("GetByID", ctx, matchID).Return(match, nil)

				return matchID, winnerID, tournamentID
			},
			expectedError:        false,
			expectedWinnersCount: 0,
			checkState: func(t *testing.T, rm *RoundManager, tournamentID uuid.UUID) {
				// No state should be created for final round
			},
		},
		{
			name: "logs round not complete when some matches pending",
			setupMocks: func(matchRepo *MockMatchRepository, roundRepo *MockRoundRepository, bracketRepo *MockBracketRepository) (uuid.UUID, uuid.UUID, uuid.UUID) {
				tournamentID := uuid.New()
				bracketID := uuid.New()
				roundID := uuid.New()
				nextMatchID := uuid.New()
				matchID := uuid.New()
				p1, p2 := uuid.New(), uuid.New()
				winnerID := p1

				match := &domain.Match{
					ID:             matchID,
					RoundID:        roundID,
					MatchNumber:    1,
					Participant1ID: p1,
					Participant2ID: &p2,
					Status:         domain.MatchStatusCompleted,
					NextMatchID:    &nextMatchID,
				}

				round := &domain.Round{
					ID:          roundID,
					BracketID:   bracketID,
					RoundNumber: 1,
				}

				bracket := &domain.Bracket{
					ID:           bracketID,
					TournamentID: tournamentID,
				}

				// 4 matches in round, only 2 completed
				allMatches := []domain.Match{
					*match,
					{ID: uuid.New(), RoundID: roundID, Status: domain.MatchStatusCompleted, NextMatchID: &nextMatchID},
					{ID: uuid.New(), RoundID: roundID, Status: domain.MatchStatusPending, NextMatchID: &nextMatchID},
					{ID: uuid.New(), RoundID: roundID, Status: domain.MatchStatusStarted, NextMatchID: &nextMatchID},
				}

				matchRepo.On("GetByID", ctx, matchID).Return(match, nil)
				roundRepo.On("GetByID", ctx, roundID).Return(round, nil)
				bracketRepo.On("GetByID", ctx, bracketID).Return(bracket, nil)
				matchRepo.On("GetByRoundID", ctx, roundID).Return(allMatches, nil)

				return matchID, winnerID, tournamentID
			},
			expectedError:        false,
			expectedWinnersCount: 1,
			checkState: func(t *testing.T, rm *RoundManager, tournamentID uuid.UUID) {
				state := rm.GetTransitionState(ctx, tournamentID)
				require.NotNil(t, state)
				assert.False(t, state.GraceStarted, "grace should not start - round not complete")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchRepo := new(MockMatchRepository)
			roundRepo := new(MockRoundRepository)
			bracketRepo := new(MockBracketRepository)

			matchID, winnerID, tournamentID := tt.setupMocks(matchRepo, roundRepo, bracketRepo)

			rm := createTestRoundManager(matchRepo, roundRepo, bracketRepo, nil)

			err := rm.OnMatchCompleted(ctx, matchID, winnerID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkState != nil {
				tt.checkState(t, rm, tournamentID)
			}

			matchRepo.AssertExpectations(t)
			roundRepo.AssertExpectations(t)
			bracketRepo.AssertExpectations(t)
		})
	}
}

// ==================== Tests for OnParticipantJoinedPreLobby ====================

func TestRoundManager_OnParticipantJoinedPreLobby(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		setupState       func(*InMemoryRoundTransitionStore, uuid.UUID) []uuid.UUID // returns expected winners
		participantID    uuid.UUID
		checkState       func(*testing.T, *RoundManager, uuid.UUID)
		expectGraceStart bool
	}{
		{
			name: "ignores participant when no pending transition exists",
			setupState: func(store *InMemoryRoundTransitionStore, tournamentID uuid.UUID) []uuid.UUID {
				// No state setup - no pending transition
				return nil
			},
			participantID: uuid.New(),
			checkState: func(t *testing.T, rm *RoundManager, tournamentID uuid.UUID) {
				state := rm.GetTransitionState(ctx, tournamentID)
				assert.Nil(t, state, "no state should exist")
			},
			expectGraceStart: false,
		},
		{
			name: "ignores participant who is not an expected winner",
			setupState: func(store *InMemoryRoundTransitionStore, tournamentID uuid.UUID) []uuid.UUID {
				expectedWinners := []uuid.UUID{uuid.New(), uuid.New()}
				store.Save(ctx, &RoundTransitionData{
					TournamentID:    tournamentID,
					CompletedRound:  1,
					NextRound:       2,
					ExpectedWinners: expectedWinners,
					JoinedWinners:   []uuid.UUID{},
					GraceStarted:    false,
				})
				return expectedWinners
			},
			participantID: uuid.New(), // Different from expected winners
			checkState: func(t *testing.T, rm *RoundManager, tournamentID uuid.UUID) {
				state := rm.GetTransitionState(ctx, tournamentID)
				require.NotNil(t, state)
				assert.Empty(t, state.JoinedWinners, "no winners should have joined")
				assert.False(t, state.GraceStarted)
			},
			expectGraceStart: false,
		},
		{
			name: "tracks winner joining pre-lobby",
			setupState: func(store *InMemoryRoundTransitionStore, tournamentID uuid.UUID) []uuid.UUID {
				expectedWinners := []uuid.UUID{uuid.New(), uuid.New()}
				store.Save(ctx, &RoundTransitionData{
					TournamentID:    tournamentID,
					CompletedRound:  1,
					NextRound:       2,
					ExpectedWinners: expectedWinners,
					JoinedWinners:   []uuid.UUID{},
					GraceStarted:    false,
				})
				return expectedWinners
			},
			participantID: uuid.UUID{}, // Will be set to first expected winner
			checkState: func(t *testing.T, rm *RoundManager, tournamentID uuid.UUID) {
				state := rm.GetTransitionState(ctx, tournamentID)
				require.NotNil(t, state)
				assert.Len(t, state.JoinedWinners, 1, "one winner should have joined")
				assert.False(t, state.GraceStarted, "grace should not start - not all winners")
			},
			expectGraceStart: false,
		},
		{
			name: "prevents double grace period start",
			setupState: func(store *InMemoryRoundTransitionStore, tournamentID uuid.UUID) []uuid.UUID {
				expectedWinners := []uuid.UUID{uuid.New()}
				store.Save(ctx, &RoundTransitionData{
					TournamentID:    tournamentID,
					CompletedRound:  1,
					NextRound:       2,
					ExpectedWinners: expectedWinners,
					JoinedWinners:   []uuid.UUID{},
					GraceStarted:    true, // Already started
				})
				return expectedWinners
			},
			participantID: uuid.UUID{}, // Will be set to expected winner
			checkState: func(t *testing.T, rm *RoundManager, tournamentID uuid.UUID) {
				state := rm.GetTransitionState(ctx, tournamentID)
				require.NotNil(t, state)
				assert.True(t, state.GraceStarted, "grace should remain true")
			},
			expectGraceStart: false, // Already started, shouldn't trigger again
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchRepo := new(MockMatchRepository)
			roundRepo := new(MockRoundRepository)
			bracketRepo := new(MockBracketRepository)
			logger := observability.NewLogger("info", "text")
			transitionStore := NewInMemoryRoundTransitionStore(logger)

			rm := createTestRoundManager(matchRepo, roundRepo, bracketRepo, transitionStore)
			tournamentID := uuid.New()

			expectedWinners := tt.setupState(transitionStore, tournamentID)

			// Use first expected winner if participantID is empty
			participantID := tt.participantID
			if participantID == uuid.Nil && len(expectedWinners) > 0 {
				participantID = expectedWinners[0]
			}

			rm.OnParticipantJoinedPreLobby(ctx, tournamentID, participantID)

			tt.checkState(t, rm, tournamentID)
		})
	}
}

// ==================== Tests for GetTransitionState ====================

func TestRoundManager_GetTransitionState(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setupState func(*InMemoryRoundTransitionStore, uuid.UUID)
		expectNil  bool
	}{
		{
			name: "returns nil when no state exists",
			setupState: func(store *InMemoryRoundTransitionStore, tournamentID uuid.UUID) {
				// No setup - no state
			},
			expectNil: true,
		},
		{
			name: "returns state when it exists",
			setupState: func(store *InMemoryRoundTransitionStore, tournamentID uuid.UUID) {
				store.Save(ctx, &RoundTransitionData{
					TournamentID:    tournamentID,
					CompletedRound:  1,
					NextRound:       2,
					ExpectedWinners: []uuid.UUID{},
					JoinedWinners:   []uuid.UUID{},
				})
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := observability.NewLogger("info", "text")
			transitionStore := NewInMemoryRoundTransitionStore(logger)
			rm := createTestRoundManager(nil, nil, nil, transitionStore)
			tournamentID := uuid.New()

			tt.setupState(transitionStore, tournamentID)

			state := rm.GetTransitionState(ctx, tournamentID)

			if tt.expectNil {
				assert.Nil(t, state)
			} else {
				assert.NotNil(t, state)
				assert.Equal(t, tournamentID, state.TournamentID)
			}
		})
	}
}

// ==================== Integration Tests for Complete Round Flow ====================

func TestRoundManager_CompleteRoundFlow(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		numMatches      int
		setupAndRun     func(*testing.T, *RoundManager, *MockMatchRepository, *MockRoundRepository, *MockBracketRepository) (tournamentID uuid.UUID, winners []uuid.UUID)
		checkFinalState func(*testing.T, *RoundManager, uuid.UUID, []uuid.UUID)
	}{
		{
			name:       "2-match round completes and all winners join triggers grace",
			numMatches: 2,
			setupAndRun: func(t *testing.T, rm *RoundManager, matchRepo *MockMatchRepository, roundRepo *MockRoundRepository, bracketRepo *MockBracketRepository) (uuid.UUID, []uuid.UUID) {
				tournamentID := uuid.New()
				bracketID := uuid.New()
				roundID := uuid.New()
				nextMatchID := uuid.New()

				// Create 2 matches with different winners
				winners := []uuid.UUID{uuid.New(), uuid.New()}
				match1ID := uuid.New()
				match2ID := uuid.New()
				p1, p2, p3, p4 := uuid.New(), uuid.New(), uuid.New(), uuid.New()

				round := &domain.Round{
					ID:          roundID,
					BracketID:   bracketID,
					RoundNumber: 1,
				}

				bracket := &domain.Bracket{
					ID:           bracketID,
					TournamentID: tournamentID,
				}

				// Match 1 - will complete first
				match1 := &domain.Match{
					ID:             match1ID,
					RoundID:        roundID,
					MatchNumber:    1,
					Participant1ID: p1,
					Participant2ID: &p2,
					Status:         domain.MatchStatusCompleted,
					NextMatchID:    &nextMatchID,
				}

				// Match 2 - will complete second
				match2 := &domain.Match{
					ID:             match2ID,
					RoundID:        roundID,
					MatchNumber:    2,
					Participant1ID: p3,
					Participant2ID: &p4,
					Status:         domain.MatchStatusCompleted,
					NextMatchID:    &nextMatchID,
				}

				// Setup mocks for match 1 completion
				matchRepo.On("GetByID", ctx, match1ID).Return(match1, nil)
				roundRepo.On("GetByID", ctx, roundID).Return(round, nil).Once()
				bracketRepo.On("GetByID", ctx, bracketID).Return(bracket, nil).Once()
				// After match 1, round not complete (match 2 still pending)
				matchRepo.On("GetByRoundID", ctx, roundID).Return([]domain.Match{
					*match1,
					{ID: match2ID, RoundID: roundID, Status: domain.MatchStatusPending, NextMatchID: &nextMatchID},
				}, nil).Once()

				// Complete match 1
				err := rm.OnMatchCompleted(ctx, match1ID, winners[0])
				require.NoError(t, err)

				// Setup mocks for match 2 completion
				matchRepo.On("GetByID", ctx, match2ID).Return(match2, nil)
				roundRepo.On("GetByID", ctx, roundID).Return(round, nil).Once()
				bracketRepo.On("GetByID", ctx, bracketID).Return(bracket, nil).Once()
				roundRepo.On("UpdateStatus", ctx, roundID, domain.RoundStatusCompleted).Return(nil).Once()
				// After match 2, round IS complete (both matches completed)
				matchRepo.On("GetByRoundID", ctx, roundID).Return([]domain.Match{
					{ID: match1ID, RoundID: roundID, Status: domain.MatchStatusCompleted, NextMatchID: &nextMatchID},
					{ID: match2ID, RoundID: roundID, Status: domain.MatchStatusCompleted, NextMatchID: &nextMatchID},
				}, nil).Once()

				// Complete match 2
				err = rm.OnMatchCompleted(ctx, match2ID, winners[1])
				require.NoError(t, err)

				return tournamentID, winners
			},
			checkFinalState: func(t *testing.T, rm *RoundManager, tournamentID uuid.UUID, winners []uuid.UUID) {
				state := rm.GetTransitionState(ctx, tournamentID)
				require.NotNil(t, state, "transition state should exist")
				assert.Len(t, state.ExpectedWinners, 2, "should have 2 expected winners")
				assert.Equal(t, 1, state.CompletedRound)
				assert.Equal(t, 2, state.NextRound)

				// Simulate winners joining pre-lobby
				rm.OnParticipantJoinedPreLobby(ctx, tournamentID, winners[0])
				state = rm.GetTransitionState(ctx, tournamentID)
				assert.Len(t, state.JoinedWinners, 1, "1 winner should have joined")
				assert.False(t, state.GraceStarted, "grace should not start yet")

				rm.OnParticipantJoinedPreLobby(ctx, tournamentID, winners[1])
				state = rm.GetTransitionState(ctx, tournamentID)
				assert.Len(t, state.JoinedWinners, 2, "2 winners should have joined")
				assert.True(t, state.GraceStarted, "grace should start when all winners join")
			},
		},
		{
			name:       "4-match round with staggered completion",
			numMatches: 4,
			setupAndRun: func(t *testing.T, rm *RoundManager, matchRepo *MockMatchRepository, roundRepo *MockRoundRepository, bracketRepo *MockBracketRepository) (uuid.UUID, []uuid.UUID) {
				tournamentID := uuid.New()
				bracketID := uuid.New()
				roundID := uuid.New()
				nextMatchID1 := uuid.New()
				nextMatchID2 := uuid.New()

				// Create 4 matches with different winners
				winners := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
				matchIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}

				round := &domain.Round{
					ID:          roundID,
					BracketID:   bracketID,
					RoundNumber: 1,
				}

				bracket := &domain.Bracket{
					ID:           bracketID,
					TournamentID: tournamentID,
				}

				// Complete matches one by one
				for i := 0; i < 4; i++ {
					p1, p2 := uuid.New(), uuid.New()
					nextMatch := &nextMatchID1
					if i >= 2 {
						nextMatch = &nextMatchID2
					}

					match := &domain.Match{
						ID:             matchIDs[i],
						RoundID:        roundID,
						MatchNumber:    i + 1,
						Participant1ID: p1,
						Participant2ID: &p2,
						Status:         domain.MatchStatusCompleted,
						NextMatchID:    nextMatch,
					}

					matchRepo.On("GetByID", ctx, matchIDs[i]).Return(match, nil)
					roundRepo.On("GetByID", ctx, roundID).Return(round, nil).Once()
					bracketRepo.On("GetByID", ctx, bracketID).Return(bracket, nil).Once()

					// Build matches state for this completion
					var allMatches []domain.Match
					for j := 0; j <= i; j++ {
						nm := &nextMatchID1
						if j >= 2 {
							nm = &nextMatchID2
						}
						allMatches = append(allMatches, domain.Match{
							ID:          matchIDs[j],
							RoundID:     roundID,
							Status:      domain.MatchStatusCompleted,
							NextMatchID: nm,
						})
					}
					for j := i + 1; j < 4; j++ {
						nm := &nextMatchID1
						if j >= 2 {
							nm = &nextMatchID2
						}
						allMatches = append(allMatches, domain.Match{
							ID:          matchIDs[j],
							RoundID:     roundID,
							Status:      domain.MatchStatusPending,
							NextMatchID: nm,
						})
					}
					matchRepo.On("GetByRoundID", ctx, roundID).Return(allMatches, nil).Once()

					// On last match, expect round status update
					if i == 3 {
						roundRepo.On("UpdateStatus", ctx, roundID, domain.RoundStatusCompleted).Return(nil).Once()
					}

					err := rm.OnMatchCompleted(ctx, matchIDs[i], winners[i])
					require.NoError(t, err)
				}

				return tournamentID, winners
			},
			checkFinalState: func(t *testing.T, rm *RoundManager, tournamentID uuid.UUID, winners []uuid.UUID) {
				state := rm.GetTransitionState(ctx, tournamentID)
				require.NotNil(t, state, "transition state should exist")
				assert.Len(t, state.ExpectedWinners, 4, "should have 4 expected winners")

				// Simulate all 4 winners joining
				for i, winner := range winners {
					rm.OnParticipantJoinedPreLobby(ctx, tournamentID, winner)
					state = rm.GetTransitionState(ctx, tournamentID)
					assert.Len(t, state.JoinedWinners, i+1)
					if i < 3 {
						assert.False(t, state.GraceStarted, "grace should not start yet")
					} else {
						assert.True(t, state.GraceStarted, "grace should start when all 4 winners join")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchRepo := new(MockMatchRepository)
			roundRepo := new(MockRoundRepository)
			bracketRepo := new(MockBracketRepository)

			rm := createTestRoundManager(matchRepo, roundRepo, bracketRepo, nil)

			tournamentID, winners := tt.setupAndRun(t, rm, matchRepo, roundRepo, bracketRepo)
			tt.checkFinalState(t, rm, tournamentID, winners)

			matchRepo.AssertExpectations(t)
			roundRepo.AssertExpectations(t)
			bracketRepo.AssertExpectations(t)
		})
	}
}

// TestRoundManager_ConcurrentWinnerJoins tests concurrent access to RoundManager
func TestRoundManager_ConcurrentWinnerJoins(t *testing.T) {
	ctx := context.Background()
	logger := observability.NewLogger("info", "text")
	transitionStore := NewInMemoryRoundTransitionStore(logger)
	rm := createTestRoundManager(nil, nil, nil, transitionStore)
	tournamentID := uuid.New()

	// Setup state with 10 expected winners
	numWinners := 10
	winners := make([]uuid.UUID, numWinners)
	for i := 0; i < numWinners; i++ {
		winners[i] = uuid.New()
	}
	transitionStore.Save(ctx, &RoundTransitionData{
		TournamentID:    tournamentID,
		CompletedRound:  1,
		NextRound:       2,
		ExpectedWinners: winners,
		JoinedWinners:   []uuid.UUID{},
		GraceStarted:    false,
	})

	// Simulate concurrent joins
	done := make(chan bool, numWinners)
	for _, winner := range winners {
		go func(w uuid.UUID) {
			rm.OnParticipantJoinedPreLobby(ctx, tournamentID, w)
			done <- true
		}(winner)
	}

	// Wait for all joins
	for i := 0; i < numWinners; i++ {
		<-done
	}

	// Verify final state
	finalState := rm.GetTransitionState(ctx, tournamentID)
	require.NotNil(t, finalState)
	assert.Len(t, finalState.JoinedWinners, numWinners, "all winners should have joined")
	assert.True(t, finalState.GraceStarted, "grace should have started")
}
