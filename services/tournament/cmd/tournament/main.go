package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/winspire/tournament/internal/application"
	"github.com/winspire/tournament/internal/config"
	httpx "github.com/winspire/tournament/internal/http"
	"github.com/winspire/tournament/internal/pubsub"
	"github.com/winspire/tournament/internal/repository"
	"github.com/winspire/tournament/internal/scheduler"
)

func main() {
	// Load .env file in development (Docker Compose will override with its own env vars)
	// Try multiple locations to handle different working directories
	envPaths := []string{".env", "services/tournament/.env", "../../.env"}
	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err == nil {
				break
			}
		}
	}

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLogLevel(cfg.LogLevel)}))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize PostgreSQL connection
	poolCfg, err := pgxpool.ParseConfig(cfg.PostgresDSN)
	if err != nil {
		logger.Error("failed to parse postgres dsn", "error", err)
		os.Exit(1)
	}
	poolCfg.MaxConns = 8
	// Use simple protocol for Supabase PgBouncer compatibility (avoids prepared statement conflicts)
	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Initialize Redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:        cfg.RedisAddr,
		DialTimeout: 5 * time.Second,
	})
	defer redisClient.Close()

	// Verify Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("redis ping failed", "error", err)
		os.Exit(1)
	}

	// Verify PostgreSQL connection
	if err := pool.Ping(ctx); err != nil {
		logger.Error("postgres ping failed", "error", err)
		os.Exit(1)
	}

	// Health check function
	healthCheck := func(ctx context.Context) error {
		if err := pool.Ping(ctx); err != nil {
			return fmt.Errorf("postgres: %w", err)
		}

		if err := redisClient.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis: %w", err)
		}

		return nil
	}

	// Initialize repositories (persistence layer)
	tournamentPersistenceRepo := repository.NewTournamentRepository(pool)
	registrationPersistenceRepo := repository.NewRegistrationRepository(pool)

	// Initialize domain repository adapters
	tournamentDomainRepo := repository.NewTournamentDomainRepository(tournamentPersistenceRepo)
	participantDomainRepo := repository.NewParticipantDomainRepository(registrationPersistenceRepo)

	// Initialize use cases
	confirmParticipationUC := application.NewConfirmParticipationUseCase(
		tournamentDomainRepo,
		participantDomainRepo,
		logger,
	)

	joinTournamentUC := application.NewJoinTournamentUseCase(
		tournamentDomainRepo,
		participantDomainRepo,
		logger,
	)

	updateStatusUC := application.NewUpdateTournamentStatusUseCase(
		tournamentDomainRepo,
		participantDomainRepo,
		logger,
	)

	// Initialize event publisher
	eventPublisher := pubsub.NewEventPublisher(redisClient)

	// Initialize game-management client
	gameManagementURL := os.Getenv("GAME_MANAGEMENT_URL")
	if gameManagementURL == "" {
		gameManagementURL = "http://localhost:8083" // Default for local dev
	}
	gameClient := application.NewGameManagementClient(gameManagementURL)

	// Initialize event subscriber for saga coordination
	eventSubscriber := pubsub.NewEventSubscriber(redisClient, logger)
	eventHandler := application.NewEventHandler(tournamentPersistenceRepo, gameClient, eventPublisher, logger)
	eventHandler.RegisterHandlers(eventSubscriber)

	// Start event subscriber in background
	go func() {
		channels := []string{
			"events:matchmaking:grace_period_started",
			"events:matchmaking:bracket_generation_failed",
		}
		logger.Info("event subscriber starting", "channels", channels)
		if err := eventSubscriber.Start(ctx, channels); err != nil {
			logger.Error("event subscriber error", "error", err)
		}
	}()

	// Create router with dependencies
	router := httpx.NewRouter(httpx.ServerDeps{
		Config:                 cfg,
		Logger:                 logger,
		HealthCheck:            healthCheck,
		Pool:                   pool,
		Redis:                  redisClient,
		EventPublisher:         eventPublisher,
		ConfirmParticipationUC: confirmParticipationUC,
		JoinTournamentUC:       joinTournamentUC,
		ParticipantRepo:        participantDomainRepo,
	})

	// Initialize and start scheduler if enabled
	var tournamentScheduler *scheduler.TournamentStatusScheduler
	if cfg.SchedulerEnabled {
		tournamentScheduler = scheduler.NewTournamentStatusScheduler(
			updateStatusUC,
			logger,
			cfg.SchedulerInterval,
		)

		go func() {
			logger.Info("scheduler starting", "interval", cfg.SchedulerInterval)
			tournamentScheduler.Run(ctx)
		}()
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.ServicePort),
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	logger.Info("shutdown signal received")

	// Stop scheduler if running
	if tournamentScheduler != nil {
		tournamentScheduler.Stop()
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownGrace)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	logger.Info("server stopped")
}

// parseLogLevel converts a string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
