package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrMiss = errors.New("cache: miss")

func New(ctx context.Context, uri string) (*redis.Client, error) {
	opt, err := redis.ParseURL(uri)
	if err != nil {
		return nil, fmt.Errorf("Error parsing redis url: %w", err)
	}

	client := redis.NewClient(opt)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Error connect to redis: %w", err)
	}

	return client, nil
}

func Close(client *redis.Client) {
	if err := client.Close(); err != nil {
		slog.Error("error disconnecting from redis", "error", err)
	}
}

func GetJSON[T any](ctx context.Context, client *redis.Client, key string) (*T, error) {
	raw, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrMiss
		}
		return nil, fmt.Errorf("cache get %q: %w", key, err)
	}

	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("cache decode %q: %w", key, err)
	}

	return &value, nil
}

func SetJSON(ctx context.Context, client *redis.Client, key string, value any, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache encode %q: %w", key, err)
	}

	if err := client.Set(ctx, key, raw, ttl).Err(); err != nil {
		return fmt.Errorf("cache set %q: %w", key, err)
	}

	return nil
}
