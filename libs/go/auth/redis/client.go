package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis client configuration
type Config struct {
	// Address is the Redis server address (host:port)
	Address string

	// Password for Redis authentication (optional)
	Password string

	// DB is the Redis database number
	DB int

	// PoolSize is the maximum number of socket connections
	PoolSize int

	// MinIdleConns is the minimum number of idle connections
	MinIdleConns int

	// ConnMaxIdleTime is the maximum idle time before closing connections
	ConnMaxIdleTime time.Duration

	// DialTimeout is the timeout for establishing new connections
	DialTimeout time.Duration

	// ReadTimeout is the timeout for socket reads
	ReadTimeout time.Duration

	// WriteTimeout is the timeout for socket writes
	WriteTimeout time.Duration
}

// DefaultConfig returns a Redis configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Address:         "localhost:6379",
		Password:        "",
		DB:              0,
		PoolSize:        10,
		MinIdleConns:    2,
		ConnMaxIdleTime: 5 * time.Minute,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
	}
}

// ProductionConfig returns a Redis configuration optimized for production
func ProductionConfig() Config {
	return Config{
		Address:         "", // Must be set
		Password:        "", // Should be set
		DB:              0,
		PoolSize:        50,
		MinIdleConns:    10,
		ConnMaxIdleTime: 5 * time.Minute,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
	}
}

// NewClient creates a new Redis client with the given configuration
func NewClient(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:            cfg.Address,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}

// NewClientFromURL creates a Redis client from a URL
// Format: redis://[:password@]host:port[/db]
func NewClientFromURL(url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}

// HealthCheck performs a health check on the Redis connection
func HealthCheck(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
