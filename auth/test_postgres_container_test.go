package auth

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func mustCreatePostgresConn(t *testing.T) *pgx.Conn {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:15.3-alpine",
		postgres.WithInitScripts("init-db.sql"),
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

	conn, err := pgx.Connect(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close(ctx) })

	return conn
}
