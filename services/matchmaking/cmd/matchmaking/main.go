// Package main is the entrypoint for the matchmaking service
package main

import (
	"context"
	"fmt"
	"log"
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
	httphandlers "github.com/winspire-core/services/matchmaking/internal/http"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/pubsub"
	"github.com/winspire-core/services/matchmaking/internal/repository"
	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
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

	// Initialize tournament service client (for fetching tournament info)
	competitionClient := application.NewCompetitionClient(cfg.CompetitionServiceURL, logger)

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
		competitionClient,
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
		matchmaking.POST("/matches/:id/claim-walkover", matchHandler.ClaimWalkover)
		matchmaking.POST("/matches/:id/complete", matchHandler.CompleteMatch)
		matchmaking.GET("/tournaments/:id/matches", matchHandler.GetMatchesForTournament)

		// WebSocket lobby endpoint
		matchmaking.GET("/matches/:id/lobby", websocketHandler.UpgradeLobbyConnection)

		// Pre-lobby endpoints (tournament waiting room)
		matchmaking.GET("/tournaments/:id/lobby", preLobbyHandler.GetPreLobbyState)
		matchmaking.GET("/tournaments/:id/lobby/ws", preLobbyWSHandler.UpgradePreLobbyConnection)
	}

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
