package authpg

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
	pool := mustCreatePostgresConn(t)
	store, err := NewCodeStore(pool)
	require.NoError(t, err)
	ctx := context.Background()
	limits := auth.CodeLimits{TTL: time.Minute, ResendAfter: time.Minute, MaxSends: 2, MaxAttempts: 5}
	hash := []byte("hash")

	age := func(t *testing.T, key string, sent, expires time.Duration) {
		t.Helper()
		_, err := pool.Exec(ctx, `UPDATE auth_otps SET sent_at = now() - $2::interval, expires_at = now() + $3::interval WHERE key = $1`, key, sent, expires)
		require.NoError(t, err)
	}
	attempts := func(t *testing.T, key string) int {
		t.Helper()
		var n int
		require.NoError(t, pool.QueryRow(ctx, `SELECT attempts FROM auth_otps WHERE key = $1`, key).Scan(&n))
		return n
	}

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

		age(t, "cap", 2*time.Minute, time.Minute)
		require.NoError(t, store.Save(ctx, "cap", hash, limits))
		assert.Equal(t, 1, attempts(t, "cap"))

		age(t, "cap", 2*time.Minute, time.Minute)
		require.ErrorIs(t, store.Save(ctx, "cap", hash, limits), auth.ErrCodeResendBlocked)
	})

	t.Run("resend while locked is blocked", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "locked", hash, limits))
		for range limits.MaxAttempts {
			_, err := store.Attempt(ctx, "locked", limits.MaxAttempts)
			require.NoError(t, err)
		}
		age(t, "locked", 2*time.Minute, time.Minute)
		require.ErrorIs(t, store.Save(ctx, "locked", hash, limits), auth.ErrCodeResendBlocked)
	})

	t.Run("expired row starts over", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "expired", hash, limits))
		_, err := store.Attempt(ctx, "expired", limits.MaxAttempts)
		require.NoError(t, err)
		age(t, "expired", 0, -time.Second)

		_, err = store.Attempt(ctx, "expired", limits.MaxAttempts)
		require.ErrorIs(t, err, auth.ErrCodeNotFound)

		require.NoError(t, store.Save(ctx, "expired", hash, limits))
		assert.Equal(t, 0, attempts(t, "expired"))
	})

	t.Run("consume rejects an expired row", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "stale", hash, limits))
		age(t, "stale", 0, -time.Second)

		require.ErrorIs(t, store.Consume(ctx, "stale", hash), auth.ErrCodeNotFound)
	})

	t.Run("delete expired keeps live rows", func(t *testing.T) {
		require.NoError(t, store.Save(ctx, "live", hash, limits))
		require.NoError(t, store.Save(ctx, "dead", hash, limits))
		age(t, "dead", 0, -time.Second)

		require.NoError(t, store.DeleteExpired(ctx))
		_, err := store.Attempt(ctx, "live", limits.MaxAttempts)
		require.NoError(t, err)
		var n int
		require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM auth_otps WHERE key = 'dead'`).Scan(&n))
		assert.Zero(t, n)
	})
}

func TestCodeStore_CustomTable(t *testing.T) {
	pool := mustCreatePostgresConn(t)
	ctx := context.Background()

	sql, err := CodeMigration("admin_otps")
	require.NoError(t, err)
	up, _, found := strings.Cut(sql, "-- +goose Down")
	require.True(t, found)
	_, err = pool.Exec(ctx, up)
	require.NoError(t, err)

	admin, err := NewCodeStore(pool, CodeStoreConfig{Table: "admin_otps"})
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

	defaultStore, err := NewCodeStore(pool)
	require.NoError(t, err)
	_, err = defaultStore.Attempt(ctx, "login:a", 3)
	require.ErrorIs(t, err, auth.ErrCodeNotFound)

	require.NoError(t, m.Verify(ctx, "login:a", code))
	require.Error(t, m.Verify(ctx, "login:a", code))
	require.NoError(t, m.DeleteExpired(ctx))
}
