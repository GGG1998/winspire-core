package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/winspire/competition/internal/domain"
)

// ParticipantDomainRepository implements domain.ParticipantRepository interface.
// It acts as an adapter between the domain layer and the persistence layer (RegistrationRepository).
type ParticipantDomainRepository struct {
	repo *RegistrationRepository
}

// NewParticipantDomainRepository creates a new adapter for participant domain operations.
func NewParticipantDomainRepository(repo *RegistrationRepository) *ParticipantDomainRepository {
	return &ParticipantDomainRepository{repo: repo}
}

// Create creates a new participant registration.
// Note: DisplayName should be set in the domain.TournamentParticipant before calling this method.
// If empty, it will fall back to userID string (not recommended).
func (r *ParticipantDomainRepository) Create(ctx context.Context, participant *domain.TournamentParticipant) error {
	displayName := participant.DisplayName
	if displayName == "" {
		// Fallback to userID string if DisplayName is not set
		// This should not happen in production if profile completion is enforced
		displayName = participant.UserID.String()
	}

	params := CreateRegistrationParams{
		TournamentID: participant.TournamentID,
		UserID:       participant.UserID,
		TeamID:       nil, // Not used in current implementation
		DisplayName:  displayName,
		AvatarURL:    participant.AvatarURL,
	}

	_, err := r.repo.Create(ctx, params)
	return err
}

// GetByTournamentAndUser retrieves a participant and converts it to a domain entity.
func (r *ParticipantDomainRepository) GetByTournamentAndUser(ctx context.Context, tournamentID, userID uuid.UUID) (*domain.TournamentParticipant, error) {
	registration, err := r.repo.GetByTournamentAndUser(ctx, tournamentID, userID)
	if err != nil {
		return nil, err
	}
	return r.toDomain(registration), nil
}

// UpdateStatus updates the status of a participant.
func (r *ParticipantDomainRepository) UpdateStatus(ctx context.Context, tournamentID, userID uuid.UUID, status string) error {
	return r.repo.UpdateStatus(ctx, tournamentID, userID, status)
}

// CountByTournamentAndStatus counts participants by tournament and status.
func (r *ParticipantDomainRepository) CountByTournamentAndStatus(ctx context.Context, tournamentID uuid.UUID, status string) (int64, error) {
	return r.repo.CountByTournamentAndStatus(ctx, tournamentID, status)
}

// CountByTournament counts all participants in a tournament.
func (r *ParticipantDomainRepository) CountByTournament(ctx context.Context, tournamentID uuid.UUID) (int64, error) {
	return r.repo.CountByTournament(ctx, tournamentID)
}

// toDomain converts a persistence Registration to a domain TournamentParticipant.
func (r *ParticipantDomainRepository) toDomain(reg *Registration) *domain.TournamentParticipant {
	return &domain.TournamentParticipant{
		ID:           reg.ID,
		TournamentID: reg.TournamentID,
		UserID:       reg.UserID,
		Status:       reg.Status,
		RegisteredAt: &reg.RegisteredAt,
		ConfirmedAt:  nil,             // Not tracked in current schema
		CheckedInAt:  reg.CheckedInAt, // Available in schema
	}
}
