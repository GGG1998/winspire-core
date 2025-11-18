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
	"github.com/winspire/winspire-core/libs/go/auth/middleware"
	"github.com/winspire/winspire-core/services/auth/internal/config"
	"github.com/winspire/winspire-core/services/auth/internal/handlers"
	"github.com/winspire/winspire-core/services/auth/internal/logger"
	"github.com/winspire/winspire-core/services/auth/internal/services"
	supabaseclient "github.com/winspire/winspire-core/services/auth/internal/supabase"
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

	// Initialize Supabase client
	supabaseClient, err := supabaseclient.New(cfg.SupabaseURL, cfg.SupabaseAnonKey)
	if err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}

	// Initialize router
	router := gin.Default()

	// Initialize services
	registrationService := services.NewRegistrationService(supabaseClient, appLogger)
	authService := services.NewAuthService(supabaseClient, appLogger)
	passwordService := services.NewPasswordService(supabaseClient, appLogger)
	oauthService := services.NewOAuthService(supabaseClient, cfg.SupabaseURL, appLogger)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(registrationService, authService, appLogger)
	passwordHandler := handlers.NewPasswordHandler(passwordService, appLogger)
	oauthHandler := handlers.NewOAuthHandler(oauthService, appLogger)

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Health check
		router.GET("/health", healthHandler.Check)

		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.GET("/verify", authHandler.VerifyEmail)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			
			// Protected routes (require JWT)
			authProtected := auth.Group("")
			authProtected.Use(middleware.ValidateJWTMiddleware(middleware.Config{
				JWTSecret: cfg.SupabaseJWTSecret,
				Issuer:    cfg.SupabaseURL,
				Audience:  "authenticated",
			}))
			{
				authProtected.POST("/logout", authHandler.Logout)
			}
		}

		// Password management routes
		password := v1.Group("/auth/password")
		{
			password.POST("/reset", passwordHandler.RequestPasswordReset)
			password.POST("/reset/confirm", passwordHandler.ConfirmPasswordReset)
		}

		// OAuth routes
		oauth := v1.Group("/auth/oauth")
		{
			oauth.GET("/:provider", oauthHandler.InitiateOAuth)
			oauth.GET("/:provider/callback", oauthHandler.OAuthCallback)
		}
	}

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

