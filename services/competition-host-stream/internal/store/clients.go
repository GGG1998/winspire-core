package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/winspire/competition-host-stream/internal/config"
)

// Clients groups external infrastructure handles shared by projections, SSE broker, etc.
type Clients struct {
	PG    *pgxpool.Pool
	Redis *redis.Client
}

// NewClients initializes Postgres + Redis connections and performs basic health checks.
func NewClients(ctx context.Context, cfg config.Config) (*Clients, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}
	poolCfg.MaxConns = 8

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:        cfg.RedisAddr,
		DialTimeout: 5 * time.Second,
	})

	c := &Clients{PG: pool, Redis: redisClient}
	if err := c.Health(ctx); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

// Health checks both backends; returns error if either is unavailable.
func (c *Clients) Health(ctx context.Context) error {
	if err := c.PG.Ping(ctx); err != nil {
		return fmt.Errorf("postgres ping: %w", err)
	}
	if err := c.Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	return nil
}

// Close releases resources.
func (c *Clients) Close() {
	if c.PG != nil {
		c.PG.Close()
	}
	if c.Redis != nil {
		_ = c.Redis.Close()
	}
}

