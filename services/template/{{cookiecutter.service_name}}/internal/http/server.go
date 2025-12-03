package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
{% if cookiecutter.use_redis == "true" %}
	"github.com/redis/go-redis/v9"
{% endif %}

	"{{ cookiecutter.go_module }}/internal/config"
	"{{ cookiecutter.go_module }}/internal/http/handlers"
)

// ServerDeps holds all dependencies for the HTTP server.
type ServerDeps struct {
	Config      config.Config
	Logger      *slog.Logger
	HealthCheck func(ctx context.Context) error
	Pool        *pgxpool.Pool
{% if cookiecutter.use_redis == "true" %}
	Redis       *redis.Client
{% endif %}
}

// NewRouter creates and configures the HTTP router.
func NewRouter(deps ServerDeps) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(LoggerMiddleware(deps.Logger))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoints
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := deps.HealthCheck(r.Context()); err != nil {
			deps.Logger.Error("health check failed", "error", err)
			http.Error(w, "unhealthy", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Add authentication middleware if configured
		if deps.Config.HasAuth() {
			r.Use(AuthMiddleware(deps.Config, deps.Logger))
		}

		// Register handlers here
		h := handlers.New(deps.Pool, deps.Logger)
		_ = h // TODO: Register your handlers
	})

	return r
}


