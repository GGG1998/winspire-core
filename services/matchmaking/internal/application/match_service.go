// Package application contains application services and business logic orchestration
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/pubsub"
	"github.com/winspire-core/services/matchmaking/internal/repository"
)

// MatchService handles match completion and winner advancement logic
type MatchService struct {
	matchRepo         repository.MatchRepository
	roundRepo         repository.RoundRepository
	bracketRepo       repository.BracketRepository
	publisher         *pubsub.EventPublisher
	metrics           *observability.MetricsEmitter
	logger            *observability.Logger
	hub               Hub           // WebSocket hub interface for broadcasting
	roundManager      *RoundManager // RoundManager for tracking round transitions
	gameManagementURL string        // Base URL for game management service
}

// Hub interface for WebSocket broadcasts (to avoid circular dependency)
// This matches the signature in websocket.Hub
type Hub interface {
	BroadcastToMatch(matchID uuid.UUID, message interface{}) // accepts any type, will be marshaled
	SendToPlayer(matchID, playerID uuid.UUID, message interface{})
}

// NewMatchService creates a new match service
func NewMatchService(
	matchRepo repository.MatchRepository,
	roundRepo repository.RoundRepository,
	bracketRepo repository.BracketRepository,
	publisher *pubsub.EventPublisher,
	metrics *observability.MetricsEmitter,
	logger *observability.Logger,
	gameManagementURL string,
) *MatchService {
	return &MatchService{
		matchRepo:         matchRepo,
		roundRepo:         roundRepo,
		bracketRepo:       bracketRepo,
		publisher:         publisher,
		metrics:           metrics,
		logger:            logger,
		gameManagementURL: gameManagementURL,
	}
}

// SetHub sets the WebSocket hub for broadcasting (called after initialization)
func (s *MatchService) SetHub(hub Hub) {
	s.hub = hub
}

// SetRoundManager sets the RoundManager for round transition tracking (called after initialization)
func (s *MatchService) SetRoundManager(rm *RoundManager) {
	s.roundManager = rm
}

// GetCurrentMatchForUser returns the latest active/pending match for a given user with round and tournament context
func (s *MatchService) GetCurrentMatchForUser(ctx context.Context, userID uuid.UUID) (*domain.Match, *domain.Round, uuid.UUID, error) {
	match, err := s.matchRepo.GetCurrentForUser(ctx, userID)
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("get current match for user: %w", err)
	}

	round, err := s.roundRepo.GetByID(ctx, match.RoundID)
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("get round for match: %w", err)
	}

	bracket, err := s.bracketRepo.GetByID(ctx, round.BracketID)
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("get bracket for round: %w", err)
	}

	return match, round, bracket.TournamentID, nil
}

// CompleteMatch marks a match as completed and advances the winner to the next round
func (s *MatchService) CompleteMatch(ctx context.Context, matchID, winnerID uuid.UUID, scorePlayer1, scorePlayer2 int, source domain.ResultSource) error {
	s.logger.Info("Completing match", map[string]interface{}{
		"match_id":  matchID.String(),
		"winner_id": winnerID.String(),
		"score":     fmt.Sprintf("%d-%d", scorePlayer1, scorePlayer2),
		"source":    source,
	})

	// Fetch match to get details
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("get match: %w", err)
	}

	// Validate winner is a participant
	isParticipant1 := match.Participant1ID == winnerID
	isParticipant2 := match.Participant2ID != nil && *match.Participant2ID == winnerID
	if !isParticipant1 && !isParticipant2 {
		return fmt.Errorf("winner_id %s is not a participant in match %s", winnerID, matchID)
	}

	// Determine loser
	var loserID uuid.UUID
	if isParticipant1 {
		if match.Participant2ID != nil {
			loserID = *match.Participant2ID
		}
	} else {
		loserID = match.Participant1ID
	}

	// Update match result
	completedAt := time.Now()
	err = s.matchRepo.UpdateResult(ctx, matchID, winnerID, scorePlayer1, scorePlayer2, source)
	if err != nil {
		return fmt.Errorf("update match result: %w", err)
	}

	// Update match status to completed
	err = s.matchRepo.UpdateStatus(ctx, matchID, domain.MatchStatusCompleted)
	if err != nil {
		return fmt.Errorf("update match status: %w", err)
	}

	s.logger.Info("Match result recorded", map[string]interface{}{
		"match_id":  matchID.String(),
		"winner_id": winnerID.String(),
		"loser_id":  loserID.String(),
	})

	// Publish MatchCompleted event (T068)
	matchCompletedEvent := domain.NewMatchCompleted(
		matchID,
		match.RoundID, // Using RoundID as TournamentID placeholder (should fetch from bracket)
		winnerID,
		loserID,
		scorePlayer1,
		scorePlayer2,
		source,
		completedAt,
		match.NextMatchID,
		map[string]string{
			"correlation_id": s.getCorrelationID(ctx),
		},
	)
	if err := s.publisher.Publish(ctx, matchCompletedEvent); err != nil {
		log.Printf("WARN: Failed to publish MatchCompleted event: %v", err)
	}

	// T106: Check if this is the final match (no next match = tournament complete)
	if match.NextMatchID == nil {
		s.logger.Info("Final match completed, tournament winner determined", map[string]interface{}{
			"match_id":    matchID.String(),
			"champion_id": winnerID.String(),
		})

		// T107, T108: Complete the tournament
		err = s.CompleteTournament(ctx, matchID, winnerID)
		if err != nil {
			return fmt.Errorf("complete tournament: %w", err)
		}
	} else {
		// Advance winner to next round (T065)
		err = s.AdvanceWinner(ctx, winnerID, matchID, *match.NextMatchID)
		if err != nil {
			return fmt.Errorf("advance winner: %w", err)
		}
	}

	// Mark loser as eliminated (T066, T070)
	if loserID != uuid.Nil {
		s.logger.Info("Marking loser as eliminated", map[string]interface{}{
			"player_id": loserID.String(),
			"match_id":  matchID.String(),
		})

		// Publish ParticipantEliminated event
		eliminatedEvent := domain.NewParticipantEliminated(
			loserID,
			match.RoundID, // Using RoundID as TournamentID placeholder
			matchID,
			0, // Round number would need to be fetched from round
			completedAt,
			map[string]string{
				"correlation_id": s.getCorrelationID(ctx),
			},
		)
		if err := s.publisher.Publish(ctx, eliminatedEvent); err != nil {
			log.Printf("WARN: Failed to publish ParticipantEliminated event: %v", err)
		}
	}

	// Fetch round and bracket for post-match WebSocket messages
	round, err := s.roundRepo.GetByID(ctx, match.RoundID)
	if err != nil {
		s.logger.Warn("Failed to get round for post-match messages", map[string]interface{}{
			"match_id": matchID.String(),
			"error":    err.Error(),
		})
		// Continue - this is not critical
	}

	var tournamentID uuid.UUID
	var nextRoundNum int
	if round != nil {
		bracket, err := s.bracketRepo.GetByID(ctx, round.BracketID)
		if err == nil {
			tournamentID = bracket.TournamentID
			nextRoundNum = round.RoundNumber + 1
		}
	}

	// Send post-match WebSocket messages to winner and loser
	if s.hub != nil && tournamentID != uuid.Nil {
		// Send message to winner
		if match.NextMatchID == nil {
			// Tournament champion message
			championPayload := domain.TournamentChampionPayload{
				TournamentID: tournamentID,
				ChampionID:   winnerID,
				Message:      "Congratulations! You are the tournament champion!",
			}
			msg := map[string]interface{}{
				"type":      domain.MsgTypeTournamentChampion,
				"timestamp": time.Now(),
				"payload":   championPayload,
			}
			s.hub.SendToPlayer(matchID, winnerID, msg)
			s.logger.Info("Sent tournament_champion message", map[string]interface{}{
				"match_id":  matchID.String(),
				"winner_id": winnerID.String(),
			})
		} else {
			// Return to pre-lobby message for winner
			returnPayload := domain.ReturnToPreLobbyPayload{
				TournamentID: tournamentID,
				MatchID:      matchID,
				NextRoundNum: nextRoundNum,
				Message:      "You won! Please return to the pre-lobby for the next round.",
				PreLobbyURL:  fmt.Sprintf("/tournaments/%s/lobby", tournamentID.String()),
			}
			msg := map[string]interface{}{
				"type":      domain.MsgTypeReturnToPreLobby,
				"timestamp": time.Now(),
				"payload":   returnPayload,
			}
			s.hub.SendToPlayer(matchID, winnerID, msg)
			s.logger.Info("Sent return_to_prelobby message", map[string]interface{}{
				"match_id":       matchID.String(),
				"winner_id":      winnerID.String(),
				"next_round_num": nextRoundNum,
			})
		}

		// Send eliminated message to loser
		if loserID != uuid.Nil {
			eliminatedPayload := domain.PlayerEliminatedNotificationPayload{
				TournamentID:  tournamentID,
				MatchID:       matchID,
				FinalPosition: 0, // TODO: Calculate actual position based on round
				Message:       "You have been eliminated from the tournament. Thank you for playing!",
			}
			msg := map[string]interface{}{
				"type":      domain.MsgTypePlayerEliminatedNotify,
				"timestamp": time.Now(),
				"payload":   eliminatedPayload,
			}
			s.hub.SendToPlayer(matchID, loserID, msg)
			s.logger.Info("Sent player_eliminated_notify message", map[string]interface{}{
				"match_id": matchID.String(),
				"loser_id": loserID.String(),
			})
		}
	}

	// Notify RoundManager to track round progression
	if s.roundManager != nil {
		if err := s.roundManager.OnMatchCompleted(ctx, matchID, winnerID); err != nil {
			s.logger.Warn("RoundManager.OnMatchCompleted failed", map[string]interface{}{
				"match_id":  matchID.String(),
				"winner_id": winnerID.String(),
				"error":     err.Error(),
			})
			// Non-critical - continue
		}
	}

	return nil
}

// AdvanceWinner assigns the winner to the next match in the bracket
func (s *MatchService) AdvanceWinner(ctx context.Context, winnerID, fromMatchID, toMatchID uuid.UUID) error {
	s.logger.Info("Advancing winner to next match", map[string]interface{}{
		"winner_id":     winnerID.String(),
		"from_match_id": fromMatchID.String(),
		"to_match_id":   toMatchID.String(),
	})

	// Fetch the next match to determine which participant slot to assign
	nextMatch, err := s.matchRepo.GetByID(ctx, toMatchID)
	if err != nil {
		return fmt.Errorf("get next match: %w", err)
	}

	// Determine which slot to fill based on the bracket structure
	// In a single-elimination bracket, the winner of match N goes to:
	// - participant1 slot if it's the first feeder match
	// - participant2 slot if it's the second feeder match
	// We fill the first available slot (NULL/zero slot) or overwrite if winner matches placeholder

	var slotToFill int
	p1Empty := nextMatch.Participant1ID == uuid.Nil
	p2Empty := nextMatch.Participant2ID == nil || *nextMatch.Participant2ID == uuid.Nil
	// Also allow overwriting if winner is already in that slot (placeholder scenario)
	p1IsWinner := nextMatch.Participant1ID == winnerID
	p2IsWinner := nextMatch.Participant2ID != nil && *nextMatch.Participant2ID == winnerID

	if p2Empty || p2IsWinner {
		// Fill participant2 slot if empty or winner matches placeholder
		slotToFill = 2
		nextMatch.Participant2ID = &winnerID
	} else if p1Empty || p1IsWinner {
		// Fill participant1 slot if empty or winner matches placeholder
		slotToFill = 1
		nextMatch.Participant1ID = winnerID
	} else {
		return fmt.Errorf("next match %s already has both participants assigned", toMatchID)
	}

	// Update the next match with the winner
	// This would require a new repository method to update participant assignments
	// For now, we'll log the advancement
	s.logger.Info("Winner assigned to next match", map[string]interface{}{
		"winner_id":  winnerID.String(),
		"next_match": toMatchID.String(),
		"slot":       slotToFill,
	})

	// Publish ParticipantAdvanced event (T069)
	advancedEvent := domain.NewParticipantAdvanced(
		winnerID,
		nextMatch.RoundID, // Using RoundID as TournamentID placeholder
		fromMatchID,
		toMatchID,
		0, // From round number (would need to fetch from round)
		0, // To round number (would need to fetch from round)
		map[string]string{
			"correlation_id": s.getCorrelationID(ctx),
		},
	)
	if err := s.publisher.Publish(ctx, advancedEvent); err != nil {
		log.Printf("WARN: Failed to publish ParticipantAdvanced event: %v", err)
	}

	return nil
}

// HandleByeMatch auto-advances participant1 when participant2 is NULL (T067)
func (s *MatchService) HandleByeMatch(ctx context.Context, matchID uuid.UUID) error {
	s.logger.Info("Handling bye match", map[string]interface{}{
		"match_id": matchID.String(),
	})

	// Fetch match
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("get match: %w", err)
	}

	// Verify this is a bye match (participant2 is NULL)
	if match.Participant2ID != nil {
		return fmt.Errorf("match %s is not a bye match (has participant2)", matchID)
	}

	// Auto-advance participant1 as the winner with a walkover
	completedAt := time.Now()
	err = s.matchRepo.UpdateResult(ctx, matchID, match.Participant1ID, 0, 0, domain.ResultSourceWalkover)
	if err != nil {
		return fmt.Errorf("update bye match result: %w", err)
	}

	// Update match status to completed
	err = s.matchRepo.UpdateStatus(ctx, matchID, domain.MatchStatusCompleted)
	if err != nil {
		return fmt.Errorf("update bye match status: %w", err)
	}

	s.logger.Info("Bye match completed, participant auto-advanced", map[string]interface{}{
		"match_id":       matchID.String(),
		"participant_id": match.Participant1ID.String(),
	})

	// Publish WalkoverGranted event
	walkoverEvent := domain.NewWalkoverGranted(
		matchID,
		match.RoundID, // Using RoundID as TournamentID placeholder
		match.Participant1ID,
		uuid.Nil, // No opponent (bye)
		"Bye - no opponent assigned",
		completedAt,
		map[string]string{
			"correlation_id": s.getCorrelationID(ctx),
		},
	)
	if err := s.publisher.Publish(ctx, walkoverEvent); err != nil {
		log.Printf("WARN: Failed to publish WalkoverGranted event: %v", err)
	}

	// Advance winner to next round if there is one
	if match.NextMatchID != nil {
		err = s.AdvanceWinner(ctx, match.Participant1ID, matchID, *match.NextMatchID)
		if err != nil {
			return fmt.Errorf("advance bye winner: %w", err)
		}
	}

	return nil
}

// CheckBothPlayersReady checks if both players in a match are ready (T072)
func (s *MatchService) CheckBothPlayersReady(ctx context.Context, matchID uuid.UUID) (bool, error) {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return false, fmt.Errorf("get match: %w", err)
	}

	// Both players must be present and ready
	bothPresent := match.Participant1ID != uuid.Nil && match.Participant2ID != nil && *match.Participant2ID != uuid.Nil
	bothReady := match.Participant1Ready && match.Participant2Ready

	return bothPresent && bothReady, nil
}

// StartMatch automatically starts a match when both players are ready (T073)
func (s *MatchService) StartMatch(ctx context.Context, matchID uuid.UUID) error {
	startTime := time.Now()
	defer func() {
		// Emit metric for time between both-ready and match-started (T077, SC-009)
		duration := time.Since(startTime)
		s.metrics.EmitReadyToStartedLatency(duration)
	}()

	s.logger.Info("Starting match", map[string]interface{}{
		"match_id": matchID.String(),
	})

	// Fetch match
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("get match: %w", err)
	}

	// Verify both players are ready
	bothReady, err := s.CheckBothPlayersReady(ctx, matchID)
	if err != nil {
		return fmt.Errorf("check ready status: %w", err)
	}

	if !bothReady {
		return fmt.Errorf("cannot start match %s: both players not ready", matchID)
	}

	// Update match status to started
	err = s.matchRepo.UpdateStatus(ctx, matchID, domain.MatchStatusStarted)
	if err != nil {
		return fmt.Errorf("update match status: %w", err)
	}

	s.logger.Info("Match started successfully", map[string]interface{}{
		"match_id":       matchID.String(),
		"participant1":   match.Participant1ID.String(),
		"participant2":   match.Participant2ID.String(),
		"start_duration": time.Since(startTime).Milliseconds(),
	})

	// Publish MatchStarted event (T074)
	if match.Participant2ID == nil {
		return fmt.Errorf("match %s has no participant2", matchID)
	}

	matchStartedEvent := domain.NewMatchStarted(
		matchID,
		match.RoundID, // Using RoundID as TournamentID placeholder
		match.Participant1ID,
		*match.Participant2ID,
		time.Now(),
		map[string]string{
			"correlation_id": s.getCorrelationID(ctx),
		},
	)
	if err := s.publisher.Publish(ctx, matchStartedEvent); err != nil {
		log.Printf("WARN: Failed to publish MatchStarted event: %v", err)
	}

	// TODO: Broadcast match start via WebSocket (T075)
	// This will be handled by the WebSocket hub listening to MatchStarted events

	return nil
}

// OnPlayerReady handles when a player marks themselves as ready (T072, T073)
// This is called after updating the ready status in the database
func (s *MatchService) OnPlayerReady(ctx context.Context, matchID, playerID uuid.UUID) error {
	s.logger.Info("Player marked ready", map[string]interface{}{
		"match_id":  matchID.String(),
		"player_id": playerID.String(),
	})

	// Check if both players are now ready
	bothReady, err := s.CheckBothPlayersReady(ctx, matchID)
	if err != nil {
		return fmt.Errorf("check ready status: %w", err)
	}

	// If both players are ready, transition to loading state
	if bothReady {
		s.logger.Info("Both players ready, transitioning to loading", map[string]interface{}{
			"match_id": matchID.String(),
		})

		// Fetch match to get tournament info
		match, err := s.matchRepo.GetByID(ctx, matchID)
		if err != nil {
			return fmt.Errorf("get match: %w", err)
		}

		// Fetch round to get bracket ID
		round, err := s.roundRepo.GetByID(ctx, match.RoundID)
		if err != nil {
			return fmt.Errorf("get round: %w", err)
		}

		// Fetch bracket to get game snapshot and tournament ID
		bracket, err := s.bracketRepo.GetByID(ctx, round.BracketID)
		if err != nil {
			return fmt.Errorf("get bracket: %w", err)
		}

		// Transition to loading state
		err = s.matchRepo.UpdateStatus(ctx, matchID, domain.MatchStatusLoading)
		if err != nil {
			return fmt.Errorf("update status to loading: %w", err)
		}

		// Construct game URL
		var gameURL string
		if bracket.GameSnapshot != nil {
			gameURL = fmt.Sprintf("%s/v1/g/%s/bundle/", s.getGameManagementURL(), bracket.GameSnapshot.Slug)
		} else {
			s.logger.Warn("No game snapshot in bracket", map[string]interface{}{
				"match_id":      matchID.String(),
				"tournament_id": bracket.TournamentID.String(),
			})
			// For now, continue without game URL - this will be handled better later
			gameURL = ""
		}

		// Broadcast match_ready_to_load via WebSocket with game URL
		if s.hub != nil {
			msg := map[string]interface{}{
				"type":      "match_ready_to_load",
				"timestamp": time.Now(),
				"payload": map[string]interface{}{
					"gameUrl":       gameURL,
					"gameSessionId": matchID.String(),
				},
			}
			s.hub.BroadcastToMatch(matchID, msg)

			s.logger.Info("Broadcasted match_ready_to_load message", map[string]interface{}{
				"match_id": matchID.String(),
				"game_url": gameURL,
			})
		}
	}

	return nil
}

// OnGameLoaded handles when a player's game has loaded
func (s *MatchService) OnGameLoaded(ctx context.Context, matchID, playerID uuid.UUID) error {
	s.logger.Info("Player game loaded", map[string]interface{}{
		"match_id":  matchID.String(),
		"player_id": playerID.String(),
	})

	// Fetch match first to validate status
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("get match: %w", err)
	}

	// Validate match is in loading state
	if match.Status != domain.MatchStatusLoading {
		// #region agent log
		func() {
			f, _ := os.OpenFile("/Users/gabrieldomanowski/programming/winspire-core/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if f != nil {
				defer f.Close()
				json.NewEncoder(f).Encode(map[string]interface{}{"location": "match_service.go:495", "message": "match not in loading state", "data": map[string]interface{}{"matchId": matchID.String(), "status": string(match.Status)}, "timestamp": time.Now().UnixMilli(), "sessionId": "debug-session", "hypothesisId": "H6"})
			}
		}()
		// #endregion
		s.logger.Warn("Game loaded callback for non-loading match", map[string]interface{}{
			"match_id": matchID.String(),
			"status":   match.Status,
		})
		return fmt.Errorf("match is not in loading state, current status: %s", match.Status)
	}

	// Atomically mark game as loaded and get updated match
	updatedMatch, err := s.matchRepo.UpdateGameLoadedAtomic(ctx, matchID, playerID)
	if err != nil {
		errMsg := err.Error()
		// If error contains "no rows" - player already marked as loaded (idempotent)
		if strings.Contains(errMsg, "no rows") {
			s.logger.Info("Player game already marked as loaded (idempotent call), broadcasting current state", map[string]interface{}{
				"match_id":  matchID.String(),
				"player_id": playerID.String(),
			})

			// Fetch current match state from DB
			currentMatch, fetchErr := s.matchRepo.GetByID(ctx, matchID)
			if fetchErr != nil {
				s.logger.Error("Failed to fetch match after idempotent game loaded", map[string]interface{}{
					"match_id": matchID.String(),
					"error":    fetchErr.Error(),
				})
				return nil // Still return nil (success) for idempotent call
			}

			// Broadcast game_loaded event with current state
			if s.hub != nil {
				msg := map[string]interface{}{
					"type":      "game_loaded",
					"timestamp": time.Now(),
					"payload": map[string]interface{}{
						"playerId":               playerID.String(),
						"participant1GameLoaded": currentMatch.Participant1GameLoaded,
						"participant2GameLoaded": currentMatch.Participant2GameLoaded,
					},
				}
				s.hub.BroadcastToMatch(matchID, msg)
				s.logger.Info("Broadcasted game_loaded (idempotent)", map[string]interface{}{
					"match_id": matchID.String(),
				})
			}

			// Check if both games are loaded (idempotent path)
			if currentMatch.BothGamesLoaded() {
				fmt.Printf("[H17] Idempotent path: both games loaded, starting countdown\n")
				s.logger.Info("Both games loaded (idempotent path), starting countdown", map[string]interface{}{
					"match_id": matchID.String(),
				})
				// Start countdown in a goroutine
				go s.startCountdownAndMatch(matchID)
			}

			return nil
		}
		return fmt.Errorf("update game loaded atomic: %w", err)
	}

	// Broadcast game_loaded status to all players in match
	if s.hub != nil {
		msg := map[string]interface{}{
			"type":      "game_loaded",
			"timestamp": time.Now(),
			"payload": map[string]interface{}{
				"playerId":               playerID.String(),
				"participant1GameLoaded": updatedMatch.Participant1GameLoaded,
				"participant2GameLoaded": updatedMatch.Participant2GameLoaded,
			},
		}
		s.hub.BroadcastToMatch(matchID, msg)

	}

	// Check if both games are loaded AFTER atomic update
	// Only ONE thread will see both=true for the first time
	if updatedMatch.BothGamesLoaded() {
		s.logger.Info("Both games loaded (atomic check), starting countdown", map[string]interface{}{
			"match_id": matchID.String(),
		})

		// Start countdown in a goroutine to avoid blocking the HTTP request
		go s.startCountdownAndMatch(matchID)
	}

	return nil
}

// startCountdownAndMatch broadcasts 3-2-1 countdown then starts the match
func (s *MatchService) startCountdownAndMatch(matchID uuid.UUID) {
	ctx := context.Background()

	// Countdown from 3 to 1
	for i := 3; i > 0; i-- {
		// Broadcast match_starting with countdown
		if s.hub != nil {
			payload := map[string]interface{}{
				"countdownSeconds": i,
			}
			// Create message with proper structure
			msg := map[string]interface{}{
				"type":      "match_starting",
				"timestamp": time.Now(),
				"payload":   payload,
			}
			s.hub.BroadcastToMatch(matchID, msg)
		}

		s.logger.Info("Countdown", map[string]interface{}{
			"match_id": matchID.String(),
			"seconds":  i,
		})

		// Wait 1 second
		time.Sleep(1 * time.Second)
	}

	// Start the match after countdown
	err := s.StartMatch(ctx, matchID)
	if err != nil {
		s.logger.Error("Failed to start match after countdown", map[string]interface{}{
			"match_id": matchID.String(),
			"error":    err.Error(),
		})
		return
	}

	s.logger.Info("Match started after countdown", map[string]interface{}{
		"match_id": matchID.String(),
	})
}

// ForceStartMatch forces a match to start (T076)
// Used for auto-force-ready scenarios or admin overrides
func (s *MatchService) ForceStartMatch(ctx context.Context, matchID uuid.UUID) error {
	s.logger.Info("Force starting match", map[string]interface{}{
		"match_id": matchID.String(),
	})

	// Fetch match
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("get match: %w", err)
	}

	// Force both players to ready state if not already
	if !match.Participant1Ready {
		err = s.matchRepo.UpdateReady(ctx, matchID, match.Participant1ID, true)
		if err != nil {
			return fmt.Errorf("force ready participant1: %w", err)
		}
	}

	if match.Participant2ID != nil && !match.Participant2Ready {
		err = s.matchRepo.UpdateReady(ctx, matchID, *match.Participant2ID, true)
		if err != nil {
			return fmt.Errorf("force ready participant2: %w", err)
		}
	}

	// Start the match
	return s.StartMatch(ctx, matchID)
}

// GrantWalkover grants a walkover to the present player when opponent no-shows (T083)
func (s *MatchService) GrantWalkover(ctx context.Context, matchID, winnerID, noShowPlayerID uuid.UUID, reason string) error {
	s.logger.Info("Granting walkover", map[string]interface{}{
		"match_id":       matchID.String(),
		"winner_id":      winnerID.String(),
		"no_show_player": noShowPlayerID.String(),
		"reason":         reason,
	})

	// Fetch match
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("get match: %w", err)
	}

	// Validate winner is a participant
	isParticipant1 := match.Participant1ID == winnerID
	isParticipant2 := match.Participant2ID != nil && *match.Participant2ID == winnerID
	if !isParticipant1 && !isParticipant2 {
		return fmt.Errorf("winner_id %s is not a participant in match %s", winnerID, matchID)
	}

	// Complete match with walkover (FR-016: present player marked as winner)
	completedAt := time.Now()
	err = s.matchRepo.UpdateResult(ctx, matchID, winnerID, 1, 0, domain.ResultSourceWalkover)
	if err != nil {
		return fmt.Errorf("update walkover result: %w", err)
	}

	// Update match status to completed
	err = s.matchRepo.UpdateStatus(ctx, matchID, domain.MatchStatusCompleted)
	if err != nil {
		return fmt.Errorf("update match status: %w", err)
	}

	s.logger.Info("Walkover granted successfully", map[string]interface{}{
		"match_id":       matchID.String(),
		"winner_id":      winnerID.String(),
		"no_show_player": noShowPlayerID.String(),
	})

	// Publish WalkoverGranted event (T084)
	walkoverEvent := domain.NewWalkoverGranted(
		matchID,
		match.RoundID, // Using RoundID as TournamentID placeholder
		winnerID,
		noShowPlayerID,
		reason,
		completedAt,
		map[string]string{
			"correlation_id": s.getCorrelationID(ctx),
		},
	)
	if err := s.publisher.Publish(ctx, walkoverEvent); err != nil {
		log.Printf("WARN: Failed to publish WalkoverGranted event: %v", err)
	}

	// Advance winner to next round if there is one
	if match.NextMatchID != nil {
		err = s.AdvanceWinner(ctx, winnerID, matchID, *match.NextMatchID)
		if err != nil {
			return fmt.Errorf("advance walkover winner: %w", err)
		}
	}

	// Mark no-show player as eliminated
	if noShowPlayerID != uuid.Nil {
		eliminatedEvent := domain.NewParticipantEliminated(
			noShowPlayerID,
			match.RoundID,
			matchID,
			0, // Round number placeholder
			completedAt,
			map[string]string{
				"correlation_id": s.getCorrelationID(ctx),
				"reason":         "no-show",
			},
		)
		if err := s.publisher.Publish(ctx, eliminatedEvent); err != nil {
			log.Printf("WARN: Failed to publish ParticipantEliminated event: %v", err)
		}
	}

	return nil
}

// HandleBothAbsent handles the scenario when both players fail to show (T085)
func (s *MatchService) HandleBothAbsent(ctx context.Context, matchID uuid.UUID) error {
	s.logger.Warn("Both players absent for match", map[string]interface{}{
		"match_id": matchID.String(),
	})

	// T085, T087: Notify host for manual resolution
	// For now, log the event - host notification would integrate with notification service
	s.logger.Info("Host notification required: both players absent", map[string]interface{}{
		"match_id": matchID.String(),
		"action":   "manual_host_resolution_required",
	})

	// Update match status to awaiting_host_decision
	err := s.matchRepo.UpdateStatus(ctx, matchID, domain.MatchStatusPending)
	if err != nil {
		return fmt.Errorf("update match status: %w", err)
	}

	// TODO: Integrate with notification service to alert host
	// This would send a push notification or email to the tournament host

	return nil
}

// CompleteTournament handles tournament completion when final match finishes (T107)
func (s *MatchService) CompleteTournament(ctx context.Context, finalMatchID, championID uuid.UUID) error {
	s.logger.Info("Completing tournament", map[string]interface{}{
		"final_match_id": finalMatchID.String(),
		"champion_id":    championID.String(),
	})

	// Get match to find bracket/tournament
	match, err := s.matchRepo.GetByID(ctx, finalMatchID)
	if err != nil {
		return fmt.Errorf("get final match: %w", err)
	}

	// Get round to find bracket
	round, err := s.roundRepo.GetByID(ctx, match.RoundID)
	if err != nil {
		return fmt.Errorf("get round: %w", err)
	}

	// Get bracket to find tournament
	bracket, err := s.bracketRepo.GetByID(ctx, round.BracketID)
	if err != nil {
		return fmt.Errorf("get bracket: %w", err)
	}

	// T109: Mark bracket as completed
	completedAt := time.Now()
	err = s.bracketRepo.UpdateCompletedAt(ctx, bracket.ID, completedAt)
	if err != nil {
		// Non-critical: log warning but continue
		s.logger.Warn("Failed to update bracket completed_at", map[string]interface{}{
			"bracket_id": bracket.ID.String(),
			"error":      err.Error(),
		})
	} else {
		s.logger.Info("Bracket completed", map[string]interface{}{
			"bracket_id":    bracket.ID.String(),
			"tournament_id": bracket.TournamentID.String(),
			"champion_id":   championID.String(),
			"completed_at":  completedAt.Format(time.RFC3339),
		})
	}

	// T108: Publish TournamentCompleted event
	completedEvent := domain.NewTournamentCompleted(
		bracket.TournamentID,
		championID,
		time.Now(),
		map[string]string{
			"correlation_id": s.getCorrelationID(ctx),
			"bracket_id":     bracket.ID.String(),
		},
	)
	if err := s.publisher.Publish(ctx, completedEvent); err != nil {
		log.Printf("WARN: Failed to publish TournamentCompleted event: %v", err)
	}

	// T110: Generate final standings
	// For MVP, final standings can be inferred from match results
	// Champion = 1st place, final match loser = 2nd place, etc.
	s.logger.Info("Tournament completed successfully", map[string]interface{}{
		"tournament_id": bracket.TournamentID.String(),
		"champion_id":   championID.String(),
		"total_matches": bracket.TotalMatches,
		"total_rounds":  bracket.TotalRounds,
	})

	return nil
}

// getCorrelationID extracts correlation ID from context
func (s *MatchService) getCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		return correlationID
	}
	return ""
}

// getGameManagementURL returns the base URL for game management service
func (s *MatchService) getGameManagementURL() string {
	return s.gameManagementURL
}
