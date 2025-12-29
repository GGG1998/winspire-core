package domain

import (
	"context"

	"github.com/google/uuid"
)

// TournamentRepository defines the interface for tournament operations in the domain layer.
type TournamentRepository interface {
	// GetByID retrieves a tournament by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Tournament, error)

	// Save persists a tournament entity.
	Save(ctx context.Context, tournament *Tournament) (*Tournament, error)

	// ListByStatus retrieves tournaments by status for scheduler processing.
	ListByStatus(ctx context.Context, statuses []string) ([]*Tournament, error)
}

// ParticipantRepository defines the interface for participant operations in the domain layer.
type ParticipantRepository interface {
	// Create creates a new participant registration.
	Create(ctx context.Context, participant *TournamentParticipant) error

	// GetByTournamentAndUser retrieves a participant by tournament ID and user ID.
	GetByTournamentAndUser(ctx context.Context, tournamentID, userID uuid.UUID) (*TournamentParticipant, error)

	// UpdateStatus updates the status of a participant.
	UpdateStatus(ctx context.Context, tournamentID, userID uuid.UUID, status string) error

	// CountByTournamentAndStatus counts participants by tournament and status.
	CountByTournamentAndStatus(ctx context.Context, tournamentID uuid.UUID, status string) (int64, error)

	// CountByTournament counts all participants in a tournament.
	CountByTournament(ctx context.Context, tournamentID uuid.UUID) (int64, error)
}
