package api

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"evertonbez/better-chat/internal/api/middleware"
	"evertonbez/better-chat/internal/api/routes"
	"evertonbez/better-chat/pkg/db"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

type APIConfig struct {
	Logger *slog.Logger
	Store  *db.Store
	Port   string
	Cache  *redis.Client
}

func New(cfg *APIConfig) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Use(requestid.New())
	app.Use(helmet.New())
	app.Use(recover.New())

	routes.New(cfg.Store, cfg.Cache).InitV1(app)

	return app
}

func Start(cfg *APIConfig) {
	app := New(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &fasthttp.Server{
		Handler: app.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", cfg.Port)); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.ShutdownWithContext(shutdownCtx); err != nil {
		log.Fatal("Server shutdowns:", err)
	}

	cfg.Logger.Info("server stopped")
}
