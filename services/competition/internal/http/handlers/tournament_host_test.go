package handlers

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/winspire/competition/internal/repository"
)

// Note: For full event publishing tests, use integration tests with miniredis

// TestTournamentStartEventPublishing tests tournament start transitions
// Note: This is a unit test for the domain logic. Event publishing is tested in integration tests.
func TestTournamentStartEventPublishing(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name             string
		initialStatus    string
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:          "Start from Registration Open",
			initialStatus: "registration_open",
			expectError:   false,
		},
		{
			name:          "Start from Registration Closed",
			initialStatus: "registration_closed",
			expectError:   false,
		},
		{
			name:             "Start from Draft",
			initialStatus:    "draft",
			expectError:      true,
			expectedErrorMsg: "cannot transition",
		},
		{
			name:             "Start from Scheduled",
			initialStatus:    "scheduled",
			expectError:      true,
			expectedErrorMsg: "cannot transition",
		},
		{
			name:             "Start from Completed",
			initialStatus:    "completed",
			expectError:      true,
			expectedErrorMsg: "cannot transition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create tournament with initial status
			tournament := &repository.Tournament{
				ID:               uuid.New(),
				HostID:           uuid.New(),
				Name:             "Test Tournament",
				Status:           tt.initialStatus,
				MinimumTeamCount: 2,
				TeamSize:         1,
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			// Test the domain logic
			err := tournament.Start()

			if tt.expectError {
				assert.Error(t, err, "Should return error for invalid transition")
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg, "Error should mention invalid transition")
				}
			} else {
				assert.NoError(t, err, "Should not return error for valid transition")
				assert.Equal(t, "starting", tournament.Status, "Status should be updated to starting (transient state before confirmation)")
				assert.NotNil(t, tournament.ActualStartTimeAt, "ActualStartTimeAt should be set")
			}
		})
	}
}

// TestTournamentStatusTransitions tests status transition validation
func TestTournamentStatusTransitions(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)

	tests := []struct {
		name          string
		initialStatus string
		action        func(*repository.Tournament) error
		expectError   bool
		errorContains string
	}{
		{
			name:          "Publish from Draft",
			initialStatus: "draft",
			action: func(t *repository.Tournament) error {
				t.ScheduledStartTimeAt = &futureTime
				return t.Publish()
			},
			expectError: false,
		},
		{
			name:          "Publish from Scheduled",
			initialStatus: "scheduled",
			action: func(t *repository.Tournament) error {
				return t.Publish()
			},
			expectError:   true,
			errorContains: "cannot transition",
		},
		{
			name:          "Open Registration from Scheduled",
			initialStatus: "scheduled",
			action: func(t *repository.Tournament) error {
				return t.OpenRegistration()
			},
			expectError: false,
		},
		{
			name:          "Start from Registration Open",
			initialStatus: "registration_open",
			action: func(t *repository.Tournament) error {
				return t.Start()
			},
			expectError: false,
		},
		{
			name:          "Start from Registration Closed",
			initialStatus: "registration_closed",
			action: func(t *repository.Tournament) error {
				return t.Start()
			},
			expectError: false,
		},
		{
			name:          "Start from Draft",
			initialStatus: "draft",
			action: func(t *repository.Tournament) error {
				return t.Start()
			},
			expectError:   true,
			errorContains: "cannot transition",
		},
		{
			name:          "Complete from Started",
			initialStatus: "started",
			action: func(t *repository.Tournament) error {
				return t.Complete()
			},
			expectError: false,
		},
		{
			name:          "Cancel from Registration Open",
			initialStatus: "registration_open",
			action: func(t *repository.Tournament) error {
				return t.Cancel()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tournament := &repository.Tournament{
				ID:               uuid.New(),
				HostID:           uuid.New(),
				Name:             "Test Tournament",
				Status:           tt.initialStatus,
				MinimumTeamCount: 2,
				TeamSize:         1,
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			err := tt.action(tournament)

			if tt.expectError {
				assert.Error(t, err, "Expected error for invalid transition")
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains, "Error should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no error for valid transition")
			}
		})
	}
}
