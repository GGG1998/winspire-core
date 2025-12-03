package application

import (
	"context"
	"log/slog"

	"github.com/winspire/competition/internal/domain"
)

// UpdateTournamentStatusUseCase handles automatic tournament status updates.
// This is used by the scheduler to manage registration_open/registration_closed status.
type UpdateTournamentStatusUseCase struct {
	tournamentRepo  domain.TournamentRepository
	participantRepo domain.ParticipantRepository
	logger          *slog.Logger
}

// NewUpdateTournamentStatusUseCase creates a new instance of UpdateTournamentStatusUseCase.
func NewUpdateTournamentStatusUseCase(
	tournamentRepo domain.TournamentRepository,
	participantRepo domain.ParticipantRepository,
	logger *slog.Logger,
) *UpdateTournamentStatusUseCase {
	return &UpdateTournamentStatusUseCase{
		tournamentRepo:  tournamentRepo,
		participantRepo: participantRepo,
		logger:          logger,
	}
}

// Execute processes all tournaments that need status updates.
// Business flow:
// 1. Get tournaments in registration_open or registration_closed status
// 2. For each tournament, count confirmed participants
// 3. Update status based on domain rules (close/reopen registration)
func (uc *UpdateTournamentStatusUseCase) Execute(ctx context.Context) error {
	// Get tournaments that might need status updates
	statuses := []string{domain.StatusRegistrationOpen, domain.StatusRegistrationClosed, domain.StatusScheduled}
	tournaments, err := uc.tournamentRepo.ListByStatus(ctx, statuses)
	if err != nil {
		uc.logger.Error("failed to list tournaments for status update", "error", err)
		return err
	}

	uc.logger.Debug("processing tournaments for status update", "count", len(tournaments))

	updatedCount := 0
	for _, tournament := range tournaments {
		if err := uc.processTournament(ctx, tournament); err != nil {
			uc.logger.Error("failed to process tournament", "error", err, "tournamentId", tournament.ID)
			// Continue processing other tournaments
			continue
		}
		updatedCount++
	}

	if updatedCount > 0 {
		uc.logger.Info("tournament status update completed", "processed", len(tournaments), "updated", updatedCount)
	}

	return nil
}

func (uc *UpdateTournamentStatusUseCase) processTournament(ctx context.Context, tournament *domain.Tournament) error {
	// Count confirmed participants
	confirmedCount, err := uc.participantRepo.CountByTournamentAndStatus(ctx, tournament.ID, domain.StatusConfirmed)
	if err != nil {
		return err
	}

	originalStatus := tournament.Status

	// Check if registration should be closed
	if tournament.ShouldCloseRegistration(confirmedCount) {
		tournament.CloseRegistration()
	}

	// Check if registration should be reopened
	if tournament.ShouldReopenRegistration(confirmedCount) {
		tournament.ReopenRegistration()
	}

	// Save if status changed
	if tournament.Status != originalStatus {
		if _, err := uc.tournamentRepo.Save(ctx, tournament); err != nil {
			return err
		}
		uc.logger.Info("tournament status updated",
			"tournamentId", tournament.ID,
			"oldStatus", originalStatus,
			"newStatus", tournament.Status,
			"confirmedCount", confirmedCount,
		)
	}

	return nil
}
