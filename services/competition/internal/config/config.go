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

	// JWT configuration for authentication
	HostJWTSecret   string
	HostJWTIssuer   string
	HostJWTAudience string

	// Scheduler configuration
	SchedulerEnabled  bool
	SchedulerInterval string

	// Logging configuration
	LogLevel string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:          valueOrDefault("APP_ENV", "development"),
		ServicePort:     intFromEnv("SERVICE_PORT", 8089),
		PostgresDSN:     valueOrDefault("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/competition_db?sslmode=disable"),

		RedisAddr:       valueOrDefault("REDIS_ADDR", "localhost:6379"),

		ReadTimeout:     durationFromEnv("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    durationFromEnv("HTTP_WRITE_TIMEOUT", 15*time.Second),
		ShutdownGrace:   durationFromEnv("SHUTDOWN_GRACE", 10*time.Second),
		HostJWTSecret:   valueOrDefault("HOST_JWT_SECRET", ""),
		HostJWTIssuer:   valueOrDefault("HOST_JWT_ISSUER", ""),
		HostJWTAudience: valueOrDefault("HOST_JWT_AUDIENCE", ""),
		
		SchedulerEnabled:  boolFromEnv("SCHEDULER_ENABLED", true),
		SchedulerInterval: valueOrDefault("SCHEDULER_INTERVAL", "*/2 * * * *"),

		LogLevel: valueOrDefault("LOG_LEVEL", "debug"),
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

func boolFromEnv(key string, def bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1"
	}
	return def
}


