package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/pubsub"
	"github.com/winspire-core/services/matchmaking/internal/websocket"
)

// EventHandler handles incoming domain events
type EventHandler struct {
	bracketService         *BracketService
	preLobbyService        *PreLobbyService
	matchAssignmentService *MatchAssignmentService
	publisher              *pubsub.EventPublisher
	logger                 *observability.Logger
	competitionClient      *CompetitionClient
	gameManagementURL      string
	hub                    *websocket.Hub
}

// NewEventHandler creates a new event handler
func NewEventHandler(
	bracketService *BracketService,
	preLobbyService *PreLobbyService,
	matchAssignmentService *MatchAssignmentService,
	publisher *pubsub.EventPublisher,
	logger *observability.Logger,
	competitionClient *CompetitionClient,
	gameManagementURL string,
	hub *websocket.Hub,
) *EventHandler {
	return &EventHandler{
		bracketService:         bracketService,
		preLobbyService:        preLobbyService,
		matchAssignmentService: matchAssignmentService,
		publisher:              publisher,
		logger:                 logger,
		competitionClient:      competitionClient,
		gameManagementURL:      gameManagementURL,
		hub:                    hub,
	}
}

// TournamentStartRequestedPayload represents the payload of TournamentStartRequested event
type TournamentStartRequestedPayload struct {
	TournamentID string                 `json:"tournament_id"`
	HostID       string                 `json:"host_id"`
	Participants []string               `json:"participants"` // Array of participant UUIDs
	GameID       string                 `json:"game_id"`
	GameSnapshot map[string]interface{} `json:"game_snapshot,omitempty"` // Optional game snapshot
	StartedAt    string                 `json:"started_at"`
}

// HandleTournamentStartRequested processes TournamentStartRequested events from tournament service
// This initiates the tournament start saga: grace period → bracket generation → confirmation
func (h *EventHandler) HandleTournamentStartRequested(ctx context.Context, eventType string, payload map[string]interface{}, metadata map[string]string) error {
	h.logger.Info("Received TournamentStartRequested event", map[string]interface{}{
		"event_type":     eventType,
		"correlation_id": metadata["correlation_id"],
	})

	// Parse payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var eventPayload TournamentStartRequestedPayload
	if err := json.Unmarshal(payloadBytes, &eventPayload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Validate payload
	if eventPayload.TournamentID == "" {
		return fmt.Errorf("tournament_id is required")
	}

	// Parse tournament ID
	tournamentID, err := uuid.Parse(eventPayload.TournamentID)
	if err != nil {
		return fmt.Errorf("parse tournament_id: %w", err)
	}

	// Add correlation ID to context
	if correlationID, ok := metadata["correlation_id"]; ok {
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
	}

	// Extract game snapshot from event payload if available
	var gameSnapshot *domain.GameSnapshot
	if eventPayload.GameSnapshot != nil {
		gameSnapshot = &domain.GameSnapshot{}
		if id, ok := eventPayload.GameSnapshot["id"].(string); ok {
			parsedID, _ := uuid.Parse(id)
			gameSnapshot.ID = parsedID
		}
		if slug, ok := eventPayload.GameSnapshot["slug"].(string); ok {
			gameSnapshot.Slug = slug
		}
		if name, ok := eventPayload.GameSnapshot["name"].(string); ok {
			gameSnapshot.Name = name
		}
		if version, ok := eventPayload.GameSnapshot["version"].(string); ok {
			gameSnapshot.Version = version
		}
		if logoURL, ok := eventPayload.GameSnapshot["logoUrl"].(string); ok {
			gameSnapshot.LogoURL = &logoURL
		}
		if description, ok := eventPayload.GameSnapshot["description"].(string); ok {
			gameSnapshot.Description = &description
		}
		if storagePath, ok := eventPayload.GameSnapshot["storagePath"].(string); ok {
			gameSnapshot.StoragePath = storagePath
		}
	}

	// Check if pre-lobby service is available
	if h.preLobbyService != nil {
		// Check if grace period is already active (duplicate event)
		isActive, err := h.preLobbyService.IsGracePeriodActive(ctx, tournamentID)
		if err != nil {
			h.logger.Warn("failed to check grace period status", map[string]interface{}{
				"tournament_id": tournamentID.String(),
				"error":         err.Error(),
			})
		}
		if isActive {
			h.logger.Info("grace period already active, ignoring duplicate TournamentStarted event", map[string]interface{}{
				"tournament_id": tournamentID.String(),
			})
			return nil
		}

		// Ensure pre-lobby exists before starting grace period
		// This fixes "pre-lobby not found" error when no one joined before tournament start
		_, err = h.preLobbyService.GetOrCreatePreLobby(ctx, tournamentID, 2)
		if err != nil {
			h.logger.Error("Failed to get or create pre-lobby", map[string]interface{}{
				"tournament_id": tournamentID.String(),
				"error":         err.Error(),
			})
			// Fall back to immediate bracket generation
		} else {
			// Start grace period - bracket will be generated after grace period ends
			err = h.preLobbyService.StartGracePeriod(ctx, tournamentID, func(tournamentID uuid.UUID, participantIDs []uuid.UUID) {
				// This callback is called when grace period ends
				h.logger.Info("Grace period ended, generating bracket", map[string]interface{}{
					"tournament_id":     tournamentID.String(),
					"participant_count": len(participantIDs),
				})

				if len(participantIDs) < 2 {
					h.logger.Error("Insufficient participants for bracket generation after grace period", map[string]interface{}{
						"tournament_id":     tournamentID.String(),
						"participant_count": len(participantIDs),
					})
					return
				}

				// Record bracket generation in activity feed
				if err := h.preLobbyService.RecordBracketGeneration(context.Background(), tournamentID); err != nil {
					h.logger.Warn("Failed to record bracket generation event", map[string]interface{}{
						"tournament_id": tournamentID.String(),
						"error":         err.Error(),
					})
				}

				// Generate bracket with participants from snapshot
				if err := h.bracketService.GenerateBracket(context.Background(), tournamentID, participantIDs, gameSnapshot); err != nil {
					h.logger.Error("Failed to generate bracket after grace period", map[string]interface{}{
						"tournament_id": tournamentID.String(),
						"error":         err.Error(),
					})

					// SAGA COMPENSATION: Cancel tournament and notify participants
					compensationCtx := context.Background()
					reason := fmt.Sprintf("Bracket generation failed: %v", err)

					// Publish BracketGenerationFailed event (triggers rollback in tournament service)
					failedEvent := domain.NewBracketGenerationFailed(
						tournamentID,
						err,
						reason,
						nil,
					)
					if pubErr := h.publisher.Publish(compensationCtx, failedEvent); pubErr != nil {
						h.logger.Error("Failed to publish BracketGenerationFailed event", map[string]interface{}{
							"tournament_id": tournamentID.String(),
							"error":         pubErr.Error(),
						})
					}

					// Cancel the pre-lobby tournament
					if cancelErr := h.preLobbyService.CancelTournamentWithError(compensationCtx, tournamentID, reason); cancelErr != nil {
						h.logger.Error("Failed to cancel tournament during compensation", map[string]interface{}{
							"tournament_id": tournamentID.String(),
							"error":         cancelErr.Error(),
						})
					}

					// Log compensation action
					h.logger.Info("Tournament cancelled via SAGA compensation", map[string]interface{}{
						"tournament_id": tournamentID.String(),
						"reason":        reason,
					})

					return
				}

				h.logger.Info("Bracket generated successfully after grace period", map[string]interface{}{
					"tournament_id": tournamentID.String(),
					"participants":  len(participantIDs),
				})
			})

			if err != nil {
				h.logger.Error("Failed to start grace period", map[string]interface{}{
					"tournament_id": tournamentID.String(),
					"error":         err.Error(),
				})
				// Fall back to immediate bracket generation
			} else {
				h.logger.Info("Grace period started successfully", map[string]interface{}{
					"tournament_id": tournamentID.String(),
				})
				return nil
			}
		}
	}

	// Fallback: Generate bracket immediately (legacy behavior when no pre-lobby service)
	// T079, T080: Validate minimum participant count (at least 2 required)
	if len(eventPayload.Participants) < 2 {
		h.logger.Error("Insufficient participants for bracket generation", map[string]interface{}{
			"tournament_id":     eventPayload.TournamentID,
			"participant_count": len(eventPayload.Participants),
			"correlation_id":    metadata["correlation_id"],
		})
		return fmt.Errorf("at least 2 participants required, got %d", len(eventPayload.Participants))
	}

	// T078: Parse participant UUIDs
	participants := make([]uuid.UUID, len(eventPayload.Participants))
	for i, participantStr := range eventPayload.Participants {
		participantID, err := uuid.Parse(participantStr)
		if err != nil {
			return fmt.Errorf("parse participant %d: %w", i, err)
		}
		participants[i] = participantID
	}

	// Record bracket generation in activity feed (if pre-lobby service available)
	if h.preLobbyService != nil {
		if err := h.preLobbyService.RecordBracketGeneration(ctx, tournamentID); err != nil {
			h.logger.Warn("Failed to record bracket generation event", map[string]interface{}{
				"tournament_id": tournamentID.String(),
				"error":         err.Error(),
			})
		}
	}

	// Generate bracket
	if err := h.bracketService.GenerateBracket(ctx, tournamentID, participants, gameSnapshot); err != nil {
		h.logger.Error("Failed to generate bracket", map[string]interface{}{
			"tournament_id":  tournamentID.String(),
			"error":          err.Error(),
			"correlation_id": metadata["correlation_id"],
		})
		return fmt.Errorf("generate bracket: %w", err)
	}

	h.logger.Info("Bracket generated successfully", map[string]interface{}{
		"tournament_id":  tournamentID.String(),
		"participants":   len(participants),
		"correlation_id": metadata["correlation_id"],
	})

	return nil
}

// HandleBracketGenerated processes BracketGenerated events
// Sends match_assigned notifications to all players in first round
func (h *EventHandler) HandleBracketGenerated(ctx context.Context, eventType string, payload map[string]interface{}, metadata map[string]string) error {
	h.logger.Info("Received BracketGenerated event", map[string]interface{}{
		"event_type": eventType,
	})

	// Parse tournament ID from payload
	tournamentIDStr, ok := payload["tournament_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid tournament_id in payload")
	}

	tournamentID, err := uuid.Parse(tournamentIDStr)
	if err != nil {
		return fmt.Errorf("parse tournament_id: %w", err)
	}

	// Broadcast match assignments for round 1 using the centralized service
	if err := h.matchAssignmentService.BroadcastMatchAssignmentsForRound(ctx, tournamentID, 1); err != nil {
		return fmt.Errorf("broadcast match assignments for round 1: %w", err)
	}

	return nil
}

// HandleMatchCreated processes MatchCreated events
// Sends match_assigned notifications to players for rounds 2+
func (h *EventHandler) HandleMatchCreated(ctx context.Context, eventType string, payload map[string]interface{}, metadata map[string]string) error {
	h.logger.Info("Received MatchCreated event", map[string]interface{}{
		"event_type": eventType,
	})

	// Parse tournament ID
	tournamentIDStr, ok := payload["tournament_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid tournament_id in payload")
	}

	tournamentID, err := uuid.Parse(tournamentIDStr)
	if err != nil {
		return fmt.Errorf("parse tournament_id: %w", err)
	}

	// Parse match ID
	matchIDStr, ok := payload["match_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid match_id in payload")
	}

	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		return fmt.Errorf("parse match_id: %w", err)
	}

	// Parse round number
	roundNumber, ok := payload["round_number"].(float64) // JSON numbers are float64
	if !ok {
		return fmt.Errorf("missing or invalid round_number in payload")
	}

	// Parse match number
	matchNumber, ok := payload["match_number"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid match_number in payload")
	}

	// Parse participant 1 ID
	participant1IDStr, ok := payload["participant1_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid participant1_id in payload")
	}

	participant1ID, err := uuid.Parse(participant1IDStr)
	if err != nil {
		return fmt.Errorf("parse participant1_id: %w", err)
	}

	// Parse participant 2 ID (nullable for BYE)
	var participant2ID *uuid.UUID
	if participant2IDStr, ok := payload["participant2_id"].(string); ok && participant2IDStr != "" {
		parsedID, err := uuid.Parse(participant2IDStr)
		if err != nil {
			return fmt.Errorf("parse participant2_id: %w", err)
		}
		participant2ID = &parsedID
	}

	// Check if this is a BYE match
	isBye, _ := payload["is_bye"].(bool)

	h.logger.Info("Processing MatchCreated event", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"match_id":      matchID.String(),
		"round_number":  int(roundNumber),
		"match_number":  int(matchNumber),
		"is_bye":        isBye,
	})

	// Send match_assigned to participant 1
	var opponent1 *PreLobbyParticipantInfo
	if participant2ID != nil && *participant2ID != uuid.Nil {
		opponent1 = &PreLobbyParticipantInfo{
			UserID:      *participant2ID,
			DisplayName: "Opponent",
		}
	}

	h.preLobbyService.BroadcastMatchAssigned(
		tournamentID,
		participant1ID,
		matchID,
		int(roundNumber),
		int(matchNumber),
		opponent1,
	)

	// Send to participant 2 (if not BYE)
	if participant2ID != nil && *participant2ID != uuid.Nil {
		opponent2 := &PreLobbyParticipantInfo{
			UserID:      participant1ID,
			DisplayName: "Opponent",
		}

		h.preLobbyService.BroadcastMatchAssigned(
			tournamentID,
			*participant2ID,
			matchID,
			int(roundNumber),
			int(matchNumber),
			opponent2,
		)

		h.logger.Info("Sent match assignments to both participants", map[string]interface{}{
			"tournament_id":   tournamentID.String(),
			"match_id":        matchID.String(),
			"participant1_id": participant1ID.String(),
			"participant2_id": participant2ID.String(),
		})
	} else {
		h.logger.Info("Sent match assignment (BYE match)", map[string]interface{}{
			"tournament_id":   tournamentID.String(),
			"match_id":        matchID.String(),
			"participant1_id": participant1ID.String(),
		})
	}

	return nil
}

// HandleMatchStarted processes MatchStarted events and broadcasts match_started WebSocket message with gameUrl
func (h *EventHandler) HandleMatchStarted(ctx context.Context, eventType string, payload map[string]interface{}, metadata map[string]string) error {
	h.logger.Info("Received MatchStarted event", map[string]interface{}{
		"event_type": eventType,
	})

	// Parse match ID
	matchIDStr, ok := payload["match_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid match_id in payload")
	}

	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		return fmt.Errorf("parse match_id: %w", err)
	}

	// Parse tournament ID
	tournamentIDStr, ok := payload["tournament_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid tournament_id in payload")
	}

	tournamentID, err := uuid.Parse(tournamentIDStr)
	if err != nil {
		return fmt.Errorf("parse tournament_id: %w", err)
	}

	h.logger.Info("Processing MatchStarted event", map[string]interface{}{
		"match_id":      matchID.String(),
		"tournament_id": tournamentID.String(),
	})

	// Fetch match to get round and bracket info
	match, err := h.bracketService.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		h.logger.Error("Failed to fetch match", map[string]interface{}{
			"match_id": matchID.String(),
			"error":    err.Error(),
		})
		return fmt.Errorf("fetch match: %w", err)
	}

	// Fetch round to get bracket ID
	round, err := h.bracketService.roundRepo.GetByID(ctx, match.RoundID)
	if err != nil {
		h.logger.Error("Failed to fetch round", map[string]interface{}{
			"round_id": match.RoundID.String(),
			"error":    err.Error(),
		})
		return fmt.Errorf("fetch round: %w", err)
	}

	// Fetch bracket to get game snapshot
	bracket, err := h.bracketService.bracketRepo.GetByID(ctx, round.BracketID)
	if err != nil {
		h.logger.Error("Failed to fetch bracket", map[string]interface{}{
			"bracket_id": round.BracketID.String(),
			"error":      err.Error(),
		})
		return fmt.Errorf("fetch bracket: %w", err)
	}

	// Try to use game snapshot from bracket (fast path - no HTTP calls)
	var gameURL string
	gameSessionID := matchID.String()

	if bracket.GameSnapshot != nil {
		gameURL = fmt.Sprintf("%s/v1/g/%s/bundle/", h.gameManagementURL, bracket.GameSnapshot.Slug)
		h.logger.Info("Using game snapshot from bracket", map[string]interface{}{
			"match_id":  matchID.String(),
			"game_slug": bracket.GameSnapshot.Slug,
		})
	}

	h.logger.Info("Game URL constructed for match", map[string]interface{}{
		"match_id":        matchID.String(),
		"game_url":        gameURL,
		"game_session_id": gameSessionID,
	})

	// Broadcast match_started via WebSocket
	if h.hub != nil {
		msg := map[string]interface{}{
			"type":      "match_started",
			"timestamp": time.Now(),
			"payload": map[string]interface{}{
				"gameUrl":       gameURL,
				"gameSessionId": gameSessionID,
			},
		}

		h.hub.BroadcastToMatch(matchID, msg)

		h.logger.Info("Broadcasted match_started message", map[string]interface{}{
			"match_id": matchID.String(),
		})
	} else {
		h.logger.Warn("WebSocket hub not available, cannot broadcast match_started", map[string]interface{}{
			"match_id": matchID.String(),
		})
	}

	return nil
}

// HandleNextRoundStartRequest processes NextRoundStartRequest events
// This is triggered when all winners from the previous round should be ready for the next round
// It starts a grace period and then broadcasts match assignments for the next round
func (h *EventHandler) HandleNextRoundStartRequest(ctx context.Context, eventType string, payload map[string]interface{}, metadata map[string]string) error {
	h.logger.Info("Received NextRoundStartRequest event", map[string]interface{}{
		"event_type":     eventType,
		"correlation_id": metadata["correlation_id"],
	})

	fmt.Println("[EventHandle] HandleNextRoundStart", payload)

	// Parse tournament ID from payload
	tournamentIDStr, ok := payload["tournament_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid tournament_id in payload")
	}

	tournamentID, err := uuid.Parse(tournamentIDStr)
	if err != nil {
		return fmt.Errorf("parse tournament_id: %w", err)
	}

	// Parse round number from payload
	roundNumberFloat, ok := payload["round_number"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid round_number in payload")
	}
	roundNumber := int(roundNumberFloat)

	h.logger.Info("Processing NextRoundStartRequest", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"round_number":  roundNumber,
	})
	fmt.Println("[EventHandler] HandleNextRoundStartRequest Pre Get")
	// 1. Ensure pre-lobby exists for the next round
	_, err = h.preLobbyService.GetOrCreatePreLobby(ctx, tournamentID, 2)
	if err != nil {
		h.logger.Error("Failed to get or create pre-lobby for next round", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"round_number":  roundNumber,
			"error":         err.Error(),
		})
		return fmt.Errorf("get or create pre-lobby: %w", err)
	}

	// 2. Reset pre-lobby status to 'waiting' before starting grace period
	// This ensures the pre-lobby is in correct state regardless of race conditions
	if err := h.preLobbyService.PrepareForNextRound(ctx, tournamentID, 2); err != nil {
		h.logger.Warn("Failed to prepare pre-lobby for next round (continuing anyway)", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"round_number":  roundNumber,
			"error":         err.Error(),
		})
	}

	fmt.Println("[EventHandler] HandleNextRoundStartRequest Pre Start Grace")
	// 3. Start grace period with callback to assign matches after it ends
	err = h.preLobbyService.StartGracePeriod(ctx, tournamentID, func(tID uuid.UUID, participantIDs []uuid.UUID) {
		h.logger.Info("Grace period ended for next round, assigning matches", map[string]interface{}{
			"tournament_id":     tID.String(),
			"round_number":      roundNumber,
			"participant_count": len(participantIDs),
		})

		// Broadcast match assignments using the centralized service
		if err := h.matchAssignmentService.BroadcastMatchAssignmentsForRound(context.Background(), tID, roundNumber); err != nil {
			h.logger.Error("Failed to broadcast match assignments for round", map[string]interface{}{
				"tournament_id": tID.String(),
				"round_number":  roundNumber,
				"error":         err.Error(),
			})
		}
	})

	if err != nil {
		h.logger.Error("Failed to start grace period for next round", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"round_number":  roundNumber,
			"error":         err.Error(),
		})
		return fmt.Errorf("start grace period: %w", err)
	}

	h.logger.Info("Grace period started for next round", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"round_number":  roundNumber,
	})

	return nil
}

// RegisterHandlers registers all event handlers with the subscriber
func (h *EventHandler) RegisterHandlers(subscriber *pubsub.EventSubscriber) {
	subscriber.Subscribe("TournamentStartRequested", h.HandleTournamentStartRequested)
	subscriber.Subscribe("BracketGenerated", h.HandleBracketGenerated)
	subscriber.Subscribe("MatchCreated", h.HandleMatchCreated)
	subscriber.Subscribe("MatchStarted", h.HandleMatchStarted)
	subscriber.Subscribe("NextRoundStartRequest", h.HandleNextRoundStartRequest)
	log.Println("[EventHandler] Registered handlers for TournamentStartRequested, BracketGenerated, MatchCreated, MatchStarted")
}
