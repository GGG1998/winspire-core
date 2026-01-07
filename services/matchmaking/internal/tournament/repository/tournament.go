package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgtypeconv "github.com/winspire/winspire-core/libs/go/pgtype"

	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
	"github.com/winspire-core/services/matchmaking/internal/tournament/domain"
)

// TournamentRepository wraps sqlc-generated queries for tournaments.
type TournamentRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// Errors
var (
	ErrTournamentNotFound = errors.New("tournament not found")
)

// NewTournamentRepository creates a new TournamentRepository.
func NewTournamentRepository(pool *pgxpool.Pool) *TournamentRepository {
	return &TournamentRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// GetByID retrieves a tournament by ID.
func (r *TournamentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tournament, error) {
	row, err := r.queries.GetTournamentByID(ctx, pgtypeconv.UUIDToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTournamentNotFound
		}
		return nil, fmt.Errorf("get tournament: %w", err)
	}

	return sqlcTournamentToDomain(row), nil
}

// GetByHostAndID retrieves a tournament ensuring it belongs to the provided host.
func (r *TournamentRepository) GetByHostAndID(ctx context.Context, hostID, tournamentID uuid.UUID) (*domain.Tournament, error) {
	params := sqlc.GetTournamentByHostAndIDParams{
		HostID: pgtypeconv.UUIDToPgtype(hostID),
		ID:     pgtypeconv.UUIDToPgtype(tournamentID),
	}

	row, err := r.queries.GetTournamentByHostAndID(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTournamentNotFound
		}
		return nil, fmt.Errorf("get tournament: %w", err)
	}

	return sqlcTournamentToDomain(row), nil
}

// ListByHost retrieves all tournaments for a specific host.
func (r *TournamentRepository) ListByHost(ctx context.Context, hostID uuid.UUID) ([]*domain.Tournament, error) {
	rows, err := r.queries.ListTournamentsByHostID(ctx, pgtypeconv.UUIDToPgtype(hostID))
	if err != nil {
		return nil, fmt.Errorf("list tournaments: %w", err)
	}

	tournaments := make([]*domain.Tournament, len(rows))
	for i, row := range rows {
		tournaments[i] = sqlcTournamentToDomain(row)
	}

	return tournaments, nil
}

// ListByStatus retrieves tournaments by status for scheduler.
func (r *TournamentRepository) ListByStatus(ctx context.Context, statuses []string) ([]*domain.Tournament, error) {
	rows, err := r.queries.ListTournamentsByStatus(ctx, statuses)
	if err != nil {
		return nil, fmt.Errorf("list tournaments by status: %w", err)
	}

	tournaments := make([]*domain.Tournament, len(rows))
	for i, row := range rows {
		tournaments[i] = sqlcTournamentToDomain(row)
	}

	return tournaments, nil
}

// CreateParams contains parameters for creating a tournament.
type CreateParams struct {
	HostID               uuid.UUID
	Name                 string
	Description          *string
	ExternalID           *string
	ScheduledStartTimeAt *time.Time
	MinimumTeamCount     *int32
	MaximumTeamCount     *int32
	TeamSize             int32
	AutoForceReady       *bool
	GameID               *uuid.UUID
	SpaceID              *uuid.UUID
	TemplateID           *uuid.UUID
	GameSnapshot         *domain.GameSnapshot
}

// Create creates a new tournament.
func (r *TournamentRepository) Create(ctx context.Context, params CreateParams) (*domain.Tournament, error) {
	var gameSnapshotJSON json.RawMessage
	if params.GameSnapshot != nil {
		data, err := json.Marshal(params.GameSnapshot)
		if err != nil {
			return nil, fmt.Errorf("marshal game snapshot: %w", err)
		}
		gameSnapshotJSON = data
	}

	sqlcParams := sqlc.CreateTournamentParams{
		HostID:                   pgtypeconv.UUIDToPgtype(params.HostID),
		Name:                     params.Name,
		Description:              pgtypeconv.StringPtrToPgtype(params.Description),
		ExternalID:               pgtypeconv.StringPtrToPgtype(params.ExternalID),
		Status:                   domain.StatusDraft,
		ScheduledStartTimeAt:     pgtypeconv.TimePtrToPgtypeTimestamptz(params.ScheduledStartTimeAt),
		RegistrationWindowOpenAt: pgtypeconv.TimePtrToPgtypeTimestamptz(nil),
		MinimumTeamCount:         pgtypeconv.Int32PtrToPgtype(params.MinimumTeamCount),
		MaximumTeamCount:         pgtypeconv.Int32PtrToPgtype(params.MaximumTeamCount),
		TeamSize:                 params.TeamSize,
		AutoForceReady:           pgtypeconv.BoolPtrToPgtype(params.AutoForceReady),
		GameID:                   pgtypeconv.UUIDPtrToPgtype(params.GameID),
		SpaceID:                  pgtypeconv.UUIDPtrToPgtype(params.SpaceID),
		TemplateID:               pgtypeconv.UUIDPtrToPgtype(params.TemplateID),
		GameSnapshot:             gameSnapshotJSON,
	}

	row, err := r.queries.CreateTournament(ctx, sqlcParams)
	if err != nil {
		return nil, fmt.Errorf("create tournament: %w", err)
	}

	return sqlcTournamentToDomain(row), nil
}

// Save persists the tournament aggregate state to the database.
func (r *TournamentRepository) Save(ctx context.Context, tournament *domain.Tournament) (*domain.Tournament, error) {
	var readyWindowJSON, prizeJSON, gameSnapshotJSON json.RawMessage

	if tournament.ReadyWindow != nil {
		data, err := json.Marshal(tournament.ReadyWindow)
		if err != nil {
			return nil, fmt.Errorf("marshal ready window: %w", err)
		}
		readyWindowJSON = data
	}

	if tournament.Prize != nil {
		data, err := json.Marshal(tournament.Prize)
		if err != nil {
			return nil, fmt.Errorf("marshal prize: %w", err)
		}
		prizeJSON = data
	}

	if tournament.GameSnapshot != nil {
		data, err := json.Marshal(tournament.GameSnapshot)
		if err != nil {
			return nil, fmt.Errorf("marshal game snapshot: %w", err)
		}
		gameSnapshotJSON = data
	}

	sqlcParams := sqlc.SaveTournamentParams{
		ID:                   pgtypeconv.UUIDToPgtype(tournament.ID),
		Name:                 tournament.Name,
		Description:          pgtypeconv.StringPtrToPgtype(tournament.Description),
		Status:               tournament.Status,
		ScheduledStartTimeAt: pgtypeconv.TimePtrToPgtypeTimestamptz(tournament.ScheduledStartTimeAt),
		ActualStartTimeAt:    pgtypeconv.TimePtrToPgtypeTimestamptz(tournament.ActualStartTimeAt),
		CompletedAt:          pgtypeconv.TimePtrToPgtypeTimestamptz(tournament.CompletedAt),
		CancelledAt:          pgtypeconv.TimePtrToPgtypeTimestamptz(tournament.CancelledAt),
		MinimumTeamCount:     pgtypeconv.Int32ToPgtype(tournament.MinimumTeamCount),
		MaximumTeamCount:     pgtypeconv.Int32PtrToPgtype(tournament.MaximumTeamCount),
		ReadyWindow:          readyWindowJSON,
		Prize:                prizeJSON,
		GameSnapshot:         gameSnapshotJSON,
	}

	row, err := r.queries.SaveTournament(ctx, sqlcParams)
	if err != nil {
		return nil, fmt.Errorf("save tournament: %w", err)
	}

	return sqlcTournamentToDomain(row), nil
}

// UpdateStatus updates tournament status.
func (r *TournamentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	params := sqlc.UpdateTournamentStatusParams{
		ID:     pgtypeconv.UUIDToPgtype(id),
		Status: status,
	}
	return r.queries.UpdateTournamentStatus(ctx, params)
}

// sqlcTournamentToDomain converts SQLC tournament to domain tournament.
func sqlcTournamentToDomain(t sqlc.TournamentTournament) *domain.Tournament {
	tournament := &domain.Tournament{
		ID:                       pgtypeconv.PgtypeToUUID(t.ID),
		HostID:                   pgtypeconv.PgtypeToUUID(t.HostID),
		Name:                     t.Name,
		Description:              pgtypeconv.PgtypeToStringPtr(t.Description),
		ExternalID:               pgtypeconv.PgtypeToStringPtr(t.ExternalID),
		Status:                   t.Status,
		ScheduledStartTimeAt:     pgtypeconv.PgtypeTimestamptzToTimePtr(t.ScheduledStartTimeAt),
		RegistrationWindowOpenAt: pgtypeconv.PgtypeTimestamptzToTimePtr(t.RegistrationWindowOpenAt),
		ActualStartTimeAt:        pgtypeconv.PgtypeTimestamptzToTimePtr(t.ActualStartTimeAt),
		CompletedAt:              pgtypeconv.PgtypeTimestamptzToTimePtr(t.CompletedAt),
		CancelledAt:              pgtypeconv.PgtypeTimestamptzToTimePtr(t.CancelledAt),
		MinimumTeamCount:         pgtypeconv.PgtypeToInt32(t.MinimumTeamCount),
		MaximumTeamCount:         pgtypeconv.PgtypeToInt32Ptr(t.MaximumTeamCount),
		TeamSize:                 t.TeamSize,
		AutoForceReady:           t.AutoForceReady.Bool,
		GameID:                   pgtypeconv.PgtypeToUUIDPtr(t.GameID),
		SpaceID:                  pgtypeconv.PgtypeToUUIDPtr(t.SpaceID),
		TemplateID:               pgtypeconv.PgtypeToUUIDPtr(t.TemplateID),
		CreatedAt:                t.CreatedAt.Time,
		UpdatedAt:                t.UpdatedAt.Time,
	}

	// Parse JSONB fields
	if len(t.ReadyWindow) > 0 {
		var rw domain.ReadyWindow
		if err := json.Unmarshal(t.ReadyWindow, &rw); err == nil {
			tournament.ReadyWindow = &rw
		}
	}

	if len(t.Prize) > 0 {
		var prize domain.Prize
		if err := json.Unmarshal(t.Prize, &prize); err == nil {
			tournament.Prize = &prize
		}
	}

	if len(t.GameSnapshot) > 0 {
		var gs domain.GameSnapshot
		if err := json.Unmarshal(t.GameSnapshot, &gs); err == nil {
			tournament.GameSnapshot = &gs
		}
	}

	return tournament
}
