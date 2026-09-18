package middleware

import (
	"context"
	"errors"
	"log/slog"

	"evertonbez/better-chat/internal/apperr"
	"evertonbez/better-chat/internal/auth"

	"github.com/gofiber/fiber/v3"
)

type SessionResolver interface {
	Resolve(ctx context.Context, token string) (*auth.Identity, error)
	Refresh(ctx context.Context, identity *auth.Identity) (*auth.Identity, bool, error)
}

type AuthConfig struct {
	Sessions SessionResolver
	Cookie   auth.CookieConfig
	Secret   string
}

func Auth(cfg AuthConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		signed := cfg.Cookie.Read(c)
		if signed == "" {
			return c.Next()
		}

		token, err := auth.UnsignToken(signed, cfg.Secret)
		if err != nil {

			cfg.Cookie.Clear(c)
			return c.Next()
		}

		identity, err := cfg.Sessions.Resolve(c, token)
		if err != nil {
			if errors.Is(err, apperr.SessionExpired) {
				cfg.Cookie.Clear(c)
				return c.Next()
			}

			return err
		}

		refreshed, didRefresh, err := cfg.Sessions.Refresh(c, identity)
		if err != nil {
			if errors.Is(err, apperr.SessionExpired) {
				cfg.Cookie.Clear(c)
				return c.Next()
			}

			slog.Warn("could not refresh session", "error", err)
			refreshed = identity
		}

		if didRefresh {
			cfg.Cookie.Write(c, auth.SignToken(refreshed.Session.Token, cfg.Secret), refreshed.Session.ExpiresAt)
		}

		auth.SetIdentity(c, refreshed)

		return c.Next()
	}
}

func RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		if _, ok := auth.FromContext(c); !ok {
			return apperr.Unauthorized
		}

		return c.Next()
	}
}

func RequireVerifiedEmail() fiber.Handler {
	return func(c fiber.Ctx) error {
		identity, ok := auth.FromContext(c)
		if !ok {
			return apperr.Unauthorized
		}

		if !identity.User.EmailVerified {
			return apperr.EmailNotVerified
		}

		return c.Next()
	}
}
