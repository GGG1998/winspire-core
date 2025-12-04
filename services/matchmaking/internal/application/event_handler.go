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
	bracketService *BracketService
	logger         *observability.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(bracketService *BracketService, logger *observability.Logger) *EventHandler {
	return &EventHandler{
		bracketService: bracketService,
		logger:         logger,
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

	// T079, T080: Validate minimum participant count (at least 2 required)
	// Note: If check-in is enabled, competition service has already filtered to checked-in participants
	// This validation ensures we don't generate brackets with insufficient players
	if len(eventPayload.Participants) < 2 {
		h.logger.Error("Insufficient participants for bracket generation", map[string]interface{}{
			"tournament_id":     eventPayload.TournamentID,
			"participant_count": len(eventPayload.Participants),
			"correlation_id":    metadata["correlation_id"],
		})
		return fmt.Errorf("at least 2 participants required, got %d", len(eventPayload.Participants))
	}

	// Parse tournament ID
	tournamentID, err := uuid.Parse(eventPayload.TournamentID)
	if err != nil {
		return fmt.Errorf("parse tournament_id: %w", err)
	}

	// T078: Parse participant UUIDs
	// The participant list is pre-filtered by competition service:
	// - If check-in disabled: all registered participants
	// - If check-in enabled: only checked-in participants
	participants := make([]uuid.UUID, len(eventPayload.Participants))
	for i, participantStr := range eventPayload.Participants {
		participantID, err := uuid.Parse(participantStr)
		if err != nil {
			return fmt.Errorf("parse participant %d: %w", i, err)
		}
		participants[i] = participantID
	}

	// Add correlation ID to context
	if correlationID, ok := metadata["correlation_id"]; ok {
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
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

// RegisterHandlers registers all event handlers with the subscriber
func (h *EventHandler) RegisterHandlers(subscriber *pubsub.EventSubscriber) {
	subscriber.Subscribe("TournamentStarted", h.HandleTournamentStarted)
	log.Println("[EventHandler] Registered handler for TournamentStarted")
}
