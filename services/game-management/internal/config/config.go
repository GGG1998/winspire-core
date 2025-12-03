package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config captures all runtime tunables provided via environment variables.
type Config struct {
	AppEnv        string
	ServicePort   int
	PostgresDSN   string
	RedisAddr     string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	ShutdownGrace time.Duration

	// Supabase Storage configuration
	SupabaseURL        string
	SupabaseServiceKey string
	GamesBucket        string

	// JWT configuration for authentication
	HostJWTSecret   string
	HostJWTIssuer   string
	HostJWTAudience string

	// Cache configuration
	BundleCacheTTL time.Duration
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:             valueOrDefault("APP_ENV", "development"),
		ServicePort:        intFromEnv("SERVICE_PORT", 8087),
		PostgresDSN:        valueOrDefault("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/game_management?sslmode=disable"),
		RedisAddr:          valueOrDefault("REDIS_ADDR", "localhost:6379"),
		ReadTimeout:        durationFromEnv("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:       durationFromEnv("HTTP_WRITE_TIMEOUT", 15*time.Second),
		ShutdownGrace:      durationFromEnv("SHUTDOWN_GRACE", 10*time.Second),
		SupabaseURL:        valueOrDefault("SUPABASE_URL", "http://localhost:54321"),
		SupabaseServiceKey: valueOrDefault("SUPABASE_SERVICE_KEY", ""),
		GamesBucket:        valueOrDefault("GAMES_BUCKET", "games"),
		HostJWTSecret:      valueOrDefault("HOST_JWT_SECRET", ""),
		HostJWTIssuer:      valueOrDefault("HOST_JWT_ISSUER", ""),
		HostJWTAudience:    valueOrDefault("HOST_JWT_AUDIENCE", ""),
		BundleCacheTTL:     durationFromEnv("BUNDLE_CACHE_TTL", 1*time.Hour),
	}

	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("POSTGRES_DSN must be provided")
	}
	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR must be provided")
	}

	return cfg, nil
}

// HasAuth indicates whether JWT middleware should be enabled.
func (c Config) HasAuth() bool {
	return c.HostJWTSecret != "" && c.HostJWTIssuer != "" && c.HostJWTAudience != ""
}

// HasSupabaseStorage indicates whether Supabase storage is configured.
func (c Config) HasSupabaseStorage() bool {
	return c.SupabaseURL != "" && c.SupabaseServiceKey != ""
}

func valueOrDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func intFromEnv(key string, def int) int {
	if val := os.Getenv(key); val != "" {
		if out, err := strconv.Atoi(val); err == nil {
			return out
		}
	}
	return def
}

func durationFromEnv(key string, def time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if out, err := time.ParseDuration(val); err == nil {
			return out
		}
	}
	return def
}

