package httpx

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	authmw "github.com/winspire/winspire-core/libs/go/auth/middleware"
	sharedhttp "github.com/winspire/winspire-core/libs/go/httpx"

	"github.com/winspire/game-management/internal/config"
	"github.com/winspire/game-management/internal/http/handlers"
	"github.com/winspire/game-management/internal/repository"
	"github.com/winspire/game-management/internal/storage"
)

// HealthFunc defines the signature for dependency health checks.
type HealthFunc func(ctx context.Context) error

// ServerDeps aggregates the dependencies required to construct the HTTP router.
type ServerDeps struct {
	Config      config.Config
	Logger      *slog.Logger
	HealthCheck HealthFunc
	GameRepo    *repository.GameRepository
	Storage     *storage.Client
}

// NewRouter builds a fully-configured Gin engine with middleware, auth, and handlers.
func NewRouter(deps ServerDeps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Use shared httpx middleware
	httpCfg := sharedhttp.DefaultConfig()
	httpCfg.ServiceName = "game-management"

	router.Use(
		sharedhttp.Recovery(deps.Logger),
		sharedhttp.CORS(httpCfg),
		sharedhttp.SecurityHeaders(httpCfg),
		sharedhttp.RequestLogger(deps.Logger),
		sharedhttp.ErrorResponder(),
	)

	// Health check endpoint (no auth required)
	router.GET("/healthz", sharedhttp.HealthCheck(deps.HealthCheck))

	// API routes with auth
	api := router.Group("/v1")
	if deps.Config.HasAuth() {
		api.Use(authmw.ValidateJWTMiddleware(authmw.Config{
			JWTSecret: deps.Config.HostJWTSecret,
			Issuer:    deps.Config.HostJWTIssuer,
			Audience:  deps.Config.HostJWTAudience,
		}))
	}

	// Public game routes
	handlers.RegisterGameRoutes(api, handlers.GameDeps{
		Repo: deps.GameRepo,
	})

	// Bundle serving routes
	handlers.RegisterBundleRoutes(api, handlers.BundleDeps{
		Repo:    deps.GameRepo,
		Storage: deps.Storage,
	})

	// Admin routes (require admin role)
	adminGroup := api.Group("/admin")
	adminGroup.Use(sharedhttp.RequireAdminRole())
	handlers.RegisterAdminRoutes(adminGroup, handlers.AdminDeps{
		Repo:    deps.GameRepo,
		Storage: deps.Storage,
	})

	return router
}
