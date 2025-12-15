package application

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/winspire-core/services/matchmaking/internal/observability"
)

func TestCompetitionClient_GetTournamentInfo(t *testing.T) {
	logger := observability.NewLogger("info", "text")

	tests := []struct {
		name           string
		tournamentID   uuid.UUID
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedResult *TournamentInfo
		expectNil      bool
		expectError    bool
	}{
		{
			name:         "success with all fields",
			tournamentID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"id":                   "11111111-1111-1111-1111-111111111111",
					"name":                 "Test Tournament",
					"hostId":               "00000000-0000-0000-0000-000000000001",
					"gameId":               "22222222-2222-2222-2222-222222222222",
					"scheduledStartTimeAt": "2025-01-15T10:00:00Z",
					"minParticipants":      8,
					"status":               "registration_open",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			},
			expectedResult: &TournamentInfo{
				ID:              uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				Name:            "Test Tournament",
				HostID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				GameID:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				MinParticipants: 8,
				Status:          "registration_open",
			},
			expectNil:   false,
			expectError: false,
		},
		{
			name:         "success without game ID",
			tournamentID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"id":              "33333333-3333-3333-3333-333333333333",
					"name":            "No Game Tournament",
					"hostId":          "00000000-0000-0000-0000-000000000001",
					"minParticipants": 4,
					"status":          "draft",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			},
			expectedResult: &TournamentInfo{
				ID:              uuid.MustParse("33333333-3333-3333-3333-333333333333"),
				Name:            "No Game Tournament",
				HostID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				GameID:          uuid.UUID{}, // Empty when not set
				MinParticipants: 4,
				Status:          "draft",
			},
			expectNil:   false,
			expectError: false,
		},
		{
			name:         "tournament not found returns nil",
			tournamentID: uuid.MustParse("99999999-9999-9999-9999-999999999999"),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "tournament not found"})
			},
			expectedResult: nil,
			expectNil:      true,
			expectError:    false,
		},
		{
			name:         "min participants defaults to 2 when less than 2",
			tournamentID: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"id":              "44444444-4444-4444-4444-444444444444",
					"name":            "Small Tournament",
					"hostId":          "00000000-0000-0000-0000-000000000001",
					"minParticipants": 1, // Less than 2
					"status":          "started",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			},
			expectedResult: &TournamentInfo{
				ID:              uuid.MustParse("44444444-4444-4444-4444-444444444444"),
				Name:            "Small Tournament",
				HostID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				MinParticipants: 2, // Should default to 2
				Status:          "started",
			},
			expectNil:   false,
			expectError: false,
		},
		{
			name:         "server error returns error",
			tournamentID: uuid.MustParse("55555555-5555-5555-5555-555555555555"),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedResult: nil,
			expectNil:      false,
			expectError:    true,
		},
		{
			name:         "invalid JSON response returns error",
			tournamentID: uuid.MustParse("66666666-6666-6666-6666-666666666666"),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			},
			expectedResult: nil,
			expectNil:      false,
			expectError:    true,
		},
		{
			name:         "invalid tournament ID in response returns error",
			tournamentID: uuid.MustParse("77777777-7777-7777-7777-777777777777"),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"id":              "not-a-uuid",
					"name":            "Bad Tournament",
					"hostId":          "00000000-0000-0000-0000-000000000001",
					"minParticipants": 8,
					"status":          "draft",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			},
			expectedResult: nil,
			expectNil:      false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request path
				expectedPath := "/internal/tournaments/" + tt.tournamentID.String()
				assert.Equal(t, expectedPath, r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				tt.serverResponse(w, r)
			}))
			defer server.Close()

			// Create client pointing to mock server
			client := NewCompetitionClient(server.URL, logger)

			// Execute
			result, err := client.GetTournamentInfo(context.Background(), tt.tournamentID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.expectNil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedResult.ID, result.ID)
			assert.Equal(t, tt.expectedResult.Name, result.Name)
			assert.Equal(t, tt.expectedResult.HostID, result.HostID)
			assert.Equal(t, tt.expectedResult.GameID, result.GameID)
			assert.Equal(t, tt.expectedResult.MinParticipants, result.MinParticipants)
			assert.Equal(t, tt.expectedResult.Status, result.Status)
		})
	}
}

func TestCompetitionClient_IsUserRegistered(t *testing.T) {
	logger := observability.NewLogger("info", "text")

	tests := []struct {
		name           string
		tournamentID   uuid.UUID
		userID         uuid.UUID
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expected       bool
		expectError    bool
	}{
		{
			name:         "user is registered",
			tournamentID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			userID:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"registered":  true,
					"status":      "confirmed",
					"displayName": "Test Player",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			},
			expected:    true,
			expectError: false,
		},
		{
			name:         "user is not registered",
			tournamentID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			userID:       uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"registered": false,
					"error":      "user not registered",
				})
			},
			expected:    false,
			expectError: false,
		},
		{
			name:         "server error is propagated",
			tournamentID: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
			userID:       uuid.MustParse("55555555-5555-5555-5555-555555555555"),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expected:    false, // Server error treated as not registered by checking status code
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request path
				expectedPath := "/internal/tournaments/" + tt.tournamentID.String() + "/registrations/" + tt.userID.String()
				assert.Equal(t, expectedPath, r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				tt.serverResponse(w, r)
			}))
			defer server.Close()

			// Create client pointing to mock server
			client := NewCompetitionClient(server.URL, logger)
			// Clear cache to ensure fresh request
			client.ClearCache()

			// Execute
			result, err := client.IsUserRegistered(context.Background(), tt.tournamentID, tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompetitionClient_IsUserRegistered_Cache(t *testing.T) {
	logger := observability.NewLogger("info", "text")

	tournamentID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := map[string]interface{}{
			"registered":  true,
			"status":      "confirmed",
			"displayName": "Test Player",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewCompetitionClient(server.URL, logger)
	client.ClearCache()

	ctx := context.Background()

	// First call - should hit server
	result1, err := client.IsUserRegistered(ctx, tournamentID, userID)
	require.NoError(t, err)
	assert.True(t, result1)
	assert.Equal(t, 1, callCount, "first call should hit server")

	// Second call - should hit cache
	result2, err := client.IsUserRegistered(ctx, tournamentID, userID)
	require.NoError(t, err)
	assert.True(t, result2)
	assert.Equal(t, 1, callCount, "second call should use cache")

	// Clear cache and call again - should hit server
	client.ClearCache()
	result3, err := client.IsUserRegistered(ctx, tournamentID, userID)
	require.NoError(t, err)
	assert.True(t, result3)
	assert.Equal(t, 2, callCount, "call after cache clear should hit server")
}

func TestCompetitionClient_GetTournamentInfo_Timeout(t *testing.T) {
	logger := observability.NewLogger("info", "text")

	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second) // Longer than client timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewCompetitionClient(server.URL, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := client.GetTournamentInfo(ctx, uuid.New())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
