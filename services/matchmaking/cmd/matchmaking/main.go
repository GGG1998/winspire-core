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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/winspire-core/services/matchmaking/internal/config"
	httphandlers "github.com/winspire-core/services/matchmaking/internal/http"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/pubsub"
	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
	"github.com/winspire-core/services/matchmaking/internal/websocket"
	authmiddleware "github.com/winspire/winspire-core/libs/go/auth/middleware"
	"github.com/winspire/winspire-core/libs/go/httpx"
)

func main() {
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

	// Initialize WebSocket hub
	hub := websocket.NewHub(nil) // Disconnect callback will be set by match service
	go hub.Run()

	// Initialize HTTP router
	router := gin.New()

	// Configure shared middleware
	httpxConfig := httpx.DefaultConfig()
	httpxConfig.ServiceName = "matchmaking"

	// Apply middleware (order matters!)
	router.Use(httpx.Recovery(slogLogger))         // Recover from panics
	router.Use(httpx.RequestLogger(slogLogger))    // Log requests with trace IDs
	router.Use(httpx.CORS(httpxConfig))            // Handle CORS
	router.Use(httpx.SecurityHeaders(httpxConfig)) // Add security headers

	// Health check endpoints (no auth required)
	dbHealthChecker := &DatabaseHealthChecker{pool: pool}
	redisHealthChecker := &RedisHealthChecker{client: redisClient}
	healthHandler := httphandlers.NewHealthHandler(dbHealthChecker, redisHealthChecker)

	router.GET("/health", healthHandler.HandleHealth)
	router.GET("/ready", healthHandler.HandleReadiness)
	router.GET("/live", healthHandler.HandleLiveness)

	// API v1 routes (auth required)
	v1 := router.Group("/v1")
	// JWT validation middleware - validates token and extracts user context
	jwtConfig := authmiddleware.Config{
		JWTSecret: cfg.JWTSecret,
		Issuer:    "winspire",
		Audience:  "winspire-api",
	}
	v1.Use(authmiddleware.ValidateJWTMiddleware(jwtConfig))
	{
		// TODO: Register bracket, match, and WebSocket handlers here
		// Example:
		// v1.GET("/tournaments/:id/bracket", bracketHandler.GetBracket)
		// v1.GET("/matches/:id", matchHandler.GetMatch)
		// v1.GET("/matches/:id/lobby", websocketHandler.UpgradeConnection)
		//
		// Use httpx.RequireRole() or httpx.RequireAuth() for additional auth checks
		// Use httpx.MustGetUser(c) in handlers to get the authenticated user
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
	// TODO: Register event handlers
	// subscriber.Subscribe("TournamentStarted", handleTournamentStarted)

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

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server stopped", nil)

	// Track unused variables to avoid compile errors
	_ = queries
	_ = publisher
	_ = hub
	_ = metrics
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
