package domain

import (
	"context"

	"github.com/google/uuid"
)

// TournamentRepository defines the interface for tournament operations in the domain layer.
type TournamentRepository interface {
	// GetByID retrieves a tournament by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Tournament, error)

	// GetByHostAndID retrieves a tournament by host ID and tournament ID.
	GetByHostAndID(ctx context.Context, hostID, tournamentID uuid.UUID) (*Tournament, error)

	// Save persists a tournament entity.
	Save(ctx context.Context, tournament *Tournament) (*Tournament, error)

	// ListByStatus retrieves tournaments by status for scheduler processing.
	ListByStatus(ctx context.Context, statuses []string) ([]*Tournament, error)

	// ListByHost retrieves tournaments for a specific host.
	ListByHost(ctx context.Context, hostID uuid.UUID) ([]*Tournament, error)
}

// ParticipantRepository defines the interface for participant operations in the domain layer.
type ParticipantRepository interface {
	// Create creates a new participant registration.
	Create(ctx context.Context, participant *TournamentParticipant) error

	// GetByTournamentAndUser retrieves a participant by tournament ID and user ID.
	GetByTournamentAndUser(ctx context.Context, tournamentID, userID uuid.UUID) (*TournamentParticipant, error)

	// UpdateStatus updates the status of a participant.
	UpdateStatus(ctx context.Context, tournamentID, userID uuid.UUID, status string) error

	// UpdateReady updates the ready status of a participant.
	UpdateReady(ctx context.Context, tournamentID, userID uuid.UUID, isReady bool) error

	// CountByTournamentAndStatus counts participants by tournament and status.
	CountByTournamentAndStatus(ctx context.Context, tournamentID uuid.UUID, status string) (int64, error)

	// CountByTournament counts all participants in a tournament.
	CountByTournament(ctx context.Context, tournamentID uuid.UUID) (int64, error)

	// CountConfirmed counts confirmed and checked-in participants.
	CountConfirmed(ctx context.Context, tournamentID uuid.UUID) (int64, error)

	// ListByTournament retrieves all participants for a tournament.
	ListByTournament(ctx context.Context, tournamentID uuid.UUID) ([]*TournamentParticipant, error)

	// ListConfirmed retrieves all confirmed/checked-in participants for a tournament.
	ListConfirmed(ctx context.Context, tournamentID uuid.UUID) ([]*TournamentParticipant, error)

	// GetConfirmedUserIDs retrieves user IDs of confirmed participants.
	GetConfirmedUserIDs(ctx context.Context, tournamentID uuid.UUID) ([]uuid.UUID, error)
}

// HostRepository defines the interface for host operations in the domain layer.
type HostRepository interface {
	// GetByID retrieves a host by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Host, error)

	// IsUserAdmin checks if a user is an admin of the host.
	IsUserAdmin(ctx context.Context, hostID, userID uuid.UUID) (bool, error)

	// GetUserRole retrieves the user's role for a host.
	GetUserRole(ctx context.Context, hostID, userID uuid.UUID) (string, error)

	// GetUserHosts retrieves all hosts the user is a member of.
	GetUserHosts(ctx context.Context, userID uuid.UUID) ([]*Host, error)
}

// Host represents a tournament host organization.
type Host struct {
	ID          uuid.UUID
	Name        string
	Description *string
	LogoURL     *string
	WebsiteURL  *string
}

// HostMember represents a user's membership in a host organization.
type HostMember struct {
	ID     uuid.UUID
	HostID uuid.UUID
	UserID uuid.UUID
	Role   string // owner, admin, member
}
