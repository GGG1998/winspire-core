// Package websocket provides WebSocket session management with Redis persistence
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SessionInfo represents metadata about a WebSocket session
type SessionInfo struct {
	MatchID     uuid.UUID `json:"match_id"`
	UserID      uuid.UUID `json:"user_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastSeen    time.Time `json:"last_seen"`
	InstanceID  string    `json:"instance_id"` // Identifier for the service instance
}

// SessionStore manages WebSocket session persistence in Redis
type SessionStore struct {
	redis      *redis.Client
	instanceID string
}

// NewSessionStore creates a new SessionStore with Redis backend
func NewSessionStore(redisClient *redis.Client, instanceID string) *SessionStore {
	return &SessionStore{
		redis:      redisClient,
		instanceID: instanceID,
	}
}

// Redis key patterns
const (
	// Set of active user IDs in a match lobby
	keyMatchSessions = "lobby:match:%s:sessions" // Set<userId>
	
	// Heartbeat timestamp for a specific session
	keySessionHeartbeat = "lobby:session:%s:%s:heartbeat" // String (timestamp)
	
	// Session metadata
	keySessionInfo = "lobby:session:%s:%s:info" // JSON
)

// TTL durations
const (
	sessionTTL   = 5 * time.Minute // Session info expires after 5 minutes
	heartbeatTTL = 1 * time.Minute // Heartbeat expires after 1 minute
)

// AddSession adds a new session to Redis
func (s *SessionStore) AddSession(ctx context.Context, matchID, userID uuid.UUID) error {
	// Add user to the match's active sessions set
	sessionsKey := fmt.Sprintf(keyMatchSessions, matchID.String())
	if err := s.redis.SAdd(ctx, sessionsKey, userID.String()).Err(); err != nil {
		return fmt.Errorf("failed to add session to set: %w", err)
	}
	
	// Set TTL on the sessions set
	if err := s.redis.Expire(ctx, sessionsKey, sessionTTL).Err(); err != nil {
		return fmt.Errorf("failed to set TTL on sessions set: %w", err)
	}

	// Store session metadata
	info := SessionInfo{
		MatchID:     matchID,
		UserID:      userID,
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
		InstanceID:  s.instanceID,
	}
	
	infoBytes, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal session info: %w", err)
	}
	
	infoKey := fmt.Sprintf(keySessionInfo, matchID.String(), userID.String())
	if err := s.redis.Set(ctx, infoKey, infoBytes, sessionTTL).Err(); err != nil {
		return fmt.Errorf("failed to set session info: %w", err)
	}

	// Initialize heartbeat
	return s.UpdateHeartbeat(ctx, matchID, userID)
}

// RemoveSession removes a session from Redis
func (s *SessionStore) RemoveSession(ctx context.Context, matchID, userID uuid.UUID) error {
	// Remove user from the match's active sessions set
	sessionsKey := fmt.Sprintf(keyMatchSessions, matchID.String())
	if err := s.redis.SRem(ctx, sessionsKey, userID.String()).Err(); err != nil {
		return fmt.Errorf("failed to remove session from set: %w", err)
	}

	// Delete session info
	infoKey := fmt.Sprintf(keySessionInfo, matchID.String(), userID.String())
	if err := s.redis.Del(ctx, infoKey).Err(); err != nil {
		return fmt.Errorf("failed to delete session info: %w", err)
	}

	// Delete heartbeat
	heartbeatKey := fmt.Sprintf(keySessionHeartbeat, matchID.String(), userID.String())
	if err := s.redis.Del(ctx, heartbeatKey).Err(); err != nil {
		return fmt.Errorf("failed to delete heartbeat: %w", err)
	}

	return nil
}

// GetActiveSessions returns list of active user IDs in a match lobby
func (s *SessionStore) GetActiveSessions(ctx context.Context, matchID uuid.UUID) ([]uuid.UUID, error) {
	sessionsKey := fmt.Sprintf(keyMatchSessions, matchID.String())
	
	members, err := s.redis.SMembers(ctx, sessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	userIDs := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		userID, err := uuid.Parse(member)
		if err != nil {
			// Skip invalid UUIDs
			continue
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// IsSessionActive checks if a user has an active session in a match
func (s *SessionStore) IsSessionActive(ctx context.Context, matchID, userID uuid.UUID) (bool, error) {
	sessionsKey := fmt.Sprintf(keyMatchSessions, matchID.String())
	
	exists, err := s.redis.SIsMember(ctx, sessionsKey, userID.String()).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return exists, nil
}

// GetSessionInfo retrieves session metadata
func (s *SessionStore) GetSessionInfo(ctx context.Context, matchID, userID uuid.UUID) (*SessionInfo, error) {
	infoKey := fmt.Sprintf(keySessionInfo, matchID.String(), userID.String())
	
	infoBytes, err := s.redis.Get(ctx, infoKey).Result()
	if err == redis.Nil {
		return nil, nil // Session not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session info: %w", err)
	}

	var info SessionInfo
	if err := json.Unmarshal([]byte(infoBytes), &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session info: %w", err)
	}

	return &info, nil
}

// UpdateHeartbeat updates the last heartbeat timestamp for a session
func (s *SessionStore) UpdateHeartbeat(ctx context.Context, matchID, userID uuid.UUID) error {
	heartbeatKey := fmt.Sprintf(keySessionHeartbeat, matchID.String(), userID.String())
	timestamp := time.Now().Unix()
	
	if err := s.redis.Set(ctx, heartbeatKey, timestamp, heartbeatTTL).Err(); err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	// Also update LastSeen in session info
	infoKey := fmt.Sprintf(keySessionInfo, matchID.String(), userID.String())
	info, err := s.GetSessionInfo(ctx, matchID, userID)
	if err != nil {
		return fmt.Errorf("failed to get session info for heartbeat update: %w", err)
	}
	
	if info != nil {
		info.LastSeen = time.Now()
		infoBytes, err := json.Marshal(info)
		if err != nil {
			return fmt.Errorf("failed to marshal session info: %w", err)
		}
		
		if err := s.redis.Set(ctx, infoKey, infoBytes, sessionTTL).Err(); err != nil {
			return fmt.Errorf("failed to update session info: %w", err)
		}
	}

	return nil
}

// GetLastHeartbeat retrieves the last heartbeat timestamp
func (s *SessionStore) GetLastHeartbeat(ctx context.Context, matchID, userID uuid.UUID) (time.Time, error) {
	heartbeatKey := fmt.Sprintf(keySessionHeartbeat, matchID.String(), userID.String())
	
	timestamp, err := s.redis.Get(ctx, heartbeatKey).Int64()
	if err == redis.Nil {
		return time.Time{}, nil // No heartbeat found
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get heartbeat: %w", err)
	}

	return time.Unix(timestamp, 0), nil
}

// CleanupStaleSessions removes sessions that haven't sent heartbeat recently
// This is a backup cleanup mechanism in addition to Redis TTL
func (s *SessionStore) CleanupStaleSessions(ctx context.Context, matchID uuid.UUID, timeout time.Duration) error {
	sessions, err := s.GetActiveSessions(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get active sessions: %w", err)
	}

	for _, userID := range sessions {
		lastHeartbeat, err := s.GetLastHeartbeat(ctx, matchID, userID)
		if err != nil {
			continue // Skip on error
		}

		if time.Since(lastHeartbeat) > timeout {
			// Session is stale, remove it
			if err := s.RemoveSession(ctx, matchID, userID); err != nil {
				// Log but continue
				continue
			}
		}
	}

	return nil
}

