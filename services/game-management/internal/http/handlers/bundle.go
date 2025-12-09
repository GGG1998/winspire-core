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

// allowIframeMiddleware allows game bundles to be embedded in iframes.
// Game bundles need to be embedded in iframes for the match lobby.
// Security Note: This removes X-Frame-Options protection for bundle routes only.
// Since games are public content meant to be embedded, this is acceptable.
func allowIframeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Don't set X-Frame-Options (allowing iframe embedding from any origin)
		// The SecurityHeaders middleware runs first and sets it to DENY,
		// but we can't easily remove headers, so we'll override with CSP

		// Use Content-Security-Policy frame-ancestors directive (modern standard)
		// This allows embedding from any origin
		c.Header("Content-Security-Policy", "frame-ancestors *")

		c.Next()
	}
}

// RegisterBundleRoutes registers routes for serving game bundles.
func RegisterBundleRoutes(group *gin.RouterGroup, deps BundleDeps) {
	// Apply middleware to allow iframe embedding for all bundle routes
	group.Use(allowIframeMiddleware())

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

		// Set cache headers for better performance
		c.Header("Cache-Control", "public, max-age=3600")

		// Set CORS headers to allow cross-origin access for game assets
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")

		c.Data(http.StatusOK, bundleFile.ContentType, bundleFile.Content)
	})
}
