package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type Client struct {
	redis *redis.Client
}

// New creates a new Redis client and verifies the connection.
func New(ctx context.Context, cfg RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Client{redis: rdb}, nil
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.redis.Close()
}
