package projections

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const upsertCupHostViewSQL = `
INSERT INTO cup_host_views (
	cup_id,
	competition_context_id,
	stage_statuses,
	attendance_counts,
	dependency_health,
	updated_at
) VALUES ($1,$2,$3,$4,$5,NOW())
ON CONFLICT (cup_id) DO UPDATE
SET stage_statuses    = EXCLUDED.stage_statuses,
	attendance_counts = EXCLUDED.attendance_counts,
	dependency_health = EXCLUDED.dependency_health,
	competition_context_id = EXCLUDED.competition_context_id,
	updated_at       = NOW();
`

// CupProjector persists host-facing cup views.
type CupProjector struct {
	pool *pgxpool.Pool
}

func NewCupProjector(pool *pgxpool.Pool) *CupProjector {
	return &CupProjector{pool: pool}
}

func (p *CupProjector) Upsert(ctx context.Context, view CupHostView) error {
	if err := validateJSON(view.StageStatuses); err != nil {
		return fmt.Errorf("stageStatuses: %w", err)
	}
	if err := validateJSON(view.AttendanceCounts); err != nil {
		return fmt.Errorf("attendanceCounts: %w", err)
	}
	if err := validateJSON(view.DependencyHealth); err != nil {
		return fmt.Errorf("dependencyHealth: %w", err)
	}

	_, err := p.pool.Exec(ctx, upsertCupHostViewSQL,
		view.CupID,
		view.CompetitionContextID,
		view.StageStatuses,
		view.AttendanceCounts,
		view.DependencyHealth,
	)
	return err
}
