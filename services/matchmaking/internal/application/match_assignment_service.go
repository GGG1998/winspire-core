// Package application contains application services and business logic orchestration
package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
)

// MatchAssignmentService handles broadcasting match assignments to participants
type MatchAssignmentService struct {
	bracketService  *BracketService
	preLobbyService *PreLobbyService
	logger          *observability.Logger
}

// NewMatchAssignmentService creates a new match assignment service
func NewMatchAssignmentService(
	bracketService *BracketService,
	preLobbyService *PreLobbyService,
	logger *observability.Logger,
) *MatchAssignmentService {
	return &MatchAssignmentService{
		bracketService:  bracketService,
		preLobbyService: preLobbyService,
		logger:          logger,
	}
}

// BroadcastMatchAssignmentsForRound sends match_assigned to all participants in a round
func (s *MatchAssignmentService) BroadcastMatchAssignmentsForRound(
	ctx context.Context,
	tournamentID uuid.UUID,
	roundNumber int,
) error {
	matches, err := s.bracketService.GetMatchesByRound(ctx, tournamentID, roundNumber)
	if err != nil {
		s.logger.Error("Failed to get matches for round", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"round_number":  roundNumber,
			"error":         err.Error(),
		})
		return fmt.Errorf("get matches for round %d: %w", roundNumber, err)
	}

	for i := range matches {
		s.BroadcastMatchAssignment(ctx, tournamentID, &matches[i], roundNumber)
	}

	s.logger.Info("Broadcast match assignments for round", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"round_number":  roundNumber,
		"match_count":   len(matches),
	})

	return nil
}

// BroadcastMatchAssignment sends match_assigned to both participants of a single match
func (s *MatchAssignmentService) BroadcastMatchAssignment(
	ctx context.Context,
	tournamentID uuid.UUID,
	match *domain.Match,
	roundNumber int,
) {
	// Send to participant 1
	if match.Participant1ID != uuid.Nil {
		var opponent *PreLobbyParticipantInfo
		if match.Participant2ID != nil && *match.Participant2ID != uuid.Nil {
			opponentDetails := s.preLobbyService.GetParticipantDetails(ctx, tournamentID, *match.Participant2ID)
			opponent = &opponentDetails
		}

		s.preLobbyService.BroadcastMatchAssigned(
			tournamentID,
			match.Participant1ID,
			match.ID,
			roundNumber,
			match.MatchNumber,
			opponent,
		)
	}

	// Send to participant 2 (if not BYE)
	if match.Participant2ID != nil && *match.Participant2ID != uuid.Nil {
		opponentDetails := s.preLobbyService.GetParticipantDetails(ctx, tournamentID, match.Participant1ID)
		opponent := &opponentDetails

		s.preLobbyService.BroadcastMatchAssigned(
			tournamentID,
			*match.Participant2ID,
			match.ID,
			roundNumber,
			match.MatchNumber,
			opponent,
		)
	}
}
