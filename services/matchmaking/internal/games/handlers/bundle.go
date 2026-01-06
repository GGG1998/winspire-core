package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/winspire-core/services/matchmaking/internal/games"
)

// BundleDeps contains dependencies for bundle handlers.
type BundleDeps struct {
	Repo    *games.GameRepository
	Storage *games.S3Client
}

// AllowIframeMiddleware allows game bundles to be embedded in iframes.
// Game bundles need to be embedded in iframes for the match lobby.
// Security Note: This removes X-Frame-Options protection for bundle routes only.
// Since games are public content meant to be embedded, this is acceptable.
func AllowIframeMiddleware() gin.HandlerFunc {
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
	bundleGroup := group.Group("")
	bundleGroup.Use(AllowIframeMiddleware())

	bundleGroup.GET("/g/:slug/bundle/*filepath", func(c *gin.Context) {
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

		// Stream file from S3
		stream, err := deps.Storage.GetBundleFileStream(c.Request.Context(), game.StoragePath, filePath)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "file not found", Details: err.Error()})
			return
		}
		defer stream.Body.Close()

		// Disable write deadline for streaming large files (Go 1.20+)
		// The server's WriteTimeout would otherwise kill the connection during large file transfers
		rc := http.NewResponseController(c.Writer)
		if err := rc.SetWriteDeadline(time.Time{}); err != nil {
			// Log but continue - some response writers may not support this
			_ = err
		}

		// Set cache headers for better performance
		c.Header("Cache-Control", "public, max-age=3600")

		// Set CORS headers to allow cross-origin access for game assets
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")

		// Set content type
		c.Header("Content-Type", stream.ContentType)

		// Set content length for proper streaming
		if stream.ContentLength > 0 {
			c.Header("Content-Length", strconv.FormatInt(stream.ContentLength, 10))
		}

		// Write status before streaming
		c.Status(http.StatusOK)

		// Stream with explicit flushing for better HTTP/2 compatibility
		// Use 32KB buffer to balance memory usage and throughput
		buf := make([]byte, 32*1024)
		flusher, canFlush := c.Writer.(http.Flusher)

		for {
			n, readErr := stream.Body.Read(buf)
			if n > 0 {
				if _, writeErr := c.Writer.Write(buf[:n]); writeErr != nil {
					return
				}
				// Flush periodically to prevent HTTP/2 stream stalls
				if canFlush {
					flusher.Flush()
				}
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return
			}
		}
	})
}
