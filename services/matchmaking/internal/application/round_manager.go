// Package application provides application services for matchmaking
package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/repository"
)

// RoundManager orchestrates tournament progression between rounds.
// It tracks which round is "pending" (waiting for winners to join pre-lobby)
// and triggers grace period when ALL expected winners have connected.
// State is persisted in Redis and can be recovered from database.
type RoundManager struct {
	matchRepo              repository.MatchRepository
	roundRepo              repository.RoundRepository
	bracketRepo            repository.BracketRepository
	preLobbyService        *PreLobbyService
	matchAssignmentService *MatchAssignmentService
	transitionStore        *RoundTransitionStore
	logger                 *observability.Logger
}

// NewRoundManager creates a new RoundManager instance
func NewRoundManager(
	matchRepo repository.MatchRepository,
	roundRepo repository.RoundRepository,
	bracketRepo repository.BracketRepository,
	preLobbyService *PreLobbyService,
	matchAssignmentService *MatchAssignmentService,
	transitionStore *RoundTransitionStore,
	logger *observability.Logger,
) *RoundManager {
	return &RoundManager{
		matchRepo:              matchRepo,
		roundRepo:              roundRepo,
		bracketRepo:            bracketRepo,
		preLobbyService:        preLobbyService,
		matchAssignmentService: matchAssignmentService,
		transitionStore:        transitionStore,
		logger:                 logger,
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

	// Add winner to expected winners in Redis
	if err := rm.transitionStore.AddExpectedWinner(ctx, bracket.TournamentID, round.RoundNumber, winnerID); err != nil {
		rm.logger.Error("[RoundManager] Failed to add expected winner", map[string]interface{}{
			"tournament_id": bracket.TournamentID.String(),
			"winner_id":     winnerID.String(),
			"error":         err.Error(),
		})
		return fmt.Errorf("add expected winner: %w", err)
	}

	rm.logger.Info("[RoundManager] Winner added to expected list (Redis)", map[string]interface{}{
		"tournament_id": bracket.TournamentID.String(),
		"winner_id":     winnerID.String(),
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
	// Expected participants = number of matches (each match produces 1 winner)
	expectedWinners := len(allMatches)
	if err := rm.preLobbyService.PrepareForNextRound(ctx, bracket.TournamentID, expectedWinners); err != nil {
		rm.logger.Error("Failed to prepare pre-lobby for next round", map[string]interface{}{
			"tournament_id":      bracket.TournamentID.String(),
			"expected_winners":   expectedWinners,
			"error":              err.Error(),
		})
	}

	if completedCount < len(allMatches) {
		rm.logger.Info("Round not complete yet", map[string]interface{}{
			"completed": completedCount,
			"total":     len(allMatches),
		})
		return nil
	}

	// Get current state to log expected winners count
	state, _ := rm.transitionStore.Get(ctx, bracket.TournamentID)
	expectedCount := 0
	if state != nil {
		expectedCount = len(state.ExpectedWinners)
	}

	// TODO: Update current round status if all player finish match

	rm.logger.Info("Round complete! Waiting for winners to join pre-lobby", map[string]interface{}{
		"tournament_id":    bracket.TournamentID.String(),
		"completed_round":  round.RoundNumber,
		"expected_winners": expectedCount,
	})

	return nil
}

// OnParticipantJoinedPreLobby is called by PreLobbyService when participant connects.
// It checks if this was an expected winner and if all winners are now present.
// When all winners have joined, it triggers the grace period.
func (rm *RoundManager) OnParticipantJoinedPreLobby(ctx context.Context, tournamentID, participantID uuid.UUID, execute func()) {
	rm.logger.Info("[RoundManager] OnParticipantJoinedPreLobby called", map[string]interface{}{
		"tournament_id":  tournamentID.String(),
		"participant_id": participantID.String(),
	})

	// Try to add as joined winner (returns true if all winners joined)
	allJoined, err := rm.transitionStore.AddJoinedWinner(ctx, tournamentID, participantID)
	if err != nil {
		fmt.Println("[RoundManager] Failed to add joined winner", map[string]interface{}{
			"tournament_id":  tournamentID.String(),
			"participant_id": participantID.String(),
			"error":          err.Error(),
		})
		return
	}

	// Get current state for logging
	state, _ := rm.transitionStore.Get(ctx, tournamentID)
	if state == nil {
		fmt.Println("[RoundManager] No pending transition for tournament", map[string]interface{}{
			"tournament_id":  tournamentID.String(),
			"participant_id": participantID.String(),
		})
		return
	}

	rm.logger.Info("Winner joined pre-lobby", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"participant":   participantID.String(),
		"joined":        len(state.JoinedWinners),
		"expected":      len(state.ExpectedWinners),
	})

	fmt.Println("[RoundManager] OnParticipantJoinedPreLobby ", allJoined)
	if allJoined {
		execute()
	}
}

// GetTransitionState returns the current transition state for a tournament (for testing)
func (rm *RoundManager) GetTransitionState(ctx context.Context, tournamentID uuid.UUID) *RoundTransitionData {
	state, _ := rm.transitionStore.Get(ctx, tournamentID)
	return state
}

// RecoverTransitions recovers pending round transitions from Redis/database on startup
func (rm *RoundManager) RecoverTransitions(ctx context.Context, tournamentIDs []uuid.UUID) error {
	rm.logger.Info("[RoundManager] Recovering round transitions", map[string]interface{}{
		"tournament_count": len(tournamentIDs),
	})

	for _, tournamentID := range tournamentIDs {
		state, err := rm.transitionStore.GetOrProject(ctx, tournamentID)
		if err != nil {
			rm.logger.Warn("[RoundManager] Failed to recover transition for tournament", map[string]interface{}{
				"tournament_id": tournamentID.String(),
				"error":         err.Error(),
			})
			continue
		}

		if state != nil {
			rm.logger.Info("[RoundManager] Recovered transition state", map[string]interface{}{
				"tournament_id":    tournamentID.String(),
				"completed_round":  state.CompletedRound,
				"expected_winners": len(state.ExpectedWinners),
				"joined_winners":   len(state.JoinedWinners),
				"grace_started":    state.GraceStarted,
			})
		}
	}

	return nil
}
