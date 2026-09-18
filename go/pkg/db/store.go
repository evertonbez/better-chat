package db

import (
	"context"
	"fmt"

	"evertonbez/better-chat/gen/dbstore"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	*dbstore.Queries
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		Queries: dbstore.New(pool),
		pool:    pool,
	}
}

func (s *Store) ExecTx(ctx context.Context, fn func(*dbstore.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer tx.Rollback(ctx) // no-op after commit

	if err := fn(s.Queries.WithTx(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
