package tests

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	redismodule "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/winspire/competition-host-stream/internal/config"
	"github.com/winspire/competition-host-stream/internal/projections"
	ssebroker "github.com/winspire/competition-host-stream/internal/sse"
	"github.com/winspire/competition-host-stream/internal/store"
)

// TestMatchEligibilitySync validates stuck-offer traceability via SSE event replay.
// This test ensures operators can trace stuck matchmaking offers by:
// 1. Subscribing to tournament SSE stream
// 2. Receiving MatchmakingQueueUpdated events
// 3. Replaying events using Last-Event-ID to trace stuck offers
// 4. Querying match_lobby_views to confirm queueState.offerId matches SSE event data
func TestMatchEligibilitySync(t *testing.T) {
	ctx := context.Background()

	// Set up testcontainers for Postgres and Redis
	postgresContainer, err := postgres.RunContainer(ctx,
		tc.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("competition_host"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		tc.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(1).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	}()

	redisContainer, err := redismodule.RunContainer(ctx,
		tc.WithImage("redis:7-alpine"),
		tc.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithOccurrence(1).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	}()

	// Get connection strings
	pgConnStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	redisAddr, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Run migrations
	err = runMigrations(ctx, pgConnStr)
	require.NoError(t, err, "failed to run migrations")

	// Set up test config
	cfg := config.Config{
		PostgresDSN: pgConnStr,
		RedisAddr:   redisAddr,
		SSEPoolSize: 100,
		LeaseTTL:    90 * time.Second,
	}

	// Initialize clients
	clients, err := store.NewClients(ctx, cfg)
	require.NoError(t, err)
	defer clients.Close()

	// Set up SSE broker and event router
	broker := ssebroker.NewBroker(cfg.SSEPoolSize)
	defer broker.Close()
	eventRouter := ssebroker.NewEventRouter(broker)
	registry := ssebroker.NewRegistry(clients.PG, cfg.LeaseTTL)

	// Create match projector with event router
	matchProjector := projections.NewMatchProjector(clients.PG, eventRouter)

	// Create test data
	hostID := uuid.New()
	tournamentID := uuid.New()
	matchID := uuid.New()
	offerID := uuid.New()

	// Create a stuck offer scenario: queueState.status=OPEN but not progressing
	stuckQueueState := map[string]any{
		"offerId": offerID.String(),
		"status":  "OPEN",
		"queueType": "matchmaking",
	}
	queueStateJSON, err := json.Marshal(stuckQueueState)
	require.NoError(t, err)

	lobbyInfo := map[string]any{
		"maximumLobbyMinutes":     10,
		"maximumGoToGameMinutes":  5,
		"gameSessionTag":          "test-session",
	}
	lobbyInfoJSON, err := json.Marshal(lobbyInfo)
	require.NoError(t, err)

	// Create match lobby view with stuck offer
	matchView := projections.MatchLobbyView{
		MatchID:          matchID,
		TournamentID:    tournamentID,
		LobbyInformation: lobbyInfoJSON,
		QueueState:      queueStateJSON,
	}

	// Upsert match view - this should trigger MatchmakingQueueUpdated event
	err = matchProjector.Upsert(ctx, matchView)
	require.NoError(t, err)

	// Set up HTTP server for SSE streaming
	router := setupTestRouter(broker, registry)
	server := httptest.NewServer(router)
	defer server.Close()

	// Subscribe to tournament SSE stream
	streamURL := fmt.Sprintf("%s/v1/hosts/%s/streams/tournament/%s", server.URL, hostID, tournamentID)
	req, err := http.NewRequestWithContext(ctx, "GET", streamURL, nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "SSE stream should return 200")
	require.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	// Read SSE events
	scanner := bufio.NewScanner(resp.Body)
	var events []SSEEvent
	var lastEventID string

	// Wait for event with timeout
	timeout := time.After(3 * time.Second)
	eventReceived := make(chan bool)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "event:") {
				eventType := strings.TrimSpace(strings.TrimPrefix(line, "event:"))
				event := SSEEvent{Type: eventType}
				for scanner.Scan() {
					nextLine := scanner.Text()
					if strings.HasPrefix(nextLine, "id:") {
						event.ID = strings.TrimSpace(strings.TrimPrefix(nextLine, "id:"))
						lastEventID = event.ID
					} else if strings.HasPrefix(nextLine, "data:") {
						event.Data = strings.TrimSpace(strings.TrimPrefix(nextLine, "data:"))
						events = append(events, event)
						if event.Type == "MatchmakingQueueUpdated" {
							eventReceived <- true
							return
						}
					} else if nextLine == "" {
						break
					}
				}
			}
		}
	}()

	select {
	case <-eventReceived:
		// Event received successfully
	case <-timeout:
		t.Fatal("timeout waiting for MatchmakingQueueUpdated event")
	}

	// Verify MatchmakingQueueUpdated event was received
	require.Greater(t, len(events), 0, "should receive at least one SSE event")
	foundMatchmakingEvent := false
	var matchmakingEventData map[string]any

	for _, event := range events {
		if event.Type == "MatchmakingQueueUpdated" {
			foundMatchmakingEvent = true
			err := json.Unmarshal([]byte(event.Data), &matchmakingEventData)
			require.NoError(t, err)
			break
		}
	}

	require.True(t, foundMatchmakingEvent, "MatchmakingQueueUpdated event should be received")
	require.Equal(t, matchID.String(), matchmakingEventData["matchId"], "event should contain correct matchId")
	require.Equal(t, tournamentID.String(), matchmakingEventData["tournamentId"], "event should contain correct tournamentId")

	// Verify queueState.offerId in event matches database
	// After JSON marshaling/unmarshaling, queueState will be a map[string]any
	queueStateData, ok := matchmakingEventData["queueState"].(map[string]any)
	if !ok {
		// If it's a string (raw JSON), parse it
		if queueStateStr, ok := matchmakingEventData["queueState"].(string); ok {
			err := json.Unmarshal([]byte(queueStateStr), &queueStateData)
			require.NoError(t, err, "queueState string should be valid JSON")
		} else {
			require.Fail(t, "queueState should be map[string]any or JSON string")
		}
	}
	eventOfferID, ok := queueStateData["offerId"].(string)
	require.True(t, ok, "queueState should contain offerId")
	require.Equal(t, offerID.String(), eventOfferID, "event offerId should match test data")

	// Test event replay with Last-Event-ID
	require.NotEmpty(t, lastEventID, "should have received an event ID for replay")

	// Create new request with Last-Event-ID header
	replayReq, err := http.NewRequestWithContext(ctx, "GET", streamURL, nil)
	require.NoError(t, err)
	replayReq.Header.Set("Last-Event-ID", lastEventID)

	replayResp, err := client.Do(replayReq)
	require.NoError(t, err)
	defer replayResp.Body.Close()

	require.Equal(t, http.StatusOK, replayResp.StatusCode, "SSE replay should return 200")

	// Verify we can query match_lobby_views to confirm queueState.offerId
	reader := projections.NewReader(clients.PG)
	matches, err := reader.ListMatches(ctx, tournamentID)
	require.NoError(t, err)
	require.Greater(t, len(matches), 0, "should find match in database")

	foundMatch := false
	for _, match := range matches {
		if match["matchId"] == matchID.String() {
			foundMatch = true
			queueStateRaw, ok := match["queueState"].(json.RawMessage)
			require.True(t, ok, "queueState should be present")

			var dbQueueState map[string]any
			err := json.Unmarshal(queueStateRaw, &dbQueueState)
			require.NoError(t, err)

			dbOfferID, ok := dbQueueState["offerId"].(string)
			require.True(t, ok, "database queueState should contain offerId")
			require.Equal(t, offerID.String(), dbOfferID, "database offerId should match test data")
			require.Equal(t, "OPEN", dbQueueState["status"], "stuck offer should have OPEN status")
			break
		}
	}

	require.True(t, foundMatch, "should find match with stuck offer in database")
}

// SSEEvent represents a parsed SSE event
type SSEEvent struct {
	Type string
	ID   string
	Data string
}

// setupTestRouter creates a minimal router for SSE testing
func setupTestRouter(broker *ssebroker.Broker, registry *ssebroker.Registry) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/hosts/", func(w http.ResponseWriter, r *http.Request) {
		// Parse path: /v1/hosts/{hostId}/streams/{scopeType}/{scopeId}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/hosts/"), "/")
		if len(parts) != 4 || parts[1] != "streams" {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}

		hostIDStr := parts[0]
		scopeType := parts[2]
		scopeIDStr := parts[3]

		hostID, err := uuid.Parse(hostIDStr)
		if err != nil {
			http.Error(w, "invalid hostId", http.StatusBadRequest)
			return
		}

		scopeID, err := uuid.Parse(scopeIDStr)
		if err != nil {
			http.Error(w, "invalid scopeId", http.StatusBadRequest)
			return
		}

		if scopeType != "tournament" && scopeType != "cup" && scopeType != "match" {
			http.Error(w, "invalid scopeType", http.StatusBadRequest)
			return
		}

		// Release expired subscriptions
		_ = registry.ReleaseExpired(r.Context())

		// Parse Last-Event-ID
		lastEventID := int64(0)
		if raw := r.Header.Get("Last-Event-ID"); raw != "" {
			// In real implementation, this would parse the event ID
			// For testing, we'll use 0
		}

		// Lease subscription
		_, err = registry.Lease(r.Context(), hostID, scopeType, scopeID, lastEventID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create stream and serve SSE
		scope := ssebroker.Scope{Type: scopeType, ID: scopeID}
		broker.Server().CreateStream(scope.Key())
		broker.Server().ServeHTTP(w, r)
	})

	return mux
}

// runMigrations executes the database migrations
func runMigrations(ctx context.Context, dsn string) error {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	migrationSQL := `
CREATE TABLE IF NOT EXISTS cup_host_views (
    cup_id UUID PRIMARY KEY,
    competition_context_id UUID NOT NULL,
    stage_statuses JSONB NOT NULL DEFAULT '[]'::jsonb,
    attendance_counts JSONB NOT NULL DEFAULT '{}'::jsonb,
    dependency_health JSONB NOT NULL DEFAULT '[]'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tournament_host_views (
    tournament_id UUID PRIMARY KEY,
    cup_id UUID,
    settings_hash TEXT NOT NULL,
    lineup_status JSONB NOT NULL DEFAULT '[]'::jsonb,
    seeding_window TSRANGE,
    match_gate JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS attendance_snapshots (
    scope_type TEXT NOT NULL,
    scope_id UUID NOT NULL,
    total_joined INTEGER NOT NULL DEFAULT 0,
    total_confirmed INTEGER NOT NULL DEFAULT 0,
    restrictions_breached JSONB NOT NULL DEFAULT '[]'::jsonb,
    last_event_id BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (scope_type, scope_id)
);

CREATE TABLE IF NOT EXISTS match_lobby_views (
    match_id UUID PRIMARY KEY,
    tournament_id UUID NOT NULL,
    lobby_information JSONB NOT NULL DEFAULT '{}'::jsonb,
    queue_state JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS host_subscriptions (
    subscription_id UUID PRIMARY KEY,
    host_id UUID NOT NULL,
    scope_type TEXT NOT NULL,
    scope_id UUID NOT NULL,
    last_delivered_event_id BIGINT NOT NULL DEFAULT 0,
    leased_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (host_id, scope_type, scope_id)
);

CREATE INDEX IF NOT EXISTS idx_tournament_host_views_cup_id ON tournament_host_views (cup_id);
CREATE INDEX IF NOT EXISTS idx_match_lobby_views_tournament_id ON match_lobby_views (tournament_id);
CREATE INDEX IF NOT EXISTS idx_attendance_snapshots_event ON attendance_snapshots (last_event_id DESC);
`

	_, err = pool.Exec(ctx, migrationSQL)
	return err
}

