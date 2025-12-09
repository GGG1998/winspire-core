package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/winspire/game-management/internal/repository"
)

// GameDeps contains dependencies for game handlers.
type GameDeps struct {
	Repo *repository.GameRepository
}

// RegisterGameRoutes registers public game routes.
func RegisterGameRoutes(group *gin.RouterGroup, deps GameDeps) {
	// List all active games
	group.GET("/games", func(c *gin.Context) {
		games, err := deps.Repo.List(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list games", Details: err.Error()})
			return
		}

		response := GamesListResponse{
			Games: make([]GameResponse, 0, len(games)),
			Total: len(games),
		}

		for _, game := range games {
			response.Games = append(response.Games, gameToResponse(game))
		}

		c.JSON(http.StatusOK, response)
	})

	// Get game by ID
	group.GET("/games/:gameId", func(c *gin.Context) {
		gameID, err := uuid.Parse(c.Param("gameId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid game ID"})
			return
		}

		game, err := deps.Repo.GetByID(c.Request.Context(), gameID)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "game not found"})
			return
		}

		// Only return active games for public API
		if !game.IsActive {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "game not found"})
			return
		}

		c.JSON(http.StatusOK, gameToResponse(*game))
	})

	// Get game by slug
	group.GET("/g/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "slug is required"})
			return
		}

		game, err := deps.Repo.GetBySlug(c.Request.Context(), slug)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "game not found"})
			return
		}

		c.JSON(http.StatusOK, gameToResponse(*game))
	})
}

func gameToResponse(game repository.Game) GameResponse {
	resp := GameResponse{
		ID:          game.ID.String(),
		Slug:        game.Slug,
		Name:        game.Name,
		Version:     game.Version,
		StoragePath: game.StoragePath,
		IsActive:    game.IsActive,
		CreatedAt:   game.CreatedAt,
		UpdatedAt:   game.UpdatedAt,
	}

	if game.GameIntegrationID != nil {
		id := game.GameIntegrationID.String()
		resp.GameIntegrationID = &id
	}
	if game.Description != nil {
		resp.Description = game.Description
	}
	if game.LogoURL != nil {
		resp.LogoURL = game.LogoURL
	}

	return resp
}
