package ssebroker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Registry manages host_subscriptions leasing so SSE connections stay stateless.
type Registry struct {
	pool *pgxpool.Pool
	ttl  time.Duration
}

// NewRegistry builds a subscription registry backed by Postgres.
func NewRegistry(pool *pgxpool.Pool, ttl time.Duration) *Registry {
	return &Registry{pool: pool, ttl: ttl}
}

// Lease upserts a host subscription and returns the subscription identifier.
func (r *Registry) Lease(ctx context.Context, hostID uuid.UUID, scopeType string, scopeID uuid.UUID, lastEventID int64) (uuid.UUID, error) {
	if r == nil || r.pool == nil {
		return uuid.Nil, fmt.Errorf("registry is not configured")
	}
	subscriptionID := uuid.New()
	query := `
INSERT INTO host_subscriptions (
	subscription_id,
	host_id,
	scope_type,
	scope_id,
	last_delivered_event_id,
	leased_at
) VALUES ($1,$2,$3,$4,$5,NOW())
ON CONFLICT (host_id, scope_type, scope_id)
DO UPDATE SET
	subscription_id = EXCLUDED.subscription_id,
	last_delivered_event_id = EXCLUDED.last_delivered_event_id,
	leased_at = NOW();`
	if _, err := r.pool.Exec(ctx, query, subscriptionID, hostID, scopeType, scopeID, lastEventID); err != nil {
		return uuid.Nil, fmt.Errorf("lease subscription: %w", err)
	}
	return subscriptionID, nil
}

// ReleaseExpired removes subscriptions whose leases exceeded TTL.
func (r *Registry) ReleaseExpired(ctx context.Context) error {
	if r == nil || r.pool == nil {
		return nil
	}
	query := `
DELETE FROM host_subscriptions
WHERE leased_at < NOW() - ($1 * INTERVAL '1 second');
`
	_, err := r.pool.Exec(ctx, query, r.ttl.Seconds())
	return err
}





