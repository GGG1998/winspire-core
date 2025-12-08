package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/winspire/competition/internal/repository"
)

// InternalDeps contains dependencies for internal service-to-service handlers.
type InternalDeps struct {
	TournamentRepo   *repository.TournamentRepository
	RegistrationRepo *repository.RegistrationRepository
	Logger           *slog.Logger
}

// RegisterInternalRoutes registers internal (service-to-service) endpoints.
// These endpoints do NOT require authentication - they are for internal use only.
// Routes: /internal/tournaments/:id, /internal/tournaments/:id/registrations/:userId
func RegisterInternalRoutes(router *gin.Engine, deps InternalDeps) {
	internal := router.Group("/internal")
	{
		// GET /internal/tournaments/:id - Get tournament by ID
		internal.GET("/tournaments/:id", func(c *gin.Context) {
			tournamentID, err := uuid.Parse(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tournament ID"})
				return
			}

			tournament, err := deps.TournamentRepo.GetByID(c.Request.Context(), tournamentID)
			if err != nil {
				if errors.Is(err, repository.ErrTournamentNotFound) {
					c.JSON(http.StatusNotFound, gin.H{"error": "tournament not found"})
					return
				}

				deps.Logger.Error("failed to get tournament",
					"error", err,
					"tournamentId", tournamentID,
				)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve tournament"})
				return
			}

			// Return simplified tournament info for matchmaking
			c.JSON(http.StatusOK, gin.H{
				"id":                   tournament.ID.String(),
				"name":                 tournament.Name,
				"status":               tournament.Status,
				"scheduledStartTimeAt": tournament.ScheduledStartTimeAt,
				"minParticipants":      tournament.MinimumTeamCount,
			})
		})

		// GET /internal/tournaments/:id/registrations/:userId - Check if user is registered
		internal.GET("/tournaments/:id/registrations/:userId", func(c *gin.Context) {
			tournamentID, err := uuid.Parse(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tournament ID"})
				return
			}

			userID, err := uuid.Parse(c.Param("userId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
				return
			}

			registration, err := deps.RegistrationRepo.GetByTournamentAndUser(c.Request.Context(), tournamentID, userID)
			if err != nil {
				if errors.Is(err, repository.ErrRegistrationNotFound) {
					c.JSON(http.StatusNotFound, gin.H{
						"registered": false,
						"error":      "user not registered",
					})
					return
				}

				deps.Logger.Error("failed to check registration",
					"error", err,
					"tournamentId", tournamentID,
					"userId", userID,
				)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check registration"})
				return
			}

			// User is registered - return registration info
			c.JSON(http.StatusOK, gin.H{
				"registered":  true,
				"status":      registration.Status,
				"displayName": registration.DisplayName,
				"avatarUrl":   registration.AvatarURL,
			})
		})
	}
}





