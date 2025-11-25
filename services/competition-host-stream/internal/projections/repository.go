package projections

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Reader fetches stored host projections.
type Reader struct {
	pool *pgxpool.Pool
}

func NewReader(pool *pgxpool.Pool) *Reader {
	return &Reader{pool: pool}
}

func (r *Reader) GetCup(ctx context.Context, cupID uuid.UUID) (map[string]any, error) {
	query := `SELECT cup_id, competition_context_id, stage_statuses, attendance_counts, dependency_health, updated_at FROM cup_host_views WHERE cup_id=$1`
	row := r.pool.QueryRow(ctx, query, cupID)

	var (
		id, competition               uuid.UUID
		stage, attendance, dependency []byte
		updatedAt                     pgtype.Timestamptz
	)
	if err := row.Scan(&id, &competition, &stage, &attendance, &dependency, &updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("cup %s not found", cupID)
		}
		return nil, err
	}

	payload := map[string]any{
		"cupId":                id,
		"competitionContextId": competition,
		"stageStatuses":        json.RawMessage(stage),
		"attendance":           json.RawMessage(attendance),
		"dependencies":         json.RawMessage(dependency),
	}
	if updatedAt.Valid {
		payload["updatedAt"] = updatedAt.Time
	}
	return payload, nil
}

func (r *Reader) GetTournament(ctx context.Context, tournamentID uuid.UUID) (map[string]any, error) {
	query := `SELECT tournament_id, cup_id, settings_hash, lineup_status, seeding_window, match_gate, updated_at FROM tournament_host_views WHERE tournament_id=$1`
	row := r.pool.QueryRow(ctx, query, tournamentID)

	var (
		id        uuid.UUID
		cupID     pgtype.UUID
		settings  string
		lineup    []byte
		window    pgtype.Text
		gate      []byte
		updatedAt pgtype.Timestamptz
	)

	if err := row.Scan(&id, &cupID, &settings, &lineup, &window, &gate, &updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tournament %s not found", tournamentID)
		}
		return nil, err
	}

	payload := map[string]any{
		"tournamentId": id,
		"settingsHash": settings,
		"lineupStatus": json.RawMessage(lineup),
		"matchGate":    json.RawMessage(gate),
	}
	if cupID.Valid {
		if cupUUID, err := uuid.FromBytes(cupID.Bytes[:]); err == nil {
			payload["cupId"] = cupUUID
		}
	}
	if window.Valid {
		payload["seedingWindow"] = window.String
	}
	if updatedAt.Valid {
		payload["updatedAt"] = updatedAt.Time
	}
	return payload, nil
}

func (r *Reader) ListMatches(ctx context.Context, tournamentID uuid.UUID) ([]map[string]any, error) {
	query := `SELECT match_id, tournament_id, lobby_information, queue_state, updated_at FROM match_lobby_views WHERE tournament_id=$1 ORDER BY updated_at DESC`
	rows, err := r.pool.Query(ctx, query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]any
	for rows.Next() {
		var (
			matchID      uuid.UUID
			tid          uuid.UUID
			lobby, queue []byte
			updatedAt    pgtype.Timestamptz
		)
		if err := rows.Scan(&matchID, &tid, &lobby, &queue, &updatedAt); err != nil {
			return nil, err
		}
		match := map[string]any{
			"matchId":          matchID,
			"tournamentId":     tid,
			"lobbyInformation": json.RawMessage(lobby),
			"queueState":       json.RawMessage(queue),
		}
		if updatedAt.Valid {
			match["updatedAt"] = updatedAt.Time
		}
		result = append(result, match)
	}
	return result, rows.Err()
}
