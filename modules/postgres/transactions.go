package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nayefradwi/nayef_go_common/core"
	"go.uber.org/zap"
)

func WithTx(ctx context.Context, pool *pgxpool.Pool, f func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		zap.L().Error("Error starting transaction", zap.Error(err))
		return core.InternalError("Error starting transaction")
	}
	defer rollbackOnPanic(tx)
	if err := f(ctx, tx); err != nil {
		tx.Rollback(ctx)
	}

	if err := tx.Commit(ctx); err != nil {
		zap.L().Error("Error committing transaction", zap.Error(err))
		return core.InternalError("Error committing transaction")
	}

	return nil
}

func rollbackOnPanic(tx pgx.Tx) {
	if r := recover(); r != nil {
		tx.Rollback(context.Background())
		panic(r)
	}
}
