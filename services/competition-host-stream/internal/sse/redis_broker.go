package ssebroker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/r3labs/sse/v2"
	"github.com/redis/go-redis/v9"
)

// RedisBroker multiplexes SSE streams across scopes using Redis Pub/Sub for horizontal scaling
type RedisBroker struct {
	server       *sse.Server
	redisClient  *redis.Client
	logger       *slog.Logger
	pubsubCancel context.CancelFunc
	wg           sync.WaitGroup

	// Metrics
	mu                sync.RWMutex
	messagesPublished int64
	messagesReceived  int64
	activeSubscribers int64
}

// RedisMessage represents a message published to Redis
type RedisMessage struct {
	ScopeKey  string `json:"scope_key"`
	EventType string `json:"event_type"`
	Data      []byte `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

// NewRedisBroker builds an SSE server backed by Redis Pub/Sub for horizontal scaling
func NewRedisBroker(poolSize int, redisClient *redis.Client, logger *slog.Logger) (*RedisBroker, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	srv := sse.New()
	srv.AutoReplay = false
	srv.BufferSize = poolSize

	broker := &RedisBroker{
		server:      srv,
		redisClient: redisClient,
		logger:      logger.With("component", "redis_broker"),
	}

	// Start Redis Pub/Sub subscriber
	if err := broker.startSubscriber(); err != nil {
		return nil, fmt.Errorf("failed to start redis subscriber: %w", err)
	}

	return broker, nil
}

// Server exposes the underlying r3labs server for HTTP handlers
func (rb *RedisBroker) Server() *sse.Server {
	return rb.server
}

// Publish pushes an event payload to subscribers via Redis Pub/Sub
// This allows horizontal scaling as all instances receive the message
func (rb *RedisBroker) Publish(ctx context.Context, scope Scope, eventType string, data []byte) error {
	msg := RedisMessage{
		ScopeKey:  scope.Key(),
		EventType: eventType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish to Redis - channel name is "sse:events"
	channel := "sse:events"
	if err := rb.redisClient.Publish(ctx, channel, msgBytes).Err(); err != nil {
		rb.logger.Error("failed to publish to redis",
			"error", err,
			"scope", scope.Key(),
			"event_type", eventType,
		)
		return fmt.Errorf("failed to publish to redis: %w", err)
	}

	// Update metrics
	rb.mu.Lock()
	rb.messagesPublished++
	rb.mu.Unlock()

	return nil
}

// startSubscriber starts a Redis Pub/Sub subscriber that forwards messages to local SSE clients
func (rb *RedisBroker) startSubscriber() error {
	ctx, cancel := context.WithCancel(context.Background())
	rb.pubsubCancel = cancel

	// Subscribe to Redis channel
	pubsub := rb.redisClient.Subscribe(ctx, "sse:events")

	// Test subscription
	if _, err := pubsub.Receive(ctx); err != nil {
		cancel()
		return fmt.Errorf("failed to subscribe to redis: %w", err)
	}

	rb.logger.Info("redis pubsub subscriber started")

	// Start goroutine to handle messages
	rb.wg.Add(1)
	go func() {
		defer rb.wg.Done()
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				rb.logger.Info("redis pubsub subscriber stopping")
				return
			case msg, ok := <-ch:
				if !ok {
					rb.logger.Warn("redis pubsub channel closed")
					return
				}

				rb.handleRedisMessage(msg.Payload)
			}
		}
	}()

	return nil
}

// handleRedisMessage processes a message received from Redis and publishes to local SSE clients
func (rb *RedisBroker) handleRedisMessage(payload string) {
	var msg RedisMessage
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		rb.logger.Error("failed to unmarshal redis message", "error", err)
		return
	}

	// Create SSE event
	event := &sse.Event{
		Event: []byte(msg.EventType),
		Data:  msg.Data,
	}

	// Publish to local SSE server
	rb.server.Publish(msg.ScopeKey, event)

	// Update metrics
	rb.mu.Lock()
	rb.messagesReceived++
	rb.mu.Unlock()

	rb.logger.Debug("published sse event",
		"scope", msg.ScopeKey,
		"event_type", msg.EventType,
		"age_seconds", time.Now().Unix()-msg.Timestamp,
	)
}

// Close shuts down the broker and Redis subscriber
func (rb *RedisBroker) Close() {
	if rb.pubsubCancel != nil {
		rb.pubsubCancel()
	}

	// Wait for subscriber to finish
	rb.wg.Wait()

	if rb.server != nil {
		rb.server.Close()
	}

	rb.logger.Info("redis broker closed")
}

// Metrics returns current broker metrics (for monitoring/CloudWatch)
func (rb *RedisBroker) Metrics() map[string]int64 {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	return map[string]int64{
		"messages_published": rb.messagesPublished,
		"messages_received":  rb.messagesReceived,
		"active_subscribers": rb.activeSubscribers,
	}
}

// ConnectionCount returns the number of active SSE connections on this instance
func (rb *RedisBroker) ConnectionCount() int {
	// Count connections across all streams
	count := 0
	// Note: r3labs/sse doesn't expose connection count directly
	// We'll need to track this separately in the handler
	return count
}
