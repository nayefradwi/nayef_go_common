package authredis

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttemptStore(t *testing.T) {
	store := NewAttemptStore(mustCreateRedisClient(t))
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

	t.Run("later hits do not extend the window", func(t *testing.T) {
		_, err := store.Hit(ctx, "window", 200*time.Millisecond)
		require.NoError(t, err)
		time.Sleep(120 * time.Millisecond)
		a, err := store.Hit(ctx, "window", 200*time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, 2, a.Count)

		time.Sleep(120 * time.Millisecond)
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
}
