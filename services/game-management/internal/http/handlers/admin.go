package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/winspire/game-management/internal/repository"
	"github.com/winspire/game-management/internal/storage"
)

// AdminDeps contains dependencies for admin handlers.
type AdminDeps struct {
	Repo    *repository.GameRepository
	Storage *storage.Client
}

// RegisterAdminRoutes registers admin-only game routes.
func RegisterAdminRoutes(group *gin.RouterGroup, deps AdminDeps) {
	// List all games (including inactive)
	group.GET("/games", func(c *gin.Context) {
		games, err := deps.Repo.ListAll(c.Request.Context())
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

	// Create a new game
	group.POST("/games", func(c *gin.Context) {
		var req CreateGameRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Details: err.Error()})
			return
		}

		// Generate storage path based on slug
		storagePath := req.Slug + "/" + req.Version + "/game.zip"

		input := repository.CreateGameInput{
			GameIntegrationID: req.GameIntegrationID,
			Slug:              req.Slug,
			Name:              req.Name,
			Description:       req.Description,
			LogoURL:           req.LogoURL,
			StoragePath:       storagePath,
			Version:           req.Version,
		}

		game, err := deps.Repo.Create(c.Request.Context(), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create game", Details: err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gameToResponse(*game))
	})

	// Update a game
	group.PUT("/games/:gameId", func(c *gin.Context) {
		gameID, err := uuid.Parse(c.Param("gameId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid game ID"})
			return
		}

		var req UpdateGameRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Details: err.Error()})
			return
		}

		input := repository.UpdateGameInput{
			ID:                gameID,
			GameIntegrationID: req.GameIntegrationID,
			Slug:              req.Slug,
			Name:              req.Name,
			Description:       req.Description,
			LogoURL:           req.LogoURL,
			Version:           req.Version,
			IsActive:          req.IsActive,
		}

		if err := deps.Repo.Update(c.Request.Context(), input); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update game", Details: err.Error()})
			return
		}

		// Fetch updated game
		game, err := deps.Repo.GetByID(c.Request.Context(), gameID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch updated game", Details: err.Error()})
			return
		}

		c.JSON(http.StatusOK, gameToResponse(*game))
	})

	// Upload game bundle
	group.PATCH("/games/:gameId/bundle", func(c *gin.Context) {
		gameID, err := uuid.Parse(c.Param("gameId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid game ID"})
			return
		}

		// Get the game to get storage path
		game, err := deps.Repo.GetByID(c.Request.Context(), gameID)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "game not found"})
			return
		}

		// Get uploaded file
		file, err := c.FormFile("bundle")
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bundle file required"})
			return
		}

		// Read file content
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to open uploaded file"})
			return
		}
		defer src.Close()

		data := make([]byte, file.Size)
		if _, err := src.Read(data); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read uploaded file"})
			return
		}

		// Upload to storage
		if err := deps.Storage.UploadGame(c.Request.Context(), game.StoragePath, data); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to upload bundle", Details: err.Error()})
			return
		}

		// Invalidate cache
		if err := deps.Storage.InvalidateCache(c.Request.Context(), gameID.String()); err != nil {
			// Log but don't fail
			c.Error(err)
		}

		c.JSON(http.StatusOK, gin.H{"message": "bundle uploaded successfully"})
	})

	// Delete (deactivate) a game
	group.DELETE("/games/:gameId", func(c *gin.Context) {
		gameID, err := uuid.Parse(c.Param("gameId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid game ID"})
			return
		}

		if err := deps.Repo.Delete(c.Request.Context(), gameID); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete game", Details: err.Error()})
			return
		}

		// Invalidate cache
		if deps.Storage != nil {
			_ = deps.Storage.InvalidateCache(c.Request.Context(), gameID.String())
		}

		c.JSON(http.StatusOK, gin.H{"message": "game deleted successfully"})
	})
}

