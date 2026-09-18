package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"evertonbez/better-chat/gen/dbstore"
	"evertonbez/better-chat/internal/apperr"
	"evertonbez/better-chat/internal/auth"
	"evertonbez/better-chat/internal/config"
	"evertonbez/better-chat/pkg/cache"
	"evertonbez/better-chat/pkg/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

type SessionOptions struct {
	TTL time.Duration

	UpdateAge time.Duration

	CacheTTL time.Duration
}

func DefaultSessionOptions() SessionOptions {
	return SessionOptions{
		TTL:       config.SESSION_TTL,
		UpdateAge: config.SESSION_UPDATE_AGE,
		CacheTTL:  config.SESSION_CACHE_TTL,
	}
}

type SessionService struct {
	store *db.Store
	cache *redis.Client
	opts  SessionOptions
}

func NewSessionService(store *db.Store, c *redis.Client, opts SessionOptions) *SessionService {
	return &SessionService{store: store, cache: c, opts: opts}
}

type CreateParams struct {
	UserID    int64
	IPAddress string
	UserAgent string
}

func (s *SessionService) Create(ctx context.Context, p CreateParams) (*auth.Identity, error) {
	token, err := auth.NewToken()
	if err != nil {
		return nil, apperr.Internal.Cause(err)
	}

	row, err := s.store.CreateSession(ctx, dbstore.CreateSessionParams{
		Token:     token,
		UserID:    p.UserID,
		ExpiresAt: timestamptz(time.Now().Add(s.opts.TTL)),
		IpAddress: text(p.IPAddress),
		UserAgent: text(p.UserAgent),
	})
	if err != nil {
		return nil, apperr.Internal.Cause(err)
	}

	user, err := s.store.GetUserByID(ctx, p.UserID)
	if err != nil {
		return nil, apperr.Internal.Cause(err)
	}

	identity := &auth.Identity{
		Session: auth.SessionFromRow(row),
		User:    auth.UserFromRow(user),
	}

	s.cacheIdentity(ctx, identity)

	return identity, nil
}

func (s *SessionService) Resolve(ctx context.Context, token string) (*auth.Identity, error) {
	now := time.Now()

	if identity, ok := s.cached(ctx, token); ok {
		if identity.Session.Expired(now) {

			s.evict(ctx, token)
			return nil, apperr.SessionExpired
		}
		return identity, nil
	}

	row, err := s.store.GetLiveSessionWithUserByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.SessionExpired
		}
		return nil, apperr.Internal.Cause(err)
	}

	identity := &auth.Identity{
		Session: auth.SessionFromRow(row.Session),
		User:    auth.UserFromRow(row.User),
	}

	if identity.Session.Expired(now) {

		return nil, apperr.SessionExpired
	}

	s.cacheIdentity(ctx, identity)

	return identity, nil
}

func (s *SessionService) Refresh(ctx context.Context, identity *auth.Identity) (*auth.Identity, bool, error) {
	if !identity.Session.NeedsRefresh(time.Now(), s.opts.UpdateAge) {
		return identity, false, nil
	}

	row, err := s.store.RefreshSession(ctx, dbstore.RefreshSessionParams{
		Token:     identity.Session.Token,
		ExpiresAt: timestamptz(time.Now().Add(s.opts.TTL)),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {

			s.evict(ctx, identity.Session.Token)
			return nil, false, apperr.SessionExpired
		}
		return nil, false, apperr.Internal.Cause(err)
	}

	refreshed := &auth.Identity{
		Session: auth.SessionFromRow(row),
		User:    identity.User,
	}

	s.cacheIdentity(ctx, refreshed)

	return refreshed, true, nil
}

func (s *SessionService) Revoke(ctx context.Context, token string) error {
	s.evict(ctx, token)

	if _, err := s.store.RevokeSessionByToken(ctx, token); err != nil {
		return apperr.Internal.Cause(err)
	}

	return nil
}

func (s *SessionService) RevokeAllForUser(ctx context.Context, userID int64) error {
	var tokens []string

	err := s.store.ExecTx(ctx, func(q *dbstore.Queries) error {
		var err error
		tokens, err = q.ListLiveSessionTokensByUserID(ctx, userID)
		if err != nil {
			return err
		}

		_, err = q.RevokeSessionsByUserID(ctx, userID)
		return err
	})
	if err != nil {
		return apperr.Internal.Cause(err)
	}

	s.evict(ctx, tokens...)

	return nil
}

func (s *SessionService) InvalidateUser(ctx context.Context, userID int64) {
	tokens, err := s.store.ListLiveSessionTokensByUserID(ctx, userID)
	if err != nil {
		slog.Warn("could not list sessions to invalidate", "error", err, "user_id", userID)
		return
	}

	s.evict(ctx, tokens...)
}

func (s *SessionService) History(ctx context.Context, userID int64) ([]auth.Session, error) {
	rows, err := s.store.ListSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Internal.Cause(err)
	}

	sessions := make([]auth.Session, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, auth.SessionFromRow(row))
	}

	return sessions, nil
}

func (s *SessionService) cached(ctx context.Context, token string) (*auth.Identity, bool) {
	if s.cache == nil {
		return nil, false
	}

	identity, err := cache.GetJSON[auth.Identity](ctx, s.cache, sessionKey(token))
	if err != nil {
		if !errors.Is(err, cache.ErrMiss) {
			slog.Warn("session cache read failed, falling back to database", "error", err)
		}
		return nil, false
	}

	identity.Session.Token = token

	return identity, true
}

func (s *SessionService) cacheIdentity(ctx context.Context, identity *auth.Identity) {
	if s.cache == nil {
		return
	}

	ttl := s.opts.CacheTTL
	if remaining := time.Until(identity.Session.ExpiresAt); remaining < ttl {
		ttl = remaining
	}
	if ttl <= 0 {
		return
	}

	if err := cache.SetJSON(ctx, s.cache, sessionKey(identity.Session.Token), identity, ttl); err != nil {
		slog.Warn("session cache write failed", "error", err)
	}
}

func (s *SessionService) evict(ctx context.Context, tokens ...string) {
	if s.cache == nil || len(tokens) == 0 {
		return
	}

	keys := make([]string, 0, len(tokens))
	for _, token := range tokens {
		keys = append(keys, sessionKey(token))
	}

	if err := s.cache.Del(ctx, keys...).Err(); err != nil {

		slog.Error("session cache eviction failed, revoked sessions may be served until TTL",
			"error", err, "keys", len(keys))
	}
}

func sessionKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "auth:session:" + hex.EncodeToString(sum[:])
}

func timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func text(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}
