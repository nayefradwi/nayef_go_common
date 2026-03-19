package otp

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func mustCreateRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	opts, err := redis.ParseURL(connStr)
	require.NoError(t, err)

	client := redis.NewClient(opts)
	t.Cleanup(func() { client.Close() })

	return client
}

func setupRepository(t *testing.T) IOtpRepository {
	t.Helper()
	client := mustCreateRedisClient(t)
	return NewRedisOtpRepository(client)
}
