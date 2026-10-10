package shared

import (
	"cmp"
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nayefradwi/nayef_go_common/auth/authpg"
	"github.com/nayefradwi/nayef_go_common/pgutil"
	"github.com/nayefradwi/nayef_go_common/redisutil"
	"github.com/redis/go-redis/v9"

	// keeps authredis in go.mod so the commented authredis swaps build without a go get
	_ "github.com/nayefradwi/nayef_go_common/auth/authredis"
)

const schema = `
CREATE TABLE IF NOT EXISTS example_migrations (name TEXT PRIMARY KEY);
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT true,
    role TEXT NOT NULL DEFAULT 'member'
);`

func Connect(ctx context.Context) *pgxpool.Pool {
	url := cmp.Or(os.Getenv("DATABASE_URL"), "postgres://postgres:postgres@localhost:5432/auth_examples?sslmode=disable")
	pool := pgutil.ConnectToPostgres(ctx, url)
	if err := migrate(ctx, pool); err != nil {
		panic(err)
	}

	return pool
}

// for the authredis store swaps
func Redis(ctx context.Context) *redis.Client {
	return redisutil.ConnectToRedis(ctx, cmp.Or(os.Getenv("REDIS_URL"), "redis://localhost:6379"))
}

// a real project copies authpg's goose files into its own migrations (ngo does);
// here a table remembers which ones ran so restarts are safe
func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, schema); err != nil {
		return err
	}

	files, err := fs.Glob(authpg.Migrations, "migrations/*.sql")
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := apply(ctx, pool, f); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
	}

	return nil
}

func apply(ctx context.Context, pool *pgxpool.Pool, file string) error {
	b, err := fs.ReadFile(authpg.Migrations, file)
	if err != nil {
		return err
	}

	up, _, _ := strings.Cut(string(b), "-- +goose Down")
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `INSERT INTO example_migrations VALUES ($1) ON CONFLICT DO NOTHING`, file)
		if err != nil || tag.RowsAffected() == 0 {
			return err
		}

		_, err = tx.Exec(ctx, up)
		return err
	})
}
