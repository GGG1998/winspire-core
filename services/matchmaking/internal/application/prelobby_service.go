// Package application contains application services and business logic orchestration
package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/pubsub"
	"github.com/winspire-core/services/matchmaking/internal/repository"
	"github.com/winspire-core/services/matchmaking/internal/websocket"
)

// PreLobbyState represents the complete state of a pre-lobby for API responses
type PreLobbyState struct {
	TournamentID     uuid.UUID
	TournamentName   string
	Status           domain.PreLobbyStatus
	Participants     []PreLobbyParticipantInfo
	ParticipantCount int
	MinParticipants  int
	StartTime        time.Time
	GracePeriod      *GracePeriodInfo
	ActivityFeed     []*domain.ActivityFeedItem
}

// PreLobbyParticipantInfo represents a participant in the pre-lobby
type PreLobbyParticipantInfo struct {
	UserID      uuid.UUID
	DisplayName string
	AvatarURL   *string
	JoinedAt    time.Time
}

// GracePeriodInfo represents grace period timing information
type GracePeriodInfo struct {
	StartTime        time.Time
	EndTime          time.Time
	RemainingSeconds int
}

// TournamentInfo contains tournament metadata from competition service
type TournamentInfo struct {
	ID              uuid.UUID
	Name            string
	StartTime       time.Time
	MinParticipants int
	Status          string
}

// PreLobbyService handles pre-lobby operations
type PreLobbyService struct {
	repo      repository.PreLobbyRepository
	hub       *websocket.Hub
	publisher *pubsub.EventPublisher
	metrics   *observability.MetricsEmitter
	logger    *observability.Logger

	// In-memory participant tracking (keyed by tournament ID)
	participants   map[uuid.UUID]map[uuid.UUID]*PreLobbyParticipantInfo
	participantsMu sync.RWMutex

	// Grace period timers (keyed by tournament ID)
	gracePeriodTimers map[uuid.UUID]*time.Timer
	timersMu          sync.Mutex
}

// NewPreLobbyService creates a new pre-lobby service
func NewPreLobbyService(
	repo repository.PreLobbyRepository,
	hub *websocket.Hub,
	publisher *pubsub.EventPublisher,
	metrics *observability.MetricsEmitter,
	logger *observability.Logger,
) *PreLobbyService {
	return &PreLobbyService{
		repo:              repo,
		hub:               hub,
		publisher:         publisher,
		metrics:           metrics,
		logger:            logger,
		participants:      make(map[uuid.UUID]map[uuid.UUID]*PreLobbyParticipantInfo),
		gracePeriodTimers: make(map[uuid.UUID]*time.Timer),
	}
}

// GetOrCreatePreLobby gets or creates a pre-lobby for a tournament
func (s *PreLobbyService) GetOrCreatePreLobby(ctx context.Context, tournamentID uuid.UUID, minParticipants int) (*domain.PreLobby, error) {
	// Try to get existing pre-lobby
	preLobby, err := s.repo.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("get pre-lobby: %w", err)
	}

	if preLobby != nil {
		return preLobby, nil
	}

	// Create new pre-lobby
	preLobby, err = domain.NewPreLobby(tournamentID, minParticipants)
	if err != nil {
		return nil, fmt.Errorf("create pre-lobby domain: %w", err)
	}

	if err := s.repo.Create(ctx, preLobby); err != nil {
		return nil, fmt.Errorf("persist pre-lobby: %w", err)
	}

	// Publish event
	event := domain.NewPreLobbyCreated(
		preLobby.ID,
		preLobby.TournamentID,
		preLobby.MinParticipants,
		preLobby.CreatedAt,
		nil,
	)
	if err := s.publisher.Publish(ctx, event); err != nil {
		s.logger.Warn("failed to publish pre-lobby created event", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
	}

	s.logger.Info("created pre-lobby", map[string]interface{}{
		"tournament_id":    tournamentID.String(),
		"prelobby_id":      preLobby.ID.String(),
		"min_participants": minParticipants,
	})

	return preLobby, nil
}

// GetState returns the current state of a pre-lobby
func (s *PreLobbyService) GetState(ctx context.Context, tournamentID uuid.UUID, tournamentInfo *TournamentInfo) (*PreLobbyState, error) {
	// Get or create pre-lobby
	preLobby, err := s.GetOrCreatePreLobby(ctx, tournamentID, tournamentInfo.MinParticipants)
	if err != nil {
		return nil, fmt.Errorf("get or create pre-lobby: %w", err)
	}

	// Get connected participants from in-memory store
	participants := s.getParticipants(tournamentID)

	// Get activity feed
	activityFeed, err := s.repo.GetRecentActivityFeed(ctx, tournamentID)
	if err != nil {
		s.logger.Warn("failed to get activity feed", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		activityFeed = []*domain.ActivityFeedItem{}
	}

	// Build state response
	state := &PreLobbyState{
		TournamentID:     tournamentID,
		TournamentName:   tournamentInfo.Name,
		Status:           preLobby.Status,
		Participants:     participants,
		ParticipantCount: len(participants),
		MinParticipants:  preLobby.MinParticipants,
		StartTime:        tournamentInfo.StartTime,
		ActivityFeed:     activityFeed,
	}

	// Add grace period info if active
	if preLobby.Status == domain.PreLobbyStatusGracePeriod && preLobby.GracePeriodEnd != nil {
		remaining := int(time.Until(*preLobby.GracePeriodEnd).Seconds())
		if remaining < 0 {
			remaining = 0
		}
		state.GracePeriod = &GracePeriodInfo{
			StartTime:        *preLobby.GracePeriodStart,
			EndTime:          *preLobby.GracePeriodEnd,
			RemainingSeconds: remaining,
		}
	}

	return state, nil
}

// getParticipants returns the list of connected participants for a tournament
func (s *PreLobbyService) getParticipants(tournamentID uuid.UUID) []PreLobbyParticipantInfo {
	s.participantsMu.RLock()
	defer s.participantsMu.RUnlock()

	tournamentParticipants, exists := s.participants[tournamentID]
	if !exists {
		return []PreLobbyParticipantInfo{}
	}

	result := make([]PreLobbyParticipantInfo, 0, len(tournamentParticipants))
	for _, p := range tournamentParticipants {
		result = append(result, *p)
	}

	return result
}

// GetParticipantCount returns the number of connected participants
func (s *PreLobbyService) GetParticipantCount(tournamentID uuid.UUID) int {
	s.participantsMu.RLock()
	defer s.participantsMu.RUnlock()

	if participants, exists := s.participants[tournamentID]; exists {
		return len(participants)
	}
	return 0
}

// AddParticipant adds a participant to the in-memory tracking
func (s *PreLobbyService) AddParticipant(tournamentID uuid.UUID, info *PreLobbyParticipantInfo) {
	s.participantsMu.Lock()
	defer s.participantsMu.Unlock()

	if s.participants[tournamentID] == nil {
		s.participants[tournamentID] = make(map[uuid.UUID]*PreLobbyParticipantInfo)
	}

	s.participants[tournamentID][info.UserID] = info
}

// RemoveParticipant removes a participant from in-memory tracking
func (s *PreLobbyService) RemoveParticipant(tournamentID, userID uuid.UUID) *PreLobbyParticipantInfo {
	s.participantsMu.Lock()
	defer s.participantsMu.Unlock()

	if participants, exists := s.participants[tournamentID]; exists {
		if info, exists := participants[userID]; exists {
			delete(participants, userID)
			// Clean up empty map
			if len(participants) == 0 {
				delete(s.participants, tournamentID)
			}
			return info
		}
	}

	return nil
}

// IsParticipantConnected checks if a participant is connected
func (s *PreLobbyService) IsParticipantConnected(tournamentID, userID uuid.UUID) bool {
	s.participantsMu.RLock()
	defer s.participantsMu.RUnlock()

	if participants, exists := s.participants[tournamentID]; exists {
		_, exists := participants[userID]
		return exists
	}
	return false
}

// GetPreLobbyStatus returns just the status of a pre-lobby
func (s *PreLobbyService) GetPreLobbyStatus(ctx context.Context, tournamentID uuid.UUID) (domain.PreLobbyStatus, error) {
	preLobby, err := s.repo.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return "", fmt.Errorf("get pre-lobby: %w", err)
	}
	if preLobby == nil {
		return domain.PreLobbyStatusWaiting, nil
	}
	return preLobby.Status, nil
}

// CanAcceptParticipants checks if the pre-lobby can accept new participants
func (s *PreLobbyService) CanAcceptParticipants(ctx context.Context, tournamentID uuid.UUID) (bool, error) {
	preLobby, err := s.repo.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return false, fmt.Errorf("get pre-lobby: %w", err)
	}
	if preLobby == nil {
		// No pre-lobby exists yet, so we can accept participants
		return true, nil
	}
	return preLobby.CanAcceptParticipants(), nil
}

// SetHub sets the WebSocket hub (called after hub initialization)
func (s *PreLobbyService) SetHub(hub *websocket.Hub) {
	s.hub = hub
}

// ============================================================================
// Broadcast Methods
// ============================================================================

// BroadcastParticipantJoined broadcasts a participant joined event to all connected clients
func (s *PreLobbyService) BroadcastParticipantJoined(tournamentID uuid.UUID, participant *PreLobbyParticipantInfo) {
	if s.hub == nil {
		s.logger.Warn("hub not set, cannot broadcast participant joined", map[string]interface{}{
			"tournament_id": tournamentID.String(),
		})
		return
	}

	payload := websocket.ParticipantJoinedPayload{
		UserID:      participant.UserID,
		DisplayName: participant.DisplayName,
		AvatarURL:   participant.AvatarURL,
		JoinedAt:    participant.JoinedAt,
	}

	msg, err := websocket.NewMessage(websocket.MessageTypeParticipantJoined, payload)
	if err != nil {
		s.logger.Error("failed to create participant joined message", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return
	}

	s.hub.BroadcastToTournament(tournamentID, msg)

	s.logger.Info("broadcast participant joined", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"user_id":       participant.UserID.String(),
		"display_name":  participant.DisplayName,
	})
}

// BroadcastParticipantLeft broadcasts a participant left event to all connected clients
func (s *PreLobbyService) BroadcastParticipantLeft(tournamentID uuid.UUID, participant *PreLobbyParticipantInfo) {
	if s.hub == nil {
		s.logger.Warn("hub not set, cannot broadcast participant left", map[string]interface{}{
			"tournament_id": tournamentID.String(),
		})
		return
	}

	payload := websocket.ParticipantLeftPayload{
		UserID:      participant.UserID,
		DisplayName: participant.DisplayName,
	}

	msg, err := websocket.NewMessage(websocket.MessageTypeParticipantLeft, payload)
	if err != nil {
		s.logger.Error("failed to create participant left message", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return
	}

	s.hub.BroadcastToTournament(tournamentID, msg)

	s.logger.Info("broadcast participant left", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"user_id":       participant.UserID.String(),
		"display_name":  participant.DisplayName,
	})
}

// BroadcastGracePeriodStarted broadcasts grace period start to all connected clients
func (s *PreLobbyService) BroadcastGracePeriodStarted(tournamentID uuid.UUID, startTime, endTime time.Time, participantCount int) {
	if s.hub == nil {
		return
	}

	payload := websocket.GracePeriodStartedPayload{
		StartTime:        startTime,
		EndTime:          endTime,
		ParticipantCount: participantCount,
	}

	msg, err := websocket.NewMessage(websocket.MessageTypeGracePeriodStarted, payload)
	if err != nil {
		s.logger.Error("failed to create grace period started message", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return
	}

	s.hub.BroadcastToTournament(tournamentID, msg)
}

// BroadcastRosterUpdated broadcasts roster update during grace period
func (s *PreLobbyService) BroadcastRosterUpdated(tournamentID uuid.UUID, action string, userID uuid.UUID, displayName string, participantCount int) {
	if s.hub == nil {
		return
	}

	payload := websocket.RosterUpdatedPayload{
		ParticipantCount: participantCount,
		Action:           action,
		UserID:           userID,
		DisplayName:      displayName,
	}

	msg, err := websocket.NewMessage(websocket.MessageTypeRosterUpdated, payload)
	if err != nil {
		s.logger.Error("failed to create roster updated message", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return
	}

	s.hub.BroadcastToTournament(tournamentID, msg)
}

// BroadcastGracePeriodEnded broadcasts grace period end to all connected clients
func (s *PreLobbyService) BroadcastGracePeriodEnded(tournamentID uuid.UUID, finalParticipantCount int, status string) {
	if s.hub == nil {
		return
	}

	payload := websocket.GracePeriodEndedPayload{
		FinalParticipantCount: finalParticipantCount,
		Status:                status,
	}

	msg, err := websocket.NewMessage(websocket.MessageTypeGracePeriodEnded, payload)
	if err != nil {
		s.logger.Error("failed to create grace period ended message", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return
	}

	s.hub.BroadcastToTournament(tournamentID, msg)
}

// BroadcastTournamentCancelled broadcasts tournament cancellation to all connected clients
func (s *PreLobbyService) BroadcastTournamentCancelled(tournamentID uuid.UUID, reason string) {
	if s.hub == nil {
		return
	}

	payload := websocket.TournamentCancelledPayload{
		Reason: reason,
	}

	msg, err := websocket.NewMessage(websocket.MessageTypeTournamentCancelled, payload)
	if err != nil {
		s.logger.Error("failed to create tournament cancelled message", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return
	}

	s.hub.BroadcastToTournament(tournamentID, msg)
}

// ============================================================================
// Grace Period Management
// ============================================================================

// GracePeriodCallback is called when grace period ends
type GracePeriodCallback func(tournamentID uuid.UUID, participantIDs []uuid.UUID)

// SetGracePeriodCallback sets the callback for when grace period ends
func (s *PreLobbyService) SetGracePeriodCallback(callback GracePeriodCallback) {
	s.timersMu.Lock()
	defer s.timersMu.Unlock()
	// Store callback for use in grace period completion
}

// StartGracePeriod starts the 30-second grace period for a tournament
func (s *PreLobbyService) StartGracePeriod(ctx context.Context, tournamentID uuid.UUID, onComplete GracePeriodCallback) error {
	// Update pre-lobby status in database
	preLobby, err := s.repo.StartGracePeriod(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("start grace period in database: %w", err)
	}

	participantCount := s.GetParticipantCount(tournamentID)

	s.logger.Info("starting grace period", map[string]interface{}{
		"tournament_id":     tournamentID.String(),
		"participant_count": participantCount,
		"grace_period_end":  preLobby.GracePeriodEnd,
	})

	// Record activity feed event
	s.RecordGracePeriodStarted(ctx, tournamentID)

	// Broadcast grace period started
	if preLobby.GracePeriodStart != nil && preLobby.GracePeriodEnd != nil {
		s.BroadcastGracePeriodStarted(tournamentID, *preLobby.GracePeriodStart, *preLobby.GracePeriodEnd, participantCount)
	}

	// Publish events
	if preLobby.GracePeriodStart != nil && preLobby.GracePeriodEnd != nil {
		// Publish PreLobbyGracePeriodStarted (for internal tracking)
		event := domain.NewPreLobbyGracePeriodStarted(
			tournamentID,
			*preLobby.GracePeriodStart,
			*preLobby.GracePeriodEnd,
			participantCount,
			nil,
		)
		if err := s.publisher.Publish(ctx, event); err != nil {
			s.logger.Warn("failed to publish grace period started event", map[string]interface{}{
				"tournament_id": tournamentID.String(),
				"error":         err.Error(),
			})
		}

		// Publish GracePeriodStarted (for saga coordination with competition service)
		sagaEvent := domain.NewGracePeriodStarted(
			tournamentID,
			*preLobby.GracePeriodStart,
			*preLobby.GracePeriodEnd,
			nil,
		)
		if err := s.publisher.Publish(ctx, sagaEvent); err != nil {
			s.logger.Warn("failed to publish GracePeriodStarted saga event", map[string]interface{}{
				"tournament_id": tournamentID.String(),
				"error":         err.Error(),
			})
		}
	}

	// Start timer for grace period completion
	s.timersMu.Lock()
	defer s.timersMu.Unlock()

	// Cancel existing timer if any
	if existingTimer, exists := s.gracePeriodTimers[tournamentID]; exists {
		existingTimer.Stop()
	}

	// Calculate remaining time
	var duration time.Duration
	if preLobby.GracePeriodEnd != nil {
		duration = time.Until(*preLobby.GracePeriodEnd)
		if duration < 0 {
			duration = 0
		}
	} else {
		duration = 30 * time.Second
	}

	// Create timer
	timer := time.AfterFunc(duration, func() {
		s.finalizeGracePeriod(context.Background(), tournamentID, onComplete)
	})
	s.gracePeriodTimers[tournamentID] = timer

	return nil
}

// finalizeGracePeriod completes the grace period and creates participant snapshot
func (s *PreLobbyService) finalizeGracePeriod(ctx context.Context, tournamentID uuid.UUID, onComplete GracePeriodCallback) {
	s.logger.Info("finalizing grace period", map[string]interface{}{
		"tournament_id": tournamentID.String(),
	})

	// Get current pre-lobby
	preLobby, err := s.repo.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		s.logger.Error("failed to get pre-lobby for finalization", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return
	}

	if preLobby == nil || preLobby.Status != domain.PreLobbyStatusGracePeriod {
		s.logger.Warn("pre-lobby not in grace period, skipping finalization", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"status":        preLobby.Status,
		})
		return
	}

	// Get connected participants
	participants := s.getParticipants(tournamentID)
	participantCount := len(participants)

	s.logger.Info("grace period finalization", map[string]interface{}{
		"tournament_id":     tournamentID.String(),
		"participant_count": participantCount,
		"min_participants":  preLobby.MinParticipants,
	})

	// Check minimum participants
	if participantCount < preLobby.MinParticipants {
		// Cancel tournament
		if err := s.cancelTournament(ctx, tournamentID, "Insufficient participants"); err != nil {
			s.logger.Error("failed to cancel tournament", map[string]interface{}{
				"tournament_id": tournamentID.String(),
				"error":         err.Error(),
			})
		}
		return
	}

	// Create participant snapshot
	domainParticipants := make([]domain.PreLobbyParticipant, len(participants))
	participantIDs := make([]uuid.UUID, len(participants))
	for i, p := range participants {
		domainParticipants[i] = domain.PreLobbyParticipant{
			UserID:      p.UserID,
			DisplayName: p.DisplayName,
			AvatarURL:   p.AvatarURL,
			JoinedAt:    p.JoinedAt,
		}
		participantIDs[i] = p.UserID
	}

	snapshot := domain.NewParticipantSnapshot(tournamentID, domainParticipants)
	if err := s.repo.CreateParticipantSnapshot(ctx, snapshot); err != nil {
		s.logger.Error("failed to create participant snapshot", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		// Continue anyway - we can still generate bracket from in-memory data
	}

	// Update status to generating_bracket
	_, err = s.repo.UpdateStatus(ctx, tournamentID, domain.PreLobbyStatusGeneratingBracket)
	if err != nil {
		s.logger.Error("failed to update pre-lobby status to generating_bracket", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
	}

	// Record activity feed event
	s.RecordGracePeriodEnded(ctx, tournamentID, participantCount)

	// Record bracket generation starting
	s.RecordBracketGeneration(ctx, tournamentID)

	// Broadcast grace period ended
	s.BroadcastGracePeriodEnded(tournamentID, participantCount, "generating_bracket")

	// Publish events
	event := domain.NewPreLobbyGracePeriodEnded(tournamentID, participantCount, "generating_bracket", nil)
	if err := s.publisher.Publish(ctx, event); err != nil {
		s.logger.Warn("failed to publish grace period ended event", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
	}

	snapshotEvent := domain.NewPreLobbyParticipantSnapshot(tournamentID, participantIDs, time.Now(), nil)
	if err := s.publisher.Publish(ctx, snapshotEvent); err != nil {
		s.logger.Warn("failed to publish participant snapshot event", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
	}

	// Trigger callback for bracket generation
	if onComplete != nil {
		onComplete(tournamentID, participantIDs)
	}

	// Clean up timer
	s.timersMu.Lock()
	delete(s.gracePeriodTimers, tournamentID)
	s.timersMu.Unlock()

	s.logger.Info("grace period finalized successfully", map[string]interface{}{
		"tournament_id":     tournamentID.String(),
		"participant_count": participantCount,
	})
}

// CancelTournamentWithError cancels the tournament with an error reason (public for SAGA compensation)
func (s *PreLobbyService) CancelTournamentWithError(ctx context.Context, tournamentID uuid.UUID, reason string) error {
	return s.cancelTournament(ctx, tournamentID, reason)
}

// cancelTournament cancels the tournament due to insufficient participants
func (s *PreLobbyService) cancelTournament(ctx context.Context, tournamentID uuid.UUID, reason string) error {
	s.logger.Info("cancelling tournament", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"reason":        reason,
	})

	// Update status to cancelled
	_, err := s.repo.UpdateStatus(ctx, tournamentID, domain.PreLobbyStatusCancelled)
	if err != nil {
		s.logger.Error("failed to update pre-lobby status to cancelled", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Record activity feed event
	s.RecordTournamentCancelled(ctx, tournamentID, reason)

	// Broadcast cancellation
	s.BroadcastTournamentCancelled(tournamentID, reason)
	s.BroadcastGracePeriodEnded(tournamentID, s.GetParticipantCount(tournamentID), "cancelled")

	// Publish event
	event := domain.NewPreLobbyCancelled(tournamentID, reason, time.Now(), nil)
	if err := s.publisher.Publish(ctx, event); err != nil {
		s.logger.Warn("failed to publish pre-lobby cancelled event", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
	}

	// Clean up timer
	s.timersMu.Lock()
	if timer, exists := s.gracePeriodTimers[tournamentID]; exists {
		timer.Stop()
		delete(s.gracePeriodTimers, tournamentID)
	}
	s.timersMu.Unlock()

	return nil
}

// RecoverActiveGracePeriods recovers grace periods that were active when service restarted
func (s *PreLobbyService) RecoverActiveGracePeriods(ctx context.Context, onComplete GracePeriodCallback) error {
	activePreLobbies, err := s.repo.GetActiveGracePeriods(ctx)
	if err != nil {
		return fmt.Errorf("get active grace periods: %w", err)
	}

	s.logger.Info("recovering active grace periods", map[string]interface{}{
		"count": len(activePreLobbies),
	})

	for _, preLobby := range activePreLobbies {
		if preLobby.GracePeriodEnd == nil {
			continue
		}

		remaining := time.Until(*preLobby.GracePeriodEnd)
		if remaining <= 0 {
			// Grace period already expired, finalize immediately
			go s.finalizeGracePeriod(ctx, preLobby.TournamentID, onComplete)
		} else {
			// Set up timer for remaining time
			s.timersMu.Lock()
			timer := time.AfterFunc(remaining, func() {
				s.finalizeGracePeriod(context.Background(), preLobby.TournamentID, onComplete)
			})
			s.gracePeriodTimers[preLobby.TournamentID] = timer
			s.timersMu.Unlock()

			s.logger.Info("recovered grace period timer", map[string]interface{}{
				"tournament_id":    preLobby.TournamentID.String(),
				"remaining_seconds": remaining.Seconds(),
			})
		}
	}

	return nil
}

// IsGracePeriodActive checks if a tournament is currently in grace period
func (s *PreLobbyService) IsGracePeriodActive(ctx context.Context, tournamentID uuid.UUID) (bool, error) {
	preLobby, err := s.repo.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return false, fmt.Errorf("get pre-lobby: %w", err)
	}
	if preLobby == nil {
		return false, nil
	}
	return preLobby.IsInGracePeriod(), nil
}

// ValidateTournamentStatus checks if tournament status allows pre-lobby access
func (s *PreLobbyService) ValidateTournamentStatus(status string) error {
	invalidStatuses := map[string]string{
		"draft":     "Tournament is still in draft state",
		"completed": "Tournament has already completed",
		"cancelled": "Tournament has been cancelled",
	}

	if msg, isInvalid := invalidStatuses[status]; isInvalid {
		return fmt.Errorf("%s", msg)
	}

	return nil
}

// ============================================================================
// Activity Feed Management
// ============================================================================

// RecordParticipantJoined records a participant joined event in the activity feed
func (s *PreLobbyService) RecordParticipantJoined(ctx context.Context, tournamentID uuid.UUID, displayName string) error {
	item := domain.NewActivityFeedItem(
		tournamentID,
		domain.ActivityEventParticipantJoined,
		displayName+" joined the lobby",
		&displayName,
	)

	if err := s.repo.AddActivityFeedEvent(ctx, item); err != nil {
		s.logger.Warn("failed to record participant joined event", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"display_name":  displayName,
			"error":         err.Error(),
		})
		return err
	}

	return nil
}

// RecordParticipantLeft records a participant left event in the activity feed
func (s *PreLobbyService) RecordParticipantLeft(ctx context.Context, tournamentID uuid.UUID, displayName string) error {
	item := domain.NewActivityFeedItem(
		tournamentID,
		domain.ActivityEventParticipantLeft,
		displayName+" left the lobby",
		&displayName,
	)

	if err := s.repo.AddActivityFeedEvent(ctx, item); err != nil {
		s.logger.Warn("failed to record participant left event", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"display_name":  displayName,
			"error":         err.Error(),
		})
		return err
	}

	return nil
}

// RecordGracePeriodStarted records a grace period started event in the activity feed
func (s *PreLobbyService) RecordGracePeriodStarted(ctx context.Context, tournamentID uuid.UUID) error {
	item := domain.NewActivityFeedItem(
		tournamentID,
		domain.ActivityEventGracePeriodStarted,
		"Tournament starting - 30 second grace period begins",
		nil,
	)

	if err := s.repo.AddActivityFeedEvent(ctx, item); err != nil {
		s.logger.Warn("failed to record grace period started event", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return err
	}

	return nil
}

// RecordGracePeriodEnded records a grace period ended event in the activity feed
func (s *PreLobbyService) RecordGracePeriodEnded(ctx context.Context, tournamentID uuid.UUID, participantCount int) error {
	item := domain.NewActivityFeedItem(
		tournamentID,
		domain.ActivityEventGracePeriodEnded,
		fmt.Sprintf("Grace period ended - %d participants confirmed", participantCount),
		nil,
	)

	if err := s.repo.AddActivityFeedEvent(ctx, item); err != nil {
		s.logger.Warn("failed to record grace period ended event", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return err
	}

	return nil
}

// RecordTournamentCancelled records a tournament cancelled event in the activity feed
func (s *PreLobbyService) RecordTournamentCancelled(ctx context.Context, tournamentID uuid.UUID, reason string) error {
	item := domain.NewActivityFeedItem(
		tournamentID,
		domain.ActivityEventTournamentCancelled,
		"Tournament cancelled: "+reason,
		nil,
	)

	if err := s.repo.AddActivityFeedEvent(ctx, item); err != nil {
		s.logger.Warn("failed to record tournament cancelled event", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return err
	}

	return nil
}

// RecordBracketGeneration records a bracket generation event in the activity feed
func (s *PreLobbyService) RecordBracketGeneration(ctx context.Context, tournamentID uuid.UUID) error {
	item := domain.NewActivityFeedItem(
		tournamentID,
		domain.ActivityEventBracketGeneration,
		"Generating tournament bracket...",
		nil,
	)

	if err := s.repo.AddActivityFeedEvent(ctx, item); err != nil {
		s.logger.Warn("failed to record bracket generation event", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return err
	}

	return nil
}

// GetActivityFeed returns the recent activity feed for a tournament
func (s *PreLobbyService) GetActivityFeed(ctx context.Context, tournamentID uuid.UUID) ([]*domain.ActivityFeedItem, error) {
	return s.repo.GetRecentActivityFeed(ctx, tournamentID)
}

// HandleParticipantCountChange should be called when participant count changes during grace period
// It broadcasts roster_updated and checks for edge cases (count drops to 0)
func (s *PreLobbyService) HandleParticipantCountChange(ctx context.Context, tournamentID uuid.UUID, action string, userID uuid.UUID, displayName string) {
	participantCount := s.GetParticipantCount(tournamentID)

	// Broadcast roster update
	s.BroadcastRosterUpdated(tournamentID, action, userID, displayName, participantCount)

	// Edge case: participant count drops to 0 during grace period
	if participantCount == 0 {
		// Check if in grace period
		preLobby, err := s.repo.GetByTournamentID(ctx, tournamentID)
		if err != nil {
			s.logger.Error("failed to get pre-lobby for participant count check", map[string]interface{}{
				"tournament_id": tournamentID.String(),
				"error":         err.Error(),
			})
			return
		}

		if preLobby != nil && preLobby.Status == domain.PreLobbyStatusGracePeriod {
			s.logger.Warn("participant count dropped to 0 during grace period, cancelling tournament", map[string]interface{}{
				"tournament_id": tournamentID.String(),
			})
			if err := s.cancelTournament(ctx, tournamentID, "All participants left during grace period"); err != nil {
				s.logger.Error("failed to cancel tournament", map[string]interface{}{
					"tournament_id": tournamentID.String(),
					"error":         err.Error(),
				})
			}
		}
	}
}

// BroadcastMatchAssigned sends match assignment to a specific participant
func (s *PreLobbyService) BroadcastMatchAssigned(tournamentID, userID, matchID uuid.UUID, roundNumber, matchNumber int, opponent *PreLobbyParticipantInfo) {
	if s.hub == nil {
		return
	}

	var opponentPayload *websocket.PreLobbyParticipantPayload
	if opponent != nil {
		opponentPayload = &websocket.PreLobbyParticipantPayload{
			UserID:      opponent.UserID,
			DisplayName: opponent.DisplayName,
			AvatarURL:   opponent.AvatarURL,
			JoinedAt:    opponent.JoinedAt,
		}
	}

	payload := websocket.MatchAssignedPayload{
		MatchID:     matchID,
		RoundNumber: roundNumber,
		MatchNumber: matchNumber,
		Opponent:    opponentPayload,
	}

	msg, err := websocket.NewMessage(websocket.MessageTypeMatchAssigned, payload)
	if err != nil {
		s.logger.Error("failed to create match assigned message", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"user_id":       userID.String(),
			"error":         err.Error(),
		})
		return
	}

	s.hub.SendToTournamentParticipant(tournamentID, userID, msg)

	s.logger.Info("sent match assignment", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"user_id":       userID.String(),
		"match_id":      matchID.String(),
		"round_number":  roundNumber,
		"match_number":  matchNumber,
	})
}

