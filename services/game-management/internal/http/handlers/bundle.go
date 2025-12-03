package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/winspire/game-management/internal/repository"
	"github.com/winspire/game-management/internal/storage"
)

// BundleDeps contains dependencies for bundle handlers.
type BundleDeps struct {
	Repo    *repository.GameRepository
	Storage *storage.Client
}

// RegisterBundleRoutes registers routes for serving game bundles.
func RegisterBundleRoutes(group *gin.RouterGroup, deps BundleDeps) {
	// Serve files from game bundle
	group.GET("/games/:gameId/bundle/*filepath", func(c *gin.Context) {
		gameID, err := uuid.Parse(c.Param("gameId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid game ID"})
			return
		}

		filePath := c.Param("filepath")
		if filePath == "" || filePath == "/" {
			filePath = "/index.html"
		}

		// Remove leading slash for internal processing
		filePath = strings.TrimPrefix(filePath, "/")

		// Get game from repository
		game, err := deps.Repo.GetByID(c.Request.Context(), gameID)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "game not found"})
			return
		}

		if !game.IsActive {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "game not found"})
			return
		}

		// Get file from bundle
		bundleFile, err := deps.Storage.GetBundleFile(c.Request.Context(), gameID.String(), game.StoragePath, filePath)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "file not found", Details: err.Error()})
			return
		}

		// Set cache headers
		c.Header("Cache-Control", "public, max-age=3600")
		c.Data(http.StatusOK, bundleFile.ContentType, bundleFile.Content)
	})

	// Alternative: serve bundle by slug
	group.GET("/g/:slug/bundle/*filepath", func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "slug is required"})
			return
		}

		filePath := c.Param("filepath")
		if filePath == "" || filePath == "/" {
			filePath = "/index.html"
		}

		// Remove leading slash for internal processing
		filePath = strings.TrimPrefix(filePath, "/")

		// Get game from repository
		game, err := deps.Repo.GetBySlug(c.Request.Context(), slug)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "game not found"})
			return
		}

		// Get file from bundle
		bundleFile, err := deps.Storage.GetBundleFile(c.Request.Context(), game.ID.String(), game.StoragePath, filePath)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "file not found", Details: err.Error()})
			return
		}

		// Set cache headers
		c.Header("Cache-Control", "public, max-age=3600")
		c.Data(http.StatusOK, bundleFile.ContentType, bundleFile.Content)
	})
}

