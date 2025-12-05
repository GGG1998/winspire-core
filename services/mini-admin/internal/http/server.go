package httpx

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	authmw "github.com/winspire/winspire-core/libs/go/auth/middleware"
	sharedhttp "github.com/winspire/winspire-core/libs/go/httpx"

	"github.com/winspire/mini-admin/internal/config"
	"github.com/winspire/mini-admin/internal/http/handlers"
	"github.com/winspire/mini-admin/internal/repository"
	"github.com/winspire/mini-admin/internal/storage"
)

// HealthFunc defines the signature for dependency health checks.
type HealthFunc func(ctx context.Context) error

// ServerDeps aggregates the dependencies required to construct the HTTP router.
type ServerDeps struct {
	Config      config.Config
	Logger      *slog.Logger
	HealthCheck HealthFunc
	GameRepo    *repository.GameRepository
	S3Client    *storage.S3Client
}

// NewRouter builds a fully-configured Gin engine with middleware, auth, and handlers.
func NewRouter(deps ServerDeps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Use shared httpx middleware
	httpCfg := sharedhttp.DefaultConfig()
	httpCfg.ServiceName = "mini-admin"

	router.Use(
		sharedhttp.Recovery(deps.Logger),
		sharedhttp.CORS(httpCfg),
		sharedhttp.SecurityHeaders(httpCfg),
		sharedhttp.RequestLogger(deps.Logger),
		sharedhttp.ErrorResponder(),
	)

	// Health check endpoint (no auth required)
	router.GET("/healthz", sharedhttp.HealthCheck(deps.HealthCheck))

	// API routes
	api := router.Group("/v1")

	// Optional authentication
	if deps.Config.HasAuth() {
		api.Use(authmw.ValidateJWTMiddleware(authmw.Config{
			JWTSecret: deps.Config.AdminJWTSecret,
			Issuer:    deps.Config.AdminJWTIssuer,
			Audience:  deps.Config.AdminJWTAudience,
		}))
	}

	// Register game handlers
	handlers.RegisterGameRoutes(api, handlers.GameDeps{
		Repo:     deps.GameRepo,
		S3Client: deps.S3Client,
	})

	return router
}

