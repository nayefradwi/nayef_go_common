package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nayefradwi/nayef_go_common/result"
	"go.uber.org/zap"
)

func Tx(ctx context.Context, pool *pgxpool.Pool, f func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		zap.L().Error("Error starting transaction", zap.Error(err))
		return result.InternalError("Error starting transaction")
	}
	defer tx.Rollback(ctx)
	if err := f(ctx, tx); err != nil {
		zap.L().Error("Error executing transaction", zap.Error(err))
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		zap.L().Error("Error committing transaction", zap.Error(err))
		return result.InternalError("Error committing transaction")
	}

	return nil
}

func TxWithData[T any](ctx context.Context, pool *pgxpool.Pool, f func(ctx context.Context, tx pgx.Tx) (T, error)) (T, error) {
	var empty T
	tx, err := pool.Begin(ctx)
	if err != nil {
		zap.L().Error("Error starting transaction", zap.Error(err))
		return empty, result.InternalError("Error starting transaction")
	}
	defer tx.Rollback(ctx)
	data, err := f(ctx, tx)
	if err != nil {
		zap.L().Error("Error executing transaction", zap.Error(err))
		return data, err
	}

	if err := tx.Commit(ctx); err != nil {
		zap.L().Error("Error committing transaction", zap.Error(err))
		return empty, result.InternalError("Error committing transaction")
	}

	return data, nil
}
