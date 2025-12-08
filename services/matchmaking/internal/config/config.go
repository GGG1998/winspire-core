// Package config handles environment variable configuration
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the matchmaking service
type Config struct {
	// Server
	Port    string
	GinMode string

	// Database
	DatabaseURL      string
	DatabaseMaxConns int32
	DatabaseMaxIdle  int32

	// Redis
	RedisURL      string
	RedisPassword string
	RedisDB       int

	// JWT Authentication
	HostJWTSecret   string
	HostJWTIssuer   string
	HostJWTAudience string

	// Game API
	GameAPIURL                     string
	GameAPIKey                     string
	GameAPIPollInterval            time.Duration
	GameAPIPollTimeout             time.Duration
	GameAPICircuitBreakerThreshold int
	GameAPICircuitBreakerTimeout   time.Duration

	// Competition Service
	CompetitionServiceURL string

	// Game Management Service
	GameManagementURL string

	// Observability
	LogLevel            string
	LogFormat           string
	CloudWatchRegion    string
	CloudWatchLogGroup  string
	CloudWatchNamespace string

	// Feature Flags
	EnableWebSocketLobbies   bool
	EnableAutoScoreRetrieval bool
	EnableDisconnectHandling bool

	// Timeouts
	LobbyJoinTimeout          time.Duration
	DisconnectReconnectWindow time.Duration
	ReadyCheckTimeout         time.Duration

	// Limits
	MaxConcurrentTournaments     int
	MaxParticipantsPerTournament int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		// Server defaults
		Port:    getEnv("PORT", "8081"),
		GinMode: getEnv("GIN_MODE", "release"),

		// Database
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		DatabaseMaxConns: getEnvAsInt32("DATABASE_MAX_CONNS", 25),
		DatabaseMaxIdle:  getEnvAsInt32("DATABASE_MAX_IDLE_CONNS", 5),

		// Redis
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379/0"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// JWT
		HostJWTSecret:   getEnv("HOST_JWT_SECRET", ""),
		HostJWTIssuer:   getEnv("HOST_JWT_ISSUER", ""),
		HostJWTAudience: getEnv("HOST_JWT_AUDIENCE", ""),

		// Game API
		GameAPIURL:                     getEnv("GAME_API_URL", ""),
		GameAPIKey:                     getEnv("GAME_API_KEY", ""),
		GameAPIPollInterval:            getEnvAsDuration("GAME_API_POLL_INTERVAL", 5*time.Second),
		GameAPIPollTimeout:             getEnvAsDuration("GAME_API_POLL_TIMEOUT", 60*time.Second),
		GameAPICircuitBreakerThreshold: getEnvAsInt("GAME_API_CIRCUIT_BREAKER_THRESHOLD", 5),
		GameAPICircuitBreakerTimeout:   getEnvAsDuration("GAME_API_CIRCUIT_BREAKER_TIMEOUT", 30*time.Second),

		// Competition Service
		CompetitionServiceURL: getEnv("COMPETITION_SERVICE_URL", "http://localhost:8080"),

		// Game Management Service
		GameManagementURL: getEnv("GAME_MANAGEMENT_URL", "http://localhost:8087"),

		// Observability
		LogLevel:            getEnv("LOG_LEVEL", "debug"),
		LogFormat:           getEnv("LOG_FORMAT", "json"),
		CloudWatchRegion:    getEnv("CLOUDWATCH_REGION", "us-east-1"),
		CloudWatchLogGroup:  getEnv("CLOUDWATCH_LOG_GROUP", "/aws/ecs/matchmaking"),
		CloudWatchNamespace: getEnv("CLOUDWATCH_NAMESPACE", "Winspire/Matchmaking"),

		// Feature Flags
		EnableWebSocketLobbies:   getEnvAsBool("ENABLE_WEBSOCKET_LOBBIES", true),
		EnableAutoScoreRetrieval: getEnvAsBool("ENABLE_AUTO_SCORE_RETRIEVAL", true),
		EnableDisconnectHandling: getEnvAsBool("ENABLE_DISCONNECT_HANDLING", true),

		// Timeouts
		LobbyJoinTimeout:          getEnvAsDuration("LOBBY_JOIN_TIMEOUT", 2*time.Minute),
		DisconnectReconnectWindow: getEnvAsDuration("DISCONNECT_RECONNECT_WINDOW", 30*time.Second),
		ReadyCheckTimeout:         getEnvAsDuration("READY_CHECK_TIMEOUT", 5*time.Minute),

		// Limits
		MaxConcurrentTournaments:     getEnvAsInt("MAX_CONCURRENT_TOURNAMENTS", 50),
		MaxParticipantsPerTournament: getEnvAsInt("MAX_PARTICIPANTS_PER_TOURNAMENT", 256),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.HostJWTSecret == "" {
		log.Println("WARN: HOST_JWT_SECRET not set, using insecure default")
		cfg.HostJWTSecret = "insecure-dev-secret"
	}

	return cfg, nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("WARN: Invalid integer for %s, using default %d", key, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsInt32(key string, defaultValue int32) int32 {
	return int32(getEnvAsInt(key, int(defaultValue)))
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Printf("WARN: Invalid boolean for %s, using default %v", key, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("WARN: Invalid duration for %s, using default %v", key, defaultValue)
		return defaultValue
	}
	return value
}
