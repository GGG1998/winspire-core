package handlers

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler holds all handler dependencies.
type Handler struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// New creates a new Handler with the given dependencies.
func New(pool *pgxpool.Pool, logger *slog.Logger) *Handler {
	return &Handler{
		pool:   pool,
		logger: logger,
	}
}

// === Common Response Types ===

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}

// SuccessResponse represents a generic success response.
type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// BaseResponse is a common structure for most API responses.
// It includes standard success/error fields that can be embedded in specific response types.
type BaseResponse struct {
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`
	// ErrorCode contains a specific error code if the operation failed.
	ErrorCode *string `json:"errorCode,omitempty"`
}

// === Tournament User Request Types ===

// JoinTournamentRequest represents the request body for joining a tournament.
// The hostId and tournamentId are provided as path parameters.
type JoinTournamentRequest struct {
	// SocialTeamID is the ID of an existing social team to join with.
	SocialTeamID *uuid.UUID `json:"socialTeamId,omitempty"`
	// TeamName is the optional name of a temporary team for this tournament.
	TeamName *string `json:"teamName,omitempty"`
	// Players is the list of user IDs who will join the tournament as a team.
	Players []uuid.UUID `json:"players" binding:"required,min=1"`
}

// === Tournament User Response Types ===

// JoinTournamentResponse represents the response for tournament join.
// JoinTournamentResponse represents the response for joining a tournament.
type JoinTournamentResponse struct {
	BaseResponse
	RoomLink string `json:"roomLink,omitempty"`
}

// LeaveTournamentResponse represents the response for leaving a tournament.
type LeaveTournamentResponse = BaseResponse

// ConfirmTournamentParticipationResponse represents the response for confirming tournament participation.
type ConfirmTournamentParticipationResponse = BaseResponse

// ConfirmMatchParticipationResponse represents the response for confirming match participation.
type ConfirmMatchParticipationResponse = BaseResponse

// === Tournament Host Request Types ===

// TournamentParametersInput contains parameters for tournament configuration.
type TournamentParametersInput struct {
	// TemplateID is the tournament template ID when creating from an existing template.
	TemplateID *uuid.UUID `json:"templateId,omitempty"`
	// TeamSize is the required size of each lineup. Must be compatible with the game.
	TeamSize *int `json:"teamSize,omitempty"`
}

// CompetitionParametersInput defines the competition parameters.
type CompetitionParametersInput struct {
	// TournamentParameters for creating tournaments.
	TournamentParameters *TournamentParametersInput `json:"tournamentParameters,omitempty"`
}

// SpaceIdentifierInput identifies a space by ID or slug.
type SpaceIdentifierInput struct {
	// ID is the space ID.
	ID *uuid.UUID `json:"id,omitempty"`
	// Slug is the space slug (e.g., from URL).
	Slug *string `json:"slug,omitempty"`
}

// GameIdentifierInput identifies a game.
type GameIdentifierInput struct {
	// ID is the game ID.
	ID *uuid.UUID `json:"id,omitempty"`
	// Slug is the game slug.
	Slug *string `json:"slug,omitempty"`
}

// CreateTournamentRequest represents the request body for creating a tournament.
// The hostId is provided as a path parameter.
type CreateTournamentRequest struct {
	// Name is the public name of the tournament (min 3, max 100 chars).
	Name string `json:"name" binding:"required,min=3,max=100"`
	// Description is the public description (max 4000 chars).
	Description *string `json:"description,omitempty" binding:"omitempty,max=4000"`
	// ExternalID can be used to guarantee idempotency.
	ExternalID *string `json:"externalId,omitempty"`
	// CompetitionParameters define the competition settings.
	CompetitionParameters CompetitionParametersInput `json:"competitionParameters" binding:"required"`
	// RegistrationWindowOpenAt is when registration opens.
	RegistrationWindowOpenAt *time.Time `json:"registrationWindowOpenAt,omitempty"`
	// ScheduledStartTimeAt is when the tournament should start (UTC RFC3339).
	ScheduledStartTimeAt *time.Time `json:"scheduledStartTimeAt,omitempty"`
	// AutoForceReady bypasses restriction settings. Defaults to true.
	AutoForceReady *bool `json:"autoForceReady,omitempty"`
	// MinimumTeamCount is the minimum teams required to start. Defaults to 2.
	MinimumTeamCount *int `json:"minimumTeamCount,omitempty"`
	// MaximumTeamCount is the maximum teams allowed. Defaults to 250.
	MaximumTeamCount *int `json:"maximumTeamCount,omitempty"`
	// Game identifies the game for this tournament.
	Game *GameIdentifierInput `json:"game,omitempty"`
	// Space identifies the hosting space.
	Space *SpaceIdentifierInput `json:"space,omitempty"`
}

// EditTournamentRequest represents the request body for editing a tournament.
// Supports partial updates including status transitions (start, cancel, open).
type EditTournamentRequest struct {
	// Name is the public name of the tournament.
	Name *string `json:"name,omitempty" binding:"omitempty,min=3,max=100"`
	// Description is the public description.
	Description *string `json:"description,omitempty" binding:"omitempty,max=4000"`
	// Status allows transitioning tournament state (open, registration_open, started, completed, cancelled).
	Status *string `json:"status,omitempty" binding:"omitempty,oneof=draft open scheduled registration_open started completed cancelled"`
	// ScheduledStartTimeAt is when the tournament should start.
	ScheduledStartTimeAt *time.Time `json:"scheduledStartTimeAt,omitempty"`
	// MinimumTeamCount is the minimum teams required to start.
	MinimumTeamCount *int `json:"minimumTeamCount,omitempty"`
	// MaximumTeamCount is the maximum teams allowed.
	MaximumTeamCount *int `json:"maximumTeamCount,omitempty"`
	// Format is the tournament format configuration.
	Format *TournamentFormat `json:"format,omitempty"`
	// ReadyWindow is the check-in window.
	ReadyWindow *TournamentReadyWindow `json:"readyWindow,omitempty"`
	// Prize is the prize configuration.
	Prize *TournamentPrize `json:"prize,omitempty"`
}

// === Tournament Host Response Types ===

// CreateTournamentResponse represents the response for tournament creation.
type CreateTournamentResponse struct {
	BaseResponse
	// TournamentID is the ID of the created tournament.
	TournamentID *string `json:"tournamentId,omitempty"`
}

// EditTournamentResponse represents the response for tournament editing.
// Also used for status transitions (start, cancel, open).
type EditTournamentResponse = BaseResponse

// TournamentListItem represents a tournament in the list response.
type TournamentListItem struct {
	// ID is the tournament ID.
	ID string `json:"id"`
	// Name is the tournament name.
	Name string `json:"name"`
	// Status is the tournament status (e.g., "draft", "open", "started", "completed", "cancelled").
	Status string `json:"status"`
	// ScheduledStartTimeAt is when the tournament is scheduled to start.
	ScheduledStartTimeAt *time.Time `json:"scheduledStartTimeAt,omitempty"`
	// CreatedAt is when the tournament was created.
	CreatedAt time.Time `json:"createdAt"`
}

// ListTournamentsResponse represents the response for listing tournaments.
type ListTournamentsResponse struct {
	// Tournaments is the list of tournaments.
	Tournaments []TournamentListItem `json:"tournaments"`
	// Count is the total number of tournaments returned.
	Count int `json:"count"`
}

// TournamentOrganizer represents the organizer/host information.
type TournamentOrganizer struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	LogoURL *string `json:"logoUrl,omitempty"`
}

// TournamentFormat represents the tournament format configuration.
type TournamentFormat struct {
	Type     string `json:"type"` // single_elimination, double_elimination, round_robin, swiss
	TeamSize int32  `json:"teamSize"`
	MaxSlots int32  `json:"maxSlots"`
	BestOf   *int32 `json:"bestOf,omitempty"`
}

// TournamentReadyWindow represents the check-in window.
type TournamentReadyWindow struct {
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
}

// TournamentPrize represents the prize configuration.
type TournamentPrize struct {
	Type        string  `json:"type"` // custom, cash, points
	Description string  `json:"description"`
	Value       *int    `json:"value,omitempty"`
	Currency    *string `json:"currency,omitempty"`
}

// TournamentMeInfo represents user-specific tournament information.
type TournamentMeInfo struct {
	ParticipationStatus *string `json:"participationStatus,omitempty"` // "registered", "confirmed", "checked_in", or null
}

// TournamentDetail represents a detailed tournament payload.
type TournamentDetail struct {
	// Core identifiers
	ID        string `json:"id"`
	HostID    string `json:"hostId"`
	CreatorID string `json:"creatorId"` // Same as HostID for frontend compatibility

	// Basic metadata
	Name         string     `json:"name"`
	Status       string     `json:"status"`
	Description  *string    `json:"description,omitempty"`
	ExternalID   *string    `json:"externalId,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	LastActivity *time.Time `json:"lastActivity,omitempty"`

	// Game information
	Game        *string `json:"game,omitempty"`        // Game slug for display
	GameLogoURL *string `json:"gameLogoUrl,omitempty"` // Game logo URL
	BannerURL   *string `json:"bannerUrl,omitempty"`   // Tournament banner image
	RoomLink    *string `json:"roomLink,omitempty"`    // Join link for participants

	// Scheduling
	ScheduledStartTimeAt     *time.Time `json:"scheduledStartTimeAt,omitempty"`
	RegistrationWindowOpenAt *time.Time `json:"registrationWindowOpenAt,omitempty"`
	ActualStartTimeAt        *time.Time `json:"actualStartTimeAt,omitempty"`
	CompletedAt              *time.Time `json:"completedAt,omitempty"`
	CancelledAt              *time.Time `json:"cancelledAt,omitempty"`

	// Configuration
	MinimumTeamCount int32  `json:"minimumTeamCount"`
	MaximumTeamCount *int32 `json:"maximumTeamCount,omitempty"`
	TeamSize         int32  `json:"teamSize"`
	AutoForceReady   bool   `json:"autoForceReady"`
	IsCompleted      bool   `json:"isCompleted"`

	// References
	GameID     *string `json:"gameId,omitempty"`
	SpaceID    *string `json:"spaceId,omitempty"`
	TemplateID *string `json:"templateId,omitempty"`

	// Rich objects
	Organizer   *TournamentOrganizer   `json:"organizer,omitempty"`
	Format      *TournamentFormat      `json:"format,omitempty"`
	ReadyWindow *TournamentReadyWindow `json:"readyWindow,omitempty"`
	Prize       *TournamentPrize       `json:"prize,omitempty"`

	// User-specific information
	Me *TournamentMeInfo `json:"me,omitempty"`

	// Participant count (for scalability - full list via separate endpoint)
	ParticipantCount int32 `json:"participantCount"`
}

// TournamentDetailResponse represents the response for a single tournament fetch.
type TournamentDetailResponse struct {
	Tournament TournamentDetail `json:"tournament"`
}

// ============================================================================
// Tournament Participants Types
// ============================================================================

// TournamentParticipant represents a participant in the tournament.
type TournamentParticipant struct {
	ID           string     `json:"id"`
	UserID       string     `json:"userId"`
	Name         string     `json:"name"`
	AvatarURL    *string    `json:"avatarUrl,omitempty"`
	IsReady      bool       `json:"isReady"`
	Status       string     `json:"status"`
	RegisteredAt time.Time  `json:"registeredAt"`
	CheckedInAt  *time.Time `json:"checkedInAt,omitempty"`
}

// ListParticipantsResponse represents the paginated response for listing participants.
type ListParticipantsResponse struct {
	Participants []TournamentParticipant `json:"participants"`
	Total        int64                   `json:"total"`
	Limit        int32                   `json:"limit"`
	Offset       int32                   `json:"offset"`
}
