package services

import (
	"context"

	"evertonbez/better-chat/gen/dbstore"
	"evertonbez/better-chat/pkg/db"

	"github.com/redis/go-redis/v9"
)

type UserService struct {
	store *db.Store
	cache *redis.Client
}

func NewUserService(store *db.Store, cache *redis.Client) *UserService {
	return &UserService{
		store: store,
		cache: cache,
	}
}

func (u *UserService) GetByEmail(ctx context.Context, email string) (*dbstore.User, error) {
	user, err := u.store.GetUserByEmail(ctx, email)
	if err != nil {
		// if errors.Is(err, pgx.ErrNoRows) {
		// 	return nil, apperr.TransferNotFound.Cause(err).With("id", id)
		// }
		// return nil, apperr.Internal.Cause(err)

		return nil, err
	}

	return &user, nil
}
