package http

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthChecker provides health check interface
type HealthChecker interface {
	Health(ctx context.Context) error
}

// HealthHandler handles health check requests
type HealthHandler struct {
	dbHealth    HealthChecker
	redisHealth HealthChecker
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(dbHealth, redisHealth HealthChecker) *HealthHandler {
	return &HealthHandler{
		dbHealth:    dbHealth,
		redisHealth: redisHealth,
	}
}

// HandleHealth returns service health status
// GET /health
func (h *HealthHandler) HandleHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	status := gin.H{
		"status":  "healthy",
		"service": "matchmaking",
		"checks":  gin.H{},
	}

	// Check database
	if h.dbHealth != nil {
		if err := h.dbHealth.Health(ctx); err != nil {
			status["checks"].(gin.H)["database"] = gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			status["status"] = "unhealthy"
		} else {
			status["checks"].(gin.H)["database"] = gin.H{
				"status": "healthy",
			}
		}
	}

	// Check Redis
	if h.redisHealth != nil {
		if err := h.redisHealth.Health(ctx); err != nil {
			status["checks"].(gin.H)["redis"] = gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			status["status"] = "unhealthy"
		} else {
			status["checks"].(gin.H)["redis"] = gin.H{
				"status": "healthy",
			}
		}
	}

	statusCode := 200
	if status["status"] == "unhealthy" {
		statusCode = 503
	}

	c.JSON(statusCode, status)
}

// HandleReadiness returns readiness status (for k8s probes)
// GET /ready
func (h *HealthHandler) HandleReadiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
	defer cancel()

	// Check critical dependencies
	if h.dbHealth != nil {
		if err := h.dbHealth.Health(ctx); err != nil {
			c.JSON(503, gin.H{
				"ready": false,
				"error": "database not ready",
			})
			return
		}
	}

	c.JSON(200, gin.H{
		"ready": true,
	})
}

// HandleLiveness returns liveness status (for k8s probes)
// GET /live
func (h *HealthHandler) HandleLiveness(c *gin.Context) {
	c.JSON(200, gin.H{
		"alive": true,
	})
}










