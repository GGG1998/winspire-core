package handlers

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/winspire/competition/internal/repository"
)

// ============================================================================
// Tournament Policies
// Encapsulate business rules for request -> repository params conversion
// ============================================================================

const (
	defaultMinimumTeamCount = 2
	defaultMaximumTeamCount = 250
)

// ApplyCreateTournamentPolicy converts and validates HTTP request to repository params.
// Encapsulates business rules for tournament creation including default values.
func ApplyCreateTournamentPolicy(hostID uuid.UUID, req CreateTournamentRequest) repository.CreateTournamentParams {
	params := repository.CreateTournamentParams{
		HostID:                   hostID,
		Name:                     req.Name,
		Description:              req.Description,
		ExternalID:               req.ExternalID,
		ScheduledStartTimeAt:     req.ScheduledStartTimeAt,
		RegistrationWindowOpenAt: req.RegistrationWindowOpenAt,
		TeamSize:                 1, // default: solo tournament
	}

	// Apply optional minimum team count
	if req.MinimumTeamCount != nil {
		minCount := int32(*req.MinimumTeamCount)
		params.MinimumTeamCount = &minCount
	} else {
		minCount := int32(defaultMinimumTeamCount)
		params.MinimumTeamCount = &minCount
	}

	// Apply optional maximum team count
	if req.MaximumTeamCount != nil {
		maxCount := int32(*req.MaximumTeamCount)
		params.MaximumTeamCount = &maxCount
	} else {
		maxCount := int32(defaultMaximumTeamCount)
		params.MaximumTeamCount = &maxCount
	}

	// Apply optional auto force ready setting
	if req.AutoForceReady != nil {
		params.AutoForceReady = req.AutoForceReady
	}

	// Apply competition parameters (team size, template)
	if req.CompetitionParameters.TournamentParameters != nil {
		if req.CompetitionParameters.TournamentParameters.TeamSize != nil {
			params.TeamSize = int32(*req.CompetitionParameters.TournamentParameters.TeamSize)
		}
		params.TemplateID = req.CompetitionParameters.TournamentParameters.TemplateID
	}

	// Apply game reference
	if req.Game != nil {
		params.GameID = req.Game.ID
	}

	// Apply space reference
	if req.Space != nil {
		params.SpaceID = req.Space.ID
	}

	// Apply game snapshot if provided
	if req.GameSnapshot != nil {
		// Parse game ID from string
		gameID, err := uuid.Parse(req.GameSnapshot.ID)
		if err == nil {
			// Convert GameSnapshotInput to repository.GameSnapshot
			snapshot := repository.GameSnapshot{
				ID:          gameID,
				Slug:        req.GameSnapshot.Slug,
				Name:        req.GameSnapshot.Name,
				Version:     req.GameSnapshot.Version,
				LogoURL:     req.GameSnapshot.LogoURL,
				Description: req.GameSnapshot.Description,
				StoragePath: req.GameSnapshot.StoragePath,
			}

			// Marshal to JSONB
			snapshotJSON, err := json.Marshal(snapshot)
			if err == nil {
				params.GameSnapshot = &snapshotJSON
			}
			// If marshaling fails, params.GameSnapshot remains nil (graceful degradation)
		}
	}

	return params
}

// ============================================================================
// Authorization Policies
// ============================================================================

// CanViewTournament checks if a user can view a specific tournament.
// Business rule: Draft tournaments can only be viewed by their creator.
// All other statuses are publicly viewable.
func CanViewTournament(userID uuid.UUID, tournament *repository.Tournament) bool {
	// Draft tournaments - only owner can view
	if tournament.Status == "draft" {
		return tournament.HostID == userID
	}

	// All other statuses (registration_open, registration_closed, started, completed, cancelled)
	// are publicly viewable
	return true
}
