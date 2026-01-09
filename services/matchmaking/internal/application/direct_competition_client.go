// Package application contains application services and business logic orchestration
package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	tournamentrepo "github.com/winspire-core/services/matchmaking/internal/tournament/repository"
)

// DirectCompetitionClient fetches tournament data directly from the database
// instead of making HTTP calls to an external service.
// This is used when tournament functionality is merged into the matchmaking service.
type DirectCompetitionClient struct {
	tournamentRepo   *tournamentrepo.TournamentRepository
	registrationRepo *tournamentrepo.RegistrationRepository
	logger           *observability.Logger
}

// NewDirectCompetitionClient creates a new direct competition client
func NewDirectCompetitionClient(
	tournamentRepo *tournamentrepo.TournamentRepository,
	registrationRepo *tournamentrepo.RegistrationRepository,
	logger *observability.Logger,
) *DirectCompetitionClient {
	return &DirectCompetitionClient{
		tournamentRepo:   tournamentRepo,
		registrationRepo: registrationRepo,
		logger:           logger,
	}
}

// GetTournamentInfo fetches tournament information directly from the database
func (c *DirectCompetitionClient) GetTournamentInfo(ctx context.Context, tournamentID uuid.UUID) (*TournamentInfo, error) {
	tournament, err := c.tournamentRepo.GetByID(ctx, tournamentID)
	if err != nil {
		if errors.Is(err, tournamentrepo.ErrTournamentNotFound) {
			return nil, nil
		}
		c.logger.Warn("failed to get tournament from database", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return nil, err
	}

	// Get start time (default to now if not set)
	var startTime time.Time
	if tournament.ScheduledStartTimeAt != nil {
		startTime = *tournament.ScheduledStartTimeAt
	} else {
		startTime = time.Now()
	}

	// Get GameID (may be nil)
	var gameID uuid.UUID
	if tournament.GameID != nil {
		gameID = *tournament.GameID
	}

	// Default min participants if not set
	minParticipants := int(tournament.MinimumTeamCount)
	if minParticipants < 2 {
		minParticipants = 2
	}

	return &TournamentInfo{
		ID:              tournament.ID,
		Name:            tournament.Name,
		HostID:          tournament.HostID,
		GameID:          gameID,
		StartTime:       startTime,
		MinParticipants: minParticipants,
		Status:          tournament.Status,
	}, nil
}

// IsUserRegistered checks if a user is registered for a tournament directly from the database
func (c *DirectCompetitionClient) IsUserRegistered(ctx context.Context, tournamentID, userID uuid.UUID) (bool, error) {
	registration, err := c.registrationRepo.GetByTournamentAndUser(ctx, tournamentID, userID)
	if err != nil {
		if errors.Is(err, tournamentrepo.ErrRegistrationNotFound) {
			return false, nil
		}
		c.logger.Warn("failed to check user registration from database", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"user_id":       userID.String(),
			"error":         err.Error(),
		})
		return false, err
	}

	// User is registered if we found a registration record
	// Check if the registration status allows access (not withdrawn)
	if registration.Status == "withdrawn" || registration.Status == "disqualified" {
		return false, nil
	}

	return true, nil
}
