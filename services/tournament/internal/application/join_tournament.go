package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/winspire/tournament/internal/domain"
)

// JoinTournamentUseCase handles the business logic for joining a tournament.
type JoinTournamentUseCase struct {
	tournamentRepo  domain.TournamentRepository
	participantRepo domain.ParticipantRepository
	logger          *slog.Logger
}

// NewJoinTournamentUseCase creates a new instance of JoinTournamentUseCase.
func NewJoinTournamentUseCase(
	tournamentRepo domain.TournamentRepository,
	participantRepo domain.ParticipantRepository,
	logger *slog.Logger,
) *JoinTournamentUseCase {
	return &JoinTournamentUseCase{
		tournamentRepo:  tournamentRepo,
		participantRepo: participantRepo,
		logger:          logger,
	}
}

// Execute allows a confirmed participant to join the tournament.
// Business flow:
// 1. Validate tournament exists and is within join window
// 2. Validate user has confirmed participation (or is registered)
// 2.5. Auto-confirm if status is 'registered'
// 3. Return room link
func (uc *JoinTournamentUseCase) Execute(ctx context.Context, tournamentID, userID, hostID uuid.UUID) (string, error) {
	// 1. Get tournament and validate join window
	tournament, err := uc.tournamentRepo.GetByID(ctx, tournamentID)
	if err != nil {
		uc.logger.Error("failed to get tournament", "error", err, "tournamentId", tournamentID)
		return "", fmt.Errorf("tournament not found: %w", err)
	}

	timeUntilStart, err := tournament.CanJoin()
	if err != nil {
		uc.logger.Warn("cannot join tournament", "error", err, "tournamentId", tournamentID)
		return "", fmt.Errorf("cannot join tournament: %w", err)
	}

	// 2. Get participant and validate confirmation
	participant, err := uc.participantRepo.GetByTournamentAndUser(ctx, tournamentID, userID)
	if err != nil {
		uc.logger.Error("failed to get participant", "error", err, "tournamentId", tournamentID, "userId", userID)
		return "", fmt.Errorf("user not registered for tournament: %w", err)
	}

	if !participant.CanJoin(timeUntilStart) {
		uc.logger.Warn("participant cannot join", "tournamentId", tournamentID, "userId", userID, "status", participant.Status)
		return "", fmt.Errorf("participation not confirmed or join window not open")
	}

	// 2.5. Auto-confirm participation if status is 'registered'
	if participant.Status == domain.StatusRegistered {
		if err := participant.ConfirmParticipation(); err != nil {
			uc.logger.Warn("failed to auto-confirm participation", "error", err, "tournamentId", tournamentID, "userId", userID)
			// Continue anyway - participant is already registered which is sufficient
		} else {
			// Update participant status to 'confirmed'
			if err := uc.participantRepo.UpdateStatus(ctx, tournamentID, userID, participant.Status); err != nil {
				uc.logger.Error("failed to update participant status to confirmed", "error", err, "tournamentId", tournamentID, "userId", userID)
				// Continue anyway - the join can still proceed
			} else {
				uc.logger.Info("auto-confirmed participation", "tournamentId", tournamentID, "userId", userID)
			}
		}
	}

	// 3. Generate room link
	// TODO: This could be enhanced to use tournament's space_id or other configuration
	roomLink := fmt.Sprintf("/tournaments/%s/lobby", tournamentID)

	uc.logger.Info("user joining tournament", "tournamentId", tournamentID, "userId", userID, "hostId", hostID, "roomLink", roomLink)

	return roomLink, nil
}
