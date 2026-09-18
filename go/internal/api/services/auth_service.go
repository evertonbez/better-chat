package services

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"sync"

	"evertonbez/better-chat/gen/dbstore"
	"evertonbez/better-chat/internal/apperr"
	"evertonbez/better-chat/internal/auth"
	"evertonbez/better-chat/pkg/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

const uniqueViolation = "23505"

type AuthService struct {
	store    *db.Store
	cache    *redis.Client
	sessions *SessionService
	users    *UserService
}

func NewAuthService(store *db.Store, c *redis.Client, sessions *SessionService, users *UserService) *AuthService {
	return &AuthService{store: store, cache: c, sessions: sessions, users: users}
}

type SignUpParams struct {
	Name      string
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

type SignInParams struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

func (a *AuthService) SignUp(ctx context.Context, p SignUpParams) (*auth.Identity, error) {
	email, err := normalizeEmail(p.Email)
	if err != nil {
		return nil, err
	}

	if err := auth.ValidatePassword(p.Password); err != nil {
		return nil, apperr.Validation.Msg("%s", err.Error()).With("field", "password")
	}

	name := strings.TrimSpace(p.Name)
	if name == "" {
		return nil, apperr.Validation.Msg("name is required").With("field", "name")
	}

	hash, err := auth.HashPassword(p.Password)
	if err != nil {
		return nil, apperr.Internal.Cause(err)
	}

	var user dbstore.User

	err = a.store.ExecTx(ctx, func(q *dbstore.Queries) error {
		user, err = q.CreateUser(ctx, dbstore.CreateUserParams{
			Uid:           text(uuid.NewString()),
			Name:          text(name),
			Email:         email,
			EmailVerified: false,
			Image:         text(""),
		})
		if err != nil {
			return err
		}

		_, err = q.CreateAccount(ctx, dbstore.CreateAccountParams{
			ProviderID: auth.ProviderCredential,
			UserID:     user.ID,
			Password:   text(hash),
		})
		return err
	})
	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return nil, apperr.EmailTaken.Cause(err).With("email", email)
		}
		return nil, apperr.Internal.Cause(err)
	}

	return a.sessions.Create(ctx, CreateParams{
		UserID:    user.ID,
		IPAddress: p.IPAddress,
		UserAgent: p.UserAgent,
	})
}

func (a *AuthService) SignIn(ctx context.Context, p SignInParams) (*auth.Identity, error) {
	email, err := normalizeEmail(p.Email)
	if err != nil {
		return nil, apperr.InvalidCredentials
	}

	user, err := a.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			equalizeTiming(p.Password)
			return nil, apperr.InvalidCredentials
		}
		return nil, apperr.Internal.Cause(err)
	}

	account, err := a.store.GetAccountByProvider(ctx, dbstore.GetAccountByProviderParams{
		UserID:     user.ID,
		ProviderID: auth.ProviderCredential,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {

			equalizeTiming(p.Password)
			return nil, apperr.InvalidCredentials
		}
		return nil, apperr.Internal.Cause(err)
	}

	if !account.Password.Valid || account.Password.String == "" {
		equalizeTiming(p.Password)
		return nil, apperr.InvalidCredentials
	}

	ok, err := auth.VerifyPassword(account.Password.String, p.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidHash) {
			return nil, apperr.Internal.Cause(err).With("user_id", user.ID)
		}
		return nil, apperr.InvalidCredentials
	}
	if !ok {
		return nil, apperr.InvalidCredentials
	}

	return a.sessions.Create(ctx, CreateParams{
		UserID:    user.ID,
		IPAddress: p.IPAddress,
		UserAgent: p.UserAgent,
	})
}

func (a *AuthService) SignOut(ctx context.Context, token string) error {
	return a.sessions.Revoke(ctx, token)
}

func (a *AuthService) SignOutAll(ctx context.Context, userID int64) error {
	return a.sessions.RevokeAllForUser(ctx, userID)
}

func (a *AuthService) ChangePassword(ctx context.Context, userID int64, current, next string) error {
	if err := auth.ValidatePassword(next); err != nil {
		return apperr.Validation.Msg("%s", err.Error()).With("field", "newPassword")
	}

	account, err := a.store.GetAccountByProvider(ctx, dbstore.GetAccountByProviderParams{
		UserID:     userID,
		ProviderID: auth.ProviderCredential,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.AccountNotFound.Cause(err).Msg("no password is set for this account")
		}
		return apperr.Internal.Cause(err)
	}

	ok, err := auth.VerifyPassword(account.Password.String, current)
	if err != nil || !ok {
		return apperr.InvalidCredentials.Msg("current password is incorrect")
	}

	hash, err := auth.HashPassword(next)
	if err != nil {
		return apperr.Internal.Cause(err)
	}

	if err := a.store.UpdateAccountPassword(ctx, dbstore.UpdateAccountPasswordParams{
		UserID:     userID,
		ProviderID: auth.ProviderCredential,
		Password:   text(hash),
	}); err != nil {
		return apperr.Internal.Cause(err)
	}

	return a.sessions.RevokeAllForUser(ctx, userID)
}

func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", apperr.Validation.Msg("email is required").With("field", "email")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return "", apperr.Validation.Cause(err).Msg("email is not a valid address").With("field", "email")
	}

	return email, nil
}

var dummyHash = sync.OnceValue(func() string {
	hash, err := auth.HashPassword(uuid.NewString())
	if err != nil {
		return ""
	}
	return hash
})

func equalizeTiming(password string) {
	if hash := dummyHash(); hash != "" {
		_, _ = auth.VerifyPassword(hash, password)
	}
}
