package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/winspire/game-management/internal/repository"
	"github.com/winspire/game-management/internal/storage"
)

// BundleDeps contains dependencies for bundle handlers.
type BundleDeps struct {
	Repo    *repository.GameRepository
	Storage *storage.S3Client
}

// RegisterBundleRoutes registers routes for serving game bundles.
func RegisterBundleRoutes(group *gin.RouterGroup, deps BundleDeps) {

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

		if deps.Storage == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "storage client not configured"})
			return
		}

		bundleFile, err := deps.Storage.GetBundleFile(c.Request.Context(), game.StoragePath, filePath)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "file not found", Details: err.Error()})
			return
		}

		// Set cache headers
		c.Header("Cache-Control", "public, max-age=3600")
		c.Data(http.StatusOK, bundleFile.ContentType, bundleFile.Content)
	})
}
