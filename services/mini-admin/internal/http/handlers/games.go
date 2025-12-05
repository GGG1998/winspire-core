package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/winspire/mini-admin/internal/repository"
	"github.com/winspire/mini-admin/internal/storage"
)

// GameDeps contains dependencies for game handlers.
type GameDeps struct {
	Repo     *repository.GameRepository
	S3Client *storage.S3Client
}

// RegisterGameRoutes registers game-related routes.
func RegisterGameRoutes(group *gin.RouterGroup, deps GameDeps) {
	// List all games
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

	// Get a single game
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

		c.JSON(http.StatusOK, gameToResponse(*game))
	})

	// Create a new game
	group.POST("/games", func(c *gin.Context) {
		var req CreateGameRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Details: err.Error()})
			return
		}

		// Generate S3 path based on slug and version
		s3Path := req.Slug + "/" + req.Version + "/"

		input := repository.CreateGameInput{
			Slug:              req.Slug,
			Name:              req.Name,
			Description:       req.Description,
			LogoURL:           req.LogoURL,
			S3Path:            s3Path,
			Version:           req.Version,
			VersioningEnabled: req.VersioningEnabled,
		}

		game, err := deps.Repo.Create(c.Request.Context(), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create game", Details: err.Error()})
			return
		}

		// If versioning is enabled, enable it on the bucket
		if req.VersioningEnabled {
			if err := deps.S3Client.EnableVersioning(c.Request.Context()); err != nil {
				// Log the error but don't fail the request
				c.Error(err)
			}
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
			Slug:              req.Slug,
			Name:              req.Name,
			Description:       req.Description,
			LogoURL:           req.LogoURL,
			Version:           req.Version,
			VersioningEnabled: req.VersioningEnabled,
			IsActive:          req.IsActive,
		}

		if err := deps.Repo.Update(c.Request.Context(), input); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update game", Details: err.Error()})
			return
		}

		// If versioning was enabled, enable it on the bucket
		if req.VersioningEnabled != nil && *req.VersioningEnabled {
			if err := deps.S3Client.EnableVersioning(c.Request.Context()); err != nil {
				c.Error(err)
			}
		}

		// Fetch updated game
		game, err := deps.Repo.GetByID(c.Request.Context(), gameID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch updated game", Details: err.Error()})
			return
		}

		c.JSON(http.StatusOK, gameToResponse(*game))
	})

	// Delete a game
	group.DELETE("/games/:gameId", func(c *gin.Context) {
		gameID, err := uuid.Parse(c.Param("gameId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid game ID"})
			return
		}

		// Get the game first to get S3 path
		game, err := deps.Repo.GetByID(c.Request.Context(), gameID)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "game not found"})
			return
		}

		// Soft delete in database
		if err := deps.Repo.Delete(c.Request.Context(), gameID); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete game", Details: err.Error()})
			return
		}

		// Optionally delete files from S3 (uncomment if you want to delete files)
		// if err := deps.S3Client.DeleteFolder(c.Request.Context(), game.S3Path); err != nil {
		//     c.Error(err)
		// }

		c.JSON(http.StatusOK, MessageResponse{Message: "game deleted successfully"})
	})

	// Upload game files
	group.POST("/games/:gameId/files", func(c *gin.Context) {
		gameID, err := uuid.Parse(c.Param("gameId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid game ID"})
			return
		}

		// Get the game to get S3 path
		game, err := deps.Repo.GetByID(c.Request.Context(), gameID)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "game not found"})
			return
		}

		// Parse multipart form
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse form", Details: err.Error()})
			return
		}

		files := form.File["files"]
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "no files provided"})
			return
		}

		// Upload files to S3
		uploadedPaths, err := deps.S3Client.UploadMultipleFiles(c.Request.Context(), game.S3Path, files)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to upload files", Details: err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "files uploaded successfully",
			"uploadedCount": len(uploadedPaths),
			"paths":         uploadedPaths,
		})
	})

	// Get S3 public URL for a game
	group.GET("/games/:gameId/url", func(c *gin.Context) {
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

		publicURL := deps.S3Client.GetPublicURL(game.S3Path)

		c.JSON(http.StatusOK, gin.H{
			"publicUrl": publicURL,
			"s3Path":    game.S3Path,
		})
	})
}

// gameToResponse converts a repository.Game to a GameResponse
func gameToResponse(g repository.Game) GameResponse {
	return GameResponse{
		ID:                g.ID.String(),
		Slug:              g.Slug,
		Name:              g.Name,
		Description:       g.Description,
		LogoURL:           g.LogoURL,
		S3Path:            g.S3Path,
		Version:           g.Version,
		VersioningEnabled: g.VersioningEnabled,
		IsActive:          g.IsActive,
		CreatedAt:         parseTime(g.CreatedAt),
		UpdatedAt:         parseTime(g.UpdatedAt),
	}
}

// parseTime parses ISO 8601 time string to time.Time
func parseTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05Z07:00", s)
	return t
}

