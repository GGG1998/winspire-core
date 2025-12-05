package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/pubsub"
)

// EventHandler handles incoming domain events
type EventHandler struct {
	bracketService  *BracketService
	preLobbyService *PreLobbyService
	logger          *observability.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(bracketService *BracketService, preLobbyService *PreLobbyService, logger *observability.Logger) *EventHandler {
	return &EventHandler{
		bracketService:  bracketService,
		preLobbyService: preLobbyService,
		logger:          logger,
	}
}

// TournamentStartedPayload represents the payload of TournamentStarted event
type TournamentStartedPayload struct {
	TournamentID string   `json:"tournament_id"`
	HostID       string   `json:"host_id"`
	Participants []string `json:"participants"` // Array of participant UUIDs
	GameID       string   `json:"game_id"`
	StartedAt    string   `json:"started_at"`
}

// HandleTournamentStarted processes TournamentStarted events from competition service
// This now starts the grace period instead of immediately generating brackets
func (h *EventHandler) HandleTournamentStarted(ctx context.Context, eventType string, payload map[string]interface{}, metadata map[string]string) error {
	h.logger.Info("Received TournamentStarted event", map[string]interface{}{
		"event_type":     eventType,
		"correlation_id": metadata["correlation_id"],
	})

	// Parse payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var eventPayload TournamentStartedPayload
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

				// Generate bracket with participants from snapshot
				if err := h.bracketService.GenerateBracket(context.Background(), tournamentID, participantIDs); err != nil {
					h.logger.Error("Failed to generate bracket after grace period", map[string]interface{}{
						"tournament_id": tournamentID.String(),
						"error":         err.Error(),
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

	// Generate bracket
	if err := h.bracketService.GenerateBracket(ctx, tournamentID, participants); err != nil {
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

	// Get all matches from round 1
	matches, err := h.bracketService.GetMatchesByRound(ctx, tournamentID, 1)
	if err != nil {
		h.logger.Error("Failed to get round 1 matches", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return fmt.Errorf("get round 1 matches: %w", err)
	}

	// Round 1 by definition has round_number = 1
	roundNumber := 1

	// Send match_assigned to each participant
	for _, match := range matches {
		// Send to participant 1
		if match.Participant1ID != uuid.Nil {
			var opponent *PreLobbyParticipantInfo
			if match.Participant2ID != nil && *match.Participant2ID != uuid.Nil {
				// Create opponent info (using placeholder for display name)
				opponent = &PreLobbyParticipantInfo{
					UserID:      *match.Participant2ID,
					DisplayName: "Opponent",
				}
			}

			h.preLobbyService.BroadcastMatchAssigned(
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
			opponent := &PreLobbyParticipantInfo{
				UserID:      match.Participant1ID,
				DisplayName: "Opponent",
			}

			h.preLobbyService.BroadcastMatchAssigned(
				tournamentID,
				*match.Participant2ID,
				match.ID,
				roundNumber,
				match.MatchNumber,
				opponent,
			)
		}
	}

	h.logger.Info("Sent match assignments", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"match_count":   len(matches),
	})

	return nil
}

// RegisterHandlers registers all event handlers with the subscriber
func (h *EventHandler) RegisterHandlers(subscriber *pubsub.EventSubscriber) {
	subscriber.Subscribe("TournamentStarted", h.HandleTournamentStarted)
	subscriber.Subscribe("BracketGenerated", h.HandleBracketGenerated)
	log.Println("[EventHandler] Registered handlers for TournamentStarted, BracketGenerated")
}
