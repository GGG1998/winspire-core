// Package application contains application services and business logic orchestration
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/observability"
)

// CompetitionClient is an HTTP client for the competition service
type CompetitionClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *observability.Logger

	// Cache for registration status (30-second TTL)
	registrationCache   map[string]registrationCacheEntry
	registrationCacheMu sync.RWMutex
}

type registrationCacheEntry struct {
	isRegistered bool
	expiresAt    time.Time
}

// NewCompetitionClient creates a new competition service client
func NewCompetitionClient(baseURL string, logger *observability.Logger) *CompetitionClient {
	return &CompetitionClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 500 * time.Millisecond, // 500ms timeout per spec
		},
		logger:            logger,
		registrationCache: make(map[string]registrationCacheEntry),
	}
}

// GetTournamentInfo fetches tournament information from the competition service
func (c *CompetitionClient) GetTournamentInfo(ctx context.Context, tournamentID uuid.UUID) (*TournamentInfo, error) {
	url := fmt.Sprintf("%s/internal/tournaments/%s", c.baseURL, tournamentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("competition service request failed", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tournamentResp struct {
		ID              string    `json:"id"`
		Name            string    `json:"name"`
		StartTime       time.Time `json:"scheduledStartTimeAt"`
		MinParticipants int       `json:"minParticipants"`
		Status          string    `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tournamentResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	parsedID, err := uuid.Parse(tournamentResp.ID)
	if err != nil {
		return nil, fmt.Errorf("parse tournament ID: %w", err)
	}

	// Default min participants if not set
	minParticipants := tournamentResp.MinParticipants
	if minParticipants < 2 {
		minParticipants = 2
	}

	return &TournamentInfo{
		ID:              parsedID,
		Name:            tournamentResp.Name,
		StartTime:       tournamentResp.StartTime,
		MinParticipants: minParticipants,
		Status:          tournamentResp.Status,
	}, nil
}

// IsUserRegistered checks if a user is registered for a tournament
func (c *CompetitionClient) IsUserRegistered(ctx context.Context, tournamentID, userID uuid.UUID) (bool, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s", tournamentID, userID)
	c.registrationCacheMu.RLock()
	if entry, exists := c.registrationCache[cacheKey]; exists && time.Now().Before(entry.expiresAt) {
		c.registrationCacheMu.RUnlock()
		return entry.isRegistered, nil
	}
	c.registrationCacheMu.RUnlock()

	// Fetch from competition service
	url := fmt.Sprintf("%s/internal/tournaments/%s/registrations/%s", c.baseURL, tournamentID, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("competition service registration check failed", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"user_id":       userID.String(),
			"error":         err.Error(),
		})
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	isRegistered := resp.StatusCode == http.StatusOK

	// Cache the result for 30 seconds
	c.registrationCacheMu.Lock()
	c.registrationCache[cacheKey] = registrationCacheEntry{
		isRegistered: isRegistered,
		expiresAt:    time.Now().Add(30 * time.Second),
	}
	c.registrationCacheMu.Unlock()

	return isRegistered, nil
}

// ClearCache clears the registration cache (useful for testing)
func (c *CompetitionClient) ClearCache() {
	c.registrationCacheMu.Lock()
	defer c.registrationCacheMu.Unlock()
	c.registrationCache = make(map[string]registrationCacheEntry)
}
