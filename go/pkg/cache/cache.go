package cache

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

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
		log.Fatal("Error disconnecting from redis", "error", err)
	}
}
