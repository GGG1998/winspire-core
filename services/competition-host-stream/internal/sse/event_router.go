package ssebroker

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// EventRouter publishes well-known domain events to SSE streams.
type EventRouter struct {
	broker *Broker
}

func NewEventRouter(b *Broker) *EventRouter {
	return &EventRouter{broker: b}
}

func (r *EventRouter) CupUpdated(ctx context.Context, cupID uuid.UUID, payload any) {
	r.publish(ctx, Scope{Type: "cup", ID: cupID}, "CupHostViewUpdated", payload)
}

func (r *EventRouter) MatchmakingQueueUpdated(ctx context.Context, tournamentID uuid.UUID, payload any) {
	r.publish(ctx, Scope{Type: "tournament", ID: tournamentID}, "MatchmakingQueueUpdated", payload)
}

func (r *EventRouter) TournamentParticipationUpdate(ctx context.Context, tournamentID uuid.UUID, payload any) {
	r.publish(ctx, Scope{Type: "tournament", ID: tournamentID}, "TournamentParticipationUpdate", payload)
}

func (r *EventRouter) publish(ctx context.Context, scope Scope, event string, payload any) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_ = r.broker.Publish(ctx, scope, event, buf)
}
