package httpx

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	authmw "github.com/winspire/winspire-core/libs/go/auth/middleware"
	sharedhttp "github.com/winspire/winspire-core/libs/go/httpx"

	"github.com/winspire/competition-host-stream/internal/config"
	"github.com/winspire/competition-host-stream/internal/http/handlers"
	"github.com/winspire/competition-host-stream/internal/projections"
	ssebroker "github.com/winspire/competition-host-stream/internal/sse"
)

// HealthFunc defines the signature for dependency health checks (Postgres + Redis).
type HealthFunc func(ctx context.Context) error

// ServerDeps aggregates the dependencies required to construct the HTTP router.
type ServerDeps struct {
	Config        config.Config
	Logger        *slog.Logger
	HealthCheck   HealthFunc
	Reader        *projections.Reader
	CupProjector  *projections.CupProjector
	TourProjector *projections.TournamentProjector
	Attendance    *projections.AttendanceProjector
	EventRouter   *ssebroker.EventRouter
	Broker        *ssebroker.Broker
	Registry      *ssebroker.Registry
}

// NewRouter builds a fully-configured Gin engine with middleware, auth, and handlers.
func NewRouter(deps ServerDeps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Use shared httpx middleware
	httpCfg := sharedhttp.DefaultConfig()
	httpCfg.ServiceName = "competition-host-stream"

	router.Use(
		sharedhttp.Recovery(deps.Logger),
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

	handlers.RegisterCupTournamentRoutes(api, handlers.CupTournamentDeps{
		Reader:              deps.Reader,
		CupProjector:        deps.CupProjector,
		TournamentProjector: deps.TourProjector,
		EventRouter:         deps.EventRouter,
	})
	handlers.RegisterMatchRoutes(api, handlers.MatchDeps{
		Reader: deps.Reader,
	})
	handlers.RegisterStreamRoute(api, handlers.StreamDeps{
		Broker:   deps.Broker,
		Registry: deps.Registry,
	})

	return router
}
