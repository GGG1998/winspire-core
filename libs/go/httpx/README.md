# httpx - Shared HTTP Middleware Library

This package provides common HTTP middleware and utilities for Winspire microservices using the Gin framework.

## Features

- **Security Headers**: CSP, HSTS, X-Frame-Options, etc.
- **Request Logging**: Structured logging with trace IDs from ALB
- **Error Handling**: Standardized JSON error responses
- **CORS**: Configurable cross-origin resource sharing
- **Authentication**: JWT validation helpers and role-based access control
- **Observability**: Metrics collection and health checks

## Usage

### Basic Setup

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/winspire/winspire-core/libs/go/httpx"
    authmw "github.com/winspire/winspire-core/libs/go/auth/middleware"
)

func main() {
    // Create config
    cfg := httpx.DefaultConfig()
    cfg.ServiceName = "my-service"
    
    // Create logger
    logger := httpx.StructuredLogger(cfg.ServiceName)
    
    // Create router
    router := gin.New()
    
    // Add middleware
    router.Use(
        httpx.Recovery(logger),
        httpx.CORS(cfg),
        httpx.SecurityHeaders(cfg),
        httpx.RequestLogger(logger),
        httpx.ErrorResponder(),
    )
    
    // Health check
    router.GET("/healthz", httpx.HealthCheck(nil))
    
    // Protected routes
    api := router.Group("/v1")
    api.Use(authmw.ValidateJWTMiddleware(authmw.Config{
        JWTSecret: cfg.JWTSecret,
        Issuer:    cfg.JWTIssuer,
        Audience:  cfg.JWTAudience,
    }))
    
    // Admin routes
    admin := api.Group("/admin")
    admin.Use(httpx.RequireAdminRole())
    
    router.Run(":8080")
}
```

### Authentication Helpers

```go
// Require authentication
router.GET("/me", httpx.RequireAuth(), func(c *gin.Context) {
    user := httpx.MustGetUser(c)
    c.JSON(200, gin.H{"user_id": user.ID})
})

// Require specific role
router.GET("/admin/users", httpx.RequireRole("admin"), func(c *gin.Context) {
    // Admin only
})

// Require any of multiple roles
router.GET("/moderator", httpx.RequireAnyRole("admin", "moderator"), func(c *gin.Context) {
    // Admin or moderator
})
```

### Configuration

**Development**:
```go
cfg := httpx.DefaultConfig()
cfg.AllowOrigins = []string{"*"}
cfg.EnableHSTS = false
```

**Production**:
```go
cfg := httpx.ProductionConfig()
cfg.AllowOrigins = []string{"https://winspire.app"}
cfg.EnableHSTS = true
cfg.JWTSecret = os.Getenv("JWT_SECRET")
```

## Migration from Service-Local Middleware

1. Replace imports:
   ```go
   - import "service/internal/http"
   + import "github.com/winspire/winspire-core/libs/go/httpx"
   ```

2. Update middleware calls:
   ```go
   - router.Use(httpx.SecurityHeaders())
   + router.Use(httpx.SecurityHeaders(cfg))
   ```

3. Remove local middleware files

## AWS Integration

### Trace IDs from ALB

The `RequestLogger` middleware automatically extracts trace IDs from:
- `X-Amzn-Trace-Id` (AWS ALB)
- `X-Request-ID` (fallback)

These IDs are included in all log entries and error responses.

### CloudWatch Structured Logging

All logs use `slog` for structured logging, compatible with CloudWatch Logs Insights:

```sql
fields @timestamp, method, path, status, latency_ms, trace_id
| filter status >= 400
| sort latency_ms desc
```

## Testing

```go
import "testing"

func TestSecurityHeaders(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    cfg := httpx.DefaultConfig()
    middleware := httpx.SecurityHeaders(cfg)
    
    middleware(c)
    
    assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
}
```


