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
type BaseResponse struct {
	Success   bool    `json:"success"`
	ErrorCode *string `json:"errorCode,omitempty"`
}

// === Tournament User Request Types ===

// JoinTournamentRequest represents the request body for joining a tournament.
type JoinTournamentRequest struct {
	SocialTeamID *uuid.UUID  `json:"socialTeamId,omitempty"`
	TeamName     *string     `json:"teamName,omitempty"`
	Players      []uuid.UUID `json:"players" binding:"required,min=1"`
}

// === Tournament User Response Types ===

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
	TemplateID *uuid.UUID `json:"templateId,omitempty"`
	TeamSize   *int       `json:"teamSize,omitempty"`
}

// CompetitionParametersInput defines the tournament configuration parameters.
type CompetitionParametersInput struct {
	TournamentParameters *TournamentParametersInput `json:"tournamentParameters,omitempty"`
}

// SpaceIdentifierInput identifies a space by ID or slug.
type SpaceIdentifierInput struct {
	ID   *uuid.UUID `json:"id,omitempty"`
	Slug *string    `json:"slug,omitempty"`
}

// GameIdentifierInput identifies a game.
type GameIdentifierInput struct {
	ID   *uuid.UUID `json:"id,omitempty"`
	Slug *string    `json:"slug,omitempty"`
}

// GameSnapshotInput represents a snapshot of game data provided by the frontend.
type GameSnapshotInput struct {
	ID          string  `json:"id" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Version     string  `json:"version" binding:"required"`
	LogoURL     *string `json:"logoUrl,omitempty"`
	Description *string `json:"description,omitempty"`
	StoragePath string  `json:"storagePath" binding:"required"`
}

// CreateTournamentRequest represents the request body for creating a tournament.
type CreateTournamentRequest struct {
	Name                     string                     `json:"name" binding:"required,min=3,max=100"`
	Description              *string                    `json:"description,omitempty" binding:"omitempty,max=4000"`
	ExternalID               *string                    `json:"externalId,omitempty"`
	CompetitionParameters    CompetitionParametersInput `json:"competitionParameters" binding:"required"`
	RegistrationWindowOpenAt *time.Time                 `json:"registrationWindowOpenAt,omitempty"`
	ScheduledStartTimeAt     *time.Time                 `json:"scheduledStartTimeAt,omitempty"`
	AutoForceReady           *bool                      `json:"autoForceReady,omitempty"`
	MinimumTeamCount         *int                       `json:"minimumTeamCount,omitempty"`
	MaximumTeamCount         *int                       `json:"maximumTeamCount,omitempty"`
	Game                     *GameIdentifierInput       `json:"game,omitempty"`
	GameSnapshot             *GameSnapshotInput         `json:"gameSnapshot,omitempty"`
	Space                    *SpaceIdentifierInput      `json:"space,omitempty"`
}

// EditTournamentRequest represents the request body for editing a tournament.
type EditTournamentRequest struct {
	Name                 *string                 `json:"name,omitempty" binding:"omitempty,min=3,max=100"`
	Description          *string                 `json:"description,omitempty" binding:"omitempty,max=4000"`
	Status               *string                 `json:"status,omitempty" binding:"omitempty,oneof=draft open scheduled registration_open started completed cancelled"`
	ScheduledStartTimeAt *time.Time              `json:"scheduledStartTimeAt,omitempty"`
	MinimumTeamCount     *int                    `json:"minimumTeamCount,omitempty"`
	MaximumTeamCount     *int                    `json:"maximumTeamCount,omitempty"`
	Format               *TournamentFormat       `json:"format,omitempty"`
	ReadyWindow          *TournamentReadyWindow  `json:"readyWindow,omitempty"`
	Prize                *TournamentPrize        `json:"prize,omitempty"`
}

// === Tournament Host Response Types ===

// CreateTournamentResponse represents the response for tournament creation.
type CreateTournamentResponse struct {
	BaseResponse
	TournamentID *string `json:"tournamentId,omitempty"`
}

// EditTournamentResponse represents the response for tournament editing.
type EditTournamentResponse = BaseResponse

// TournamentListItem represents a tournament in the list response.
type TournamentListItem struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Status               string     `json:"status"`
	ScheduledStartTimeAt *time.Time `json:"scheduledStartTimeAt,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
}

// ListTournamentsResponse represents the response for listing tournaments.
type ListTournamentsResponse struct {
	Tournaments []TournamentListItem `json:"tournaments"`
	Count       int                  `json:"count"`
}

// TournamentOrganizer represents the organizer/host information.
type TournamentOrganizer struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	LogoURL *string `json:"logoUrl,omitempty"`
}

// TournamentFormat represents the tournament format configuration.
type TournamentFormat struct {
	Type     string `json:"type"`
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
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Value       *int    `json:"value,omitempty"`
	Currency    *string `json:"currency,omitempty"`
}

// TournamentMeInfo represents user-specific tournament information.
type TournamentMeInfo struct {
	ParticipationStatus *string    `json:"participationStatus,omitempty"`
	Match               *MatchInfo `json:"match,omitempty"`
}

// MatchInfo represents the current/last active match for the user
type MatchInfo struct {
	MatchID      string     `json:"matchId"`
	TournamentID string     `json:"tournamentId"`
	Status       string     `json:"status"`
	Round        int32      `json:"round"`
	Table        int32      `json:"table"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// TournamentDetail represents a detailed tournament payload.
type TournamentDetail struct {
	ID        string `json:"id"`
	HostID    string `json:"hostId"`
	CreatorID string `json:"creatorId"`

	Name         string     `json:"name"`
	Status       string     `json:"status"`
	Description  *string    `json:"description,omitempty"`
	ExternalID   *string    `json:"externalId,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	LastActivity *time.Time `json:"lastActivity,omitempty"`

	Game        *string `json:"game,omitempty"`
	GameLogoURL *string `json:"gameLogoUrl,omitempty"`
	BannerURL   *string `json:"bannerUrl,omitempty"`
	RoomLink    *string `json:"roomLink,omitempty"`

	ScheduledStartTimeAt     *time.Time `json:"scheduledStartTimeAt,omitempty"`
	RegistrationWindowOpenAt *time.Time `json:"registrationWindowOpenAt,omitempty"`
	ActualStartTimeAt        *time.Time `json:"actualStartTimeAt,omitempty"`
	CompletedAt              *time.Time `json:"completedAt,omitempty"`
	CancelledAt              *time.Time `json:"cancelledAt,omitempty"`

	MinimumTeamCount int32  `json:"minimumTeamCount"`
	MaximumTeamCount *int32 `json:"maximumTeamCount,omitempty"`
	TeamSize         int32  `json:"teamSize"`
	AutoForceReady   bool   `json:"autoForceReady"`
	IsCompleted      bool   `json:"isCompleted"`

	GameID     *string `json:"gameId,omitempty"`
	SpaceID    *string `json:"spaceId,omitempty"`
	TemplateID *string `json:"templateId,omitempty"`

	Organizer   *TournamentOrganizer   `json:"organizer,omitempty"`
	Format      *TournamentFormat      `json:"format,omitempty"`
	ReadyWindow *TournamentReadyWindow `json:"readyWindow,omitempty"`
	Prize       *TournamentPrize       `json:"prize,omitempty"`

	Me *TournamentMeInfo `json:"me,omitempty"`

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
