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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/winspire/game-management/internal/config"
	httpx "github.com/winspire/game-management/internal/http"
	"github.com/winspire/game-management/internal/repository"
	"github.com/winspire/game-management/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize PostgreSQL connection
	poolCfg, err := pgxpool.ParseConfig(cfg.PostgresDSN)
	if err != nil {
		logger.Error("failed to parse postgres dsn", "error", err)
		os.Exit(1)
	}
	poolCfg.MaxConns = 8

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

	// Verify connections
	if err := pool.Ping(ctx); err != nil {
		logger.Error("postgres ping failed", "error", err)
		os.Exit(1)
	}
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("redis ping failed", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	gameRepo := repository.NewGameRepository(pool)

	s3Client, err := storage.NewS3Client(cfg.AWSRegion, cfg.AWSS3Bucket, cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey)
	if err != nil {
		logger.Error("failed to create S3 client", "error", err)
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

	// Create router
	router := httpx.NewRouter(httpx.ServerDeps{
		Config:      cfg,
		Logger:      logger,
		HealthCheck: healthCheck,
		GameRepo:    gameRepo,
		Storage:     s3Client,
	})

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

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownGrace)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	logger.Info("server stopped")
}
