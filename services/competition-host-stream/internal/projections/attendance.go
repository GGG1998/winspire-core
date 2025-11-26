package projections

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const upsertAttendanceSQL = `
INSERT INTO attendance_snapshots (
	scope_type,
	scope_id,
	total_joined,
	total_confirmed,
	restrictions_breached,
	last_event_id,
	updated_at
) VALUES ($1,$2,$3,$4,$5,$6,NOW())
ON CONFLICT (scope_type, scope_id) DO UPDATE
SET total_joined          = EXCLUDED.total_joined,
	total_confirmed       = EXCLUDED.total_confirmed,
	restrictions_breached = EXCLUDED.restrictions_breached,
	last_event_id         = EXCLUDED.last_event_id,
	updated_at            = NOW();
`

// AttendanceProjector keeps running participation stats for hosts.
type AttendanceProjector struct {
	pool *pgxpool.Pool
}

func NewAttendanceProjector(pool *pgxpool.Pool) *AttendanceProjector {
	return &AttendanceProjector{pool: pool}
}

func (p *AttendanceProjector) Upsert(ctx context.Context, snapshot AttendanceSnapshot) error {
	if snapshot.ScopeType == "" {
		return fmt.Errorf("scopeType required")
	}
	if snapshot.TotalConfirmed > snapshot.TotalJoined {
		return fmt.Errorf("totalConfirmed exceeds totalJoined")
	}
	if err := validateJSON(snapshot.RestrictionsBreached); err != nil {
		return fmt.Errorf("restrictions: %w", err)
	}

	_, err := p.pool.Exec(ctx, upsertAttendanceSQL,
		snapshot.ScopeType,
		snapshot.ScopeID,
		snapshot.TotalJoined,
		snapshot.TotalConfirmed,
		snapshot.RestrictionsBreached,
		snapshot.LastEventID,
	)
	return err
}


