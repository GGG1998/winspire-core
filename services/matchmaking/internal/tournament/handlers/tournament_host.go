package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	sharedhttp "github.com/winspire/winspire-core/libs/go/httpx"

	"github.com/winspire-core/services/matchmaking/internal/tournament/domain"
	"github.com/winspire-core/services/matchmaking/internal/tournament/repository"
)

// StartTournamentFunc is a function type for starting a tournament.
// This replaces the event-based TournamentStartRequested pattern with direct invocation.
type StartTournamentFunc func(ctx gin.Context, tournamentID uuid.UUID, participantIDs []uuid.UUID) error

// TournamentHostDeps contains dependencies for tournament host handlers.
type TournamentHostDeps struct {
	HostRepo           *repository.HostRepository
	TournamentRepo     *repository.TournamentRepository
	RegistrationRepo   *repository.RegistrationRepository
	Logger             *slog.Logger
	// StartTournamentFunc is called directly instead of publishing TournamentStartRequested event
	StartTournament    StartTournamentFunc
}

// RegisterTournamentHostRoutes registers host/admin-facing tournament routes.
// Routes are scoped under /:hostId and require host admin authorization.
func RegisterTournamentHostRoutes(group *gin.RouterGroup, deps TournamentHostDeps) {
	// User's hosts endpoint (not scoped by hostId)
	// GET /v1/hosts - Get all hosts for the authenticated user
	group.GET("/hosts", func(c *gin.Context) {
		// Get user from context (set by JWT middleware)
		user, ok := sharedhttp.GetUser(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
			return
		}

		userID, err := uuid.Parse(string(user.ID))
		if err != nil {
			deps.Logger.Error("invalid user ID in JWT", "userId", user.ID)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "invalid user context"})
			return
		}

		// Get all hosts for user
		hosts, err := deps.HostRepo.GetUserHosts(c.Request.Context(), userID)
		if err != nil {
			deps.Logger.Error("failed to get user hosts", "error", err, "userId", userID)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve hosts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"hosts": hosts,
			"count": len(hosts),
		})
	})

	// Host-scoped tournament routes: /v1/:hostId/tournaments/...
	hostGroup := group.Group("/:hostId")

	// Apply host authorization middleware to all routes in this group
	hostGroup.Use(hostAuthMiddleware(deps.HostRepo, deps.Logger))

	tournaments := hostGroup.Group("/tournaments")
	{
		// GET /v1/:hostId/tournaments - List all tournaments for the host
		tournaments.GET("", func(c *gin.Context) {
			hostID, _ := uuid.Parse(c.Param("hostId"))

			deps.Logger.Info("list tournaments request", "hostId", hostID)

			// Get tournaments from repository
			tournamentItems, err := deps.TournamentRepo.ListByHost(c.Request.Context(), hostID)
			if err != nil {
				deps.Logger.Error("failed to list tournaments", "error", err, "hostId", hostID)
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error: "failed to retrieve tournaments",
				})
				return
			}

			// Get current user for authorization checks
			user, _ := sharedhttp.GetUser(c)
			userID, _ := uuid.Parse(string(user.ID))

			// Filter tournaments based on permissions (draft tournaments are private)
			visibleTournaments := make([]*domain.Tournament, 0)
			for _, t := range tournamentItems {
				if CanViewTournament(userID, t) {
					visibleTournaments = append(visibleTournaments, t)
				}
			}

			// Convert to response format
			responseItems := make([]TournamentListItem, len(visibleTournaments))
			for i, t := range visibleTournaments {
				responseItems[i] = TournamentListItem{
					ID:                   t.ID.String(),
					Name:                 t.Name,
					Status:               t.Status,
					ScheduledStartTimeAt: t.ScheduledStartTimeAt,
					CreatedAt:            t.CreatedAt,
				}
			}

			c.JSON(http.StatusOK, ListTournamentsResponse{
				Tournaments: responseItems,
				Count:       len(responseItems),
			})
		})

		// POST /v1/:hostId/tournaments - Create a tournament
		tournaments.POST("", func(c *gin.Context) {
			hostID, _ := uuid.Parse(c.Param("hostId"))

			var req CreateTournamentRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "invalid request body",
					Details: err.Error(),
				})
				return
			}

			// Apply policy to convert request to repository params
			params := ApplyCreateTournamentPolicy(hostID, req)

			// Create tournament in database
			tournament, err := deps.TournamentRepo.Create(c.Request.Context(), params)
			if err != nil {
				deps.Logger.Error("failed to create tournament", "error", err, "hostId", hostID)
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error: "failed to create tournament",
				})
				return
			}

			deps.Logger.Info("tournament created",
				"hostId", hostID,
				"tournamentId", tournament.ID,
				"name", req.Name,
			)

			tournamentID := tournament.ID.String()
			c.JSON(http.StatusCreated, CreateTournamentResponse{
				BaseResponse: BaseResponse{
					Success: true,
				},
				TournamentID: &tournamentID,
			})
		})

		// PUT /v1/:hostId/tournaments/:tournamentId - Edit a tournament (DDD aggregate pattern)
		// Supports partial updates including status transitions (start, cancel, open, complete)
		tournaments.PUT("/:tournamentId", func(c *gin.Context) {
			hostID, _ := uuid.Parse(c.Param("hostId"))

			tournamentID, err := uuid.Parse(c.Param("tournamentId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid tournament ID"})
				return
			}

			var req EditTournamentRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "invalid request body",
					Details: err.Error(),
				})
				return
			}

			// Load tournament aggregate
			tournament, err := deps.TournamentRepo.GetByHostAndID(c.Request.Context(), hostID, tournamentID)
			if err != nil {
				if errors.Is(err, repository.ErrTournamentNotFound) {
					c.JSON(http.StatusNotFound, ErrorResponse{Error: "tournament not found"})
					return
				}
				deps.Logger.Error("failed to get tournament",
					"error", err,
					"hostId", hostID,
					"tournamentId", tournamentID,
				)
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve tournament"})
				return
			}

			// Apply status transitions using aggregate methods
			if req.Status != nil {
				switch *req.Status {
				case "open":
					if err := tournament.Publish(); err != nil {
						deps.Logger.Warn("failed to publish tournament",
							"error", err,
							"tournamentId", tournamentID,
							"currentStatus", tournament.Status,
						)
						c.JSON(http.StatusBadRequest, ErrorResponse{
							Error:   "cannot publish tournament",
							Details: err.Error(),
						})
						return
					}
				case "registration_open":
					if err := tournament.OpenRegistration(); err != nil {
						deps.Logger.Warn("failed to open registration",
							"error", err,
							"tournamentId", tournamentID,
							"currentStatus", tournament.Status,
						)
						c.JSON(http.StatusBadRequest, ErrorResponse{
							Error:   "cannot open registration",
							Details: err.Error(),
						})
						return
					}
				case "started":
					if err := tournament.RequestStart(); err != nil {
						deps.Logger.Warn("failed to start tournament",
							"error", err,
							"tournamentId", tournamentID,
							"currentStatus", tournament.Status,
						)
						c.JSON(http.StatusBadRequest, ErrorResponse{
							Error:   "cannot start tournament",
							Details: err.Error(),
						})
						return
					}

					// DIRECT FUNCTION CALL instead of pub/sub event
					// Get registered participants for the tournament
					if deps.StartTournament != nil {
						participantIDs, err := deps.RegistrationRepo.GetConfirmedUserIDs(c.Request.Context(), tournamentID)
						if err != nil {
							deps.Logger.Warn("failed to get participants for tournament start",
								"error", err,
								"tournamentId", tournamentID,
							)
							participantIDs = []uuid.UUID{}
						}

						// Call start tournament function directly (replaces TournamentStartRequested event)
						if err := deps.StartTournament(*c, tournamentID, participantIDs); err != nil {
							deps.Logger.Error("failed to start tournament via direct call",
								"error", err,
								"tournamentId", tournamentID,
							)
							// Don't fail the request - tournament status is already "starting"
							// The saga will handle retries
						} else {
							deps.Logger.Info("tournament start initiated via direct call",
								"tournamentId", tournamentID,
								"participantCount", len(participantIDs),
							)
						}
					}
				case "completed":
					tournament.Complete()
				case "cancelled":
					tournament.Cancel()
				default:
					c.JSON(http.StatusBadRequest, ErrorResponse{
						Error: "invalid status value",
					})
					return
				}
			}

			// Apply detail updates using aggregate method
			if req.Name != nil || req.Description != nil || req.ScheduledStartTimeAt != nil ||
				req.MinimumTeamCount != nil || req.MaximumTeamCount != nil {

				var minCount *int32
				if req.MinimumTeamCount != nil {
					val := int32(*req.MinimumTeamCount)
					minCount = &val
				}

				var maxCount *int32
				if req.MaximumTeamCount != nil {
					val := int32(*req.MaximumTeamCount)
					maxCount = &val
				}

				tournament.UpdateDetails(req.Name, req.Description, req.ScheduledStartTimeAt, minCount, maxCount)
			}

			// Update JSONB fields if provided
			if req.ReadyWindow != nil {
				tournament.ReadyWindow = &domain.ReadyWindow{
					StartsAt: &req.ReadyWindow.StartsAt,
					EndsAt:   &req.ReadyWindow.EndsAt,
				}
			}

			if req.Prize != nil {
				var value float64
				if req.Prize.Value != nil {
					value = float64(*req.Prize.Value)
				}
				var currency string
				if req.Prize.Currency != nil {
					currency = *req.Prize.Currency
				}
				tournament.Prize = &domain.Prize{
					Type:        req.Prize.Type,
					Description: req.Prize.Description,
					Value:       value,
					Currency:    currency,
				}
			}

			// Save aggregate state
			updatedTournament, err := deps.TournamentRepo.Save(c.Request.Context(), tournament)
			if err != nil {
				deps.Logger.Error("failed to save tournament",
					"error", err,
					"tournamentId", tournamentID,
				)
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update tournament"})
				return
			}

			deps.Logger.Info("tournament updated",
				"hostId", hostID,
				"tournamentId", tournamentID,
				"status", updatedTournament.Status,
			)

			c.JSON(http.StatusOK, EditTournamentResponse{
				Success: true,
			})
		})
	}
}

// parsePositiveInt32 parses a string to int32
func parsePositiveInt32(s string) (int32, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}

// hostAuthMiddleware creates a middleware that checks if the authenticated user
// is an admin/owner of the host specified in the path.
// Only users with user_type = "streamer" can be hosts and manage tournaments.
// If the hostID matches the userID and the host doesn't exist, it auto-creates a personal host.
func hostAuthMiddleware(hostRepo *repository.HostRepository, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse hostId from path
		hostID, err := uuid.Parse(c.Param("hostId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid host ID"})
			c.Abort()
			return
		}

		// Get user from context (set by JWT middleware)
		user, ok := sharedhttp.GetUser(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
			c.Abort()
			return
		}

		// Check if user is a streamer - only streamers can be hosts
		if user.UserType != "streamer" {
			logger.Warn("non-streamer user attempted to access host routes",
				"userId", user.ID,
				"userType", user.UserType,
				"hostId", hostID,
			)
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error: "only streamers can host tournaments",
				Code:  "NOT_STREAMER",
			})
			c.Abort()
			return
		}

		// Parse user ID
		userID, err := uuid.Parse(string(user.ID))
		if err != nil {
			logger.Error("invalid user ID in JWT", "userId", user.ID)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "invalid user context"})
			c.Abort()
			return
		}

		// Get or create host for user (auto-creates if hostID == userID)
		host, err := hostRepo.GetOrCreateHostForUser(c.Request.Context(), hostID, userID)
		if err != nil {
			logger.Error("failed to get or create host",
				"error", err,
				"hostId", hostID,
				"userId", userID,
			)
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error: "insufficient permissions",
				Code:  "NOT_HOST_ADMIN",
			})
			c.Abort()
			return
		}

		// Store host in context for later use
		c.Set("host", host)

		c.Next()
	}
}
