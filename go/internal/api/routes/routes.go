package routes

import (
	"evertonbez/better-chat/pkg/db"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

type RouteParams struct {
	Store *db.Store
	Cache *redis.Client
}

func New(store *db.Store, cache *redis.Client) *RouteParams {
	return &RouteParams{
		Store: store,
		Cache: cache,
	}
}

func (r *RouteParams) InitV1(app *fiber.App) {
	// auth := handlers.NewAuthHandler(r.Cache)

	// v1 := app.Group("/api/v1")
}
