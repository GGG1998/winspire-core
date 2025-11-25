package projections

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	ssebroker "github.com/winspire/competition-host-stream/internal/sse"
)

const upsertMatchLobbySQL = `
INSERT INTO match_lobby_views (
	match_id,
	tournament_id,
	lobby_information,
	queue_state,
	updated_at
) VALUES ($1,$2,$3,$4,NOW())
ON CONFLICT (match_id) DO UPDATE
SET tournament_id    = EXCLUDED.tournament_id,
	lobby_information = EXCLUDED.lobby_information,
	queue_state       = EXCLUDED.queue_state,
	updated_at        = NOW();
`

type MatchProjector struct {
	pool        *pgxpool.Pool
	eventRouter *ssebroker.EventRouter
}

func NewMatchProjector(pool *pgxpool.Pool, eventRouter *ssebroker.EventRouter) *MatchProjector {
	return &MatchProjector{
		pool:        pool,
		eventRouter: eventRouter,
	}
}

func (p *MatchProjector) Upsert(ctx context.Context, view MatchLobbyView) error {
	if err := validateJSON(view.LobbyInformation); err != nil {
		return fmt.Errorf("lobbyInformation: %w", err)
	}
	if err := validateJSON(view.QueueState); err != nil {
		return fmt.Errorf("queueState: %w", err)
	}
	_, err := p.pool.Exec(ctx, upsertMatchLobbySQL,
		view.MatchID,
		view.TournamentID,
		view.LobbyInformation,
		view.QueueState,
	)
	if err != nil {
		return err
	}

	// Route MatchmakingQueueUpdated event through SSE broker
	if p.eventRouter != nil {
		payload := map[string]any{
			"matchId":          view.MatchID,
			"tournamentId":     view.TournamentID,
			"lobbyInformation": json.RawMessage(view.LobbyInformation),
			"queueState":       json.RawMessage(view.QueueState),
		}
		p.eventRouter.MatchmakingQueueUpdated(ctx, view.TournamentID, payload)
	}

	return nil
}
