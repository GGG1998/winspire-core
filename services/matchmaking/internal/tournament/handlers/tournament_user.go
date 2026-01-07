package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sharedhttp "github.com/winspire/winspire-core/libs/go/httpx"

	"github.com/winspire-core/services/matchmaking/internal/tournament/repository"
)

// ConfirmParticipationFunc is a function type for confirming participation.
type ConfirmParticipationFunc func(ctx *gin.Context, tournamentID, userID uuid.UUID, displayName string) error

// JoinTournamentFunc is a function type for joining a tournament.
type JoinTournamentFunc func(ctx *gin.Context, tournamentID, userID, hostID uuid.UUID) (string, error)

// TournamentUserDeps contains dependencies for tournament user handlers.
type TournamentUserDeps struct {
	Pool                   *pgxpool.Pool
	Logger                 *slog.Logger
	TournamentRepo         *repository.TournamentRepository
	RegistrationRepo       *repository.RegistrationRepository
	ConfirmParticipationFn ConfirmParticipationFunc
	JoinTournamentFn       JoinTournamentFunc
}

// RegisterTournamentUserRoutes registers player-facing tournament routes.
// Routes are scoped under /:hostId for multi-tenant support.
func RegisterTournamentUserRoutes(group *gin.RouterGroup, deps TournamentUserDeps) {
	// Host-scoped tournament routes: /v1/:hostId/tournaments/...
	hostGroup := group.Group("/:hostId")

	tournaments := hostGroup.Group("/tournaments")
	{
		// GET /v1/:hostId/tournaments/:tournamentId - Get tournament details (any authenticated user)
		// Authorization: Draft tournaments only visible to owner, all others are public
		tournaments.GET("/:tournamentId", func(c *gin.Context) {
			tournamentID, err := uuid.Parse(c.Param("tournamentId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid tournament ID"})
				return
			}

			// Use GetByID to allow viewing tournaments from any host
			tournament, err := deps.TournamentRepo.GetByID(c.Request.Context(), tournamentID)
			if err != nil {
				if errors.Is(err, repository.ErrTournamentNotFound) {
					c.JSON(http.StatusNotFound, ErrorResponse{Error: "tournament not found"})
					return
				}

				deps.Logger.Error("failed to get tournament detail",
					"error", err,
					"tournamentId", tournamentID,
				)
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error: "failed to retrieve tournament",
				})
				return
			}

			// Check if user can view this tournament (draft tournaments are private)
			user, _ := sharedhttp.GetUser(c)
			userID, _ := uuid.Parse(string(user.ID))
			if !CanViewTournament(userID, tournament) {
				c.JSON(http.StatusForbidden, ErrorResponse{
					Error: "You do not have permission to view this tournament",
					Code:  "FORBIDDEN",
				})
				return
			}

			// Get participant count
			registeredCount, err := deps.RegistrationRepo.CountByTournamentAndStatus(c.Request.Context(), tournamentID, "registered")
			if err != nil {
				deps.Logger.Warn("failed to count registered participants", "error", err, "tournamentId", tournamentID)
				registeredCount = 0
			}

			detail := newTournamentDetail(tournament)
			detail.ParticipantCount = int32(registeredCount)

			// Get user's participation status
			if user.ID != "" {
				participant, err := deps.RegistrationRepo.GetByTournamentAndUser(c.Request.Context(), tournamentID, userID)
				if err != nil {
					if !errors.Is(err, repository.ErrRegistrationNotFound) {
						deps.Logger.Warn("failed to get user participation status", "error", err, "tournamentId", tournamentID, "userId", userID)
					}
					// User is not a participant - leave Me as nil
				} else {
					// User is a participant - set participation status
					status := participant.Status
					detail.Me = &TournamentMeInfo{
						ParticipationStatus: &status,
					}
				}
			}

			c.JSON(http.StatusOK, TournamentDetailResponse{
				Tournament: detail,
			})
		})

		// GET /v1/:hostId/tournaments/:tournamentId/participants - List tournament participants (paginated)
		// Any authenticated user can view participants of a public tournament
		tournaments.GET("/:tournamentId/participants", func(c *gin.Context) {
			tournamentID, err := uuid.Parse(c.Param("tournamentId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid tournament ID"})
				return
			}

			// Parse pagination params
			limit := int32(50) // Default limit
			offset := int32(0) // Default offset

			if limitStr := c.Query("limit"); limitStr != "" {
				if l, err := parsePositiveInt32(limitStr); err == nil && l > 0 && l <= 100 {
					limit = l
				}
			}
			if offsetStr := c.Query("offset"); offsetStr != "" {
				if o, err := parsePositiveInt32(offsetStr); err == nil && o >= 0 {
					offset = o
				}
			}

			// Verify tournament exists and user can view it
			tournament, err := deps.TournamentRepo.GetByID(c.Request.Context(), tournamentID)
			if err != nil {
				if errors.Is(err, repository.ErrTournamentNotFound) {
					c.JSON(http.StatusNotFound, ErrorResponse{Error: "tournament not found"})
					return
				}
				deps.Logger.Error("failed to verify tournament",
					"error", err,
					"tournamentId", tournamentID,
				)
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve tournament"})
				return
			}

			// Check if user can view this tournament
			user, _ := sharedhttp.GetUser(c)
			userID, _ := uuid.Parse(string(user.ID))
			if !CanViewTournament(userID, tournament) {
				c.JSON(http.StatusForbidden, ErrorResponse{
					Error: "You do not have permission to view this tournament",
					Code:  "FORBIDDEN",
				})
				return
			}

			// Get participants
			registrations, err := deps.RegistrationRepo.ListByTournament(c.Request.Context(), tournamentID)
			if err != nil {
				deps.Logger.Error("failed to list participants",
					"error", err,
					"tournamentId", tournamentID,
				)
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve participants"})
				return
			}

			// Get total count
			total, err := deps.RegistrationRepo.CountByTournament(c.Request.Context(), tournamentID)
			if err != nil {
				deps.Logger.Error("failed to count participants",
					"error", err,
					"tournamentId", tournamentID,
				)
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to count participants"})
				return
			}

			// Apply pagination manually (since we don't have paginated query)
			start := int(offset)
			end := start + int(limit)
			if start > len(registrations) {
				start = len(registrations)
			}
			if end > len(registrations) {
				end = len(registrations)
			}
			paginatedRegs := registrations[start:end]

			// Convert to response format
			participants := make([]TournamentParticipant, len(paginatedRegs))
			for i, reg := range paginatedRegs {
				var registeredAt *time.Time
				if reg.RegisteredAt != nil {
					registeredAt = reg.RegisteredAt
				}
				participants[i] = TournamentParticipant{
					ID:           reg.ID.String(),
					UserID:       reg.UserID.String(),
					Name:         reg.DisplayName,
					AvatarURL:    reg.AvatarURL,
					IsReady:      reg.IsReady,
					Status:       reg.Status,
					RegisteredAt: *registeredAt,
					CheckedInAt:  reg.CheckedInAt,
				}
			}

			c.JSON(http.StatusOK, ListParticipantsResponse{
				Participants: participants,
				Total:        total,
				Limit:        limit,
				Offset:       offset,
			})
		})

		// POST /v1/:hostId/tournaments/:tournamentId/join - Join a tournament (requires confirmed participation)
		tournaments.POST("/:tournamentId/join", func(c *gin.Context) {
			hostID, err := uuid.Parse(c.Param("hostId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid host ID"})
				return
			}

			tournamentID, err := uuid.Parse(c.Param("tournamentId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid tournament ID"})
				return
			}

			// Get user from context (set by JWT middleware)
			user, ok := sharedhttp.GetUser(c)
			if !ok {
				c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
				return
			}
			userID, err := uuid.Parse(string(user.ID))
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "invalid user ID"})
				return
			}

			// Execute use case via function
			if deps.JoinTournamentFn == nil {
				c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "join tournament not implemented"})
				return
			}

			roomLink, err := deps.JoinTournamentFn(c, tournamentID, userID, hostID)
			if err != nil {
				deps.Logger.Error("failed to join tournament", "error", err, "tournamentId", tournamentID, "userId", userID)
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}

			deps.Logger.Info("user joined tournament", "hostId", hostID, "tournamentId", tournamentID, "userId", userID)

			c.JSON(http.StatusOK, JoinTournamentResponse{
				BaseResponse: BaseResponse{
					Success: true,
				},
				RoomLink: roomLink,
			})
		})

		// POST /v1/:hostId/tournaments/:tournamentId/leave - Leave a tournament
		tournaments.POST("/:tournamentId/leave", func(c *gin.Context) {
			hostID, err := uuid.Parse(c.Param("hostId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid host ID"})
				return
			}

			tournamentID, err := uuid.Parse(c.Param("tournamentId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid tournament ID"})
				return
			}

			// TODO: Implement leave logic using use case
			deps.Logger.Info("tournament leave request",
				"hostId", hostID,
				"tournamentId", tournamentID,
			)

			c.JSON(http.StatusOK, LeaveTournamentResponse{
				Success: true,
			})
		})

		// POST /v1/:hostId/tournaments/:tournamentId/confirm-participation - Register for tournament (sets status to 'registered')
		tournaments.POST("/:tournamentId/confirm-participation", func(c *gin.Context) {
			hostID, err := uuid.Parse(c.Param("hostId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid host ID"})
				return
			}

			tournamentID, err := uuid.Parse(c.Param("tournamentId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid tournament ID"})
				return
			}

			// Get user from context (set by JWT middleware)
			user, ok := sharedhttp.GetUser(c)
			if !ok {
				c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
				return
			}
			userID, err := uuid.Parse(string(user.ID))
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "invalid user ID"})
				return
			}

			// Use nickname from JWT (enriched by custom access token hook)
			// Fallback to email if nickname is empty
			displayName := user.Nickname
			if displayName == "" {
				displayName = string(user.Email)
			}

			// Execute use case via function
			if deps.ConfirmParticipationFn == nil {
				c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "confirm participation not implemented"})
				return
			}

			if err := deps.ConfirmParticipationFn(c, tournamentID, userID, displayName); err != nil {
				deps.Logger.Error("failed to register participation", "error", err, "tournamentId", tournamentID, "userId", userID)
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}

			deps.Logger.Info("participation registered", "hostId", hostID, "tournamentId", tournamentID, "userId", userID)

			c.JSON(http.StatusOK, ConfirmTournamentParticipationResponse{
				Success: true,
			})
		})
	}

	matches := hostGroup.Group("/matches")
	{
		// POST /v1/:hostId/matches/:matchId/confirm-participation - Confirm participation in match
		matches.POST("/:matchId/confirm-participation", func(c *gin.Context) {
			hostID, err := uuid.Parse(c.Param("hostId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid match ID"})
				return
			}

			matchID, err := uuid.Parse(c.Param("matchId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid match ID"})
				return
			}

			// TODO: Implement match confirmation logic using use case
			deps.Logger.Info("match confirm participation request",
				"hostId", hostID,
				"matchId", matchID,
			)

			c.JSON(http.StatusOK, ConfirmMatchParticipationResponse{
				Success: true,
			})
		})
	}
}
