package authpg

import (
	"context"
	"io/fs"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func mustCreatePostgresConn(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:15.3-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })

	applyMigrations(t, pool)
	return pool
}

// runs the shipped files so a broken migration fails the tests
func applyMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	files, err := fs.Glob(Migrations, "migrations/*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, files)

	for _, f := range files {
		b, err := fs.ReadFile(Migrations, f)
		require.NoError(t, err)
		up, _, found := strings.Cut(string(b), "-- +goose Down")
		require.True(t, found, f+" has no goose Down section")
		_, err = pool.Exec(context.Background(), up)
		require.NoError(t, err, f)
	}
}
