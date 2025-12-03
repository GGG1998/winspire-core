package ssebroker

import (
	"context"

	"github.com/r3labs/sse/v2"
)

// Publisher defines the interface for publishing SSE events
// This allows swapping between in-memory and Redis-backed brokers
type Publisher interface {
	// Publish pushes an event payload to subscribers of the given scope
	Publish(ctx context.Context, scope Scope, eventType string, data []byte) error

	// Server exposes the underlying SSE server for HTTP handlers
	Server() *sse.Server

	// Close shuts down the broker
	Close()
}

// Ensure both implementations satisfy the interface
var (
	_ Publisher = (*Broker)(nil)
	_ Publisher = (*RedisBroker)(nil)
)
