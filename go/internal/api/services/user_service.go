package services

import (
	"context"
	"errors"
	"strings"

	"evertonbez/better-chat/gen/dbstore"
	"evertonbez/better-chat/internal/apperr"
	"evertonbez/better-chat/internal/auth"
	"evertonbez/better-chat/pkg/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

type UserService struct {
	store    *db.Store
	cache    *redis.Client
	sessions *SessionService
}

func NewUserService(store *db.Store, c *redis.Client, sessions *SessionService) *UserService {
	return &UserService{store: store, cache: c, sessions: sessions}
}

func (u *UserService) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	user, err := u.store.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.UserNotFound.Cause(err).With("email", email)
		}
		return nil, apperr.Internal.Cause(err)
	}

	return ptr(auth.UserFromRow(user)), nil
}

func (u *UserService) GetByID(ctx context.Context, id int64) (*auth.User, error) {
	user, err := u.store.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.UserNotFound.Cause(err).With("id", id)
		}
		return nil, apperr.Internal.Cause(err)
	}

	return ptr(auth.UserFromRow(user)), nil
}

func (u *UserService) GetByUID(ctx context.Context, uid string) (*auth.User, error) {
	user, err := u.store.GetUserByUID(ctx, text(uid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.UserNotFound.Cause(err).With("uid", uid)
		}
		return nil, apperr.Internal.Cause(err)
	}

	return ptr(auth.UserFromRow(user)), nil
}

type UpdateProfileParams struct {
	Name  *string
	Image *string
}

func (u *UserService) UpdateProfile(ctx context.Context, userID int64, p UpdateProfileParams) (*auth.User, error) {
	arg := dbstore.UpdateUserParams{ID: userID}

	if p.Name != nil {
		name := strings.TrimSpace(*p.Name)
		if name == "" {
			return nil, apperr.Validation.Msg("name cannot be empty").With("field", "name")
		}
		arg.Name = pgtype.Text{String: name, Valid: true}
	}

	if p.Image != nil {

		arg.Image = pgtype.Text{String: *p.Image, Valid: true}
	}

	user, err := u.store.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.UserNotFound.Cause(err).With("id", userID)
		}
		return nil, apperr.Internal.Cause(err)
	}

	u.sessions.InvalidateUser(ctx, userID)

	return ptr(auth.UserFromRow(user)), nil
}

func ptr[T any](v T) *T { return &v }
