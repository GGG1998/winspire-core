package handlers

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/winspire/game-management/internal/repository"
	"github.com/winspire/game-management/internal/storage"
)

// AdminDeps contains dependencies for admin handlers.
type AdminDeps struct {
	Repo    *repository.GameRepository
	Storage *storage.S3Client
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

		// Generate normalized storage path based on slug/version
		storagePath := normalizeStoragePath(path.Join(req.Slug, req.Version))

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

	// Upload assets for an existing game
	group.POST("/games/:gameId/files", func(c *gin.Context) {
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

		if deps.Storage == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "storage client not configured"})
			return
		}

		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid multipart form", Details: err.Error()})
			return
		}

		files := form.File["files"]
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "at least one file is required"})
			return
		}

		basePath := normalizeStoragePath(game.StoragePath)
		uploaded, err := deps.Storage.UploadMultipleFiles(c.Request.Context(), basePath, files)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to upload files", Details: err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":         "files uploaded successfully",
			"uploadedCount":   len(uploaded),
			"uploadedPaths":   uploaded,
			"storageBasePath": basePath,
		})
	})

	// Upload assets for a new game before creation
	group.POST("/games/uploads", func(c *gin.Context) {
		if deps.Storage == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "storage client not configured"})
			return
		}

		slug := strings.TrimSpace(c.PostForm("slug"))
		version := strings.TrimSpace(c.PostForm("version"))
		// filePath contains the relative path (e.g., "Build/game.js") since browsers strip paths from filenames
		filePath := strings.TrimSpace(c.PostForm("filePath"))

		if slug == "" || version == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "slug and version are required"})
			return
		}

		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid multipart form", Details: err.Error()})
			return
		}

		files := form.File["files"]
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "at least one file is required"})
			return
		}

		basePath := normalizeStoragePath(path.Join(slug, version))

		// If filePath is provided (single file upload with path), use it
		// Otherwise fall back to UploadMultipleFiles for batch uploads
		var uploaded []string
		if filePath != "" && len(files) == 1 {
			uploaded, err = deps.Storage.UploadMultipleFilesWithPath(c.Request.Context(), basePath, files, filePath)
		} else {
			uploaded, err = deps.Storage.UploadMultipleFiles(c.Request.Context(), basePath, files)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to upload files", Details: err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":         fmt.Sprintf("uploaded %d file(s)", len(uploaded)),
			"uploadedCount":   len(uploaded),
			"uploadedPaths":   uploaded,
			"storageBasePath": basePath,
		})
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
		// TODO: implement cache invalidation
		// if deps.Storage != nil {
		// 	_ = deps.Storage.InvalidateCache(c.Request.Context(), gameID.String())
		// }

		c.JSON(http.StatusOK, gin.H{"message": "game deleted successfully"})
	})
}

func normalizeStoragePath(p string) string {
	clean := strings.Trim(p, "/")
	if clean == "" {
		return clean
	}

	if strings.HasSuffix(strings.ToLower(clean), ".zip") {
		dir := path.Dir(clean)
		if dir == "." {
			return ""
		}
		return dir
	}
	return clean
}
