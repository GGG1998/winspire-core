package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
{% if cookiecutter.use_redis == "true" %}
	"time"

	"github.com/redis/go-redis/v9"
{% endif %}
	"github.com/jackc/pgx/v5/pgxpool"

	"{{ cookiecutter.go_module }}/internal/config"
	httpx "{{ cookiecutter.go_module }}/internal/http"
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

{% if cookiecutter.use_redis == "true" %}
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
{% endif %}

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
{% if cookiecutter.use_redis == "true" %}
		if err := redisClient.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis: %w", err)
		}
{% endif %}
		return nil
	}

	// Create router
	router := httpx.NewRouter(httpx.ServerDeps{
		Config:      cfg,
		Logger:      logger,
		HealthCheck: healthCheck,
		Pool:        pool,
{% if cookiecutter.use_redis == "true" %}
		Redis:       redisClient,
{% endif %}
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


