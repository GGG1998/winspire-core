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
	"github.com/winspire/winspire-core/services/auth/internal/config"
	"github.com/winspire/winspire-core/services/auth/internal/db"
	"github.com/winspire/winspire-core/services/auth/internal/handlers"
	"github.com/winspire/winspire-core/services/auth/internal/logger"
	"github.com/winspire/winspire-core/services/auth/internal/supabase"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.New(cfg.Env)

	// Set Gin mode based on environment
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database connection
	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize Supabase client
	supabaseClient, err := supabase.New(cfg.SupabaseURL, cfg.SupabaseAnonKey)
	if err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}
	_ = supabaseClient // Will be used in services

	// Initialize router
	router := gin.Default()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()

	// Health check endpoint
	router.GET("/health", healthHandler.Check)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		appLogger.Info("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	appLogger.Info("Server exited")
}

