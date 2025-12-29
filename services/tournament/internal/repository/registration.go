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

	"github.com/winspire/tournament/internal/store/sqlc"
)

// RegistrationRepository wraps sqlc-generated queries for tournament registrations.
type RegistrationRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// ErrRegistrationNotFound is returned when a registration cannot be found.
var ErrRegistrationNotFound = errors.New("registration not found")

// ErrAlreadyRegistered is returned when user is already registered.
var ErrAlreadyRegistered = errors.New("user already registered")

// NewRegistrationRepository creates a new RegistrationRepository.
func NewRegistrationRepository(pool *pgxpool.Pool) *RegistrationRepository {
	return &RegistrationRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// Registration represents a tournament participant registration.
type Registration struct {
	ID           uuid.UUID
	TournamentID uuid.UUID
	UserID       uuid.UUID
	TeamID       *uuid.UUID

	Status      string
	DisplayName string
	AvatarURL   *string

	CheckedInAt *time.Time
	IsReady     bool

	RegisteredAt time.Time
	UpdatedAt    time.Time
}

// CreateRegistrationParams contains parameters for creating a registration.
type CreateRegistrationParams struct {
	TournamentID uuid.UUID
	UserID       uuid.UUID
	TeamID       *uuid.UUID
	DisplayName  string
	AvatarURL    *string
}

// Create creates a new registration.
func (r *RegistrationRepository) Create(ctx context.Context, params CreateRegistrationParams) (*Registration, error) {
	sqlcParams := sqlc.CreateTournamentRegistrationParams{
		TournamentID: pgtypeconv.UUIDToPgtype(params.TournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(params.UserID),
		TeamID:       pgtypeconv.UUIDPtrToPgtype(params.TeamID),
		Status:       "pending",
		DisplayName:  params.DisplayName,
		AvatarUrl:    pgtypeconv.StringPtrToPgtype(params.AvatarURL),
	}

	sqlcReg, err := r.queries.CreateTournamentRegistration(ctx, sqlcParams)
	if err != nil {
		return nil, fmt.Errorf("create registration: %w", err)
	}

	return sqlcRegistrationToRegistration(sqlcReg), nil
}

// GetByTournamentAndUser retrieves a registration by tournament and user ID.
func (r *RegistrationRepository) GetByTournamentAndUser(ctx context.Context, tournamentID, userID uuid.UUID) (*Registration, error) {
	params := sqlc.GetTournamentRegistrationParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(userID),
	}

	sqlcReg, err := r.queries.GetTournamentRegistration(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRegistrationNotFound
		}
		return nil, fmt.Errorf("get registration: %w", err)
	}

	return sqlcRegistrationToRegistration(sqlcReg), nil
}

// ListByTournament retrieves paginated registrations for a tournament.
func (r *RegistrationRepository) ListByTournament(ctx context.Context, tournamentID uuid.UUID, limit, offset int32) ([]Registration, error) {
	params := sqlc.ListTournamentRegistrationsParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		Limit:        limit,
		Offset:       offset,
	}

	sqlcRegs, err := r.queries.ListTournamentRegistrations(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list registrations: %w", err)
	}

	regs := make([]Registration, len(sqlcRegs))
	for i, sqlcReg := range sqlcRegs {
		regs[i] = *sqlcRegistrationToRegistration(sqlcReg)
	}

	return regs, nil
}

// CountByTournament counts total registrations for a tournament.
func (r *RegistrationRepository) CountByTournament(ctx context.Context, tournamentID uuid.UUID) (int64, error) {
	count, err := r.queries.CountTournamentRegistrations(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return 0, fmt.Errorf("count registrations: %w", err)
	}
	return count, nil
}

// CountReadyByTournament counts ready participants for a tournament.
func (r *RegistrationRepository) CountReadyByTournament(ctx context.Context, tournamentID uuid.UUID) (int64, error) {
	count, err := r.queries.CountReadyParticipants(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return 0, fmt.Errorf("count ready participants: %w", err)
	}
	return count, nil
}

// CheckIn marks a participant as checked in.
func (r *RegistrationRepository) CheckIn(ctx context.Context, tournamentID, userID uuid.UUID) error {
	params := sqlc.CheckInParticipantParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(userID),
	}

	if err := r.queries.CheckInParticipant(ctx, params); err != nil {
		return fmt.Errorf("check in participant: %w", err)
	}
	return nil
}

// Withdraw withdraws a registration.
func (r *RegistrationRepository) Withdraw(ctx context.Context, tournamentID, userID uuid.UUID) error {
	params := sqlc.WithdrawRegistrationParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(userID),
	}

	if err := r.queries.WithdrawRegistration(ctx, params); err != nil {
		return fmt.Errorf("withdraw registration: %w", err)
	}
	return nil
}

// SetReady sets the ready status for a participant.
func (r *RegistrationRepository) SetReady(ctx context.Context, tournamentID, userID uuid.UUID, isReady bool) error {
	params := sqlc.UpdateRegistrationReadyParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(userID),
		IsReady:      pgtypeconv.BoolPtrToPgtype(&isReady),
	}

	if err := r.queries.UpdateRegistrationReady(ctx, params); err != nil {
		return fmt.Errorf("update ready status: %w", err)
	}
	return nil
}

// UpdateStatus updates the registration status for a participant.
func (r *RegistrationRepository) UpdateStatus(ctx context.Context, tournamentID, userID uuid.UUID, status string) error {
	params := sqlc.UpdateRegistrationStatusParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(userID),
		Status:       status,
	}

	if err := r.queries.UpdateRegistrationStatus(ctx, params); err != nil {
		return fmt.Errorf("update registration status: %w", err)
	}
	return nil
}

// CountByTournamentAndStatus counts registrations for a tournament by status.
func (r *RegistrationRepository) CountByTournamentAndStatus(ctx context.Context, tournamentID uuid.UUID, status string) (int64, error) {
	params := sqlc.CountTournamentRegistrationsByStatusParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		Status:       status,
	}

	count, err := r.queries.CountTournamentRegistrationsByStatus(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("count registrations by status: %w", err)
	}
	return count, nil
}

// Delete removes a registration.
func (r *RegistrationRepository) Delete(ctx context.Context, tournamentID, userID uuid.UUID) error {
	params := sqlc.DeleteRegistrationParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(userID),
	}

	if err := r.queries.DeleteRegistration(ctx, params); err != nil {
		return fmt.Errorf("delete registration: %w", err)
	}
	return nil
}

// ============================================================================
// Type Converters
// ============================================================================

func sqlcRegistrationToRegistration(r sqlc.TournamentRegistration) *Registration {
	return &Registration{
		ID:           pgtypeconv.PgtypeToUUID(r.ID),
		TournamentID: pgtypeconv.PgtypeToUUID(r.TournamentID),
		UserID:       pgtypeconv.PgtypeToUUID(r.UserID),
		TeamID:       pgtypeconv.PgtypeToUUIDPtr(r.TeamID),
		Status:       r.Status,
		DisplayName:  r.DisplayName,
		AvatarURL:    pgtypeconv.PgtypeToStringPtr(r.AvatarUrl),
		CheckedInAt:  pgtypeconv.PgtypeTimestamptzToTimePtr(r.CheckedInAt),
		IsReady:      r.IsReady.Bool,
		RegisteredAt: r.RegisteredAt.Time,
		UpdatedAt:    r.UpdatedAt.Time,
	}
}
