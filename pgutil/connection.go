package pgutil

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectToPostgres(ctx context.Context, url string) *pgxpool.Pool {
	slog.Info("connecting to postgres")
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create pgx pool", "error", err.Error())
		panic(err)
	}

	if err := pool.Ping(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to connect to postgres database", "error", err.Error())
		panic(err)
	}

	return pool
}
