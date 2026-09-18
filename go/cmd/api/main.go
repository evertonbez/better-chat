package main

import (
	"context"
	"log/slog"
	"os"

	"evertonbez/better-chat/internal/api"
	"evertonbez/better-chat/internal/config"
	"evertonbez/better-chat/pkg/cache"
	"evertonbez/better-chat/pkg/db"
)

func main() {
	config.Load(os.Getenv("ENV"))

	slog.Info("starting API", "port", config.PORT, "env", config.ENV)

	ctx := context.Background()

	pool, err := db.New(ctx, config.DATABASE_URL)
	if err != nil {
		slog.Error("could not connect to database", "error", err)
		os.Exit(1)
	}

	slog.Info("database connected", "env", config.ENV)

	c, err := cache.New(ctx, config.REDIS_URL)
	if err != nil {
		slog.Error("could not connect to cache client", "error", err)
		os.Exit(1)
	}

	slog.Info("cache connected", "env", config.ENV)

	defer cache.Close(c)

	api.Start(&api.APIConfig{
		Store: db.NewStore(pool),
		Port:  config.PORT,
		Cache: c,
	})
}
