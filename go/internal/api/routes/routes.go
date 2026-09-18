package routes

import (
	"evertonbez/better-chat/internal/api/handlers"
	"evertonbez/better-chat/internal/api/middleware"
	"evertonbez/better-chat/internal/api/services"
	"evertonbez/better-chat/internal/auth"
	"evertonbez/better-chat/internal/config"
	"evertonbez/better-chat/pkg/db"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

type RouteParams struct {
	Store *db.Store
	Cache *redis.Client

	cookie      auth.CookieConfig
	sessions    *services.SessionService
	users       *services.UserService
	authService *services.AuthService
}

func New(store *db.Store, cache *redis.Client) *RouteParams {
	opts := services.DefaultSessionOptions()
	cookie := auth.NewCookieConfig(config.COOKIE_NAME, config.COOKIE_DOMAIN, config.COOKIE_SECURE, opts.TTL)

	sessions := services.NewSessionService(store, cache, opts)
	users := services.NewUserService(store, cache, sessions)

	return &RouteParams{
		Store:       store,
		Cache:       cache,
		cookie:      cookie,
		sessions:    sessions,
		users:       users,
		authService: services.NewAuthService(store, cache, sessions, users),
	}
}

func (r *RouteParams) InitV1(app *fiber.App) {
	authHandler := handlers.NewAuthHandler(r.authService, r.sessions, r.cookie, config.AUTH_SECRET)
	userHandler := handlers.NewUserHandler(r.users, r.sessions)

	v1 := app.Group("/api/v1", middleware.Auth(middleware.AuthConfig{
		Sessions: r.sessions,
		Cookie:   r.cookie,
		Secret:   config.AUTH_SECRET,
	}))

	requireAuth := middleware.RequireAuth()

	v1.Route("/auth", func(api fiber.Router) {
		api.Post("/sign-up/email", authHandler.SignUp)
		api.Post("/sign-in/email", authHandler.SignIn)
		api.Post("/sign-out", authHandler.SignOut)
		api.Get("/session", authHandler.GetSession)

		api.Post("/sign-out/all", requireAuth, authHandler.SignOutAll)
		api.Post("/change-password", requireAuth, authHandler.ChangePassword)
	})

	v1.Route("/users", func(api fiber.Router) {
		api.Get("/me", requireAuth, userHandler.Me)
		api.Patch("/me", requireAuth, userHandler.UpdateMe)
		api.Get("/me/sessions", requireAuth, userHandler.Sessions)
	})
}
