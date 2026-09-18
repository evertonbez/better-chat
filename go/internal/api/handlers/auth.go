package handlers

import "github.com/redis/go-redis/v9"

type AuthHandler struct {
	Cache *redis.Client
}

func NewAuthHandler(cache *redis.Client) *AuthHandler {
	return &AuthHandler{
		Cache: cache,
	}
}
