package auth

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestIdentitySurvivesCacheRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)

	original := Identity{
		Session: Session{
			ID:        42,
			Token:     "raw-token",
			UserID:    7,
			ExpiresAt: now.Add(time.Hour),
			IPAddress: "203.0.113.4",
			UserAgent: "curl/8",
			CreatedAt: now,
			UpdatedAt: now,
		},
		User: User{
			ID:            7,
			UID:           "b6f0c5b2-0000-4000-8000-000000000000",
			Name:          "Ana",
			Email:         "ana@example.com",
			EmailVerified: true,
			Image:         "https://example.com/a.png",
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}

	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Identity
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Session.Token != "" {
		t.Fatal("the raw session token must not be stored in the cached payload")
	}
	decoded.Session.Token = original.Session.Token

	if decoded != original {
		t.Fatalf("cache round trip lost data:\n got %+v\nwant %+v", decoded, original)
	}
}

func TestCachedPayloadCarriesNoCredentials(t *testing.T) {
	raw, err := json.Marshal(Identity{
		Session: Session{Token: "raw-token"},
		User:    User{Email: "ana@example.com"},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	for _, forbidden := range []string{"password", "raw-token"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("cached payload contains %q: %s", forbidden, raw)
		}
	}
}

func TestSessionExpiredAndNeedsRefresh(t *testing.T) {
	now := time.Now()

	s := Session{ExpiresAt: now.Add(time.Hour), UpdatedAt: now}
	if s.Expired(now) {
		t.Fatal("a session with time left was reported expired")
	}
	if s.NeedsRefresh(now, time.Hour) {
		t.Fatal("a just-written session should not be refreshed yet")
	}
	if !s.NeedsRefresh(now.Add(2*time.Hour), time.Hour) {
		t.Fatal("a session untouched for longer than updateAge should refresh")
	}

	expired := Session{ExpiresAt: now.Add(-time.Second)}
	if !expired.Expired(now) {
		t.Fatal("a past expiry was not reported expired")
	}

	boundary := Session{ExpiresAt: now}
	if !boundary.Expired(now) {
		t.Fatal("a session expiring exactly now should be treated as expired")
	}
}
