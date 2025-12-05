package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

// EventHandler is a function that handles an incoming event
type EventHandler func(ctx context.Context, eventType string, payload map[string]interface{}, metadata map[string]string) error

// EventSubscriber subscribes to Redis channels and routes events to handlers
type EventSubscriber struct {
	redisClient *redis.Client
	logger      *slog.Logger
	handlers    map[string][]EventHandler
}

// NewEventSubscriber creates a new event subscriber
func NewEventSubscriber(redisClient *redis.Client, logger *slog.Logger) *EventSubscriber {
	return &EventSubscriber{
		redisClient: redisClient,
		logger:      logger,
		handlers:    make(map[string][]EventHandler),
	}
}

// Subscribe registers a handler for a specific event type
func (s *EventSubscriber) Subscribe(eventType string, handler EventHandler) {
	if s.handlers[eventType] == nil {
		s.handlers[eventType] = []EventHandler{}
	}
	s.handlers[eventType] = append(s.handlers[eventType], handler)
	s.logger.Info("registered event handler", "event_type", eventType)
}

// Start begins listening to Redis channels
func (s *EventSubscriber) Start(ctx context.Context, channels []string) error {
	if len(channels) == 0 {
		return fmt.Errorf("no channels to subscribe to")
	}

	// Subscribe to Redis Pub/Sub
	pubsub := s.redisClient.Subscribe(ctx, channels...)
	defer pubsub.Close()

	// Wait for subscription confirmation
	if _, err := pubsub.Receive(ctx); err != nil {
		return fmt.Errorf("subscribe to channels: %w", err)
	}

	s.logger.Info("subscribed to event channels", "channels", channels)

	// Start consuming messages
	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg == nil {
				s.logger.Info("channel closed, stopping subscriber")
				return nil
			}

			if err := s.handleMessage(ctx, msg); err != nil {
				s.logger.Error("error handling message", "error", err)
				// Continue processing other messages
			}

		case <-ctx.Done():
			s.logger.Info("context cancelled, stopping subscriber")
			return ctx.Err()
		}
	}
}

// handleMessage processes a single Redis message
func (s *EventSubscriber) handleMessage(ctx context.Context, msg *redis.Message) error {
	// Parse event from JSON
	var event struct {
		EventID        string                 `json:"event_id"`
		EventType      string                 `json:"event_type"`
		AggregateID    string                 `json:"aggregate_id"`
		AggregateType  string                 `json:"aggregate_type"`
		BoundedContext string                 `json:"bounded_context"`
		Timestamp      string                 `json:"timestamp"`
		Payload        map[string]interface{} `json:"payload"`
		Metadata       map[string]string      `json:"metadata"`
	}

	if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
		return fmt.Errorf("unmarshal event from channel %s: %w", msg.Channel, err)
	}

	s.logger.Info("received event", 
		"event_type", event.EventType,
		"channel", msg.Channel,
		"event_id", event.EventID,
	)

	// Find and execute handlers for this event type
	handlers, exists := s.handlers[event.EventType]
	if !exists || len(handlers) == 0 {
		s.logger.Warn("no handlers registered for event type", "event_type", event.EventType)
		return nil
	}

	// Execute all registered handlers
	for i, handler := range handlers {
		if err := handler(ctx, event.EventType, event.Payload, event.Metadata); err != nil {
			s.logger.Error("handler failed", 
				"handler_index", i,
				"event_type", event.EventType,
				"error", err,
			)
			// Continue executing other handlers
		}
	}

	return nil
}

// Close closes the subscriber
func (s *EventSubscriber) Close() error {
	if s.redisClient != nil {
		return s.redisClient.Close()
	}
	return nil
}


