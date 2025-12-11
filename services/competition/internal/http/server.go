package httpx

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	authmw "github.com/winspire/winspire-core/libs/go/auth/middleware"
	sharedhttp "github.com/winspire/winspire-core/libs/go/httpx"

	"github.com/winspire/competition/internal/application"
	"github.com/winspire/competition/internal/config"
	"github.com/winspire/competition/internal/domain"
	"github.com/winspire/competition/internal/http/handlers"
	"github.com/winspire/competition/internal/pubsub"
	"github.com/winspire/competition/internal/repository"
)

// HealthFunc defines the signature for dependency health checks.
type HealthFunc func(ctx context.Context) error

// ServerDeps aggregates the dependencies required to construct the HTTP router.
type ServerDeps struct {
	Config                 config.Config
	Logger                 *slog.Logger
	HealthCheck            HealthFunc
	Pool                   *pgxpool.Pool
	Redis                  *redis.Client
	EventPublisher         *pubsub.EventPublisher
	ConfirmParticipationUC *application.ConfirmParticipationUseCase
	JoinTournamentUC       *application.JoinTournamentUseCase
	ParticipantRepo        domain.ParticipantRepository
}

// NewRouter builds a fully-configured Gin engine with middleware, auth, and handlers.
func NewRouter(deps ServerDeps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Use shared httpx middleware
	httpCfg := sharedhttp.DefaultConfig()
	httpCfg.ServiceName = "competition"

	router.Use(
		sharedhttp.Recovery(deps.Logger),
		sharedhttp.CORS(httpCfg),
		sharedhttp.SecurityHeaders(httpCfg),
		sharedhttp.RequestLogger(deps.Logger),
		sharedhttp.ErrorResponder(),
	)

	// Health check endpoint (no auth required)
	router.GET("/healthz", sharedhttp.HealthCheck(deps.HealthCheck))

	// Readiness probe
	router.GET("/readyz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Initialize persistence repositories
	hostRepo := repository.NewHostRepository(deps.Pool)
	tournamentPersistenceRepo := repository.NewTournamentRepository(deps.Pool)
	registrationPersistenceRepo := repository.NewRegistrationRepository(deps.Pool)

	// Register internal routes (no auth - service-to-service communication)
	handlers.RegisterInternalRoutes(router, handlers.InternalDeps{
		TournamentRepo:   tournamentPersistenceRepo,
		RegistrationRepo: registrationPersistenceRepo,
		Logger:           deps.Logger,
	})

	// API routes with auth
	api := router.Group("/v1")
	if deps.Config.HasAuth() {
		api.Use(authmw.ValidateJWTMiddleware(authmw.Config{
			JWTSecret: deps.Config.HostJWTSecret,
			Issuer:    deps.Config.HostJWTIssuer,
			Audience:  deps.Config.HostJWTAudience,
		}))
	}

	// Register tournament user routes (player-facing endpoints)
	// Routes: /v1/:hostId/tournaments/:tournamentId/join, leave, confirm-participation
	//         /v1/:hostId/matches/:matchId/confirm-participation
	handlers.RegisterTournamentUserRoutes(api, handlers.TournamentUserDeps{
		Pool:                   deps.Pool,
		Logger:                 deps.Logger,
		ConfirmParticipationUC: deps.ConfirmParticipationUC,
		JoinTournamentUC:       deps.JoinTournamentUC,
	})

	// Register tournament host routes (host/admin endpoints)
	// Routes: /v1/:hostId/tournaments (create, edit, start, cancel)
	// These routes require host admin authorization
	handlers.RegisterTournamentHostRoutes(api, handlers.TournamentHostDeps{
		HostRepo:           hostRepo,
		TournamentRepo:     tournamentPersistenceRepo,
		RegistrationRepo:   registrationPersistenceRepo,
		ParticipantRepo:    deps.ParticipantRepo,
		EventPublisher:     deps.EventPublisher,
		Logger:             deps.Logger,
		MatchmakingBaseURL: deps.Config.MatchmakingBaseURL,
	})

	return router
}
