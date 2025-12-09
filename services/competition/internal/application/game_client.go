package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/winspire/competition/internal/repository"
)

// GameManagementClient is an HTTP client for the game-management service
type GameManagementClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewGameManagementClient creates a new game-management service client
func NewGameManagementClient(baseURL string) *GameManagementClient {
	return &GameManagementClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 2 * time.Second, // Reasonable timeout for internal service calls
		},
	}
}

// GetGameByID fetches game information from game-management service and returns a GameSnapshot
func (c *GameManagementClient) GetGameByID(ctx context.Context, gameID uuid.UUID) (*repository.GameSnapshot, error) {
	url := fmt.Sprintf("%s/v1/games/%s", c.baseURL, gameID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
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
		ID          string  `json:"id"`
		Slug        string  `json:"slug"`
		Name        string  `json:"name"`
		Version     string  `json:"version"`
		LogoURL     *string `json:"logoUrl"`
		Description *string `json:"description"`
		StoragePath string  `json:"storagePath"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&gameResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	parsedID, err := uuid.Parse(gameResp.ID)
	if err != nil {
		return nil, fmt.Errorf("parse game ID: %w", err)
	}

	return &repository.GameSnapshot{
		ID:          parsedID,
		Slug:        gameResp.Slug,
		Name:        gameResp.Name,
		Version:     gameResp.Version,
		LogoURL:     gameResp.LogoURL,
		Description: gameResp.Description,
		StoragePath: gameResp.StoragePath,
	}, nil
}

