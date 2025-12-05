package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sharedhttp "github.com/winspire/winspire-core/libs/go/httpx"

	"github.com/winspire/competition/internal/application"
)

// TournamentUserDeps contains dependencies for tournament user handlers.
type TournamentUserDeps struct {
	Pool                   *pgxpool.Pool
	Logger                 *slog.Logger
	ConfirmParticipationUC *application.ConfirmParticipationUseCase
	JoinTournamentUC       *application.JoinTournamentUseCase
}

// RegisterTournamentUserRoutes registers player-facing tournament routes.
// Routes are scoped under /:hostId for multi-tenant support.
func RegisterTournamentUserRoutes(group *gin.RouterGroup, deps TournamentUserDeps) {
	// Host-scoped tournament routes: /v1/:hostId/tournaments/...
	hostGroup := group.Group("/:hostId")

	tournaments := hostGroup.Group("/tournaments")
	{
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

			// Execute use case
			roomLink, err := deps.JoinTournamentUC.Execute(c.Request.Context(), tournamentID, userID, hostID)
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
			// 1. Validate tournament exists and belongs to host
			// 2. Validate user is signed up
			// 3. Remove participation record
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
			// Fallback to email if nickname is empty (shouldn't happen if profile completion is enforced)
			displayName := user.Nickname
			if displayName == "" {
				displayName = string(user.Email)
			}

			// Execute use case with displayName
			if err := deps.ConfirmParticipationUC.Execute(c.Request.Context(), tournamentID, userID, displayName); err != nil {
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
			// 1. Validate match exists and belongs to host
			// 2. Validate confirmation is required
			// 3. Validate user is a participant in the match
			// 4. Update match participation status
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
