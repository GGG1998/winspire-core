package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config captures all runtime tunables that should be provided via env vars (or App Runner secrets).
type Config struct {
	AppEnv          string
	ServicePort     int
	PostgresDSN     string
	RedisAddr       string
	SSEPoolSize     int
	LeaseTTL        time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownGrace   time.Duration
	HostJWTSecret   string
	HostJWTIssuer   string
	HostJWTAudience string
}

// Load reads configuration from environment variables, applying sensible defaults for local dev.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:          valueOrDefault("APP_ENV", "development"),
		ServicePort:     intFromEnv("SERVICE_PORT", 8086),
		PostgresDSN:     valueOrDefault("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/competition_host?sslmode=disable"),
		RedisAddr:       valueOrDefault("REDIS_ADDR", "localhost:6379"),
		SSEPoolSize:     intFromEnv("SSE_POOL_SIZE", 5000),
		LeaseTTL:        durationFromEnv("LEASE_TTL", 90*time.Second),
		ReadTimeout:     durationFromEnv("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    durationFromEnv("HTTP_WRITE_TIMEOUT", 15*time.Second),
		ShutdownGrace:   durationFromEnv("SHUTDOWN_GRACE", 10*time.Second),
		HostJWTSecret:   valueOrDefault("HOST_JWT_SECRET", ""),
		HostJWTIssuer:   valueOrDefault("HOST_JWT_ISSUER", ""),
		HostJWTAudience: valueOrDefault("HOST_JWT_AUDIENCE", ""),
	}

	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("POSTGRES_DSN must be provided")
	}
	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR must be provided")
	}
	if cfg.SSEPoolSize <= 0 {
		return Config{}, fmt.Errorf("SSE_POOL_SIZE must be positive")
	}
	return cfg, nil
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

// HasAuth indicates whether JWT middleware should be enabled.
func (c Config) HasAuth() bool {
	return c.HostJWTSecret != "" && c.HostJWTIssuer != "" && c.HostJWTAudience != ""
}
