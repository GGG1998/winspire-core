// Package application contains business logic for the matchmaking service
package application

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/pubsub"
	"github.com/winspire-core/services/matchmaking/internal/repository"
)

// DisconnectService handles player disconnection and reconnection logic
type DisconnectService struct {
	matchRepo     repository.MatchRepository
	matchService  *MatchService
	publisher     *pubsub.EventPublisher
	logger        *observability.Logger
	metrics       *observability.MetricsEmitter

	// Track active reconnection timers
	reconnectTimers map[uuid.UUID]*time.Timer
	timersMu        sync.RWMutex
}

// NewDisconnectService creates a new disconnect service
func NewDisconnectService(
	matchRepo repository.MatchRepository,
	matchService *MatchService,
	publisher *pubsub.EventPublisher,
	logger *observability.Logger,
	metrics *observability.MetricsEmitter,
) *DisconnectService {
	return &DisconnectService{
		matchRepo:       matchRepo,
		matchService:    matchService,
		publisher:       publisher,
		logger:          logger,
		metrics:         metrics,
		reconnectTimers: make(map[uuid.UUID]*time.Timer),
	}
}

// HandleDisconnect handles a player disconnection (T112, T113, T114, T115)
// Implements CS:GO-style disconnect: opponent gets +1 point, 30s timer starts
func (s *DisconnectService) HandleDisconnect(ctx context.Context, matchID, playerID uuid.UUID) error {
	s.logger.Info("Handling player disconnect", map[string]interface{}{
		"match_id":  matchID.String(),
		"player_id": playerID.String(),
	})

	// Fetch match
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("get match: %w", err)
	}

	// Only handle disconnects for active matches
	if match.Status != domain.MatchStatusStarted {
		s.logger.Info("Match not active, ignoring disconnect", map[string]interface{}{
			"match_id":     matchID.String(),
			"match_status": match.Status,
		})
		return nil
	}

	// Determine opponent
	var opponentID uuid.UUID
	if playerID == match.Participant1ID && match.Participant2ID != nil {
		opponentID = *match.Participant2ID
	} else if match.Participant2ID != nil && playerID == *match.Participant2ID {
		opponentID = match.Participant1ID
	} else {
		return fmt.Errorf("player %s is not a participant in match %s", playerID, matchID)
	}

	// T113: Award +1 point to opponent (FR-016b)
	if match.ScorePlayer1 == nil {
		zero := 0
		match.ScorePlayer1 = &zero
	}
	if match.ScorePlayer2 == nil {
		zero := 0
		match.ScorePlayer2 = &zero
	}

	if opponentID == match.Participant1ID {
		*match.ScorePlayer1++
	} else {
		*match.ScorePlayer2++
	}

	// Update score in database
	err = s.matchRepo.UpdateScore(ctx, matchID, *match.ScorePlayer1, *match.ScorePlayer2)
	if err != nil {
		return fmt.Errorf("update score: %w", err)
	}

	s.logger.Info("Awarded point to opponent on disconnect", map[string]interface{}{
		"match_id":    matchID.String(),
		"opponent_id": opponentID.String(),
		"score":       fmt.Sprintf("%d-%d", *match.ScorePlayer1, *match.ScorePlayer2),
	})

	// T114: Pause match (set status='paused', store disconnected_at)
	now := time.Now()
	err = s.matchRepo.UpdateStatus(ctx, matchID, domain.MatchStatusPaused)
	if err != nil {
		return fmt.Errorf("pause match: %w", err)
	}

	err = s.matchRepo.UpdateDisconnectedPlayer(ctx, matchID, playerID, now)
	if err != nil {
		return fmt.Errorf("update disconnect info: %w", err)
	}

	// T119: Publish PlayerConnectionLost event
	disconnectEvent := domain.NewPlayerConnectionLost(
		matchID,
		matchID, // Using matchID as tournamentID placeholder
		playerID,
		now,
		true, // point awarded
		map[string]string{
			"correlation_id": s.getCorrelationID(ctx),
			"opponent_id":    opponentID.String(),
		},
	)
	if err := s.publisher.Publish(ctx, disconnectEvent); err != nil {
		log.Printf("WARN: Failed to publish PlayerConnectionLost event: %v", err)
	}

	// T115: Start 30-second reconnection timer (FR-016c)
	s.startReconnectTimer(matchID, playerID, 30*time.Second)

	// T122: Log disconnect event for observability
	s.logger.Info("Player disconnected, 30s reconnect timer started", map[string]interface{}{
		"match_id":        matchID.String(),
		"player_id":       playerID.String(),
		"disconnected_at": now.Format(time.RFC3339),
		"timeout_seconds": 30,
	})

	return nil
}

// HandleReconnect handles a player reconnection (T116)
func (s *DisconnectService) HandleReconnect(ctx context.Context, matchID, playerID uuid.UUID) error {
	s.logger.Info("Handling player reconnect", map[string]interface{}{
		"match_id":  matchID.String(),
		"player_id": playerID.String(),
	})

	// Fetch match
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("get match: %w", err)
	}

	// Verify match is paused
	if match.Status != domain.MatchStatusPaused {
		s.logger.Info("Match not paused, ignoring reconnect", map[string]interface{}{
			"match_id":     matchID.String(),
			"match_status": match.Status,
		})
		return nil
	}

	// Verify player is the disconnected player
	if match.DisconnectedPlayerID == nil || *match.DisconnectedPlayerID != playerID {
		return fmt.Errorf("player %s is not the disconnected player", playerID)
	}

	// Check if reconnect is within 30-second window (FR-016f)
	if match.DisconnectedAt != nil {
		elapsed := time.Since(*match.DisconnectedAt)
		if elapsed > 30*time.Second {
			s.logger.Warn("Reconnect attempt after 30s timeout", map[string]interface{}{
				"match_id":       matchID.String(),
				"player_id":      playerID.String(),
				"elapsed_seconds": elapsed.Seconds(),
			})
			return fmt.Errorf("reconnect timeout exceeded: %v", elapsed)
		}

		s.logger.Info("Player reconnected within timeout", map[string]interface{}{
			"match_id":        matchID.String(),
			"player_id":       playerID.String(),
			"elapsed_seconds": elapsed.Seconds(),
		})
	}

	// Cancel reconnection timer
	s.cancelReconnectTimer(matchID)

	// T116: Resume match (FR-016f)
	err = s.matchRepo.UpdateStatus(ctx, matchID, domain.MatchStatusStarted)
	if err != nil {
		return fmt.Errorf("resume match: %w", err)
	}

	// Clear disconnect info
	err = s.matchRepo.ClearDisconnectedPlayer(ctx, matchID)
	if err != nil {
		return fmt.Errorf("clear disconnect info: %w", err)
	}

	// T120: Publish PlayerConnectionRestored event
	now := time.Now()
	elapsedSeconds := 0
	if match.DisconnectedAt != nil {
		elapsedSeconds = int(now.Sub(*match.DisconnectedAt).Seconds())
	}
	restoreEvent := domain.NewPlayerConnectionRestored(
		matchID,
		matchID, // Using matchID as tournamentID placeholder
		playerID,
		now,
		elapsedSeconds,
		map[string]string{
			"correlation_id": s.getCorrelationID(ctx),
		},
	)
	if err := s.publisher.Publish(ctx, restoreEvent); err != nil {
		log.Printf("WARN: Failed to publish PlayerConnectionRestored event: %v", err)
	}

	s.logger.Info("Player reconnected successfully, match resumed", map[string]interface{}{
		"match_id":  matchID.String(),
		"player_id": playerID.String(),
	})

	return nil
}

// startReconnectTimer starts a 30-second timer for reconnection (T115)
func (s *DisconnectService) startReconnectTimer(matchID, playerID uuid.UUID, timeout time.Duration) {
	s.timersMu.Lock()
	defer s.timersMu.Unlock()

	// Cancel existing timer if any
	if existingTimer, exists := s.reconnectTimers[matchID]; exists {
		existingTimer.Stop()
	}

	// Create new timer (T117: disqualification on timeout)
	timer := time.AfterFunc(timeout, func() {
		ctx := context.Background()
		s.handleReconnectTimeout(ctx, matchID, playerID)
	})

	s.reconnectTimers[matchID] = timer

	s.logger.Info("Reconnect timer started", map[string]interface{}{
		"match_id":        matchID.String(),
		"player_id":       playerID.String(),
		"timeout_seconds": timeout.Seconds(),
	})
}

// cancelReconnectTimer cancels an active reconnection timer
func (s *DisconnectService) cancelReconnectTimer(matchID uuid.UUID) {
	s.timersMu.Lock()
	defer s.timersMu.Unlock()

	if timer, exists := s.reconnectTimers[matchID]; exists {
		timer.Stop()
		delete(s.reconnectTimers, matchID)
		s.logger.Info("Reconnect timer cancelled", map[string]interface{}{
			"match_id": matchID.String(),
		})
	}
}

// handleReconnectTimeout handles the case when player doesn't reconnect within 30s (T117)
// Grant walkover to opponent per FR-016d
func (s *DisconnectService) handleReconnectTimeout(ctx context.Context, matchID, disconnectedPlayerID uuid.UUID) {
	s.logger.Warn("Reconnect timeout, granting walkover", map[string]interface{}{
		"match_id":  matchID.String(),
		"player_id": disconnectedPlayerID.String(),
	})

	// Fetch match
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		s.logger.Error("Failed to fetch match for reconnect timeout", map[string]interface{}{
			"match_id": matchID.String(),
			"error":    err.Error(),
		})
		return
	}

	// Determine opponent (winner)
	var opponentID uuid.UUID
	if disconnectedPlayerID == match.Participant1ID && match.Participant2ID != nil {
		opponentID = *match.Participant2ID
	} else if match.Participant2ID != nil && disconnectedPlayerID == *match.Participant2ID {
		opponentID = match.Participant1ID
	} else {
		s.logger.Error("Invalid participant for reconnect timeout", map[string]interface{}{
			"match_id":  matchID.String(),
			"player_id": disconnectedPlayerID.String(),
		})
		return
	}

	// T117: Grant walkover to opponent (FR-016d)
	err = s.matchService.GrantWalkover(ctx, matchID, opponentID, disconnectedPlayerID, "disconnect timeout (30s)")
	if err != nil {
		s.logger.Error("Failed to grant walkover on reconnect timeout", map[string]interface{}{
			"match_id": matchID.String(),
			"error":    err.Error(),
		})
		return
	}

	// Clean up timer
	s.timersMu.Lock()
	delete(s.reconnectTimers, matchID)
	s.timersMu.Unlock()

	s.logger.Info("Walkover granted due to disconnect timeout", map[string]interface{}{
		"match_id":  matchID.String(),
		"winner_id": opponentID.String(),
		"loser_id":  disconnectedPlayerID.String(),
	})
}

// getCorrelationID extracts correlation ID from context
func (s *DisconnectService) getCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		return correlationID
	}
	return ""
}

