package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgtypeconv "github.com/winspire/winspire-core/libs/go/pgtype"

	"github.com/winspire/competition/internal/store/sqlc"
)

// TournamentRepository wraps sqlc-generated queries with business logic.
type TournamentRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// ErrTournamentNotFound is returned when a tournament cannot be located.
var ErrTournamentNotFound = errors.New("tournament not found")

// NewTournamentRepository creates a new TournamentRepository.
func NewTournamentRepository(pool *pgxpool.Pool) *TournamentRepository {
	return &TournamentRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// Tournament is the business domain model.
type Tournament struct {
	ID          uuid.UUID
	HostID      uuid.UUID
	Name        string
	Description *string
	ExternalID  *string
	Status      string

	// Timing
	ScheduledStartTimeAt     *time.Time
	RegistrationWindowOpenAt *time.Time
	ActualStartTimeAt        *time.Time
	CompletedAt              *time.Time
	CancelledAt              *time.Time

	// Configuration
	MinimumTeamCount int32
	MaximumTeamCount *int32
	TeamSize         int32
	AutoForceReady   bool

	// Optional references
	GameID     *uuid.UUID
	SpaceID    *uuid.UUID
	TemplateID *uuid.UUID

	// Rich data (stored as JSONB)
	ReadyWindow []byte // JSONB
	Prize       []byte // JSONB

	// Metadata
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TournamentListItem is a simplified view for list responses.
type TournamentListItem struct {
	ID                   uuid.UUID
	Name                 string
	Status               string
	ScheduledStartTimeAt *time.Time
	CreatedAt            time.Time
}

// ============================================================================
// Tournament Aggregate Business Methods (DDD)
// ============================================================================

// Domain errors for business rule violations
var (
	ErrInvalidStatusTransition    = errors.New("invalid status transition")
	ErrCannotEditCompleted        = errors.New("cannot edit completed tournament")
	ErrCannotEditCancelled        = errors.New("cannot edit cancelled tournament")
	ErrCannotCancelCompleted      = errors.New("cannot cancel completed tournament")
	ErrMustBeOpen                 = errors.New("tournament must be in registration_open or registration_closed status to start")
	ErrMustBeDraft                = errors.New("tournament must be in draft status to publish")
	ErrMustBeScheduled            = errors.New("tournament must be in scheduled status to open registration")
	ErrScheduledStartTimeRequired = errors.New("scheduled start time is required to publish tournament")
	ErrStartTimeInPast            = errors.New("cannot publish tournament with past start time")
)

// UpdateDetailsParams contains parameters for updating tournament details.
type UpdateDetailsParams struct {
	Name                 *string
	Description          *string
	ScheduledStartTimeAt *time.Time
	MinimumTeamCount     *int32
	MaximumTeamCount     *int32
}

// UpdateDetails updates tournament metadata while preserving state.
// Business rule: cannot edit completed or cancelled tournaments.
func (t *Tournament) UpdateDetails(params UpdateDetailsParams) error {
	// Enforce invariants
	if t.Status == "completed" {
		return ErrCannotEditCompleted
	}
	if t.Status == "cancelled" {
		return ErrCannotEditCancelled
	}

	// Apply updates
	if params.Name != nil {
		t.Name = *params.Name
	}
	if params.Description != nil {
		t.Description = params.Description
	}
	if params.ScheduledStartTimeAt != nil {
		t.ScheduledStartTimeAt = params.ScheduledStartTimeAt
	}
	if params.MinimumTeamCount != nil {
		t.MinimumTeamCount = *params.MinimumTeamCount
	}
	if params.MaximumTeamCount != nil {
		t.MaximumTeamCount = params.MaximumTeamCount
	}

	return nil
}

// Publish transitions tournament from draft to scheduled status.
// Business rule: tournament must be in draft status and have a future start time.
func (t *Tournament) Publish() error {
	if t.Status != "draft" {
		return fmt.Errorf("%w: cannot transition from %s to scheduled", ErrInvalidStatusTransition, t.Status)
	}

	// Validate scheduled start time exists
	if t.ScheduledStartTimeAt == nil {
		return ErrScheduledStartTimeRequired
	}

	// Validate scheduled start time is in the future
	now := time.Now()
	if !t.ScheduledStartTimeAt.After(now) {
		return ErrStartTimeInPast
	}

	t.Status = "scheduled"
	return nil
}

// OpenRegistration transitions tournament from scheduled to registration_open status.
// Business rule: tournament must be in scheduled status.
func (t *Tournament) OpenRegistration() error {
	if t.Status != "scheduled" {
		return fmt.Errorf("%w: cannot transition from %s to registration_open", ErrInvalidStatusTransition, t.Status)
	}

	t.Status = "registration_open"
	return nil
}

// Start transitions tournament to started state.
// Business rule: tournament must be in registration_open or registration_closed status.
func (t *Tournament) Start() error {
	if t.Status != "registration_open" && t.Status != "registration_closed" {
		return fmt.Errorf("%w: cannot transition from %s to started", ErrMustBeOpen, t.Status)
	}

	now := time.Now()
	t.Status = "started"
	t.ActualStartTimeAt = &now
	return nil
}

// Complete transitions tournament to completed state.
// Business rule: tournament must be in started status.
func (t *Tournament) Complete() error {
	if t.Status != "started" {
		return fmt.Errorf("%w: cannot transition from %s to completed", ErrInvalidStatusTransition, t.Status)
	}

	now := time.Now()
	t.Status = "completed"
	t.CompletedAt = &now
	return nil
}

// Cancel transitions tournament to cancelled state.
// Business rule: cannot cancel completed tournaments.
func (t *Tournament) Cancel() error {
	if t.Status == "completed" {
		return ErrCannotCancelCompleted
	}

	// Valid cancellation from draft, open, or started
	now := time.Now()
	t.Status = "cancelled"
	t.CancelledAt = &now
	return nil
}

// GetByID retrieves a tournament by ID.
func (r *TournamentRepository) GetByID(ctx context.Context, id uuid.UUID) (*Tournament, error) {
	row, err := r.queries.GetTournamentByID(ctx, pgtypeconv.UUIDToPgtype(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrTournamentNotFound
		}
		return nil, fmt.Errorf("get tournament: %w", err)
	}

	return &Tournament{
		ID:                       pgtypeconv.PgtypeToUUID(row.ID),
		HostID:                   pgtypeconv.PgtypeToUUID(row.HostID),
		Name:                     row.Name,
		Description:              pgtypeconv.PgtypeToStringPtr(row.Description),
		ExternalID:               pgtypeconv.PgtypeToStringPtr(row.ExternalID),
		Status:                   row.Status,
		ScheduledStartTimeAt:     pgtypeconv.PgtypeTimestamptzToTimePtr(row.ScheduledStartTimeAt),
		RegistrationWindowOpenAt: pgtypeconv.PgtypeTimestamptzToTimePtr(row.RegistrationWindowOpenAt),
		ActualStartTimeAt:        pgtypeconv.PgtypeTimestamptzToTimePtr(row.ActualStartTimeAt),
		CompletedAt:              pgtypeconv.PgtypeTimestamptzToTimePtr(row.CompletedAt),
		CancelledAt:              pgtypeconv.PgtypeTimestamptzToTimePtr(row.CancelledAt),
		MinimumTeamCount:         pgtypeconv.PgtypeToInt32(row.MinimumTeamCount),
		MaximumTeamCount:         pgtypeconv.PgtypeToInt32Ptr(row.MaximumTeamCount),
		TeamSize:                 row.TeamSize,
		AutoForceReady:           row.AutoForceReady.Bool,
		GameID:                   pgtypeconv.PgtypeToUUIDPtr(row.GameID),
		SpaceID:                  pgtypeconv.PgtypeToUUIDPtr(row.SpaceID),
		TemplateID:               pgtypeconv.PgtypeToUUIDPtr(row.TemplateID),
		ReadyWindow:              row.ReadyWindow,
		Prize:                    row.Prize,
		CreatedAt:                row.CreatedAt.Time,
		UpdatedAt:                row.UpdatedAt.Time,
	}, nil
}

// GetByHostAndID retrieves a tournament ensuring it belongs to the provided host.
func (r *TournamentRepository) GetByHostAndID(ctx context.Context, hostID, tournamentID uuid.UUID) (*Tournament, error) {
	params := sqlc.GetTournamentByHostAndIDParams{
		HostID: pgtypeconv.UUIDToPgtype(hostID),
		ID:     pgtypeconv.UUIDToPgtype(tournamentID),
	}

	row, err := r.queries.GetTournamentByHostAndID(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrTournamentNotFound
		}
		return nil, fmt.Errorf("get tournament: %w", err)
	}

	return &Tournament{
		ID:                       pgtypeconv.PgtypeToUUID(row.ID),
		HostID:                   pgtypeconv.PgtypeToUUID(row.HostID),
		Name:                     row.Name,
		Description:              pgtypeconv.PgtypeToStringPtr(row.Description),
		ExternalID:               pgtypeconv.PgtypeToStringPtr(row.ExternalID),
		Status:                   row.Status,
		ScheduledStartTimeAt:     pgtypeconv.PgtypeTimestamptzToTimePtr(row.ScheduledStartTimeAt),
		RegistrationWindowOpenAt: pgtypeconv.PgtypeTimestamptzToTimePtr(row.RegistrationWindowOpenAt),
		ActualStartTimeAt:        pgtypeconv.PgtypeTimestamptzToTimePtr(row.ActualStartTimeAt),
		CompletedAt:              pgtypeconv.PgtypeTimestamptzToTimePtr(row.CompletedAt),
		CancelledAt:              pgtypeconv.PgtypeTimestamptzToTimePtr(row.CancelledAt),
		MinimumTeamCount:         pgtypeconv.PgtypeToInt32(row.MinimumTeamCount),
		MaximumTeamCount:         pgtypeconv.PgtypeToInt32Ptr(row.MaximumTeamCount),
		TeamSize:                 row.TeamSize,
		AutoForceReady:           row.AutoForceReady.Bool,
		GameID:                   pgtypeconv.PgtypeToUUIDPtr(row.GameID),
		SpaceID:                  pgtypeconv.PgtypeToUUIDPtr(row.SpaceID),
		TemplateID:               pgtypeconv.PgtypeToUUIDPtr(row.TemplateID),
		ReadyWindow:              row.ReadyWindow,
		Prize:                    row.Prize,
		CreatedAt:                row.CreatedAt.Time,
		UpdatedAt:                row.UpdatedAt.Time,
	}, nil
}

// ListByHostID retrieves all tournaments for a specific host.
func (r *TournamentRepository) ListByHostID(ctx context.Context, hostID uuid.UUID) ([]TournamentListItem, error) {
	sqlcTournaments, err := r.queries.ListTournamentsByHostID(ctx, pgtypeconv.UUIDToPgtype(hostID))
	if err != nil {
		return nil, fmt.Errorf("list tournaments: %w", err)
	}

	items := make([]TournamentListItem, len(sqlcTournaments))
	for i, t := range sqlcTournaments {
		items[i] = TournamentListItem{
			ID:                   pgtypeconv.PgtypeToUUID(t.ID),
			Name:                 t.Name,
			Status:               t.Status,
			ScheduledStartTimeAt: pgtypeconv.PgtypeTimestamptzToTimePtr(t.ScheduledStartTimeAt),
			CreatedAt:            t.CreatedAt.Time,
		}
	}

	return items, nil
}

// CreateTournamentParams contains parameters for creating a tournament.
type CreateTournamentParams struct {
	HostID      uuid.UUID
	Name        string
	Description *string
	ExternalID  *string

	ScheduledStartTimeAt     *time.Time
	RegistrationWindowOpenAt *time.Time

	MinimumTeamCount *int32
	MaximumTeamCount *int32
	TeamSize         int32
	AutoForceReady   *bool

	GameID     *uuid.UUID
	SpaceID    *uuid.UUID
	TemplateID *uuid.UUID
}

// Create creates a new tournament.
func (r *TournamentRepository) Create(ctx context.Context, params CreateTournamentParams) (*Tournament, error) {
	sqlcParams := sqlc.CreateTournamentParams{
		HostID:                   pgtypeconv.UUIDToPgtype(params.HostID),
		Name:                     params.Name,
		Description:              pgtypeconv.StringPtrToPgtype(params.Description),
		ExternalID:               pgtypeconv.StringPtrToPgtype(params.ExternalID),
		Status:                   "draft",
		ScheduledStartTimeAt:     pgtypeconv.TimePtrToPgtypeTimestamptz(params.ScheduledStartTimeAt),
		RegistrationWindowOpenAt: pgtypeconv.TimePtrToPgtypeTimestamptz(params.RegistrationWindowOpenAt),
		MinimumTeamCount:         pgtypeconv.Int32PtrToPgtype(params.MinimumTeamCount),
		MaximumTeamCount:         pgtypeconv.Int32PtrToPgtype(params.MaximumTeamCount),
		TeamSize:                 params.TeamSize,
		AutoForceReady:           pgtypeconv.BoolPtrToPgtype(params.AutoForceReady),
		GameID:                   pgtypeconv.UUIDPtrToPgtype(params.GameID),
		SpaceID:                  pgtypeconv.UUIDPtrToPgtype(params.SpaceID),
		TemplateID:               pgtypeconv.UUIDPtrToPgtype(params.TemplateID),
	}

	row, err := r.queries.CreateTournament(ctx, sqlcParams)
	if err != nil {
		return nil, fmt.Errorf("create tournament: %w", err)
	}

	return &Tournament{
		ID:                       pgtypeconv.PgtypeToUUID(row.ID),
		HostID:                   pgtypeconv.PgtypeToUUID(row.HostID),
		Name:                     row.Name,
		Description:              pgtypeconv.PgtypeToStringPtr(row.Description),
		ExternalID:               pgtypeconv.PgtypeToStringPtr(row.ExternalID),
		Status:                   row.Status,
		ScheduledStartTimeAt:     pgtypeconv.PgtypeTimestamptzToTimePtr(row.ScheduledStartTimeAt),
		RegistrationWindowOpenAt: pgtypeconv.PgtypeTimestamptzToTimePtr(row.RegistrationWindowOpenAt),
		ActualStartTimeAt:        pgtypeconv.PgtypeTimestamptzToTimePtr(row.ActualStartTimeAt),
		CompletedAt:              pgtypeconv.PgtypeTimestamptzToTimePtr(row.CompletedAt),
		CancelledAt:              pgtypeconv.PgtypeTimestamptzToTimePtr(row.CancelledAt),
		MinimumTeamCount:         pgtypeconv.PgtypeToInt32(row.MinimumTeamCount),
		MaximumTeamCount:         pgtypeconv.PgtypeToInt32Ptr(row.MaximumTeamCount),
		TeamSize:                 row.TeamSize,
		AutoForceReady:           row.AutoForceReady.Bool,
		GameID:                   pgtypeconv.PgtypeToUUIDPtr(row.GameID),
		SpaceID:                  pgtypeconv.PgtypeToUUIDPtr(row.SpaceID),
		TemplateID:               pgtypeconv.PgtypeToUUIDPtr(row.TemplateID),
		ReadyWindow:              row.ReadyWindow,
		Prize:                    row.Prize,
		CreatedAt:                row.CreatedAt.Time,
		UpdatedAt:                row.UpdatedAt.Time,
	}, nil
}

// UpdateTournamentParams contains parameters for updating a tournament.
// Deprecated: Use Save method with Tournament aggregate instead.
type UpdateTournamentParams struct {
	Name                 *string
	Description          *string
	ScheduledStartTimeAt *time.Time
	MinimumTeamCount     *int32
	MaximumTeamCount     *int32
}

// Update updates tournament information.
// Deprecated: Use Save method with Tournament aggregate instead.
func (r *TournamentRepository) Update(ctx context.Context, id uuid.UUID, params UpdateTournamentParams) (*Tournament, error) {
	sqlcParams := sqlc.UpdateTournamentParams{
		ID:                   pgtypeconv.UUIDToPgtype(id),
		Name:                 pgtypeconv.StringPtrToPgtype(params.Name),
		Description:          pgtypeconv.StringPtrToPgtype(params.Description),
		ScheduledStartTimeAt: pgtypeconv.TimePtrToPgtypeTimestamptz(params.ScheduledStartTimeAt),
		MinimumTeamCount:     pgtypeconv.Int32PtrToPgtype(params.MinimumTeamCount),
		MaximumTeamCount:     pgtypeconv.Int32PtrToPgtype(params.MaximumTeamCount),
	}

	row, err := r.queries.UpdateTournament(ctx, sqlcParams)
	if err != nil {
		return nil, fmt.Errorf("update tournament: %w", err)
	}

	return &Tournament{
		ID:                       pgtypeconv.PgtypeToUUID(row.ID),
		HostID:                   pgtypeconv.PgtypeToUUID(row.HostID),
		Name:                     row.Name,
		Description:              pgtypeconv.PgtypeToStringPtr(row.Description),
		ExternalID:               pgtypeconv.PgtypeToStringPtr(row.ExternalID),
		Status:                   row.Status,
		ScheduledStartTimeAt:     pgtypeconv.PgtypeTimestamptzToTimePtr(row.ScheduledStartTimeAt),
		RegistrationWindowOpenAt: pgtypeconv.PgtypeTimestamptzToTimePtr(row.RegistrationWindowOpenAt),
		ActualStartTimeAt:        pgtypeconv.PgtypeTimestamptzToTimePtr(row.ActualStartTimeAt),
		CompletedAt:              pgtypeconv.PgtypeTimestamptzToTimePtr(row.CompletedAt),
		CancelledAt:              pgtypeconv.PgtypeTimestamptzToTimePtr(row.CancelledAt),
		MinimumTeamCount:         pgtypeconv.PgtypeToInt32(row.MinimumTeamCount),
		MaximumTeamCount:         pgtypeconv.PgtypeToInt32Ptr(row.MaximumTeamCount),
		TeamSize:                 row.TeamSize,
		AutoForceReady:           row.AutoForceReady.Bool,
		GameID:                   pgtypeconv.PgtypeToUUIDPtr(row.GameID),
		SpaceID:                  pgtypeconv.PgtypeToUUIDPtr(row.SpaceID),
		TemplateID:               pgtypeconv.PgtypeToUUIDPtr(row.TemplateID),
		ReadyWindow:              nil, // UpdateTournament doesn't return these fields
		Prize:                    nil,
		CreatedAt:                row.CreatedAt.Time,
		UpdatedAt:                row.UpdatedAt.Time,
	}, nil
}

// Save persists the tournament aggregate state to the database.
// This method saves the full aggregate state after business methods have been applied.
func (r *TournamentRepository) Save(ctx context.Context, tournament *Tournament) (*Tournament, error) {
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
		ReadyWindow:          tournament.ReadyWindow,
		Prize:                tournament.Prize,
	}

	row, err := r.queries.SaveTournament(ctx, sqlcParams)
	if err != nil {
		return nil, fmt.Errorf("save tournament: %w", err)
	}

	return &Tournament{
		ID:                       pgtypeconv.PgtypeToUUID(row.ID),
		HostID:                   pgtypeconv.PgtypeToUUID(row.HostID),
		Name:                     row.Name,
		Description:              pgtypeconv.PgtypeToStringPtr(row.Description),
		ExternalID:               pgtypeconv.PgtypeToStringPtr(row.ExternalID),
		Status:                   row.Status,
		ScheduledStartTimeAt:     pgtypeconv.PgtypeTimestamptzToTimePtr(row.ScheduledStartTimeAt),
		RegistrationWindowOpenAt: pgtypeconv.PgtypeTimestamptzToTimePtr(row.RegistrationWindowOpenAt),
		ActualStartTimeAt:        pgtypeconv.PgtypeTimestamptzToTimePtr(row.ActualStartTimeAt),
		CompletedAt:              pgtypeconv.PgtypeTimestamptzToTimePtr(row.CompletedAt),
		CancelledAt:              pgtypeconv.PgtypeTimestamptzToTimePtr(row.CancelledAt),
		MinimumTeamCount:         pgtypeconv.PgtypeToInt32(row.MinimumTeamCount),
		MaximumTeamCount:         pgtypeconv.PgtypeToInt32Ptr(row.MaximumTeamCount),
		TeamSize:                 row.TeamSize,
		AutoForceReady:           row.AutoForceReady.Bool,
		GameID:                   pgtypeconv.PgtypeToUUIDPtr(row.GameID),
		SpaceID:                  pgtypeconv.PgtypeToUUIDPtr(row.SpaceID),
		TemplateID:               pgtypeconv.PgtypeToUUIDPtr(row.TemplateID),
		ReadyWindow:              row.ReadyWindow,
		Prize:                    row.Prize,
		CreatedAt:                row.CreatedAt.Time,
		UpdatedAt:                row.UpdatedAt.Time,
	}, nil
}

// Start marks a tournament as started.
func (r *TournamentRepository) Start(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.StartTournament(ctx, pgtypeconv.UUIDToPgtype(id)); err != nil {
		return fmt.Errorf("start tournament: %w", err)
	}
	return nil
}

// Cancel marks a tournament as cancelled.
func (r *TournamentRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.CancelTournament(ctx, pgtypeconv.UUIDToPgtype(id)); err != nil {
		return fmt.Errorf("cancel tournament: %w", err)
	}
	return nil
}

// ============================================================================
// Type Converters - using shared converters from converters.go
// ============================================================================

func sqlcTournamentToTournament(t sqlc.Tournament) *Tournament {
	return &Tournament{
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
		ReadyWindow:              t.ReadyWindow,
		Prize:                    t.Prize,
		CreatedAt:                t.CreatedAt.Time,
		UpdatedAt:                t.UpdatedAt.Time,
	}
}
