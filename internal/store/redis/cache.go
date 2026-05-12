package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Get retrieves a string value by key.
func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

// Set stores a string value with optional TTL.
func (rc *RedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return rc.client.Set(ctx, key, value, ttl).Err()
}

// GetJSON retrieves a value and unmarshals it into dest.
func (rc *RedisClient) GetJSON(ctx context.Context, key string, dest any) error {
	data, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// SetJSON marshals value to JSON and stores it with optional TTL.
func (rc *RedisClient) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return rc.client.Set(ctx, key, data, ttl).Err()
}

// Exists checks if a key exists.
func (rc *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	n, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Del deletes one or more keys.
func (rc *RedisClient) Del(ctx context.Context, keys ...string) error {
	return rc.client.Del(ctx, keys...).Err()
}

// IsNil returns true if the error is a redis.Nil (key not found).
func IsNil(err error) bool {
	return err == redis.Nil
}
