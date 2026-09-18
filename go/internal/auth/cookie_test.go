package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestCookieNameGetsSecurePrefix(t *testing.T) {
	insecure := NewCookieConfig("session", "", false, time.Hour)
	if got := insecure.CookieName(); got != "session" {
		t.Fatalf("CookieName() = %q, want %q", got, "session")
	}

	secure := NewCookieConfig("session", "", true, time.Hour)
	if got := secure.CookieName(); got != "__Secure-session" {
		t.Fatalf("CookieName() = %q, want %q", got, "__Secure-session")
	}
}

func setCookieHeader(t *testing.T, fn func(c fiber.Ctx)) string {
	t.Helper()

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		fn(c)
		return c.SendStatus(fiber.StatusOK)
	})

	res, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	return res.Header.Get("Set-Cookie")
}

func TestWriteSetsHardenedAttributes(t *testing.T) {
	cfg := NewCookieConfig("session", "", true, time.Hour)

	header := setCookieHeader(t, func(c fiber.Ctx) {
		cfg.Write(c, "token.signature", time.Now().Add(time.Hour))
	})

	for _, want := range []string{"__Secure-session=token.signature", "HttpOnly", "secure", "SameSite=Lax", "path=/"} {
		if !strings.Contains(strings.ToLower(header), strings.ToLower(want)) {
			t.Fatalf("Set-Cookie %q is missing %q", header, want)
		}
	}
}

func TestClearMatchesWrite(t *testing.T) {
	cfg := NewCookieConfig("session", "app.example.com", true, time.Hour)

	written := setCookieHeader(t, func(c fiber.Ctx) {
		cfg.Write(c, "token.signature", time.Now().Add(time.Hour))
	})
	cleared := setCookieHeader(t, func(c fiber.Ctx) { cfg.Clear(c) })

	for _, attr := range []string{"path=/", "domain=app.example.com", "httponly", "secure", "samesite=lax"} {
		if !strings.Contains(strings.ToLower(written), attr) {
			t.Fatalf("Write did not set %q: %s", attr, written)
		}
		if !strings.Contains(strings.ToLower(cleared), attr) {
			t.Fatalf("Clear did not set %q, so the browser will keep the cookie: %s", attr, cleared)
		}
	}

	if !strings.Contains(strings.ToLower(cleared), "max-age=0") {
		t.Fatalf("Clear did not expire the cookie: %s", cleared)
	}
}

func TestReadReturnsCookieValue(t *testing.T) {
	cfg := NewCookieConfig("session", "", false, time.Hour)

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(cfg.Read(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "token.signature"})

	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	buf := make([]byte, 64)
	n, _ := res.Body.Read(buf)
	if got := string(buf[:n]); got != "token.signature" {
		t.Fatalf("Read() = %q, want %q", got, "token.signature")
	}
}
