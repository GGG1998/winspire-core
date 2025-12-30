package httpx

import (
	"context"
	"crypto/subtle"
	"log/slog"

	"github.com/gin-gonic/gin"
	authmw "github.com/winspire/winspire-core/libs/go/auth/middleware"
	"github.com/winspire/winspire-core/libs/go/auth/types"
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
	Storage     *storage.S3Client
}

// NewRouter builds a fully-configured Gin engine with middleware, auth, and handlers.
func NewRouter(deps ServerDeps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Use shared httpx middleware
	httpCfg := sharedhttp.DefaultConfig()
	httpCfg.ServiceName = "game-management"
	httpCfg.AllowHeaders = append(httpCfg.AllowHeaders, "X-Internal-Service-Key")
	httpCfg.AllowOrigins = deps.Config.CORSAllowedOrigins
	httpCfg.AllowCredentials = true

	router.Use(
		sharedhttp.Recovery(deps.Logger),
		sharedhttp.CORS(httpCfg),
		sharedhttp.SecurityHeaders(httpCfg),
		sharedhttp.RequestLogger(deps.Logger),
		sharedhttp.ErrorResponder(),
	)

	// Health check endpoint (no auth required)
	router.GET("/healthz", sharedhttp.HealthCheck(deps.HealthCheck))

	// Public API routes (no auth required - games need to be publicly accessible)
	publicAPI := router.Group("/v1")
	
	// Public bundle serving routes
	handlers.RegisterBundleRoutes(publicAPI, handlers.BundleDeps{
		Repo:    deps.GameRepo,
		Storage: deps.Storage,
	})
	
	// Public game routes (no auth required)
	handlers.RegisterGameRoutes(publicAPI, handlers.GameDeps{
		Repo: deps.GameRepo,
	})

	// API routes with auth
	api := router.Group("/v1")
	if deps.Config.HasAuth() {
		api.Use(authOrInternalMiddleware(authmw.Config{
			JWTSecret: deps.Config.HostJWTSecret,
			Issuer:    deps.Config.HostJWTIssuer,
			Audience:  deps.Config.HostJWTAudience,
		}, deps.Config.InternalServiceAPIKey))
	}

	// Admin routes (require admin role)
	adminGroup := api.Group("/admin")
	adminGroup.Use(sharedhttp.RequireAdminRole())
	handlers.RegisterAdminRoutes(adminGroup, handlers.AdminDeps{
		Repo:    deps.GameRepo,
		Storage: deps.Storage,
	})

	return router
}

func authOrInternalMiddleware(cfg authmw.Config, internalKey string) gin.HandlerFunc {
	validateJWT := authmw.ValidateJWTMiddleware(cfg)

	return func(c *gin.Context) {
		if internalKey != "" {
			provided := c.GetHeader("X-Internal-Service-Key")
			if provided != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(internalKey)) == 1 {
				user := &types.UserContext{
					ID:       types.UserID("internal-service"),
					Nickname: "internal-service",
					UserType: "service",
					Roles:    []types.Role{"admin"},
				}
				c.Request = c.Request.WithContext(types.WithUserContext(c.Request.Context(), user))
				c.Set("user", user)
				c.Next()
				return
			}
		}

		validateJWT(c)
	}
}
