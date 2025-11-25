package ssebroker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/r3labs/sse/v2"
)

// Scope uniquely identifies a stream of events (cup/tournament/match).
type Scope struct {
	Type string
	ID   uuid.UUID
}

// Key returns the stream identifier used by the SSE server.
func (s Scope) Key() string {
	return fmt.Sprintf("%s:%s", s.Type, s.ID.String())
}

// Broker multiplexes SSE streams across scopes.
type Broker struct {
	server *sse.Server
}

// NewBroker builds an SSE server with bounded history.
func NewBroker(poolSize int) *Broker {
	srv := sse.New()
	srv.AutoReplay = false
	srv.BufferSize = poolSize
	return &Broker{server: srv}
}

// Server exposes the underlying r3labs server for HTTP handlers.
func (b *Broker) Server() *sse.Server {
	return b.server
}

// Publish pushes an event payload to subscribers of the given scope.
func (b *Broker) Publish(ctx context.Context, scope Scope, eventType string, data []byte) error {
	message := &sse.Event{
		Event: []byte(eventType),
		Data:  data,
	}
	b.server.Publish(scope.Key(), message)
	return nil
}

// Close shuts down the broker.
func (b *Broker) Close() {
	if b.server != nil {
		b.server.Close()
	}
}
