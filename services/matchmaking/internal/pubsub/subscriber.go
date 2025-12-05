package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// EventHandler is a function that handles an incoming event
type EventHandler func(ctx context.Context, eventType string, payload map[string]interface{}, metadata map[string]string) error

// EventSubscriber subscribes to Redis channels and routes events to handlers
type EventSubscriber struct {
	redisClient *redis.Client
	handlers    map[string][]EventHandler
}

// NewEventSubscriber creates a new event subscriber
func NewEventSubscriber(redisClient *redis.Client) *EventSubscriber {
	return &EventSubscriber{
		redisClient: redisClient,
		handlers:    make(map[string][]EventHandler),
	}
}

// Subscribe registers a handler for a specific event type
func (s *EventSubscriber) Subscribe(eventType string, handler EventHandler) {
	if s.handlers[eventType] == nil {
		s.handlers[eventType] = []EventHandler{}
	}
	s.handlers[eventType] = append(s.handlers[eventType], handler)
	log.Printf("[EventSubscriber] Registered handler for event type: %s", eventType)
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

	log.Printf("[EventSubscriber] Subscribed to channels: %v", channels)

	// Start consuming messages
	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg == nil {
				log.Println("[EventSubscriber] Channel closed, stopping subscriber")
				return nil
			}

			if err := s.handleMessage(ctx, msg); err != nil {
				log.Printf("[EventSubscriber] ERROR handling message: %v", err)
				// Continue processing other messages
			}

		case <-ctx.Done():
			log.Println("[EventSubscriber] Context cancelled, stopping subscriber")
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

	log.Printf("[EventSubscriber] Received %s from channel %s (event_id=%s)", 
		event.EventType, msg.Channel, event.EventID)

	// Find and execute handlers for this event type
	handlers, exists := s.handlers[event.EventType]
	if !exists || len(handlers) == 0 {
		log.Printf("[EventSubscriber] WARN: No handlers registered for event type %s", event.EventType)
		return nil
	}

	// Execute all registered handlers
	for i, handler := range handlers {
		if err := handler(ctx, event.EventType, event.Payload, event.Metadata); err != nil {
			log.Printf("[EventSubscriber] ERROR: Handler %d for %s failed: %v", i, event.EventType, err)
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


