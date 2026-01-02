package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	sharedhttp "github.com/winspire/winspire-core/libs/go/httpx"

	"github.com/winspire/tournament/internal/domain"
	"github.com/winspire/tournament/internal/pubsub"
	"github.com/winspire/tournament/internal/repository"
)

// TournamentHostDeps contains dependencies for tournament host handlers.
type TournamentHostDeps struct {
	HostRepo           *repository.HostRepository
	TournamentRepo     *repository.TournamentRepository
	RegistrationRepo   *repository.RegistrationRepository
	ParticipantRepo    domain.ParticipantRepository
	EventPublisher     *pubsub.EventPublisher
	Logger             *slog.Logger
	MatchmakingBaseURL string
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
			tournamentItems, err := deps.TournamentRepo.ListByHostID(c.Request.Context(), hostID)
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
			visibleTournaments := make([]repository.TournamentListItem, 0)
			for _, t := range tournamentItems {
				// Convert TournamentListItem to Tournament for authorization check
				fullTournament := &repository.Tournament{
					ID:     t.ID,
					HostID: hostID,
					Status: t.Status,
				}
				if CanViewTournament(userID, fullTournament) {
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

		// NOTE: GET /:tournamentId moved to tournament_user.go (no host auth required for viewing)

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
					if err := tournament.Start(); err != nil {
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

					// Publish TournamentStartRequested event (command/intent) after status change to "starting"
					// This initiates the tournament start saga: grace period → bracket generation → confirmation
					if deps.EventPublisher != nil {
						// Get registered participants count for the tournament
						participantCount, err := deps.ParticipantRepo.CountByTournamentAndStatus(c.Request.Context(), tournamentID, "registered")
						if err != nil {
							deps.Logger.Warn("failed to count participants for event",
								"error", err,
								"tournamentId", tournamentID,
							)
							participantCount = 0
						}

						// Get registrations to extract participant IDs
						participantIDs := []string{}
						if deps.RegistrationRepo != nil {
							// Use registration repository to get all registered users
							registrations, err := deps.RegistrationRepo.ListByTournament(c.Request.Context(), tournamentID, 1000, 0)
							if err != nil {
								deps.Logger.Warn("failed to fetch registrations for event",
									"error", err,
									"tournamentId", tournamentID,
								)
							} else {
								for _, reg := range registrations {
									if reg.Status == "registered" {
										participantIDs = append(participantIDs, reg.UserID.String())
									}
								}
							}
						}

						// Prepare event payload
						eventPayload := map[string]interface{}{
							"tournament_id": tournamentID.String(),
							"host_id":       tournament.HostID.String(),
							"participants":  participantIDs,
							"started_at":    time.Now().Format(time.RFC3339),
						}

						// Add game_id if available
						// TODO clean, we use GameSnapshot
						if tournament.GameID != nil {
							eventPayload["game_id"] = tournament.GameID.String()
						}

						// Add game snapshot if available (for performance optimization)
						if tournament.GameSnapshot != nil {
							var gameSnapshotMap map[string]interface{}
							if err := json.Unmarshal(tournament.GameSnapshot, &gameSnapshotMap); err == nil {
								eventPayload["game_snapshot"] = gameSnapshotMap
							}
						}

						// Publish TournamentStartRequested event (represents intent, not fact)
						event := pubsub.DomainEvent{
							EventID:       uuid.New().String(),
							EventType:     "TournamentStartRequested",
							AggregateID:   tournamentID.String(),
							AggregateType: "Tournament",
							Timestamp:     time.Now(),
							Payload:       eventPayload,
							Metadata: map[string]string{
								"correlation_id": c.GetString("correlation_id"),
							},
						}

						if err := deps.EventPublisher.Publish(context.Background(), pubsub.ChannelTournamentStartRequested, event); err != nil {
							deps.Logger.Error("failed to publish TournamentStartRequested event",
								"error", err,
								"tournamentId", tournamentID,
							)
							// Don't fail the request if event publishing fails
						} else {
							deps.Logger.Info("published TournamentStartRequested event",
								"tournamentId", tournamentID,
								"participantCount", participantCount,
							)
						}
					}
				case "completed":
					if err := tournament.Complete(); err != nil {
						deps.Logger.Warn("failed to complete tournament",
							"error", err,
							"tournamentId", tournamentID,
							"currentStatus", tournament.Status,
						)
						c.JSON(http.StatusBadRequest, ErrorResponse{
							Error:   "cannot complete tournament",
							Details: err.Error(),
						})
						return
					}
				case "cancelled":
					if err := tournament.Cancel(); err != nil {
						deps.Logger.Warn("failed to cancel tournament",
							"error", err,
							"tournamentId", tournamentID,
							"currentStatus", tournament.Status,
						)
						c.JSON(http.StatusBadRequest, ErrorResponse{
							Error:   "cannot cancel tournament",
							Details: err.Error(),
						})
						return
					}
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

				updateParams := repository.UpdateDetailsParams{
					Name:                 req.Name,
					Description:          req.Description,
					ScheduledStartTimeAt: req.ScheduledStartTimeAt,
					MinimumTeamCount:     minCount,
					MaximumTeamCount:     maxCount,
				}

				if err := tournament.UpdateDetails(updateParams); err != nil {
					deps.Logger.Warn("failed to update tournament details",
						"error", err,
						"tournamentId", tournamentID,
						"status", tournament.Status,
					)
					c.JSON(http.StatusBadRequest, ErrorResponse{
						Error:   "cannot update tournament",
						Details: err.Error(),
					})
					return
				}
			}

			// Update JSONB fields if provided
			if req.ReadyWindow != nil {
				readyWindowJSON, err := json.Marshal(req.ReadyWindow)
				if err != nil {
					deps.Logger.Warn("failed to marshal ready window",
						"error", err,
						"tournamentId", tournamentID,
					)
					c.JSON(http.StatusBadRequest, ErrorResponse{
						Error:   "invalid ready window format",
						Details: err.Error(),
					})
					return
				}
				tournament.ReadyWindow = readyWindowJSON
			}

			if req.Prize != nil {
				prizeJSON, err := json.Marshal(req.Prize)
				if err != nil {
					deps.Logger.Warn("failed to marshal prize",
						"error", err,
						"tournamentId", tournamentID,
					)
					c.JSON(http.StatusBadRequest, ErrorResponse{
						Error:   "invalid prize format",
						Details: err.Error(),
					})
					return
				}
				tournament.Prize = prizeJSON
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

		// NOTE: GET /:tournamentId/participants moved to tournament_user.go (no host auth required for viewing)
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

// newTournamentDetail converts repository model into API response payload.
func newTournamentDetail(t *repository.Tournament) TournamentDetail {
	hostID := t.HostID.String()
	tournamentID := t.ID.String()

	// Generate room link for participants
	roomLink := fmt.Sprintf("/tournaments/%s/lobby", tournamentID)

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
		ParticipantCount:         0, // TODO: Fetch from tournament_registrations table
	}

	// Set last activity
	lastActivity := t.UpdatedAt
	detail.LastActivity = &lastActivity

	// Set game information
	// TODO: Fetch game details from games table/service when GameID is present
	if t.GameID != nil {
		gameID := t.GameID.String()
		detail.GameID = &gameID
		// Default game info - should be fetched from games repository
		gameSlug := "packman"
		detail.Game = &gameSlug
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

	// Parse ready window from JSONB
	if len(t.ReadyWindow) > 0 {
		var readyWindow TournamentReadyWindow
		if err := json.Unmarshal(t.ReadyWindow, &readyWindow); err == nil {
			detail.ReadyWindow = &readyWindow
		}
	}

	// Parse prize from JSONB
	if len(t.Prize) > 0 {
		var prize TournamentPrize
		if err := json.Unmarshal(t.Prize, &prize); err == nil {
			detail.Prize = &prize
		}
	}

	return detail
}

// fetchCurrentMatch calls matchmaking /me endpoint and returns current match info (if any)
func fetchCurrentMatch(ctx context.Context, baseURL, authHeader string, logger *slog.Logger) (*MatchInfo, error) {
	base := strings.TrimRight(baseURL, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/v1/matchmaking/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("matchmaking /me returned status %d", resp.StatusCode)
	}

	var payload struct {
		Match *MatchInfo `json:"match"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	// If no active match, return nil without error
	if payload.Match == nil {
		return nil, nil
	}

	// ensure required IDs are non-empty
	if payload.Match.MatchID == "" || payload.Match.TournamentID == "" {
		logger.Warn("matchmaking /me returned match without IDs")
		return nil, nil
	}

	return payload.Match, nil
}
