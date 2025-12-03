package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/winspire/libs/go/auth/jwt"

	"{{ cookiecutter.go_module }}/internal/config"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// UserContextKey is the key for storing user info in the request context.
	UserContextKey contextKey = "user"
)

// LoggerMiddleware returns a middleware that logs HTTP requests.
func LoggerMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				logger.Info("http request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"bytes", ww.BytesWritten(),
					"duration", time.Since(start).String(),
					"request_id", middleware.GetReqID(r.Context()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// AuthMiddleware returns a middleware that validates JWT tokens.
func AuthMiddleware(cfg config.Config, logger *slog.Logger) func(next http.Handler) http.Handler {
	parser := jwt.NewParser(cfg.HostJWTSecret, cfg.HostJWTIssuer, cfg.HostJWTAudience)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			claims, err := parser.Parse(parts[1])
			if err != nil {
				logger.Debug("jwt parse error", "error", err)
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the user claims from the request context.
func GetUserFromContext(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*jwt.Claims)
	return claims, ok
}


