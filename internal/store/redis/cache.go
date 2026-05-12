package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Get retrieves a string value by key.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.redis.Get(ctx, key).Result()
}

// Set stores a string value with optional TTL.
func (c *Client) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.redis.Set(ctx, key, value, ttl).Err()
}

// GetJSON retrieves a value and unmarshals it into dest.
func (c *Client) GetJSON(ctx context.Context, key string, dest any) error {
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// SetJSON marshals value to JSON and stores it with optional TTL.
func (c *Client) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.redis.Set(ctx, key, data, ttl).Err()
}

// Exists checks if a key exists.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Del deletes one or more keys.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.redis.Del(ctx, keys...).Err()
}

// IsNil returns true if the error is a redis.Nil (key not found).
func IsNil(err error) bool {
	return err == redis.Nil
}
