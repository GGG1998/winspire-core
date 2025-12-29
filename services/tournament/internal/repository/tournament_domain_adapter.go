package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/winspire/tournament/internal/domain"
)

// TournamentDomainRepository implements domain.TournamentRepository interface.
// It acts as an adapter between the domain layer and the persistence layer.
type TournamentDomainRepository struct {
	repo *TournamentRepository
}

// NewTournamentDomainRepository creates a new adapter for tournament domain operations.
func NewTournamentDomainRepository(repo *TournamentRepository) *TournamentDomainRepository {
	return &TournamentDomainRepository{repo: repo}
}

// GetByID retrieves a tournament and converts it to a domain entity.
func (r *TournamentDomainRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tournament, error) {
	t, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.toDomain(t), nil
}

// Save persists a domain tournament entity.
func (r *TournamentDomainRepository) Save(ctx context.Context, tournament *domain.Tournament) (*domain.Tournament, error) {
	// Convert domain entity to persistence model
	persistenceModel := r.fromDomain(tournament)

	// Save using persistence repository
	saved, err := r.repo.Save(ctx, persistenceModel)
	if err != nil {
		return nil, err
	}

	return r.toDomain(saved), nil
}

// ListByStatus retrieves tournaments by status.
func (r *TournamentDomainRepository) ListByStatus(ctx context.Context, statuses []string) ([]*domain.Tournament, error) {
	tournaments, err := r.repo.ListByStatus(ctx, statuses)
	if err != nil {
		return nil, fmt.Errorf("list tournaments by status: %w", err)
	}

	domainTournaments := make([]*domain.Tournament, len(tournaments))
	for i, t := range tournaments {
		domainTournaments[i] = r.toDomain(t)
	}

	return domainTournaments, nil
}

// toDomain converts a persistence Tournament to a domain Tournament.
func (r *TournamentDomainRepository) toDomain(t *Tournament) *domain.Tournament {
	return &domain.Tournament{
		ID:                   t.ID,
		HostID:               t.HostID,
		Name:                 t.Name,
		Description:          t.Description,
		Status:               t.Status,
		ScheduledStartTimeAt: t.ScheduledStartTimeAt,
		MaximumTeamCount:     t.MaximumTeamCount,
		MinimumTeamCount:     t.MinimumTeamCount,
		TeamSize:             t.TeamSize,
		AutoForceReady:       t.AutoForceReady,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}
}

// fromDomain converts a domain Tournament to a persistence Tournament.
func (r *TournamentDomainRepository) fromDomain(t *domain.Tournament) *Tournament {
	return &Tournament{
		ID:                   t.ID,
		HostID:               t.HostID,
		Name:                 t.Name,
		Description:          t.Description,
		Status:               t.Status,
		ScheduledStartTimeAt: t.ScheduledStartTimeAt,
		MaximumTeamCount:     t.MaximumTeamCount,
		MinimumTeamCount:     t.MinimumTeamCount,
		TeamSize:             t.TeamSize,
		AutoForceReady:       t.AutoForceReady,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}
}
