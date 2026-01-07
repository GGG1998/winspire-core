package handlers

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/winspire-core/services/matchmaking/internal/tournament/domain"
	"github.com/winspire-core/services/matchmaking/internal/tournament/repository"
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
func ApplyCreateTournamentPolicy(hostID uuid.UUID, req CreateTournamentRequest) repository.CreateParams {
	params := repository.CreateParams{
		HostID:               hostID,
		Name:                 req.Name,
		Description:          req.Description,
		ExternalID:           req.ExternalID,
		ScheduledStartTimeAt: req.ScheduledStartTimeAt,
		TeamSize:             1, // default: solo tournament
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

	// Apply tournament configuration parameters (team size, template)
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
			// Convert GameSnapshotInput to domain.GameSnapshot
			snapshot := domain.GameSnapshot{
				ID:          gameID,
				Slug:        req.GameSnapshot.Slug,
				Name:        req.GameSnapshot.Name,
				Version:     req.GameSnapshot.Version,
				StoragePath: req.GameSnapshot.StoragePath,
			}
			// Handle optional pointer fields
			if req.GameSnapshot.LogoURL != nil {
				snapshot.LogoURL = *req.GameSnapshot.LogoURL
			}
			if req.GameSnapshot.Description != nil {
				snapshot.Description = *req.GameSnapshot.Description
			}
			params.GameSnapshot = &snapshot
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
func CanViewTournament(userID uuid.UUID, tournament *domain.Tournament) bool {
	// Draft tournaments - only owner can view
	if tournament.Status == "draft" {
		return tournament.HostID == userID
	}

	// All other statuses (registration_open, registration_closed, started, completed, cancelled)
	// are publicly viewable
	return true
}

// newTournamentDetail converts domain model into API response payload.
func newTournamentDetail(t *domain.Tournament) TournamentDetail {
	hostID := t.HostID.String()
	tournamentID := t.ID.String()

	// Generate room link for participants
	roomLink := "/tournaments/" + tournamentID + "/lobby"

	detail := TournamentDetail{
		ID:                       tournamentID,
		HostID:                   hostID,
		CreatorID:                hostID, // CreatorID = HostID for frontend compatibility
		Name:                     t.Name,
		Status:                   t.Status,
		Description:              t.Description,
		ExternalID:               t.ExternalID,
		CreatedAt:                t.CreatedAt,
		UpdatedAt:                t.UpdatedAt,
		RoomLink:                 &roomLink,
		ScheduledStartTimeAt:     t.ScheduledStartTimeAt,
		RegistrationWindowOpenAt: t.RegistrationWindowOpenAt,
		ActualStartTimeAt:        t.ActualStartTimeAt,
		CompletedAt:              t.CompletedAt,
		CancelledAt:              t.CancelledAt,
		MinimumTeamCount:         t.MinimumTeamCount,
		MaximumTeamCount:         t.MaximumTeamCount,
		TeamSize:                 t.TeamSize,
		AutoForceReady:           t.AutoForceReady,
		IsCompleted:              t.CompletedAt != nil || t.Status == "completed",
		LastActivity:             nil,
		ParticipantCount:         0,
	}

	// Set last activity
	lastActivity := t.UpdatedAt
	detail.LastActivity = &lastActivity

	// Set game information from snapshot
	if t.GameSnapshot != nil {
		gameID := t.GameSnapshot.ID.String()
		detail.GameID = &gameID
		detail.Game = &t.GameSnapshot.Slug
		if t.GameSnapshot.LogoURL != "" {
			detail.GameLogoURL = &t.GameSnapshot.LogoURL
		}
	} else if t.GameID != nil {
		gameID := t.GameID.String()
		detail.GameID = &gameID
	}

	if t.SpaceID != nil {
		spaceID := t.SpaceID.String()
		detail.SpaceID = &spaceID
	}

	if t.TemplateID != nil {
		templateID := t.TemplateID.String()
		detail.TemplateID = &templateID
	}

	// Set tournament format based on configuration
	maxSlots := int32(16) // Default max slots
	if t.MaximumTeamCount != nil {
		maxSlots = *t.MaximumTeamCount
	}
	detail.Format = &TournamentFormat{
		Type:     "single_elimination", // Default format
		TeamSize: t.TeamSize,
		MaxSlots: maxSlots,
	}

	// Parse ready window from domain
	if t.ReadyWindow != nil && t.ReadyWindow.StartsAt != nil && t.ReadyWindow.EndsAt != nil {
		detail.ReadyWindow = &TournamentReadyWindow{
			StartsAt: *t.ReadyWindow.StartsAt,
			EndsAt:   *t.ReadyWindow.EndsAt,
		}
	}

	// Parse prize from domain
	if t.Prize != nil {
		var value *int
		if t.Prize.Value > 0 {
			v := int(t.Prize.Value)
			value = &v
		}
		var currency *string
		if t.Prize.Currency != "" {
			currency = &t.Prize.Currency
		}
		detail.Prize = &TournamentPrize{
			Type:        t.Prize.Type,
			Description: t.Prize.Description,
			Value:       value,
			Currency:    currency,
		}
	}

	return detail
}

// marshalJSONB marshals a value to JSON for JSONB storage.
func marshalJSONB(v any) (json.RawMessage, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}
