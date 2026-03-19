package postgres

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectToPostgres(ctx context.Context, url string) *pgxpool.Pool {
	slog.Info("connecting to postgres")
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		slog.ErrorContext(ctx, "failed to connect to postgres", "error", err.Error())
		panic(err)
	}

	return pool
}
