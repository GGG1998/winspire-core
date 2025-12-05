package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/winspire/competition/internal/pubsub"
	"github.com/winspire/competition/internal/repository"
)

// EventHandler handles incoming domain events from other services
type EventHandler struct {
	tournamentRepo *repository.TournamentRepository
	publisher      *pubsub.EventPublisher
	logger         *slog.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(
	tournamentRepo *repository.TournamentRepository,
	publisher *pubsub.EventPublisher,
	logger *slog.Logger,
) *EventHandler {
	return &EventHandler{
		tournamentRepo: tournamentRepo,
		publisher:      publisher,
		logger:         logger,
	}
}

// GracePeriodStartedPayload represents the payload from matchmaking when grace period starts
type GracePeriodStartedPayload struct {
	TournamentID string    `json:"tournament_id"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

// HandleGracePeriodStarted processes GracePeriodStarted events from matchmaking service
// This confirms that the tournament start saga has progressed successfully
// Transitions tournament status from "starting" → "started"
func (h *EventHandler) HandleGracePeriodStarted(ctx context.Context, eventType string, payload map[string]interface{}, metadata map[string]string) error {
	h.logger.Info("received GracePeriodStarted event", 
		"event_type", eventType,
		"correlation_id", metadata["correlation_id"],
	)

	// Parse payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var eventPayload GracePeriodStartedPayload
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

	// Get tournament
	tournament, err := h.tournamentRepo.GetByID(ctx, tournamentID)
	if err != nil {
		h.logger.Error("failed to get tournament", 
			"tournament_id", tournamentID.String(),
			"error", err,
		)
		return fmt.Errorf("get tournament: %w", err)
	}

	// Confirm start (transitions "starting" → "started")
	if err := tournament.ConfirmStart(); err != nil {
		h.logger.Warn("failed to confirm start", 
			"tournament_id", tournamentID.String(),
			"current_status", tournament.Status,
			"error", err,
		)
		return fmt.Errorf("confirm start: %w", err)
	}

	// Save tournament state
	if _, err := h.tournamentRepo.Save(ctx, tournament); err != nil {
		h.logger.Error("failed to save tournament after confirming start", 
			"tournament_id", tournamentID.String(),
			"error", err,
		)
		return fmt.Errorf("save tournament: %w", err)
	}

	h.logger.Info("tournament start confirmed", 
		"tournament_id", tournamentID.String(),
		"status", tournament.Status,
	)

	// Publish TournamentStarted event (the fact, not the intent)
	event := pubsub.DomainEvent{
		EventID:       uuid.New().String(),
		EventType:     "TournamentStarted",
		AggregateID:   tournamentID.String(),
		AggregateType: "Tournament",
		Timestamp:     time.Now(),
		Payload: map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"started_at":    time.Now().Format(time.RFC3339),
		},
		Metadata: map[string]string{
			"correlation_id": metadata["correlation_id"],
		},
	}

	if err := h.publisher.Publish(ctx, pubsub.ChannelTournamentStarted, event); err != nil {
		h.logger.Error("failed to publish TournamentStarted event", 
			"tournament_id", tournamentID.String(),
			"error", err,
		)
		// Don't fail the handler if event publishing fails
	} else {
		h.logger.Info("published TournamentStarted event", 
			"tournament_id", tournamentID.String(),
		)
	}

	return nil
}

// BracketGenerationFailedPayload represents the payload when bracket generation fails
type BracketGenerationFailedPayload struct {
	TournamentID string `json:"tournament_id"`
	Error        string `json:"error"`
	Reason       string `json:"reason"`
}

// HandleBracketGenerationFailed processes BracketGenerationFailed events from matchmaking
// This is a compensating action for the tournament start saga
// Transitions tournament status from "starting" → "registration_open" (rollback)
func (h *EventHandler) HandleBracketGenerationFailed(ctx context.Context, eventType string, payload map[string]interface{}, metadata map[string]string) error {
	h.logger.Warn("received BracketGenerationFailed event", 
		"event_type", eventType,
		"correlation_id", metadata["correlation_id"],
	)

	// Parse payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var eventPayload BracketGenerationFailedPayload
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

	// Get tournament
	tournament, err := h.tournamentRepo.GetByID(ctx, tournamentID)
	if err != nil {
		h.logger.Error("failed to get tournament for rollback", 
			"tournament_id", tournamentID.String(),
			"error", err,
		)
		return fmt.Errorf("get tournament: %w", err)
	}

	// Rollback start (transitions "starting" → "registration_open")
	if err := tournament.RollbackStart(); err != nil {
		h.logger.Warn("failed to rollback start", 
			"tournament_id", tournamentID.String(),
			"current_status", tournament.Status,
			"error", err,
		)
		return fmt.Errorf("rollback start: %w", err)
	}

	// Save tournament state
	if _, err := h.tournamentRepo.Save(ctx, tournament); err != nil {
		h.logger.Error("failed to save tournament after rollback", 
			"tournament_id", tournamentID.String(),
			"error", err,
		)
		return fmt.Errorf("save tournament: %w", err)
	}

	h.logger.Info("tournament start rolled back", 
		"tournament_id", tournamentID.String(),
		"status", tournament.Status,
		"reason", eventPayload.Reason,
	)

	// Publish TournamentStartFailed event
	event := pubsub.DomainEvent{
		EventID:       uuid.New().String(),
		EventType:     "TournamentStartFailed",
		AggregateID:   tournamentID.String(),
		AggregateType: "Tournament",
		Timestamp:     time.Now(),
		Payload: map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         eventPayload.Error,
			"reason":        eventPayload.Reason,
			"failed_at":     time.Now().Format(time.RFC3339),
		},
		Metadata: map[string]string{
			"correlation_id": metadata["correlation_id"],
		},
	}

	if err := h.publisher.Publish(ctx, pubsub.ChannelTournamentStartFailed, event); err != nil {
		h.logger.Error("failed to publish TournamentStartFailed event", 
			"tournament_id", tournamentID.String(),
			"error", err,
		)
		// Don't fail the handler if event publishing fails
	} else {
		h.logger.Info("published TournamentStartFailed event", 
			"tournament_id", tournamentID.String(),
		)
	}

	return nil
}

// RegisterHandlers registers all event handlers with the subscriber
func (h *EventHandler) RegisterHandlers(subscriber *pubsub.EventSubscriber) {
	subscriber.Subscribe("GracePeriodStarted", h.HandleGracePeriodStarted)
	subscriber.Subscribe("BracketGenerationFailed", h.HandleBracketGenerationFailed)
	h.logger.Info("registered event handlers for GracePeriodStarted, BracketGenerationFailed")
}


