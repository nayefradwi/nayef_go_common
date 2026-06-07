package locking

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryAcquireReleaseRoundTrip(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()

	err := locker.AcquireLock(ctx, "key", DefaultLockParams)
	require.NoError(t, err)

	err = locker.ReleaseLock(ctx, "key")
	require.NoError(t, err)
}

func TestInMemoryReleaseUnknownKey(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()

	err := locker.ReleaseLock(ctx, "never-acquired")
	assert.NoError(t, err)
}

func TestInMemoryDoubleRelease(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()

	require.NoError(t, locker.AcquireLock(ctx, "k", DefaultLockParams))
	require.NoError(t, locker.ReleaseLock(ctx, "k"))
	assert.NoError(t, locker.ReleaseLock(ctx, "k"))
}

func TestInMemoryReleaseLocksMixedKnownUnknown(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()

	require.NoError(t, locker.AcquireLock(ctx, "a", DefaultLockParams))
	require.NoError(t, locker.AcquireLock(ctx, "c", DefaultLockParams))

	require.NoError(t, locker.ReleaseLocks(ctx, []string{"a", "b-unknown", "c"}))

	// Both known keys should be reacquirable.
	assert.NoError(t, locker.AcquireLock(ctx, "a", DefaultLockParams))
	assert.NoError(t, locker.AcquireLock(ctx, "c", DefaultLockParams))
}

func TestInMemorySecondAcquireBlocksUntilRelease(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()
	key := "blocking-key"

	params := NewLockParams(0, 50*time.Millisecond, 200)
	require.NoError(t, locker.AcquireLock(ctx, key, params))

	acquired := make(chan struct{})
	go func() {
		err := locker.AcquireLock(ctx, key, params)
		assert.NoError(t, err)
		close(acquired)
	}()

	time.Sleep(150 * time.Millisecond)
	select {
	case <-acquired:
		t.Fatal("second goroutine acquired lock while first still held it")
	default:
	}

	require.NoError(t, locker.ReleaseLock(ctx, key))

	select {
	case <-acquired:
	case <-time.After(2 * time.Second):
		t.Fatal("second goroutine did not acquire lock after release")
	}
}

func TestInMemoryConcurrentCounter(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()
	key := "counter-key"
	n := 50

	params := NewLockParams(0, 20*time.Millisecond, 1000)

	var counter int64
	var wg sync.WaitGroup
	wg.Add(n)

	for range n {
		go func() {
			defer wg.Done()
			err := locker.RunWithLock(ctx, key, params, func() error {
				val := atomic.LoadInt64(&counter)
				time.Sleep(2 * time.Millisecond)
				atomic.StoreInt64(&counter, val+1)
				return nil
			})
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(n), counter)
}

func TestInMemoryChannelIdentityAcrossReleaseAcquire(t *testing.T) {
	// Regression for the channel-swap race: if ReleaseLock deleted the map
	// entry, a goroutine spinning in AcquireLock's retry loop could send on
	// the orphaned channel while a new caller obtained a fresh channel — two
	// holders simultaneously. We assert mutual exclusion via an in-critical-
	// section counter that must never exceed 1.
	locker := NewInMemoryLocker()
	ctx := context.Background()
	key := "identity-key"

	var inCritical int64
	var maxObserved int64
	updateMax := func() {
		cur := atomic.AddInt64(&inCritical, 1)
		for {
			prev := atomic.LoadInt64(&maxObserved)
			if cur <= prev || atomic.CompareAndSwapInt64(&maxObserved, prev, cur) {
				break
			}
		}
	}

	hold := func(d time.Duration) {
		updateMax()
		time.Sleep(d)
		atomic.AddInt64(&inCritical, -1)
	}

	retryParams := NewLockParams(0, 20*time.Millisecond, 500)

	require.NoError(t, locker.AcquireLock(ctx, key, retryParams))
	hold(0) // bump in/out so maxObserved is at least 1

	g2Done := make(chan struct{})
	go func() {
		err := locker.AcquireLock(ctx, key, retryParams)
		assert.NoError(t, err)
		hold(50 * time.Millisecond)
		require.NoError(t, locker.ReleaseLock(ctx, key))
		close(g2Done)
	}()

	// Give G2 time to enter the retry loop with the cached channel reference.
	time.Sleep(100 * time.Millisecond)
	require.NoError(t, locker.ReleaseLock(ctx, key))

	// G3 races G2 with a single fast attempt.
	go func() {
		singleShot := NewLockParams(0, 10*time.Millisecond, 200)
		err := locker.AcquireLock(ctx, key, singleShot)
		if err == nil {
			hold(50 * time.Millisecond)
			require.NoError(t, locker.ReleaseLock(ctx, key))
		}
	}()

	<-g2Done
	time.Sleep(200 * time.Millisecond)
	assert.LessOrEqual(t, atomic.LoadInt64(&maxObserved), int64(1),
		"two goroutines were inside the critical section at once — channel identity broke")
}

func TestInMemoryAcquireReturnsCtxErrWhenCancelled(t *testing.T) {
	locker := NewInMemoryLocker()
	key := "ctx-cancel-key"
	require.NoError(t, locker.AcquireLock(context.Background(), key, DefaultLockParams))

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- locker.AcquireLock(ctx, key, NewLockParams(0, 100*time.Millisecond, 100))
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("AcquireLock did not return after ctx cancellation")
	}
}

func TestInMemoryAcquirePreCancelledCtx(t *testing.T) {
	locker := NewInMemoryLocker()
	key := "pre-cancelled-key"
	require.NoError(t, locker.AcquireLock(context.Background(), key, DefaultLockParams))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := locker.AcquireLock(ctx, key, NewLockParams(0, 50*time.Millisecond, 5))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestInMemoryAcquireRespectsCtxDeadline(t *testing.T) {
	locker := NewInMemoryLocker()
	key := "ctx-deadline-key"
	require.NoError(t, locker.AcquireLock(context.Background(), key, DefaultLockParams))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := locker.AcquireLock(ctx, key, NewLockParams(0, 10*time.Millisecond, 1000))
	elapsed := time.Since(start)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, elapsed, 500*time.Millisecond, "did not exit promptly on ctx deadline")
}

func TestInMemoryAcquireFailsAfterMaxRetries(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()
	key := "retries-key"

	require.NoError(t, locker.AcquireLock(ctx, key, DefaultLockParams))

	err := locker.AcquireLock(ctx, key, NewLockParams(0, 20*time.Millisecond, 2))
	assert.Error(t, err)
	assert.NotErrorIs(t, err, context.Canceled)
	assert.NotErrorIs(t, err, context.DeadlineExceeded)
}

func TestInMemoryAcquireWaitTimeZeroDoesNotBusyLoop(t *testing.T) {
	locker := NewInMemoryLocker()
	key := "wait-zero-key"
	require.NoError(t, locker.AcquireLock(context.Background(), key, DefaultLockParams))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := locker.AcquireLock(ctx, key, NewLockParams(0, 0, 5))
	elapsed := time.Since(start)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.GreaterOrEqual(t, elapsed, 80*time.Millisecond,
		"returned too quickly — looks like a busy loop instead of waiting on ctx")
}

func TestInMemoryAcquireMaxRetriesZeroSingleAttempt(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()
	key := "max-zero-key"
	require.NoError(t, locker.AcquireLock(ctx, key, DefaultLockParams))

	start := time.Now()
	err := locker.AcquireLock(ctx, key, NewLockParams(0, 30*time.Millisecond, 0))
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.GreaterOrEqual(t, elapsed, 25*time.Millisecond, "did not run a wait cycle")
	assert.Less(t, elapsed, 200*time.Millisecond, "ran more than one wait cycle")
}

func TestInMemoryTTLExpiryReleasesLock(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()
	key := "ttl-key"

	require.NoError(t, locker.AcquireLock(ctx, key, NewLockParams(50*time.Millisecond, 10*time.Millisecond, 1)))
	time.Sleep(150 * time.Millisecond)

	err := locker.AcquireLock(ctx, key, DefaultLockParams)
	assert.NoError(t, err)
}

func TestInMemoryManualReleaseBeforeTTL(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()
	key := "manual-before-ttl-key"

	require.NoError(t, locker.AcquireLock(ctx, key, NewLockParams(500*time.Millisecond, 10*time.Millisecond, 1)))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, locker.ReleaseLock(ctx, key))

	require.NoError(t, locker.AcquireLock(ctx, key, DefaultLockParams))

	// Let the lingering TTL goroutine fire; its drain should be a harmless no-op.
	time.Sleep(600 * time.Millisecond)

	// Still releasable, no panic.
	assert.NoError(t, locker.ReleaseLock(ctx, key))
}

func TestInMemoryTTLZeroMeansNoExpiry(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()
	key := "ttl-zero-key"

	require.NoError(t, locker.AcquireLock(ctx, key, NewLockParams(0, 10*time.Millisecond, 1)))
	time.Sleep(100 * time.Millisecond)

	err := locker.AcquireLock(ctx, key, NewLockParams(0, 20*time.Millisecond, 2))
	assert.Error(t, err)
}

func TestInMemoryAcquireLocksAllOrNothing(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()

	require.NoError(t, locker.AcquireLock(ctx, "B", DefaultLockParams))

	err := locker.AcquireLocks(ctx, []string{"A", "B"}, NewLockParams(0, 20*time.Millisecond, 2))
	assert.Error(t, err)

	// A must have been rolled back.
	require.NoError(t, locker.ReleaseLock(ctx, "A"))
	assert.NoError(t, locker.AcquireLock(ctx, "A", DefaultLockParams))
}

func TestInMemoryRunWithLocksReleasesEverythingOnSuccess(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()
	keys := []string{"rwl-1", "rwl-2", "rwl-3"}

	executed := false
	err := locker.RunWithLocks(ctx, keys, DefaultLockParams, func() error {
		executed = true
		return nil
	})
	require.NoError(t, err)
	assert.True(t, executed)

	for _, k := range keys {
		assert.NoError(t, locker.AcquireLock(ctx, k, DefaultLockParams), "key %s not reacquirable", k)
	}
}

func TestInMemoryRunWithLocksReleasesEverythingOnError(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()
	keys := []string{"rwl-err-1", "rwl-err-2"}

	err := locker.RunWithLocks(ctx, keys, DefaultLockParams, func() error {
		return assert.AnError
	})
	assert.ErrorIs(t, err, assert.AnError)

	for _, k := range keys {
		assert.NoError(t, locker.AcquireLock(ctx, k, DefaultLockParams), "key %s not reacquirable", k)
	}
}

func TestInMemoryStressManyGoroutinesManyKeys(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()
	const goroutines = 100
	const iterations = 50
	keys := []string{"s-0", "s-1", "s-2", "s-3", "s-4"}

	params := NewLockParams(0, 5*time.Millisecond, 2000)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for j := range iterations {
				key := keys[(id+j)%len(keys)]
				err := locker.RunWithLock(ctx, key, params, func() error { return nil })
				assert.NoError(t, err)
			}
		}(g)
	}
	wg.Wait()
}

// RED: TTL goroutine drains whatever token is currently in the channel, not
// the one its original holder put there. After G1 releases manually, G2
// acquires, then G1's stale TTL fires and frees G2's lock — letting G3 in
// while G2 still believes it holds the lock.
func TestInMemoryTTLDoesNotReleaseSubsequentHoldersLock(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()
	key := "stale-ttl-key"

	require.NoError(t, locker.AcquireLock(ctx, key, NewLockParams(100*time.Millisecond, 10*time.Millisecond, 1)))
	require.NoError(t, locker.ReleaseLock(ctx, key))

	require.NoError(t, locker.AcquireLock(ctx, key, NewLockParams(0, 10*time.Millisecond, 1)))
	defer locker.ReleaseLock(ctx, key)

	time.Sleep(200 * time.Millisecond)

	err := locker.AcquireLock(ctx, key, NewLockParams(0, 20*time.Millisecond, 1))
	assert.Error(t, err, "third acquirer got the lock — stale TTL from first holder released the second holder's lock")
}

// RED: TTL release is tied to the acquirer's ctx. If the caller cancels its
// ctx after acquiring, releaseAfter exits on <-ctx.Done() before the timer
// fires and the lock is never freed — defeating the point of TTL.
func TestInMemoryTTLFiresEvenWhenAcquirerCtxCancelled(t *testing.T) {
	locker := NewInMemoryLocker()
	key := "ttl-ctx-key"

	ctx, cancel := context.WithCancel(context.Background())
	require.NoError(t, locker.AcquireLock(ctx, key, NewLockParams(50*time.Millisecond, 10*time.Millisecond, 1)))
	cancel()

	time.Sleep(200 * time.Millisecond)

	err := locker.AcquireLock(context.Background(), key, NewLockParams(0, 10*time.Millisecond, 1))
	assert.NoError(t, err, "TTL did not release the lock after the acquirer's ctx was cancelled")
}

// RED: makeCh inserts into the map but nothing ever deletes. A workload with
// many unique keys grows the map without bound.
func TestInMemoryDoesNotLeakChannelsForUniqueKeys(t *testing.T) {
	locker := NewInMemoryLocker().(*InMemoryLocker)
	ctx := context.Background()

	const n = 1000
	for i := range n {
		key := fmt.Sprintf("leak-%d", i)
		require.NoError(t, locker.AcquireLock(ctx, key, DefaultLockParams))
		require.NoError(t, locker.ReleaseLock(ctx, key))
	}

	locker.mu.Lock()
	size := len(locker.locks)
	locker.mu.Unlock()

	assert.Less(t, size, n/10, "map retained %d entries after acquiring+releasing %d unique keys", size, n)
}

func TestInMemoryStressAcquireReleaseWithTTL(t *testing.T) {
	locker := NewInMemoryLocker()
	ctx := context.Background()
	const goroutines = 50
	const iterations = 30
	keys := []string{"st-0", "st-1", "st-2"}

	params := NewLockParams(20*time.Millisecond, 5*time.Millisecond, 2000)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for j := range iterations {
				key := keys[(id+j)%len(keys)]
				err := locker.RunWithLock(ctx, key, params, func() error { return nil })
				assert.NoError(t, err)
			}
		}(g)
	}
	wg.Wait()
}
