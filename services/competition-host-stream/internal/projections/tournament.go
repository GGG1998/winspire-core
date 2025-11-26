package projections

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const upsertTournamentHostViewSQL = `
INSERT INTO tournament_host_views (
	tournament_id,
	cup_id,
	settings_hash,
	lineup_status,
	seeding_window,
	match_gate,
	updated_at
) VALUES ($1,$2,$3,$4,$5,$6,NOW())
ON CONFLICT (tournament_id) DO UPDATE
SET cup_id        = EXCLUDED.cup_id,
	settings_hash = EXCLUDED.settings_hash,
	lineup_status = EXCLUDED.lineup_status,
	seeding_window = EXCLUDED.seeding_window,
	match_gate    = EXCLUDED.match_gate,
	updated_at    = NOW();
`

// TournamentProjector writes tournament host rows.
type TournamentProjector struct {
	pool *pgxpool.Pool
}

func NewTournamentProjector(pool *pgxpool.Pool) *TournamentProjector {
	return &TournamentProjector{pool: pool}
}

func (p *TournamentProjector) Upsert(ctx context.Context, view TournamentHostView) error {
	if view.SettingsHash == "" {
		return fmt.Errorf("settingsHash required")
	}
	if err := validateJSON(view.LineupStatus); err != nil {
		return fmt.Errorf("lineupStatus: %w", err)
	}
	if err := validateJSON(view.MatchGate); err != nil {
		return fmt.Errorf("matchGate: %w", err)
	}

	var cupID interface{}
	if view.CupID != nil {
		cupID = *view.CupID
	}

	_, err := p.pool.Exec(ctx, upsertTournamentHostViewSQL,
		view.TournamentID,
		cupID,
		view.SettingsHash,
		view.LineupStatus,
		view.SeedingWindow,
		view.MatchGate,
	)
	return err
}

func PtrUUID(id uuid.UUID) *uuid.UUID {
	return &id
}

