// Package main is the entrypoint for the matchmaking service
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/winspire-core/services/matchmaking/internal/application"
	"github.com/winspire-core/services/matchmaking/internal/config"
	"github.com/winspire-core/services/matchmaking/internal/games"
	gamehandlers "github.com/winspire-core/services/matchmaking/internal/games/handlers"
	httphandlers "github.com/winspire-core/services/matchmaking/internal/http"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/pubsub"
	"github.com/winspire-core/services/matchmaking/internal/repository"
	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
	tournamentapp "github.com/winspire-core/services/matchmaking/internal/tournament/application"
	tournamenthandlers "github.com/winspire-core/services/matchmaking/internal/tournament/handlers"
	tournamentrepo "github.com/winspire-core/services/matchmaking/internal/tournament/repository"
	"github.com/winspire-core/services/matchmaking/internal/websocket"
	authmiddleware "github.com/winspire/winspire-core/libs/go/auth/middleware"
	"github.com/winspire/winspire-core/libs/go/httpx"
)

func main() {
	// Load .env file in development (Docker Compose will override with its own env vars)
	// Try multiple locations to handle different working directories
	envPaths := []string{".env", "services/matchmaking/.env", "../../.env"}
	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err == nil {
				break
			}
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize structured logger for HTTP middleware
	slogLogger := httpx.StructuredLogger("matchmaking")

	// Initialize custom logger for application logging
	logger := observability.NewLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info("Starting matchmaking service", map[string]interface{}{
		"port": cfg.Port,
		"mode": cfg.GinMode,
	})

	// Initialize metrics
	metrics := observability.NewMetricsEmitter(cfg.CloudWatchNamespace)

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize database connection pool
	ctx := context.Background()
	dbConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to parse database URL: %v", err)
	}
	dbConfig.MaxConns = cfg.DatabaseMaxConns
	dbConfig.MinConns = cfg.DatabaseMaxIdle
	// Use simple protocol for Supabase PgBouncer compatibility (avoids prepared statement conflicts)
	dbConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Failed to create database pool: %v", err)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	logger.Info("Database connected", map[string]interface{}{
		"max_conns": cfg.DatabaseMaxConns,
	})

	// Initialize SQLC queries
	queries := sqlc.New(pool)

	// Initialize Redis client
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	if cfg.RedisPassword != "" {
		redisOpts.Password = cfg.RedisPassword
	}
	redisOpts.DB = cfg.RedisDB

	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	logger.Info("Redis connected", map[string]interface{}{
		"db": cfg.RedisDB,
	})

	// Initialize event publisher
	publisher := pubsub.NewEventPublisher(redisClient)

	// Initialize S3 client for game bundle storage (optional - may be nil if not configured)
	var s3Client *games.S3Client
	if cfg.AWSS3Bucket != "" {
		var s3Err error
		s3Client, s3Err = games.NewS3Client(
			cfg.AWSRegion,
			cfg.AWSS3Bucket,
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			cfg.AWSEndpoint, // For LocalStack
		)
		if s3Err != nil {
			logger.Warn("S3 client initialization failed - game bundle features disabled", map[string]interface{}{
				"error": s3Err.Error(),
			})
		} else {
			logger.Info("S3 client initialized", map[string]interface{}{
				"bucket": cfg.AWSS3Bucket,
				"region": cfg.AWSRegion,
			})
		}
	}

	// Initialize game repository
	gameRepo := games.NewGameRepository(queries, pool)

	// Initialize repositories
	bracketRepo := repository.NewBracketRepository(queries, pool, pool) // queries, db (DBTX), pool
	roundRepo := repository.NewRoundRepository(queries)
	matchRepo := repository.NewMatchRepository(queries)
	preLobbyRepo := repository.NewPreLobbyRepository(queries, pool, pool)

	// Initialize bracket service
	bracketService := application.NewBracketService(
		bracketRepo,
		roundRepo,
		matchRepo,
		publisher,
		metrics,
		logger,
	)

	// Initialize WebSocket hub with disconnect callback (T112)
	// Generate instance ID for this service instance
	hostname, err := os.Hostname()
	if err != nil {
		hostname = uuid.New().String()
	}
	instanceID := fmt.Sprintf("%s-%s", hostname, uuid.New().String()[:8])
	log.Printf("[Main] Service instance ID: %s", instanceID)

	// Create session store for WebSocket session persistence in Redis
	sessionStore := websocket.NewSessionStore(redisClient, instanceID)

	// Hub must be created before matchService so we can pass it
	// Initial callback is a placeholder - will be updated after disconnectService is created
	disconnectCallback := func(matchID, playerID uuid.UUID, disconnectedAt time.Time) {
		log.Printf("Player disconnected: match=%s player=%s", matchID, playerID)
	}
	hub := websocket.NewHub(sessionStore, matchRepo, disconnectCallback)
	go hub.Run()

	// Initialize match service
	matchService := application.NewMatchService(
		matchRepo,
		roundRepo,
		bracketRepo,
		publisher,
		metrics,
		logger,
		cfg.GameManagementURL,
	)
	// Set hub on match service after creation
	matchService.SetHub(hub)

	// Initialize disconnect service (T112-T122: CS:GO-style disconnect handling)
	disconnectService := application.NewDisconnectService(matchRepo, matchService, publisher, logger, metrics)

	// Update disconnect callback now that we have disconnectService
	hub.SetDisconnectCallback(func(matchID, playerID uuid.UUID, disconnectedAt time.Time) {
		disconnectCtx := context.Background()
		err := disconnectService.HandleDisconnect(disconnectCtx, matchID, playerID)
		if err != nil {
			log.Printf("ERROR: Failed to handle disconnect: %v", err)
		}
	})

	// Initialize pre-lobby participant store (Redis-backed)
	preLobbyParticipantStore := application.NewRedisPreLobbyParticipantStore(redisClient)

	// Initialize pre-lobby service
	preLobbyService := application.NewPreLobbyService(
		preLobbyRepo,
		matchRepo,
		roundRepo,
		nil, // Hub will be set after initialization
		publisher,
		metrics,
		logger,
		preLobbyParticipantStore,
	)

	// Initialize HTTP router
	router := gin.New()

	// Configure shared middleware
	httpxConfig := httpx.DefaultConfig()
	httpxConfig.ServiceName = "matchmaking"
	httpxConfig.AllowOrigins = cfg.CORSAllowedOrigins
	httpxConfig.AllowCredentials = true

	// Apply middleware (order matters!)
	router.Use(httpx.Recovery(slogLogger))      // Recover from panics
	router.Use(httpx.RequestLogger(slogLogger)) // Log requests with trace IDs

	// T135: Rate limiting (100 requests per minute per IP)
	rateLimiter := httphandlers.NewRateLimiter(100, time.Minute)
	router.Use(rateLimiter.Middleware()) // Rate limiting to prevent abuse

	router.Use(httpx.CORS(httpxConfig))            // Handle CORS
	router.Use(httpx.SecurityHeaders(httpxConfig)) // Add security headers

	// Health check endpoints (no auth required)
	dbHealthChecker := &DatabaseHealthChecker{pool: pool}
	redisHealthChecker := &RedisHealthChecker{client: redisClient}
	healthHandler := httphandlers.NewHealthHandler(dbHealthChecker, redisHealthChecker)

	router.GET("/health", healthHandler.HandleHealth)
	router.GET("/ready", healthHandler.HandleReadiness)
	router.GET("/live", healthHandler.HandleLiveness)

	// ==========================================
	// Game Management Routes (merged from game-management service)
	// ==========================================

	// Public game routes (no auth required)
	v1 := router.Group("/v1")
	gamehandlers.RegisterGameRoutes(v1, gamehandlers.GameDeps{
		Repo: gameRepo,
	})

	// Bundle streaming routes (no auth, allows iframe embedding)
	gamehandlers.RegisterBundleRoutes(v1, gamehandlers.BundleDeps{
		Repo:    gameRepo,
		Storage: s3Client,
	})

	// Admin game routes (internal API key auth)
	admin := router.Group("/v1/admin")
	admin.Use(InternalAPIKeyMiddleware(cfg.GameManagementInternalAPIKey))
	gamehandlers.RegisterAdminRoutes(admin, gamehandlers.AdminDeps{
		Repo:    gameRepo,
		Storage: s3Client,
	})

	// ==========================================
	// Tournament Integration (merged from tournament service)
	// ==========================================

	// Create slog logger for tournament handlers
	tournamentLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Initialize tournament repositories
	tournamentRepo := tournamentrepo.NewTournamentRepository(pool)
	registrationRepo := tournamentrepo.NewRegistrationRepository(pool)
	hostRepo := tournamentrepo.NewHostRepository(pool)

	// Initialize tournament use cases
	confirmParticipationUC := tournamentapp.NewConfirmParticipationUseCase(
		tournamentRepo,
		registrationRepo,
		tournamentLogger,
	)

	joinTournamentUC := tournamentapp.NewJoinTournamentUseCase(
		tournamentRepo,
		registrationRepo,
		tournamentLogger,
	)

	// Initialize tournament service client (for fetching tournament info)
	// Uses direct database access since tournament is merged into matchmaking
	competitionClient := application.NewDirectCompetitionClient(tournamentRepo, registrationRepo, logger)

	// Set hub on pre-lobby service (after hub is created)
	preLobbyService.SetHub(hub)

	// Initialize match assignment service (centralizes match assignment broadcasting)
	matchAssignmentService := application.NewMatchAssignmentService(
		bracketService,
		preLobbyService,
		logger,
	)

	// Initialize round transition store (Redis + DB projection for state persistence)
	roundTransitionStore := application.NewRoundTransitionStore(
		redisClient,
		matchRepo,
		roundRepo,
		bracketRepo,
		logger,
	)

	// Initialize round manager (orchestrates round transitions)
	roundManager := application.NewRoundManager(
		matchRepo,
		roundRepo,
		bracketRepo,
		preLobbyService,
		matchAssignmentService,
		roundTransitionStore,
		logger,
	)

	// Wire up round manager to match service and pre-lobby service
	matchService.SetRoundManager(roundManager)
	preLobbyService.SetRoundManager(roundManager)

	// Initialize event handler with pre-lobby service for grace period support
	eventHandler := application.NewEventHandler(
		bracketService,
		preLobbyService,
		matchAssignmentService,
		publisher,
		logger,
		cfg.GameManagementURL,
		hub,
	)

	// Initialize HTTP handlers
	bracketHandler := httphandlers.NewBracketHandler(bracketRepo, roundRepo, matchRepo)
	matchHandler := httphandlers.NewMatchHandler(matchRepo, roundRepo, bracketRepo, preLobbyService, matchService, hub, competitionClient)
	websocketHandler := httphandlers.NewWebSocketHandler(hub, matchRepo, publisher, sessionStore, disconnectService)
	preLobbyHandler := httphandlers.NewPreLobbyHandler(preLobbyService, competitionClient, logger)
	preLobbyWSHandler := httphandlers.NewPreLobbyWebSocketHandler(preLobbyService, competitionClient, hub, logger)

	// Set tournament disconnect callback on hub
	hub.SetTournamentDisconnectCallback(preLobbyWSHandler.HandleTournamentDisconnect)

	// API matchmaking routes (auth required)
	matchmaking := router.Group("/v1/matchmaking")
	// JWT validation middleware - validates token and extracts user context
	jwtConfig := authmiddleware.Config{
		JWTSecret: cfg.HostJWTSecret,
		Issuer:    cfg.HostJWTIssuer,
		Audience:  cfg.HostJWTAudience,
	}
	matchmaking.Use(authmiddleware.ValidateJWTMiddleware(jwtConfig))
	{
		// Bracket endpoints
		matchmaking.GET("/brackets/:id", bracketHandler.GetBracket)
		matchmaking.GET("/tournaments/:id/bracket", bracketHandler.GetBracketByTournament)

		// Match endpoints
		matchmaking.GET("/me", matchHandler.GetCurrentUserMatch)
		matchmaking.GET("/matches/:id", matchHandler.GetMatch)
		matchmaking.POST("/matches/:id/ready", matchHandler.MarkPlayerReady)
		matchmaking.POST("/matches/:id/game-loaded", matchHandler.MarkGameLoaded)
		matchmaking.GET("/matches/:id/game-loaded", matchHandler.GetGameLoadedStatus)
		matchmaking.POST("/matches/:id/claim-walkover", matchHandler.ClaimWalkover)
		matchmaking.POST("/matches/:id/complete", matchHandler.CompleteMatch)
		matchmaking.GET("/tournaments/:id/matches", matchHandler.GetMatchesForTournament)

		// WebSocket lobby endpoint
		matchmaking.GET("/matches/:id/lobby", websocketHandler.UpgradeLobbyConnection)

		// Pre-lobby endpoints (tournament waiting room)
		matchmaking.GET("/tournaments/:id/lobby", preLobbyHandler.GetPreLobbyState)
		matchmaking.GET("/tournaments/:id/lobby/ws", preLobbyWSHandler.UpgradePreLobbyConnection)
	}

	// ==========================================
	// Tournament Routes (merged from tournament service)
	// ==========================================

	// Create StartTournament callback that integrates with preLobbyService
	startTournamentFn := func(c gin.Context, tournamentID uuid.UUID, participantIDs []uuid.UUID) error {
		tournamentLogger.Info("Starting tournament via direct call",
			"tournamentId", tournamentID.String(),
			"participantCount", len(participantIDs),
		)

		// Create pre-lobby first (minimum 2 participants for a tournament)
		_, err := preLobbyService.GetOrCreatePreLobby(c.Request.Context(), tournamentID, 2)
		if err != nil {
			return fmt.Errorf("failed to create pre-lobby: %w", err)
		}

		// Create grace period callback
		gracePeriodCallback := func(tID uuid.UUID, pIDs []uuid.UUID) {
			tournamentLogger.Info("Grace period ended, generating bracket",
				"tournamentId", tID.String(),
				"participantCount", len(pIDs),
			)
			if len(pIDs) >= 2 {
				if err := bracketService.GenerateBracket(context.Background(), tID, pIDs, nil); err != nil {
					tournamentLogger.Error("Failed to generate bracket",
						"tournamentId", tID.String(),
						"error", err.Error(),
					)
				}
			}
		}

		// Start grace period
		if err := preLobbyService.StartGracePeriod(c.Request.Context(), tournamentID, gracePeriodCallback); err != nil {
			return fmt.Errorf("failed to start grace period: %w", err)
		}

		return nil
	}

	// Tournament host routes (admin) - /v1/:hostId/tournaments/...
	tournamentV1 := router.Group("/v1")
	tournamentV1.Use(authmiddleware.ValidateJWTMiddleware(jwtConfig))
	tournamenthandlers.RegisterTournamentHostRoutes(tournamentV1, tournamenthandlers.TournamentHostDeps{
		HostRepo:         hostRepo,
		TournamentRepo:   tournamentRepo,
		RegistrationRepo: registrationRepo,
		Logger:           tournamentLogger,
		StartTournament:  startTournamentFn,
	})

	// Tournament user routes (player) - /v1/:hostId/tournaments/...
	tournamenthandlers.RegisterTournamentUserRoutes(tournamentV1, tournamenthandlers.TournamentUserDeps{
		Pool:             pool,
		Logger:           tournamentLogger,
		TournamentRepo:   tournamentRepo,
		RegistrationRepo: registrationRepo,
		ConfirmParticipationFn: func(c *gin.Context, tournamentID, userID uuid.UUID, displayName string) error {
			return confirmParticipationUC.Execute(c.Request.Context(), tournamentID, userID, displayName)
		},
		JoinTournamentFn: func(c *gin.Context, tournamentID, userID, hostID uuid.UUID) (string, error) {
			return joinTournamentUC.Execute(c.Request.Context(), tournamentID, userID, hostID)
		},
	})

	logger.Info("Tournament routes registered", nil)

	// Create HTTP server
	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start HTTP server in goroutine
	go func() {
		logger.Info("HTTP server starting", map[string]interface{}{
			"addr": addr,
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Start event subscriber in goroutine
	subscriber := pubsub.NewEventSubscriber(redisClient)

	// Register event handlers
	eventHandler.RegisterHandlers(subscriber)

	// Recover active grace periods on startup (T030)
	gracePeriodCallback := func(tournamentID uuid.UUID, participantIDs []uuid.UUID) {
		logger.Info("Grace period callback triggered", map[string]interface{}{
			"tournament_id":     tournamentID.String(),
			"participant_count": len(participantIDs),
		})
		if len(participantIDs) >= 2 {
			if err := bracketService.GenerateBracket(context.Background(), tournamentID, participantIDs, nil); err != nil {
				logger.Error("Failed to generate bracket from grace period callback", map[string]interface{}{
					"tournament_id": tournamentID.String(),
					"error":         err.Error(),
				})
			}
		}
	}
	if err := preLobbyService.RecoverActiveGracePeriods(ctx, gracePeriodCallback); err != nil {
		logger.Warn("Failed to recover active grace periods", map[string]interface{}{
			"error": err.Error(),
		})
	}

	go func() {
		channels := pubsub.GetSubscriptionChannels()
		logger.Info("Event subscriber starting", map[string]interface{}{
			"channels": channels,
		})
		if err := subscriber.Start(ctx, channels); err != nil {
			log.Printf("Event subscriber stopped: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...", nil)

	// Send "server_restarting" message to all connected WebSocket clients
	// This gives clients a heads-up to expect reconnection
	log.Printf("[Shutdown] Notifying all WebSocket clients of server restart...")
	hub.BroadcastServerRestarting()

	// Give clients 2 seconds to receive the message and start reconnect logic
	time.Sleep(2 * time.Second)

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server stopped", nil)
}

// DatabaseHealthChecker implements HealthChecker for database
type DatabaseHealthChecker struct {
	pool *pgxpool.Pool
}

func (d *DatabaseHealthChecker) Health(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

// RedisHealthChecker implements HealthChecker for Redis
type RedisHealthChecker struct {
	client *redis.Client
}

func (r *RedisHealthChecker) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// InternalAPIKeyMiddleware validates the X-Internal-API-Key header for admin routes.
// This is used for service-to-service authentication and admin operations.
func InternalAPIKeyMiddleware(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth if no key is configured (dev mode)
		if expectedKey == "" {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-Internal-API-Key")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing X-Internal-API-Key header",
			})
			return
		}

		if apiKey != expectedKey {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "invalid API key",
			})
			return
		}

		c.Next()
	}
}
