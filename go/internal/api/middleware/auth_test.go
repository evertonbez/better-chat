package middleware

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"evertonbez/better-chat/internal/apperr"
	"evertonbez/better-chat/internal/auth"

	"github.com/gofiber/fiber/v3"
)

const testSecret = "test-secret"

type stubSessions struct {
	identity     *auth.Identity
	resolveErr   error
	refreshed    bool
	refreshErr   error
	resolveCalls []string
}

func (s *stubSessions) Resolve(_ context.Context, token string) (*auth.Identity, error) {
	s.resolveCalls = append(s.resolveCalls, token)
	if s.resolveErr != nil {
		return nil, s.resolveErr
	}
	return s.identity, nil
}

func (s *stubSessions) Refresh(_ context.Context, identity *auth.Identity) (*auth.Identity, bool, error) {
	if s.refreshErr != nil {
		return nil, false, s.refreshErr
	}
	return identity, s.refreshed, nil
}

func newIdentity() *auth.Identity {
	now := time.Now()
	return &auth.Identity{
		Session: auth.Session{
			ID:        1,
			Token:     "raw-token",
			UserID:    7,
			ExpiresAt: now.Add(time.Hour),
			CreatedAt: now,
			UpdatedAt: now,
		},
		User: auth.User{ID: 7, UID: "user-uid", Email: "a@b.com", Name: "Ana"},
	}
}

func newApp(t *testing.T, sessions SessionResolver, extra ...fiber.Handler) (*fiber.App, auth.CookieConfig) {
	t.Helper()

	cookie := auth.NewCookieConfig("session", "", false, time.Hour)

	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Use(Auth(AuthConfig{Sessions: sessions, Cookie: cookie, Secret: testSecret}))
	for _, h := range extra {
		app.Use(h)
	}
	app.Get("/probe", func(c fiber.Ctx) error {
		if id, ok := auth.FromContext(c); ok {
			return c.JSON(fiber.Map{"authenticated": true, "userId": id.User.UID})
		}
		return c.JSON(fiber.Map{"authenticated": false})
	})

	return app, cookie
}

func request(t *testing.T, app *fiber.App, cookieValue string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	if cookieValue != "" {
		req.AddCookie(&http.Cookie{Name: "session", Value: cookieValue})
	}

	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	return res
}

func TestAuthWithoutCookieLeavesRequestAnonymous(t *testing.T) {
	sessions := &stubSessions{identity: newIdentity()}
	app, _ := newApp(t, sessions)

	res := request(t, app, "")

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if len(sessions.resolveCalls) != 0 {
		t.Fatalf("no cookie should mean no lookup, got %d", len(sessions.resolveCalls))
	}
}

func TestAuthRejectsForgedCookieWithoutLookup(t *testing.T) {
	sessions := &stubSessions{identity: newIdentity()}
	app, _ := newApp(t, sessions)

	res := request(t, app, "raw-token.not-a-real-signature")

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if len(sessions.resolveCalls) != 0 {
		t.Fatal("a cookie that fails the signature check must not reach the session store")
	}
	if !clearsCookie(res) {
		t.Fatal("expected the unusable cookie to be cleared")
	}
}

func TestAuthResolvesValidCookie(t *testing.T) {
	sessions := &stubSessions{identity: newIdentity()}
	app, _ := newApp(t, sessions)

	res := request(t, app, auth.SignToken("raw-token", testSecret))

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if len(sessions.resolveCalls) != 1 || sessions.resolveCalls[0] != "raw-token" {
		t.Fatalf("expected the raw token to be resolved once, got %v", sessions.resolveCalls)
	}
	if body := readBody(t, res); !strings.Contains(body, `"authenticated":true`) {
		t.Fatalf("request was not authenticated: %s", body)
	}
}

func TestAuthClearsCookieForExpiredSession(t *testing.T) {
	sessions := &stubSessions{resolveErr: apperr.SessionExpired}
	app, _ := newApp(t, sessions)

	res := request(t, app, auth.SignToken("raw-token", testSecret))

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if !clearsCookie(res) {
		t.Fatal("expected the expired session cookie to be cleared")
	}
}

func TestAuthPropagatesInfrastructureFailures(t *testing.T) {

	sessions := &stubSessions{resolveErr: apperr.Internal.Cause(errors.New("pool exhausted"))}
	app, _ := newApp(t, sessions)

	res := request(t, app, auth.SignToken("raw-token", testSecret))

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", res.StatusCode)
	}
}

func TestAuthRewritesCookieOnRefresh(t *testing.T) {
	sessions := &stubSessions{identity: newIdentity(), refreshed: true}
	app, _ := newApp(t, sessions)

	res := request(t, app, auth.SignToken("raw-token", testSecret))

	setCookie := res.Header.Get("Set-Cookie")
	if setCookie == "" {
		t.Fatal("a refreshed session should be written back to the client")
	}
	if !strings.Contains(setCookie, auth.SignToken("raw-token", testSecret)) {
		t.Fatalf("refreshed cookie does not carry the signed token: %s", setCookie)
	}
}

func TestAuthKeepsSessionWhenRefreshFails(t *testing.T) {
	sessions := &stubSessions{identity: newIdentity(), refreshErr: errors.New("update failed")}
	app, _ := newApp(t, sessions)

	res := request(t, app, auth.SignToken("raw-token", testSecret))

	if body := readBody(t, res); !strings.Contains(body, `"authenticated":true`) {
		t.Fatalf("a failed expiry extension must not sign the user out: %s", body)
	}
}

func TestRequireAuthBlocksAnonymousRequests(t *testing.T) {
	sessions := &stubSessions{identity: newIdentity()}
	app, _ := newApp(t, sessions, RequireAuth())

	res := request(t, app, "")

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.StatusCode)
	}
	if body := readBody(t, res); !strings.Contains(body, string(apperr.CodeUnauthorized)) {
		t.Fatalf("unexpected error body: %s", body)
	}
}

func TestRequireAuthAllowsAuthenticatedRequests(t *testing.T) {
	sessions := &stubSessions{identity: newIdentity()}
	app, _ := newApp(t, sessions, RequireAuth())

	res := request(t, app, auth.SignToken("raw-token", testSecret))

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
}

func TestRequireVerifiedEmail(t *testing.T) {
	identity := newIdentity()
	sessions := &stubSessions{identity: identity}
	app, _ := newApp(t, sessions, RequireAuth(), RequireVerifiedEmail())

	res := request(t, app, auth.SignToken("raw-token", testSecret))
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("unverified email: status = %d, want 403", res.StatusCode)
	}

	identity.User.EmailVerified = true
	res = request(t, app, auth.SignToken("raw-token", testSecret))
	if res.StatusCode != http.StatusOK {
		t.Fatalf("verified email: status = %d, want 200", res.StatusCode)
	}
}

func clearsCookie(res *http.Response) bool {
	for _, c := range res.Cookies() {
		if c.Name == "session" && (c.MaxAge < 0 || c.Value == "") {
			return true
		}
	}
	return false
}

func readBody(t *testing.T, res *http.Response) string {
	t.Helper()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	_ = res.Body.Close()

	return string(body)
}
