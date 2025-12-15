// Package application provides application services for matchmaking
package application

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/repository"
)

// RoundTransitionState tracks winners joining pre-lobby after round completes
type RoundTransitionState struct {
	mu              sync.Mutex
	TournamentID    uuid.UUID
	CompletedRound  int                    // Round N that just finished
	NextRound       int                    // Round N+1 to start
	ExpectedWinners map[uuid.UUID]struct{} // Winners who should join pre-lobby
	JoinedWinners   map[uuid.UUID]struct{} // Winners who have connected
	GraceStarted    bool                   // Prevent double-start
}

// RoundManager orchestrates tournament progression between rounds.
// It tracks which round is "pending" (waiting for winners to join pre-lobby)
// and triggers grace period when ALL expected winners have connected.
type RoundManager struct {
	matchRepo       repository.MatchRepository
	roundRepo       repository.RoundRepository
	bracketRepo     repository.BracketRepository
	preLobbyService *PreLobbyService
	logger          *observability.Logger

	// In-memory state: pending round transitions
	// Key: tournamentID, Value: *RoundTransitionState
	pendingTransitions sync.Map
}

// NewRoundManager creates a new RoundManager instance
func NewRoundManager(
	matchRepo repository.MatchRepository,
	roundRepo repository.RoundRepository,
	bracketRepo repository.BracketRepository,
	preLobbyService *PreLobbyService,
	logger *observability.Logger,
) *RoundManager {
	return &RoundManager{
		matchRepo:       matchRepo,
		roundRepo:       roundRepo,
		bracketRepo:     bracketRepo,
		preLobbyService: preLobbyService,
		logger:          logger,
	}
}

// OnMatchCompleted is called by MatchService after each match completion.
// It tracks the winner and checks if the round is complete.
// If complete, it registers expected winners for the next round.
func (rm *RoundManager) OnMatchCompleted(ctx context.Context, matchID, winnerID uuid.UUID) error {
	rm.logger.Info("[RoundManager] OnMatchCompleted called", map[string]interface{}{
		"match_id":  matchID.String(),
		"winner_id": winnerID.String(),
	})

	match, err := rm.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		rm.logger.Error("[RoundManager] Failed to get match", map[string]interface{}{
			"match_id": matchID.String(),
			"error":    err.Error(),
		})
		return fmt.Errorf("get match: %w", err)
	}

	// Skip final round - tournament completion handled elsewhere
	if match.NextMatchID == nil {
		rm.logger.Info("Final match completed, tournament ends", nil)
		return nil
	}

	round, err := rm.roundRepo.GetByID(ctx, match.RoundID)
	if err != nil {
		return fmt.Errorf("get round: %w", err)
	}

	bracket, err := rm.bracketRepo.GetByID(ctx, round.BracketID)
	if err != nil {
		rm.logger.Error("[RoundManager] Failed to get bracket", map[string]interface{}{
			"bracket_id": round.BracketID.String(),
			"error":      err.Error(),
		})
		return fmt.Errorf("get bracket: %w", err)
	}

	rm.logger.Info("[RoundManager] Match context retrieved", map[string]interface{}{
		"tournament_id": bracket.TournamentID.String(),
		"round_id":      round.ID.String(),
		"round_number":  round.RoundNumber,
		"match_id":      matchID.String(),
		"winner_id":     winnerID.String(),
	})

	// Get or create transition state for this tournament
	state := rm.getOrCreateTransitionState(bracket.TournamentID, round.RoundNumber)

	state.mu.Lock()
	defer state.mu.Unlock()

	// Add this winner to expected list
	state.ExpectedWinners[winnerID] = struct{}{}

	rm.logger.Info("[RoundManager] Winner added to expected list", map[string]interface{}{
		"tournament_id":    bracket.TournamentID.String(),
		"winner_id":        winnerID.String(),
		"expected_winners": len(state.ExpectedWinners),
	})

	// Check if ALL matches in round are done
	allMatches, err := rm.matchRepo.GetByRoundID(ctx, round.ID)
	if err != nil {
		return fmt.Errorf("get matches by round: %w", err)
	}

	completedCount := 0
	for _, m := range allMatches {
		if m.Status == domain.MatchStatusCompleted {
			completedCount++
		}
	}

	// Reset pre-lobby to waiting so winners can join for next round
	if err := rm.preLobbyService.PrepareForNextRound(ctx, bracket.TournamentID); err != nil {
		rm.logger.Error("Failed to prepare pre-lobby for next round", map[string]interface{}{
			"tournament_id": bracket.TournamentID.String(),
			"error":         err.Error(),
		})
	}

	if completedCount < len(allMatches) {
		rm.logger.Info("Round not complete yet", map[string]interface{}{
			"completed": completedCount,
			"total":     len(allMatches),
		})
		return nil
	}

	rm.logger.Info("Round complete! Waiting for winners to join pre-lobby", map[string]interface{}{
		"tournament_id":    bracket.TournamentID.String(),
		"completed_round":  round.RoundNumber,
		"expected_winners": len(state.ExpectedWinners),
	})

	return nil
}

// OnParticipantJoinedPreLobby is called by PreLobbyService when participant connects.
// It checks if this was an expected winner and if all winners are now present.
// When all winners have joined, it triggers the grace period.
func (rm *RoundManager) OnParticipantJoinedPreLobby(ctx context.Context, tournamentID, participantID uuid.UUID) {
	rm.logger.Info("[RoundManager] OnParticipantJoinedPreLobby called", map[string]interface{}{
		"tournament_id":  tournamentID.String(),
		"participant_id": participantID.String(),
	})

	val, exists := rm.pendingTransitions.Load(tournamentID)
	if !exists {
		// No pending transition for this tournament
		rm.logger.Warn("[RoundManager] No pending transition for tournament - participant joined but round not tracked", map[string]interface{}{
			"tournament_id":  tournamentID.String(),
			"participant_id": participantID.String(),
		})
		return
	}

	state := val.(*RoundTransitionState)
	state.mu.Lock()
	defer state.mu.Unlock()

	rm.logger.Info("[RoundManager] Found pending transition state", map[string]interface{}{
		"tournament_id":    tournamentID.String(),
		"completed_round":  state.CompletedRound,
		"next_round":       state.NextRound,
		"expected_winners": len(state.ExpectedWinners),
		"joined_winners":   len(state.JoinedWinners),
		"grace_started":    state.GraceStarted,
	})

	// Check if this is an expected winner
	if _, isExpected := state.ExpectedWinners[participantID]; !isExpected {
		rm.logger.Warn("[RoundManager] Participant is NOT an expected winner", map[string]interface{}{
			"tournament_id":  tournamentID.String(),
			"participant_id": participantID.String(),
		})
		return // Not a winner we're waiting for
	}

	// Mark winner as joined
	state.JoinedWinners[participantID] = struct{}{}

	rm.logger.Info("Winner joined pre-lobby", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"participant":   participantID.String(),
		"joined":        len(state.JoinedWinners),
		"expected":      len(state.ExpectedWinners),
	})

	// Check if ALL winners have joined
	// TAG: REFACTOR, TODO: It should be observated by scheduler or handled in different way. What if player has never join?
	if len(state.JoinedWinners) >= len(state.ExpectedWinners) {
		rm.startGracePeriodForNextRound(ctx, state)
	}
}

// GetTransitionState returns the current transition state for a tournament (for testing)
func (rm *RoundManager) GetTransitionState(tournamentID uuid.UUID) *RoundTransitionState {
	val, exists := rm.pendingTransitions.Load(tournamentID)
	if !exists {
		return nil
	}
	return val.(*RoundTransitionState)
}

// getOrCreateTransitionState retrieves or creates a transition state for a tournament
func (rm *RoundManager) getOrCreateTransitionState(tournamentID uuid.UUID, completedRound int) *RoundTransitionState {
	val, _ := rm.pendingTransitions.LoadOrStore(tournamentID, &RoundTransitionState{
		TournamentID:    tournamentID,
		CompletedRound:  completedRound,
		NextRound:       completedRound + 1,
		ExpectedWinners: make(map[uuid.UUID]struct{}),
		JoinedWinners:   make(map[uuid.UUID]struct{}),
		GraceStarted:    false,
	})

	return val.(*RoundTransitionState)
}

// startGracePeriodForNextRound starts the grace period when all winners are present
func (rm *RoundManager) startGracePeriodForNextRound(ctx context.Context, state *RoundTransitionState) {
	rm.logger.Info("[RoundManager] startGracePeriodForNextRound called", map[string]interface{}{
		"tournament_id": state.TournamentID.String(),
		"next_round":    state.NextRound,
		"grace_started": state.GraceStarted,
	})

	if state.GraceStarted {
		rm.logger.Warn("[RoundManager] Grace period already started, skipping", map[string]interface{}{
			"tournament_id": state.TournamentID.String(),
		})
		return // Prevent double-start
	}
	state.GraceStarted = true

	rm.logger.Info("[RoundManager] All winners joined! Triggering grace period", map[string]interface{}{
		"tournament_id":    state.TournamentID.String(),
		"next_round":       state.NextRound,
		"expected_winners": len(state.ExpectedWinners),
		"joined_winners":   len(state.JoinedWinners),
	})

	// Callback when grace period ends
	onComplete := func(tID uuid.UUID, presentParticipantIDs []uuid.UUID) {
		rm.logger.Info("[RoundManager] Grace period completed, assigning matches", map[string]interface{}{
			"tournament_id":   tID.String(),
			"present_players": len(presentParticipantIDs),
		})
		rm.assignMatchesForRound(context.Background(), tID, state.NextRound, presentParticipantIDs)
		rm.pendingTransitions.Delete(tID) // Cleanup
	}

	// REUSE existing PreLobbyService.StartGracePeriod if available
	if rm.preLobbyService != nil {
		rm.logger.Info("[RoundManager] Calling PreLobbyService.StartGracePeriod", map[string]interface{}{
			"tournament_id": state.TournamentID.String(),
		})
		err := rm.preLobbyService.StartGracePeriod(ctx, state.TournamentID, onComplete)
		if err != nil {
			rm.logger.Error("[RoundManager] StartGracePeriod failed", map[string]interface{}{
				"tournament_id": state.TournamentID.String(),
				"error":         err.Error(),
			})
		} else {
			rm.logger.Info("[RoundManager] StartGracePeriod called successfully", map[string]interface{}{
				"tournament_id": state.TournamentID.String(),
			})
		}
	} else {
		rm.logger.Warn("[RoundManager] preLobbyService is nil, cannot start grace period", nil)
	}
}

// assignMatchesForRound assigns matches after grace period ends
func (rm *RoundManager) assignMatchesForRound(ctx context.Context, tournamentID uuid.UUID, roundNum int, presentIDs []uuid.UUID) {
	// TODO: Implement
}
