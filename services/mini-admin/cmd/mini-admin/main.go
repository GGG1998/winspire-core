package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/winspire/mini-admin/internal/config"
	httpx "github.com/winspire/mini-admin/internal/http"
	"github.com/winspire/mini-admin/internal/repository"
	"github.com/winspire/mini-admin/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

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

	// Verify PostgreSQL connection
	if err := pool.Ping(ctx); err != nil {
		logger.Error("postgres ping failed", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	gameRepo := repository.NewGameRepository(pool)

	// Initialize S3 storage client
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
		return nil
	}

	// Create router
	router := httpx.NewRouter(httpx.ServerDeps{
		Config:      cfg,
		Logger:      logger,
		HealthCheck: healthCheck,
		GameRepo:    gameRepo,
		S3Client:    s3Client,
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
