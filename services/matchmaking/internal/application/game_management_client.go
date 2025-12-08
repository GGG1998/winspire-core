// Package application contains application services and business logic orchestration
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/observability"
)

// GameManagementClient is an HTTP client for the game-management service
type GameManagementClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *observability.Logger
}

// GameInfo represents game information from game-management service
type GameInfo struct {
	ID          uuid.UUID
	Slug        string
	Name        string
	Version     string
	StoragePath string
}

// NewGameManagementClient creates a new game-management service client
func NewGameManagementClient(baseURL string, logger *observability.Logger) *GameManagementClient {
	return &GameManagementClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 1 * time.Second,
		},
		logger: logger,
	}
}

// GetGameByID fetches game information from game-management service
func (c *GameManagementClient) GetGameByID(ctx context.Context, gameID uuid.UUID) (*GameInfo, error) {
	url := fmt.Sprintf("%s/v1/games/%s", c.baseURL, gameID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("game-management service request failed", map[string]interface{}{
			"game_id": gameID.String(),
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var gameResp struct {
		ID      string `json:"id"`
		Slug    string `json:"slug"`
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&gameResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	parsedID, err := uuid.Parse(gameResp.ID)
	if err != nil {
		return nil, fmt.Errorf("parse game ID: %w", err)
	}

	return &GameInfo{
		ID:      parsedID,
		Slug:    gameResp.Slug,
		Name:    gameResp.Name,
		Version: gameResp.Version,
	}, nil
}
