package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgtypeconv "github.com/winspire/winspire-core/libs/go/pgtype"

	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
	"github.com/winspire-core/services/matchmaking/internal/tournament/domain"
)

// RegistrationRepository wraps sqlc-generated queries for tournament registrations.
type RegistrationRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// Errors
var (
	ErrRegistrationNotFound = errors.New("registration not found")
)

// NewRegistrationRepository creates a new RegistrationRepository.
func NewRegistrationRepository(pool *pgxpool.Pool) *RegistrationRepository {
	return &RegistrationRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// Create creates a new registration.
func (r *RegistrationRepository) Create(ctx context.Context, participant *domain.TournamentParticipant) error {
	params := sqlc.CreateTournamentRegistrationParams{
		TournamentID: pgtypeconv.UUIDToPgtype(participant.TournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(participant.UserID),
		TeamID:       pgtypeconv.UUIDPtrToPgtype(participant.TeamID),
		Status:       participant.Status,
		DisplayName:  participant.DisplayName,
		AvatarUrl:    pgtypeconv.StringPtrToPgtype(participant.AvatarURL),
	}

	_, err := r.queries.CreateTournamentRegistration(ctx, params)
	if err != nil {
		return fmt.Errorf("create registration: %w", err)
	}

	return nil
}

// GetByTournamentAndUser retrieves a registration by tournament and user ID.
func (r *RegistrationRepository) GetByTournamentAndUser(ctx context.Context, tournamentID, userID uuid.UUID) (*domain.TournamentParticipant, error) {
	params := sqlc.GetTournamentRegistrationParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(userID),
	}

	row, err := r.queries.GetTournamentRegistration(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRegistrationNotFound
		}
		return nil, fmt.Errorf("get registration: %w", err)
	}

	return sqlcRegistrationToDomain(row), nil
}

// ListByTournament retrieves all registrations for a tournament.
func (r *RegistrationRepository) ListByTournament(ctx context.Context, tournamentID uuid.UUID) ([]*domain.TournamentParticipant, error) {
	rows, err := r.queries.ListAllTournamentRegistrations(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return nil, fmt.Errorf("list registrations: %w", err)
	}

	participants := make([]*domain.TournamentParticipant, len(rows))
	for i, row := range rows {
		participants[i] = sqlcRegistrationToDomain(row)
	}

	return participants, nil
}

// ListConfirmed retrieves all confirmed/checked-in participants for a tournament.
func (r *RegistrationRepository) ListConfirmed(ctx context.Context, tournamentID uuid.UUID) ([]*domain.TournamentParticipant, error) {
	rows, err := r.queries.ListConfirmedTournamentRegistrations(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return nil, fmt.Errorf("list confirmed registrations: %w", err)
	}

	participants := make([]*domain.TournamentParticipant, len(rows))
	for i, row := range rows {
		participants[i] = sqlcRegistrationToDomain(row)
	}

	return participants, nil
}

// GetConfirmedUserIDs retrieves user IDs of confirmed participants.
func (r *RegistrationRepository) GetConfirmedUserIDs(ctx context.Context, tournamentID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.queries.GetConfirmedTournamentUserIDs(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return nil, fmt.Errorf("get confirmed user IDs: %w", err)
	}

	userIDs := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		userIDs[i] = pgtypeconv.PgtypeToUUID(row)
	}

	return userIDs, nil
}

// UpdateStatus updates the registration status for a participant.
func (r *RegistrationRepository) UpdateStatus(ctx context.Context, tournamentID, userID uuid.UUID, status string) error {
	params := sqlc.UpdateTournamentRegistrationStatusParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(userID),
		Status:       status,
	}

	if err := r.queries.UpdateTournamentRegistrationStatus(ctx, params); err != nil {
		return fmt.Errorf("update registration status: %w", err)
	}
	return nil
}

// UpdateReady updates the ready status for a participant.
func (r *RegistrationRepository) UpdateReady(ctx context.Context, tournamentID, userID uuid.UUID, isReady bool) error {
	params := sqlc.UpdateTournamentRegistrationReadyParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(userID),
		IsReady:      pgtypeconv.BoolPtrToPgtype(&isReady),
	}

	if err := r.queries.UpdateTournamentRegistrationReady(ctx, params); err != nil {
		return fmt.Errorf("update ready status: %w", err)
	}
	return nil
}

// CountByTournament counts all participants in a tournament.
func (r *RegistrationRepository) CountByTournament(ctx context.Context, tournamentID uuid.UUID) (int64, error) {
	count, err := r.queries.CountTournamentRegistrations(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return 0, fmt.Errorf("count registrations: %w", err)
	}
	return count, nil
}

// CountByTournamentAndStatus counts participants by tournament and status.
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

// CountConfirmed counts confirmed and checked-in participants.
func (r *RegistrationRepository) CountConfirmed(ctx context.Context, tournamentID uuid.UUID) (int64, error) {
	count, err := r.queries.CountConfirmedTournamentRegistrations(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return 0, fmt.Errorf("count confirmed registrations: %w", err)
	}
	return count, nil
}

// CheckIn marks a participant as checked in.
func (r *RegistrationRepository) CheckIn(ctx context.Context, tournamentID, userID uuid.UUID) error {
	params := sqlc.CheckInTournamentParticipantParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(userID),
	}

	if err := r.queries.CheckInTournamentParticipant(ctx, params); err != nil {
		return fmt.Errorf("check in participant: %w", err)
	}
	return nil
}

// Withdraw withdraws a registration.
func (r *RegistrationRepository) Withdraw(ctx context.Context, tournamentID, userID uuid.UUID) error {
	params := sqlc.WithdrawTournamentRegistrationParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(userID),
	}

	if err := r.queries.WithdrawTournamentRegistration(ctx, params); err != nil {
		return fmt.Errorf("withdraw registration: %w", err)
	}
	return nil
}

// Delete removes a registration.
func (r *RegistrationRepository) Delete(ctx context.Context, tournamentID, userID uuid.UUID) error {
	params := sqlc.DeleteTournamentRegistrationParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		UserID:       pgtypeconv.UUIDToPgtype(userID),
	}

	if err := r.queries.DeleteTournamentRegistration(ctx, params); err != nil {
		return fmt.Errorf("delete registration: %w", err)
	}
	return nil
}

// sqlcRegistrationToDomain converts SQLC registration to domain participant.
func sqlcRegistrationToDomain(r sqlc.TournamentRegistration) *domain.TournamentParticipant {
	return &domain.TournamentParticipant{
		ID:           pgtypeconv.PgtypeToUUID(r.ID),
		TournamentID: pgtypeconv.PgtypeToUUID(r.TournamentID),
		UserID:       pgtypeconv.PgtypeToUUID(r.UserID),
		TeamID:       pgtypeconv.PgtypeToUUIDPtr(r.TeamID),
		Status:       r.Status,
		DisplayName:  r.DisplayName,
		AvatarURL:    pgtypeconv.PgtypeToStringPtr(r.AvatarUrl),
		IsReady:      r.IsReady.Bool,
		RegisteredAt: pgtypeconv.PgtypeTimestamptzToTimePtr(r.RegisteredAt),
		CheckedInAt:  pgtypeconv.PgtypeTimestamptzToTimePtr(r.CheckedInAt),
		UpdatedAt:    r.UpdatedAt.Time,
	}
}
