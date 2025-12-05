// Package gameapi provides integration with the external Game API for score retrieval
package gameapi

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/pubsub"
	"github.com/winspire-core/services/matchmaking/internal/repository"
)

// PollerService handles polling the Game API for match results
type PollerService struct {
	gameAPIClient *GameAPIClient
	matchRepo     repository.MatchRepository
	publisher     *pubsub.EventPublisher
	metrics       *observability.MetricsEmitter
	logger        *observability.Logger
}

// NewPollerService creates a new poller service
func NewPollerService(
	gameAPIClient *GameAPIClient,
	matchRepo repository.MatchRepository,
	publisher *pubsub.EventPublisher,
	metrics *observability.MetricsEmitter,
	logger *observability.Logger,
) *PollerService {
	return &PollerService{
		gameAPIClient: gameAPIClient,
		matchRepo:     matchRepo,
		publisher:     publisher,
		metrics:       metrics,
		logger:        logger,
	}
}

// StartPolling begins polling for a match result (T089, T090)
// Called when match status changes to 'started'
func (s *PollerService) StartPolling(ctx context.Context, matchID uuid.UUID, gameMatchID string) error {
	s.logger.Info("Starting Game API polling", map[string]interface{}{
		"match_id":      matchID.String(),
		"game_match_id": gameMatchID,
	})

	pollStart := time.Now()
	pollInterval := 5 * time.Second  // T090: 5-second polling interval
	maxDuration := 60 * time.Second  // T093: 60-second timeout
	pollCount := 0

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	timeout := time.After(maxDuration)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Polling cancelled", map[string]interface{}{
				"match_id":   matchID.String(),
				"poll_count": pollCount,
			})
			return ctx.Err()

		case <-timeout:
			// T093: Timeout - flag for manual host entry
			s.logger.Warn("Game API polling timeout", map[string]interface{}{
				"match_id":   matchID.String(),
				"poll_count": pollCount,
				"duration":   time.Since(pollStart).Seconds(),
			})

			// T098: Audit log for timeout
			s.logPollingAttempt(matchID, gameMatchID, pollCount, nil, "timeout", false)

			// Update match to require manual entry
			// TODO: Add match status for "awaiting_manual_entry" or similar
			s.logger.Info("Match requires manual score entry by host", map[string]interface{}{
				"match_id": matchID.String(),
			})

			return fmt.Errorf("polling timeout after %d attempts", pollCount)

		case <-ticker.C:
			pollCount++

			// T090: Poll Game API
			result, err := s.gameAPIClient.PollMatchResult(ctx, gameMatchID)
			if err != nil {
				s.logger.Warn("Game API poll failed", map[string]interface{}{
					"match_id":    matchID.String(),
					"poll_count":  pollCount,
					"error":       err.Error(),
				})

				// T098: Audit log for failed attempt
				s.logPollingAttempt(matchID, gameMatchID, pollCount, nil, "error", false)

				// Continue polling unless circuit breaker is open
				if s.gameAPIClient.GetCircuitState() == CircuitOpen {
					s.logger.Error("Circuit breaker open, stopping polling", map[string]interface{}{
						"match_id": matchID.String(),
					})
					return fmt.Errorf("circuit breaker open: %w", err)
				}
				continue
			}

			// T098: Audit log for attempt
			s.logPollingAttempt(matchID, gameMatchID, pollCount, result, "success", result != nil)

			if result == nil {
				// Result not ready yet, continue polling
				s.logger.Info("Game API result not ready, continuing polling", map[string]interface{}{
					"match_id":   matchID.String(),
					"poll_count": pollCount,
				})
				continue
			}

			// T091, T092: Process retrieved result
			return s.processResult(ctx, matchID, gameMatchID, result, pollStart, pollCount)
		}
	}
}

// processResult handles a successfully retrieved result from Game API (T091, T092)
func (s *PollerService) processResult(ctx context.Context, matchID uuid.UUID, gameMatchID string, result *MatchResult, pollStart time.Time, pollCount int) error {
	duration := time.Since(pollStart)

	s.logger.Info("Game API result retrieved", map[string]interface{}{
		"match_id":          matchID.String(),
		"game_match_id":     gameMatchID,
		"poll_count":        pollCount,
		"duration_seconds":  duration.Seconds(),
		"fraud_check_passed": result.FraudCheckPassed,
	})

	// T096: Emit latency metric (SC-012: <10s target)
	s.metrics.EmitGameAPIPollingLatency(duration)

	// T095: Emit success rate metric (SC-011: 95% target)
	s.metrics.EmitAutomaticScoreRetrievalRate(1.0) // Success

	// T092: Handle fraud validation failure
	if !result.FraudCheckPassed {
		s.logger.Warn("Fraud check failed, flagging for host review", map[string]interface{}{
			"match_id":            matchID.String(),
			"fraud_check_details": result.FraudCheckDetails,
		})

		// TODO: Update match status to require host review (FR-027)
		// For now, log the event
		s.logger.Info("Match requires host review due to fraud detection", map[string]interface{}{
			"match_id": matchID.String(),
			"details":  result.FraudCheckDetails,
		})

		return fmt.Errorf("fraud check failed: %s", result.FraudCheckDetails)
	}

	// Parse winner ID
	winnerID, err := uuid.Parse(result.WinnerID)
	if err != nil {
		return fmt.Errorf("parse winner ID: %w", err)
	}

	// T097: Publish ScoreRetrieved event
	scoreRetrievedEvent := domain.NewScoreRetrieved(
		matchID,
		matchID, // Using matchID as tournamentID placeholder
		winnerID,
		result.ScorePlayer1,
		result.ScorePlayer2,
		gameMatchID,
		result.CompletedAt,
		map[string]string{
			"correlation_id":     s.getCorrelationID(ctx),
			"poll_count":         fmt.Sprintf("%d", pollCount),
			"retrieval_duration": fmt.Sprintf("%.2fs", duration.Seconds()),
		},
	)
	if err := s.publisher.Publish(ctx, scoreRetrievedEvent); err != nil {
		log.Printf("WARN: Failed to publish ScoreRetrieved event: %v", err)
	}

	s.logger.Info("Score retrieval successful", map[string]interface{}{
		"match_id":      matchID.String(),
		"winner_id":     winnerID.String(),
		"score":         fmt.Sprintf("%d-%d", result.ScorePlayer1, result.ScorePlayer2),
		"poll_count":    pollCount,
		"duration":      duration.Seconds(),
	})

	// T091: Result successfully retrieved and validated
	// The match completion will be handled by the calling service
	return nil
}

// logPollingAttempt logs each polling attempt for audit trail (T098)
func (s *PollerService) logPollingAttempt(matchID uuid.UUID, gameMatchID string, pollCount int, result *MatchResult, status string, hasResult bool) {
	logData := map[string]interface{}{
		"match_id":       matchID.String(),
		"game_match_id":  gameMatchID,
		"poll_count":     pollCount,
		"status":         status,
		"has_result":     hasResult,
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	if result != nil {
		logData["fraud_check_passed"] = result.FraudCheckPassed
		logData["winner_id"] = result.WinnerID
		logData["score"] = fmt.Sprintf("%d-%d", result.ScorePlayer1, result.ScorePlayer2)
	}

	s.logger.Info("Game API polling attempt", logData)
}

// getCorrelationID extracts correlation ID from context
func (s *PollerService) getCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		return correlationID
	}
	return ""
}


