// Package pubsub handles Redis Pub/Sub for domain events
package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// EventPublisher publishes domain events to Redis
type EventPublisher struct {
	client *redis.Client
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(client *redis.Client) *EventPublisher {
	return &EventPublisher{
		client: client,
	}
}

// DomainEvent represents a domain event structure
type DomainEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Timestamp     time.Time              `json:"timestamp"`
	Payload       map[string]interface{} `json:"payload"`
	Metadata      map[string]string      `json:"metadata"`
}

// Publish publishes a domain event to Redis
func (p *EventPublisher) Publish(ctx context.Context, channel string, event DomainEvent) error {
	// Add timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Serialize event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	// Publish to Redis channel
	if err := p.client.Publish(ctx, channel, eventJSON).Err(); err != nil {
		return fmt.Errorf("publish to Redis: %w", err)
	}

	return nil
}











