package httpx

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsCollector is an interface for collecting HTTP metrics
type MetricsCollector interface {
	RecordRequest(method, path string, status int, duration time.Duration)
	RecordError(method, path string, errorType string)
}

// Metrics middleware collects HTTP metrics if a collector is provided
func Metrics(collector MetricsCollector) gin.HandlerFunc {
	if collector == nil {
		// No-op if no collector is provided
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		// Record request metrics
		collector.RecordRequest(method, path, status, duration)

		// Record error metrics if there are errors
		if len(c.Errors) > 0 {
			lastErr := c.Errors.Last()
			errorType := "unknown"
			if lastErr != nil {
				// Convert ErrorType to string representation
				switch lastErr.Type {
				case gin.ErrorTypeBind:
					errorType = "bind"
				case gin.ErrorTypeRender:
					errorType = "render"
				case gin.ErrorTypePrivate:
					errorType = "private"
				case gin.ErrorTypePublic:
					errorType = "public"
				case gin.ErrorTypeAny:
					errorType = "any"
				default:
					errorType = "unknown"
				}
			}
			collector.RecordError(method, path, errorType)
		}
	}
}

// HealthCheck returns a simple health check handler
func HealthCheck(healthCheckFn func(context.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		if healthCheckFn == nil {
			c.JSON(200, gin.H{"status": "ok"})
			return
		}

		if err := healthCheckFn(c.Request.Context()); err != nil {
			c.JSON(503, gin.H{
				"status": "degraded",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{"status": "ok"})
	}
}

// StructuredLogger creates a new slog logger with common fields
func StructuredLogger(serviceName string) *slog.Logger {
	return slog.Default().With(
		"service", serviceName,
	)
}

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, "trace_id", traceID)
}

// GetTraceID extracts the trace ID from the context
func GetTraceID(ctx context.Context) string {
	traceID, _ := ctx.Value("trace_id").(string)
	return traceID
}
