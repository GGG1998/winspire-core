// Package gameapi provides integration with the external Game API for score retrieval
package gameapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// MatchResult represents a validated match result from the Game API
type MatchResult struct {
	MatchID           string    `json:"match_id"`
	WinnerID          string    `json:"winner_id"`
	LoserID           string    `json:"loser_id"`
	ScorePlayer1      int       `json:"score_player1"`
	ScorePlayer2      int       `json:"score_player2"`
	CompletedAt       time.Time `json:"completed_at"`
	FraudCheckPassed  bool      `json:"fraud_check_passed"`
	FraudCheckDetails string    `json:"fraud_check_details,omitempty"`
}

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

// GameAPIClient handles communication with the Game API
type GameAPIClient struct {
	baseURL    string
	httpClient *http.Client

	// T094: Circuit breaker for Game API calls
	circuitState        CircuitBreakerState
	consecutiveFailures int
	lastFailureTime     time.Time
	circuitOpenDuration time.Duration
	failureThreshold    int
	mu                  sync.RWMutex
}

// NewGameAPIClient creates a new Game API client
func NewGameAPIClient(baseURL string) *GameAPIClient {
	return &GameAPIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		circuitState:        CircuitClosed,
		consecutiveFailures: 0,
		circuitOpenDuration: 30 * time.Second, // Circuit stays open for 30s before retry
		failureThreshold:    5,                // Open after 5 consecutive failures
	}
}

// PollMatchResult polls the Game API for a validated match result (T088, T090)
// Returns the result if available, nil if not yet ready, error if polling failed
func (c *GameAPIClient) PollMatchResult(ctx context.Context, gameMatchID string) (*MatchResult, error) {
	// T094: Check circuit breaker state
	if !c.canMakeRequest() {
		return nil, fmt.Errorf("circuit breaker open: Game API unavailable")
	}

	// Build request URL: GET /api/matches/:id/result
	url := fmt.Sprintf("%s/api/matches/%s/result", c.baseURL, gameMatchID)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.recordFailure()
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Winspire-Matchmaking/1.0")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.recordFailure()
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// Handle response status codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Result is ready
		var result MatchResult
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.recordFailure()
			return nil, fmt.Errorf("read response body: %w", err)
		}

		if err := json.Unmarshal(body, &result); err != nil {
			c.recordFailure()
			return nil, fmt.Errorf("unmarshal result: %w", err)
		}

		c.recordSuccess()
		return &result, nil

	case http.StatusNotFound:
		// Match not found in Game API yet
		c.recordSuccess() // Not a failure, just not ready
		return nil, nil

	case http.StatusAccepted:
		// Match is still in progress
		c.recordSuccess() // Not a failure
		return nil, nil

	case http.StatusTooManyRequests:
		// Rate limited
		c.recordFailure()
		return nil, fmt.Errorf("rate limited by Game API")

	default:
		// Unexpected error
		c.recordFailure()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
}

// canMakeRequest checks if circuit breaker allows requests (T094)
func (c *GameAPIClient) canMakeRequest() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	switch c.circuitState {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if enough time has passed to try half-open
		if time.Since(c.lastFailureTime) > c.circuitOpenDuration {
			c.mu.RUnlock()
			c.mu.Lock()
			c.circuitState = CircuitHalfOpen
			c.mu.Unlock()
			c.mu.RLock()
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// recordSuccess records a successful API call (T094)
func (c *GameAPIClient) recordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.consecutiveFailures = 0
	if c.circuitState == CircuitHalfOpen {
		c.circuitState = CircuitClosed
		log.Println("[GameAPI] Circuit breaker closed (recovered)")
	}
}

// recordFailure records a failed API call and updates circuit breaker (T094)
func (c *GameAPIClient) recordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.consecutiveFailures++
	c.lastFailureTime = time.Now()

	// Open circuit if threshold exceeded
	if c.consecutiveFailures >= c.failureThreshold {
		c.circuitState = CircuitOpen
		log.Printf("[GameAPI] Circuit breaker opened after %d consecutive failures", c.consecutiveFailures)
	}
}

// GetCircuitState returns the current circuit breaker state
func (c *GameAPIClient) GetCircuitState() CircuitBreakerState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.circuitState
}

// GetConsecutiveFailures returns the current count of consecutive failures
func (c *GameAPIClient) GetConsecutiveFailures() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.consecutiveFailures
}
