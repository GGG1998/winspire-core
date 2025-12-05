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
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	ShutdownGrace time.Duration

	// AWS S3 configuration
	AWSRegion          string
	AWSS3Bucket        string
	AWSAccessKeyID     string
	AWSSecretAccessKey string

	// JWT configuration for authentication
	AdminJWTSecret   string
	AdminJWTIssuer   string
	AdminJWTAudience string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:             valueOrDefault("APP_ENV", "development"),
		ServicePort:        intFromEnv("SERVICE_PORT", 8088),
		PostgresDSN:        valueOrDefault("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/mini_admin?sslmode=disable"),
		ReadTimeout:        durationFromEnv("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:       durationFromEnv("HTTP_WRITE_TIMEOUT", 15*time.Second),
		ShutdownGrace:      durationFromEnv("SHUTDOWN_GRACE", 10*time.Second),
		AWSRegion:          valueOrDefault("AWS_REGION", "eu-central-1"),
		AWSS3Bucket:        valueOrDefault("AWS_S3_BUCKET", "mini-admin-games"),
		AWSAccessKeyID:     valueOrDefault("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: valueOrDefault("AWS_SECRET_ACCESS_KEY", ""),
		AdminJWTSecret:     valueOrDefault("ADMIN_JWT_SECRET", ""),
		AdminJWTIssuer:     valueOrDefault("ADMIN_JWT_ISSUER", ""),
		AdminJWTAudience:   valueOrDefault("ADMIN_JWT_AUDIENCE", ""),
	}

	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("POSTGRES_DSN must be provided")
	}
	if cfg.AWSS3Bucket == "" {
		return Config{}, fmt.Errorf("AWS_S3_BUCKET must be provided")
	}

	return cfg, nil
}

// HasAuth indicates whether JWT middleware should be enabled.
func (c Config) HasAuth() bool {
	return c.AdminJWTSecret != "" && c.AdminJWTIssuer != "" && c.AdminJWTAudience != ""
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
