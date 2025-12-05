package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/winspire-core/services/matchmaking/internal/domain"
)

// EventPublisher publishes domain events to Redis channels
type EventPublisher struct {
	redisClient *redis.Client
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(redisClient *redis.Client) *EventPublisher {
	return &EventPublisher{
		redisClient: redisClient,
	}
}

// Publish publishes a domain event to its Redis channel
func (p *EventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	// Generate channel name from event type
	channel := ChannelName(event.BoundedContext(), event.EventType())

	// Serialize event to JSON
	payload, err := json.Marshal(map[string]interface{}{
		"event_id":        event.EventID(),
		"event_type":      event.EventType(),
		"aggregate_id":    event.AggregateID(),
		"aggregate_type":  event.AggregateType(),
		"bounded_context": event.BoundedContext(),
		"timestamp":       event.Timestamp(),
		"payload":         event.Payload(),
		"metadata":        event.Metadata(),
	})
	if err != nil {
		return fmt.Errorf("marshal event %s: %w", event.EventType(), err)
	}

	// Publish to Redis channel
	if err := p.redisClient.Publish(ctx, channel, payload).Err(); err != nil {
		return fmt.Errorf("publish event %s to channel %s: %w", event.EventType(), channel, err)
	}

	log.Printf("[EventPublisher] Published %s to channel %s (event_id=%s)", 
		event.EventType(), channel, event.EventID())

	return nil
}

// PublishBatch publishes multiple events in a pipeline for better performance
func (p *EventPublisher) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	pipe := p.redisClient.Pipeline()

	for _, event := range events {
		channel := ChannelName(event.BoundedContext(), event.EventType())

		payload, err := json.Marshal(map[string]interface{}{
			"event_id":        event.EventID(),
			"event_type":      event.EventType(),
			"aggregate_id":    event.AggregateID(),
			"aggregate_type":  event.AggregateType(),
			"bounded_context": event.BoundedContext(),
			"timestamp":       event.Timestamp(),
			"payload":         event.Payload(),
			"metadata":        event.Metadata(),
		})
		if err != nil {
			return fmt.Errorf("marshal event %s: %w", event.EventType(), err)
		}

		pipe.Publish(ctx, channel, payload)
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("execute publish pipeline: %w", err)
	}

	log.Printf("[EventPublisher] Published batch of %d events", len(events))

	return nil
}

// Close closes the Redis client connection
func (p *EventPublisher) Close() error {
	if p.redisClient != nil {
		return p.redisClient.Close()
	}
	return nil
}

// Health checks Redis connection health
func (p *EventPublisher) Health(ctx context.Context) error {
	if err := p.redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}
	return nil
}



