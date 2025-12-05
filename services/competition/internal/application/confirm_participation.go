package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/winspire/competition/internal/domain"
	"github.com/winspire/competition/internal/repository"
)

// ConfirmParticipationUseCase handles the business logic for confirming tournament participation.
type ConfirmParticipationUseCase struct {
	tournamentRepo  domain.TournamentRepository
	participantRepo domain.ParticipantRepository
	logger          *slog.Logger
}

// NewConfirmParticipationUseCase creates a new instance of ConfirmParticipationUseCase.
func NewConfirmParticipationUseCase(
	tournamentRepo domain.TournamentRepository,
	participantRepo domain.ParticipantRepository,
	logger *slog.Logger,
) *ConfirmParticipationUseCase {
	return &ConfirmParticipationUseCase{
		tournamentRepo:  tournamentRepo,
		participantRepo: participantRepo,
		logger:          logger,
	}
}

// Execute registers a user's participation in a tournament.
// Business flow:
// 1. Validate tournament exists and is open for registrations
// 2. Get or create participant (upsert pattern)
// 3. Register participation using domain logic
// 4. Update participant status in database
// 5. Check if all places are taken and close registration if needed
func (uc *ConfirmParticipationUseCase) Execute(ctx context.Context, tournamentID, userID uuid.UUID, displayName string) error {
	// 1. Get tournament and validate registration window
	tournament, err := uc.tournamentRepo.GetByID(ctx, tournamentID)
	if err != nil {
		uc.logger.Error("failed to get tournament", "error", err, "tournamentId", tournamentID)
		return fmt.Errorf("tournament not found: %w", err)
	}

	if err := tournament.CanConfirmParticipation(); err != nil {
		uc.logger.Warn("cannot register for tournament", "error", err, "tournamentId", tournamentID, "status", tournament.Status)
		return fmt.Errorf("registration window not open: %w", err)
	}

	// 2. Get or create participant
	participant, err := uc.participantRepo.GetByTournamentAndUser(ctx, tournamentID, userID)
	if err != nil {
		// If not found, create new participant with pending status
		if errors.Is(err, repository.ErrRegistrationNotFound) {
			participant = &domain.TournamentParticipant{
				ID:           uuid.New(),
				TournamentID: tournamentID,
				UserID:       userID,
				Status:       domain.StatusPending,
				DisplayName:  displayName, // Set from user profile/JWT
				AvatarURL:    nil,         // TODO: Get from user profile
			}
			if err := uc.participantRepo.Create(ctx, participant); err != nil {
				uc.logger.Error("failed to create participant", "error", err, "tournamentId", tournamentID, "userId", userID)
				return fmt.Errorf("failed to create registration: %w", err)
			}
			uc.logger.Info("created new participant", "tournamentId", tournamentID, "userId", userID, "displayName", displayName)
		} else {
			uc.logger.Error("failed to get participant", "error", err, "tournamentId", tournamentID, "userId", userID)
			return fmt.Errorf("database error: %w", err)
		}
	}

	// 3. Register participation using domain logic
	if err := participant.RegisterParticipation(); err != nil {
		uc.logger.Warn("cannot register participation", "error", err, "tournamentId", tournamentID, "userId", userID, "currentStatus", participant.Status)
		return fmt.Errorf("cannot register participation: %w", err)
	}

	// 4. Update participant status
	if err := uc.participantRepo.UpdateStatus(ctx, tournamentID, userID, participant.Status); err != nil {
		uc.logger.Error("failed to update participant status", "error", err, "tournamentId", tournamentID, "userId", userID)
		return fmt.Errorf("failed to register participation: %w", err)
	}

	uc.logger.Info("participation registered", "tournamentId", tournamentID, "userId", userID)

	// 5. Check if all places are taken and close registration
	registeredCount, err := uc.participantRepo.CountByTournamentAndStatus(ctx, tournamentID, domain.StatusRegistered)
	if err != nil {
		uc.logger.Warn("failed to count registered participants", "error", err, "tournamentId", tournamentID)
		// Continue without closing registration
		return nil
	}

	if tournament.ShouldCloseRegistration(registeredCount) {
		tournament.CloseRegistration()
		if _, err := uc.tournamentRepo.Save(ctx, tournament); err != nil {
			uc.logger.Error("failed to close registration", "error", err, "tournamentId", tournamentID)
			// Don't fail the whole operation if auto-close fails
			return nil
		}
		uc.logger.Info("registration auto-closed - all places taken", "tournamentId", tournamentID, "registeredCount", registeredCount)
	}

	return nil
}
