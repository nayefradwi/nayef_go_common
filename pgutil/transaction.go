package pgutil

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RunWithTx[Q any](ctx context.Context, pool *pgxpool.Pool, bind func(pgx.Tx) Q, fn func(Q) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(bind(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func RunInTx[Q any, R any](ctx context.Context, pool *pgxpool.Pool, bind func(pgx.Tx) Q, fn func(Q) (R, error)) (R, error) {
	var zero R
	tx, err := pool.Begin(ctx)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback(ctx)
	result, err := fn(bind(tx))
	if err != nil {
		return zero, err
	}
	if err := tx.Commit(ctx); err != nil {
		return zero, err
	}
	return result, nil
}
