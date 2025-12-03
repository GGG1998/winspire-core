package httpx

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds common security headers to responses
func SecurityHeaders(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic security headers (always applied)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		// HSTS for production
		if cfg.EnableHSTS {
			c.Header("Strict-Transport-Security", "max-age="+strconv.Itoa(cfg.HSTSMaxAge)+"; includeSubDomains")
		}

		// Content Security Policy
		if cfg.ContentSecurityPolicy != "" {
			c.Header("Content-Security-Policy", cfg.ContentSecurityPolicy)
		}

		c.Next()
	}
}

// RequestLogger logs incoming requests with timing information and trace IDs
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Extract trace ID from ALB or generate one
		traceID := c.Request.Header.Get("X-Amzn-Trace-Id")
		if traceID == "" {
			traceID = c.Request.Header.Get("X-Request-ID")
		}

		// Store trace ID for later use
		if traceID != "" {
			c.Set("trace_id", traceID)
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		// Build structured log
		attrs := []any{
			"method", method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", clientIP,
		}

		if traceID != "" {
			attrs = append(attrs, "trace_id", traceID)
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, "error", c.Errors.String())
		}

		logger.InfoContext(c.Request.Context(), "http_request", attrs...)
	}
}

// ErrorResponder converts Gin errors into standardized JSON error responses
func ErrorResponder() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		// Get the last error
		err := c.Errors.Last()
		status := c.Writer.Status()

		// Default to 500 if status is not an error status
		if status < 400 {
			status = http.StatusInternalServerError
		}

		// Get trace ID if available
		traceID, _ := c.Get("trace_id")

		// Build error response
		response := gin.H{
			"error":  err.Error(),
			"status": status,
			"path":   c.Request.URL.Path,
			"method": c.Request.Method,
		}

		if traceID != nil {
			response["trace_id"] = traceID
		}

		c.JSON(status, response)
	}
}

// CORS adds Cross-Origin Resource Sharing headers based on configuration
func CORS(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if len(cfg.AllowOrigins) > 0 {
			allowed := false
			for _, allowedOrigin := range cfg.AllowOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if !allowed {
				c.Next()
				return
			}

			// Set the actual origin instead of "*" if credentials are enabled
			if cfg.AllowCredentials && origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			} else if len(cfg.AllowOrigins) == 1 && cfg.AllowOrigins[0] == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		// Set other CORS headers
		if len(cfg.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
		}

		if len(cfg.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
		}

		if len(cfg.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
		}

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if cfg.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Recovery is a middleware that recovers from panics and logs them
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
