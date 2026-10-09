package authpg

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttemptStore(t *testing.T) {
	pool := mustCreatePostgresConn(t)
	store := NewAttemptStore(pool)
	ctx := context.Background()

	t.Run("concurrent hits get unique counts", func(t *testing.T) {
		const n = 50
		counts := make(chan int, n)
		var wg sync.WaitGroup
		for range n {
			wg.Go(func() {
				a, err := store.Hit(ctx, "concurrent", time.Minute)
				assert.NoError(t, err)
				counts <- a.Count
			})
		}
		wg.Wait()
		close(counts)

		seen := map[int]bool{}
		for c := range counts {
			assert.False(t, seen[c], "count %d returned twice", c)
			seen[c] = true
		}
		for i := 1; i <= n; i++ {
			assert.True(t, seen[i], "count %d missing", i)
		}
	})

	t.Run("window expiry restarts the count", func(t *testing.T) {
		_, err := store.Hit(ctx, "window", 200*time.Millisecond)
		require.NoError(t, err)
		a, err := store.Hit(ctx, "window", 200*time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, 2, a.Count)

		time.Sleep(300 * time.Millisecond)
		a, err = store.Hit(ctx, "window", 200*time.Millisecond)
		require.NoError(t, err)
		assert.Equal(t, 1, a.Count)
		assert.True(t, a.ResetAt.After(time.Now()))
	})

	t.Run("reset clears the key", func(t *testing.T) {
		_, err := store.Hit(ctx, "reset", time.Minute)
		require.NoError(t, err)
		require.NoError(t, store.Reset(ctx, "reset"))
		a, err := store.Hit(ctx, "reset", time.Minute)
		require.NoError(t, err)
		assert.Equal(t, 1, a.Count)
	})

	t.Run("delete expired keeps live rows", func(t *testing.T) {
		_, err := store.Hit(ctx, "expired", time.Millisecond)
		require.NoError(t, err)
		_, err = store.Hit(ctx, "live", time.Hour)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)

		require.NoError(t, store.DeleteExpired(ctx))

		var keys []string
		rows, err := pool.Query(ctx, `SELECT key FROM auth_attempts WHERE key IN ('expired', 'live')`)
		require.NoError(t, err)
		for rows.Next() {
			var k string
			require.NoError(t, rows.Scan(&k))
			keys = append(keys, k)
		}
		require.NoError(t, rows.Err())
		assert.Equal(t, []string{"live"}, keys)
	})
}
