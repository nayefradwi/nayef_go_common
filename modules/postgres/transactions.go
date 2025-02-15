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
	defer tx.Rollback(ctx)
	if err := f(ctx, tx); err != nil {
		zap.L().Error("Error executing transaction", zap.Error(err))
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		zap.L().Error("Error committing transaction", zap.Error(err))
		return core.InternalError("Error committing transaction")
	}

	return nil
}
