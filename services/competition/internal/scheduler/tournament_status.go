package scheduler

import (
	"context"
	"log/slog"

	"github.com/robfig/cron/v3"
	"github.com/winspire/competition/internal/application"
)

// TournamentStatusScheduler periodically updates tournament statuses.
type TournamentStatusScheduler struct {
	useCase  *application.UpdateTournamentStatusUseCase
	logger   *slog.Logger
	interval string
	cron     *cron.Cron
}

// NewTournamentStatusScheduler creates a new tournament status scheduler.
func NewTournamentStatusScheduler(
	useCase *application.UpdateTournamentStatusUseCase,
	logger *slog.Logger,
	interval string,
) *TournamentStatusScheduler {
	return &TournamentStatusScheduler{
		useCase:  useCase,
		logger:   logger,
		interval: interval,
		cron:     cron.New(),
	}
}

// Start begins the scheduled task execution.
func (s *TournamentStatusScheduler) Start(ctx context.Context) error {
	s.logger.Info("tournament status scheduler starting", "interval", s.interval)

	// Add the job
	_, err := s.cron.AddFunc(s.interval, func() {
		s.logger.Debug("running tournament status update")

		if err := s.useCase.Execute(ctx); err != nil {
			s.logger.Error("tournament status update failed", "error", err)
		}
	})
	if err != nil {
		return err
	}

	// Start the cron scheduler
	s.cron.Start()

	s.logger.Info("tournament status scheduler started")
	return nil
}

// Stop gracefully stops the scheduler.
func (s *TournamentStatusScheduler) Stop() {
	s.logger.Info("tournament status scheduler stopping")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("tournament status scheduler stopped")
}

// Run starts the scheduler and blocks until context is cancelled.
func (s *TournamentStatusScheduler) Run(ctx context.Context) error {
	if err := s.Start(ctx); err != nil {
		return err
	}

	// Wait for context cancellation
	<-ctx.Done()

	s.Stop()
	return nil
}


