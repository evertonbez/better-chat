package auth

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

type CookieConfig struct {
	Name     string
	Domain   string
	Path     string
	Secure   bool
	SameSite string
	TTL      time.Duration
}

func NewCookieConfig(name, domain string, secure bool, ttl time.Duration) CookieConfig {
	return CookieConfig{
		Name:     name,
		Domain:   domain,
		Path:     "/",
		Secure:   secure,
		SameSite: fiber.CookieSameSiteLaxMode,
		TTL:      ttl,
	}
}

func (c CookieConfig) CookieName() string {
	if c.Secure {
		return "__Secure-" + c.Name
	}
	return c.Name
}

func (c CookieConfig) Read(ctx fiber.Ctx) string {
	return ctx.Cookies(c.CookieName())
}

func (c CookieConfig) Write(ctx fiber.Ctx, signedToken string, expiresAt time.Time) {
	ctx.Cookie(&fiber.Cookie{
		Name:     c.CookieName(),
		Value:    signedToken,
		Path:     c.Path,
		Domain:   c.Domain,
		Expires:  expiresAt,
		Secure:   c.Secure,
		HTTPOnly: true,
		SameSite: c.SameSite,
	})
}

func (c CookieConfig) Clear(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     c.CookieName(),
		Value:    "",
		Path:     c.Path,
		Domain:   c.Domain,
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
		Secure:   c.Secure,
		HTTPOnly: true,
		SameSite: c.SameSite,
	})
}
