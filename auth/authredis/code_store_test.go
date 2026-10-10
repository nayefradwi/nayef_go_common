package authredis

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeStore(t *testing.T) {
	client := mustCreateRedisClient(t)
	store, err := NewCodeStore(client)
	require.NoError(t, err)
	ctx := context.Background()
	limits := auth.CodeLimits{TTL: time.Minute, ResendAfter: time.Minute, MaxSends: 2, MaxAttempts: 5}
	hash := []byte("hash")

	sentAgo := func(t *testing.T, key string, ago time.Duration) {
		t.Helper()
		sent := time.Now().Add(-ago).UnixMilli()
		require.NoError(t, client.HSet(ctx, defaultCodePrefix+key, "sent_at", sent).Err())
	}
	expire := func(t *testing.T, key string) {
		t.Helper()
		require.NoError(t, client.PExpire(ctx, defaultCodePrefix+key, time.Millisecond).Err())
		time.Sleep(5 * time.Millisecond)
	}
	attempts := func(t *testing.T, key string) int {
		t.Helper()
		n, err := client.HGet(ctx, defaultCodePrefix+key, "attempts").Int()
		require.NoError(t, err)
		return n
	}

	t.Run("save and resend set the ttl", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "ttl", hash, limits))
		ttl, err := client.PTTL(ctx, defaultCodePrefix+"ttl").Result()
		require.NoError(t, err)
		assert.Positive(t, ttl)

		require.NoError(t, client.PExpire(ctx, defaultCodePrefix+"ttl", time.Second).Err())
		sentAgo(t, "ttl", 2*time.Minute)
		require.NoError(t, store.Save(ctx, "ttl", hash, limits))
		ttl, err = client.PTTL(ctx, defaultCodePrefix+"ttl").Result()
		require.NoError(t, err)
		assert.Greater(t, ttl, time.Second)
	})

	t.Run("concurrent attempts stop at max", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "attempts", hash, limits))
		var ok atomic.Int32
		var wg sync.WaitGroup
		for range 20 {
			wg.Go(func() {
				_, err := store.Attempt(ctx, "attempts", limits.MaxAttempts)
				if err == nil {
					ok.Add(1)
					return
				}
				assert.ErrorIs(t, err, auth.ErrCodeNotFound)
			})
		}
		wg.Wait()
		assert.EqualValues(t, limits.MaxAttempts, ok.Load())
	})

	t.Run("concurrent consume succeeds once", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "consume", hash, limits))
		var ok atomic.Int32
		var wg sync.WaitGroup
		for range 20 {
			wg.Go(func() {
				if store.Consume(ctx, "consume", hash) == nil {
					ok.Add(1)
				}
			})
		}
		wg.Wait()
		assert.EqualValues(t, 1, ok.Load())
	})

	t.Run("consume with a replaced hash fails", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "replaced", []byte("new"), limits))
		require.ErrorIs(t, store.Consume(ctx, "replaced", hash), auth.ErrCodeNotFound)
	})

	t.Run("resend too soon is blocked", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "soon", hash, limits))
		require.ErrorIs(t, store.Save(ctx, "soon", hash, limits), auth.ErrCodeResendBlocked)
	})

	t.Run("resend keeps attempts and stops at the send cap", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "cap", hash, limits))
		_, err := store.Attempt(ctx, "cap", limits.MaxAttempts)
		require.NoError(t, err)

		sentAgo(t, "cap", 2*time.Minute)
		require.NoError(t, store.Save(ctx, "cap", hash, limits))
		assert.Equal(t, 1, attempts(t, "cap"))

		sentAgo(t, "cap", 2*time.Minute)
		require.ErrorIs(t, store.Save(ctx, "cap", hash, limits), auth.ErrCodeResendBlocked)
	})

	t.Run("resend while locked is blocked", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "locked", hash, limits))
		for range limits.MaxAttempts {
			_, err := store.Attempt(ctx, "locked", limits.MaxAttempts)
			require.NoError(t, err)
		}
		sentAgo(t, "locked", 2*time.Minute)
		require.ErrorIs(t, store.Save(ctx, "locked", hash, limits), auth.ErrCodeResendBlocked)
	})

	t.Run("expired key starts over", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "expired", hash, limits))
		_, err := store.Attempt(ctx, "expired", limits.MaxAttempts)
		require.NoError(t, err)
		expire(t, "expired")

		_, err = store.Attempt(ctx, "expired", limits.MaxAttempts)
		require.ErrorIs(t, err, auth.ErrCodeNotFound)

		require.NoError(t, store.Save(ctx, "expired", hash, limits))
		assert.Equal(t, 0, attempts(t, "expired"))
	})

	t.Run("consume rejects an expired key", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "stale", hash, limits))
		expire(t, "stale")

		require.ErrorIs(t, store.Consume(ctx, "stale", hash), auth.ErrCodeNotFound)
	})
}

func TestCodeStore_CustomPrefix(t *testing.T) {
	client := mustCreateRedisClient(t)
	ctx := context.Background()

	_, err := NewCodeStore(client, CodeStoreConfig{Prefix: "Bad Prefix"})
	require.Error(t, err)

	admin, err := NewCodeStore(client, CodeStoreConfig{Prefix: "admin:otps:"})
	require.NoError(t, err)
	m, err := auth.NewOtpManager(admin, auth.OtpConfig{
		Secret:     []byte(strings.Repeat("s", 32)),
		Length:     6,
		CodeLimits: auth.CodeLimits{TTL: time.Minute, MaxSends: 1, MaxAttempts: 3},
	})
	require.NoError(t, err)

	code, err := m.Issue(ctx, "login:a")
	require.NoError(t, err)
	_, err = m.Issue(ctx, "login:a")
	require.ErrorIs(t, err, auth.ErrCodeResendBlocked)

	defaultStore, err := NewCodeStore(client)
	require.NoError(t, err)
	_, err = defaultStore.Attempt(ctx, "login:a", 3)
	require.ErrorIs(t, err, auth.ErrCodeNotFound)

	require.NoError(t, m.Verify(ctx, "login:a", code))
	require.Error(t, m.Verify(ctx, "login:a", code))
}
