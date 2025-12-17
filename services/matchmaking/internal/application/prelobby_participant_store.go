// Package application provides application services for matchmaking
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// PreLobbyParticipantStore defines the interface for participant storage
type PreLobbyParticipantStore interface {
	// AddParticipant adds a participant to a tournament's pre-lobby
	AddParticipant(ctx context.Context, tournamentID uuid.UUID, info *PreLobbyParticipantInfo) error

	// RemoveParticipant removes a participant and returns their info (nil if not found)
	RemoveParticipant(ctx context.Context, tournamentID, userID uuid.UUID) (*PreLobbyParticipantInfo, error)

	// IsParticipantConnected checks if a participant exists in the pre-lobby
	IsParticipantConnected(ctx context.Context, tournamentID, userID uuid.UUID) (bool, error)

	// GetParticipantCount returns the number of participants in a pre-lobby
	GetParticipantCount(ctx context.Context, tournamentID uuid.UUID) (int, error)

	// GetParticipants returns all participants in a pre-lobby
	GetParticipants(ctx context.Context, tournamentID uuid.UUID) ([]PreLobbyParticipantInfo, error)

	// GetParticipantDetails returns details for a specific participant (nil if not found)
	GetParticipantDetails(ctx context.Context, tournamentID, userID uuid.UUID) (*PreLobbyParticipantInfo, error)

	// RefreshTTL refreshes the TTL for a participant (call on heartbeat)
	RefreshTTL(ctx context.Context, tournamentID, userID uuid.UUID) error
}

// Redis key patterns for pre-lobby participants
const (
	// Set of participant IDs in a tournament pre-lobby
	keyPreLobbyParticipants = "prelobby:%s:participants" // Set<userID>

	// Individual participant metadata
	keyPreLobbyParticipant = "prelobby:%s:participant:%s" // JSON string
)

// TTL for participant data (10 minutes as per requirements)
const prelobbyParticipantTTL = 10 * time.Minute

// RedisPreLobbyParticipantStore implements PreLobbyParticipantStore using Redis
type RedisPreLobbyParticipantStore struct {
	redis *redis.Client
}

// NewRedisPreLobbyParticipantStore creates a new Redis-backed participant store
func NewRedisPreLobbyParticipantStore(redisClient *redis.Client) *RedisPreLobbyParticipantStore {
	return &RedisPreLobbyParticipantStore{
		redis: redisClient,
	}
}

// AddParticipant adds a participant to the pre-lobby
func (s *RedisPreLobbyParticipantStore) AddParticipant(ctx context.Context, tournamentID uuid.UUID, info *PreLobbyParticipantInfo) error {
	setKey := fmt.Sprintf(keyPreLobbyParticipants, tournamentID.String())
	infoKey := fmt.Sprintf(keyPreLobbyParticipant, tournamentID.String(), info.UserID.String())

	// Marshal participant info
	infoBytes, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal participant info: %w", err)
	}

	// Use pipeline for atomic operations
	pipe := s.redis.Pipeline()
	pipe.SAdd(ctx, setKey, info.UserID.String())
	pipe.Expire(ctx, setKey, prelobbyParticipantTTL)
	pipe.Set(ctx, infoKey, infoBytes, prelobbyParticipantTTL)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("add participant to redis: %w", err)
	}

	return nil
}

// RemoveParticipant removes a participant and returns their info
func (s *RedisPreLobbyParticipantStore) RemoveParticipant(ctx context.Context, tournamentID, userID uuid.UUID) (*PreLobbyParticipantInfo, error) {
	// Get participant info before removal
	info, err := s.GetParticipantDetails(ctx, tournamentID, userID)
	if err != nil {
		return nil, err
	}

	if info == nil {
		return nil, nil // Not found
	}

	setKey := fmt.Sprintf(keyPreLobbyParticipants, tournamentID.String())
	infoKey := fmt.Sprintf(keyPreLobbyParticipant, tournamentID.String(), userID.String())

	// Use pipeline for atomic removal
	pipe := s.redis.Pipeline()
	pipe.SRem(ctx, setKey, userID.String())
	pipe.Del(ctx, infoKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("remove participant from redis: %w", err)
	}

	return info, nil
}

// IsParticipantConnected checks if a participant exists in the pre-lobby
func (s *RedisPreLobbyParticipantStore) IsParticipantConnected(ctx context.Context, tournamentID, userID uuid.UUID) (bool, error) {
	setKey := fmt.Sprintf(keyPreLobbyParticipants, tournamentID.String())
	exists, err := s.redis.SIsMember(ctx, setKey, userID.String()).Result()
	if err != nil {
		return false, fmt.Errorf("check participant membership: %w", err)
	}
	return exists, nil
}

// GetParticipantCount returns the number of participants
func (s *RedisPreLobbyParticipantStore) GetParticipantCount(ctx context.Context, tournamentID uuid.UUID) (int, error) {
	setKey := fmt.Sprintf(keyPreLobbyParticipants, tournamentID.String())
	count, err := s.redis.SCard(ctx, setKey).Result()
	if err != nil {
		return 0, fmt.Errorf("get participant count: %w", err)
	}
	return int(count), nil
}

// GetParticipants returns all participants in a pre-lobby
func (s *RedisPreLobbyParticipantStore) GetParticipants(ctx context.Context, tournamentID uuid.UUID) ([]PreLobbyParticipantInfo, error) {
	setKey := fmt.Sprintf(keyPreLobbyParticipants, tournamentID.String())

	// Get all member IDs
	members, err := s.redis.SMembers(ctx, setKey).Result()
	if err != nil {
		return nil, fmt.Errorf("get participant IDs: %w", err)
	}

	if len(members) == 0 {
		return []PreLobbyParticipantInfo{}, nil
	}

	// Get details for each participant using pipeline for efficiency
	pipe := s.redis.Pipeline()
	cmds := make([]*redis.StringCmd, len(members))
	for i, memberID := range members {
		infoKey := fmt.Sprintf(keyPreLobbyParticipant, tournamentID.String(), memberID)
		cmds[i] = pipe.Get(ctx, infoKey)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// Some keys may not exist, that's okay - we'll skip them
	}

	result := make([]PreLobbyParticipantInfo, 0, len(members))
	for _, cmd := range cmds {
		infoBytes, err := cmd.Result()
		if err == redis.Nil {
			continue // Skip missing info
		}
		if err != nil {
			continue // Skip on error
		}

		var info PreLobbyParticipantInfo
		if err := json.Unmarshal([]byte(infoBytes), &info); err != nil {
			continue // Skip invalid JSON
		}

		result = append(result, info)
	}

	return result, nil
}

// GetParticipantDetails returns details for a specific participant
func (s *RedisPreLobbyParticipantStore) GetParticipantDetails(ctx context.Context, tournamentID, userID uuid.UUID) (*PreLobbyParticipantInfo, error) {
	infoKey := fmt.Sprintf(keyPreLobbyParticipant, tournamentID.String(), userID.String())

	infoBytes, err := s.redis.Get(ctx, infoKey).Result()
	if err == redis.Nil {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("get participant info: %w", err)
	}

	var info PreLobbyParticipantInfo
	if err := json.Unmarshal([]byte(infoBytes), &info); err != nil {
		return nil, fmt.Errorf("unmarshal participant info: %w", err)
	}

	return &info, nil
}

// RefreshTTL refreshes the TTL for a participant (call on heartbeat/activity)
func (s *RedisPreLobbyParticipantStore) RefreshTTL(ctx context.Context, tournamentID, userID uuid.UUID) error {
	setKey := fmt.Sprintf(keyPreLobbyParticipants, tournamentID.String())
	infoKey := fmt.Sprintf(keyPreLobbyParticipant, tournamentID.String(), userID.String())

	// Refresh TTL on both keys using a pipeline for efficiency
	pipe := s.redis.Pipeline()
	pipe.Expire(ctx, setKey, prelobbyParticipantTTL)
	pipe.Expire(ctx, infoKey, prelobbyParticipantTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("refresh participant TTL: %w", err)
	}

	return nil
}

// InMemoryPreLobbyParticipantStore implements PreLobbyParticipantStore for testing
type InMemoryPreLobbyParticipantStore struct {
	participants map[uuid.UUID]map[uuid.UUID]*PreLobbyParticipantInfo
}

// NewInMemoryPreLobbyParticipantStore creates a new in-memory participant store for testing
func NewInMemoryPreLobbyParticipantStore() *InMemoryPreLobbyParticipantStore {
	return &InMemoryPreLobbyParticipantStore{
		participants: make(map[uuid.UUID]map[uuid.UUID]*PreLobbyParticipantInfo),
	}
}

func (s *InMemoryPreLobbyParticipantStore) AddParticipant(_ context.Context, tournamentID uuid.UUID, info *PreLobbyParticipantInfo) error {
	if s.participants[tournamentID] == nil {
		s.participants[tournamentID] = make(map[uuid.UUID]*PreLobbyParticipantInfo)
	}
	s.participants[tournamentID][info.UserID] = info
	return nil
}

func (s *InMemoryPreLobbyParticipantStore) RemoveParticipant(_ context.Context, tournamentID, userID uuid.UUID) (*PreLobbyParticipantInfo, error) {
	if participants, exists := s.participants[tournamentID]; exists {
		if info, exists := participants[userID]; exists {
			delete(participants, userID)
			if len(participants) == 0 {
				delete(s.participants, tournamentID)
			}
			return info, nil
		}
	}
	return nil, nil
}

func (s *InMemoryPreLobbyParticipantStore) IsParticipantConnected(_ context.Context, tournamentID, userID uuid.UUID) (bool, error) {
	if participants, exists := s.participants[tournamentID]; exists {
		_, exists := participants[userID]
		return exists, nil
	}
	return false, nil
}

func (s *InMemoryPreLobbyParticipantStore) GetParticipantCount(_ context.Context, tournamentID uuid.UUID) (int, error) {
	if participants, exists := s.participants[tournamentID]; exists {
		return len(participants), nil
	}
	return 0, nil
}

func (s *InMemoryPreLobbyParticipantStore) GetParticipants(_ context.Context, tournamentID uuid.UUID) ([]PreLobbyParticipantInfo, error) {
	tournamentParticipants, exists := s.participants[tournamentID]
	if !exists {
		return []PreLobbyParticipantInfo{}, nil
	}
	result := make([]PreLobbyParticipantInfo, 0, len(tournamentParticipants))
	for _, p := range tournamentParticipants {
		result = append(result, *p)
	}
	return result, nil
}

func (s *InMemoryPreLobbyParticipantStore) GetParticipantDetails(_ context.Context, tournamentID, userID uuid.UUID) (*PreLobbyParticipantInfo, error) {
	if participants, exists := s.participants[tournamentID]; exists {
		if info, found := participants[userID]; found {
			return info, nil
		}
	}
	return nil, nil
}

func (s *InMemoryPreLobbyParticipantStore) RefreshTTL(_ context.Context, _, _ uuid.UUID) error {
	return nil // No-op for in-memory store
}
