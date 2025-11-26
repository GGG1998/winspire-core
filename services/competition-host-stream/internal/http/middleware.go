package httpx

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs basic request information using slog so App Runner logs stay structured.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger.InfoContext(c.Request.Context(), "http_request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

// ErrorResponder ensures all errors bubble up in a consistent JSON envelope.
func ErrorResponder() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		status := c.Writer.Status()
		if status < 400 {
			status = http.StatusInternalServerError
		}

		c.JSON(status, gin.H{
			"error":   err.Error(),
			"status":  status,
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
			"traceId": c.Request.Header.Get("X-Amzn-Trace-Id"),
		})
	}
}

// SecurityHeaders applies a small set of defaults suitable for APIs.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	}
}


