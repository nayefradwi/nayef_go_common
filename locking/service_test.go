package locking

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcquireAndReleaseLock(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()

	err := locker.AcquireLock(ctx, "test-key", DefaultLockParams)
	require.NoError(t, err)

	err = locker.ReleaseLock(ctx, "test-key")
	require.NoError(t, err)
}

func TestReleaseLockUnknownKey(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()

	err := locker.ReleaseLock(ctx, "never-acquired")
	assert.NoError(t, err)
}

func TestDoubleRelease(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()

	err := locker.AcquireLock(ctx, "double-key", DefaultLockParams)
	require.NoError(t, err)

	err = locker.ReleaseLock(ctx, "double-key")
	require.NoError(t, err)

	err = locker.ReleaseLock(ctx, "double-key")
	assert.NoError(t, err)
}

func TestAcquireAndReleaseLocks(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()
	keys := []string{"multi-1", "multi-2", "multi-3"}

	err := locker.AcquireLocks(ctx, keys, DefaultLockParams)
	require.NoError(t, err)

	err = locker.ReleaseLocks(ctx, keys)
	require.NoError(t, err)
}

func TestLockContention(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()
	key := "contention-key"

	params := NewLockParams(5*time.Second, 100*time.Millisecond, 50)

	err := locker.AcquireLock(ctx, key, params)
	require.NoError(t, err)

	acquired := make(chan struct{})
	go func() {
		err := locker.AcquireLock(ctx, key, params)
		assert.NoError(t, err)
		close(acquired)
	}()

	// Give the goroutine time to start waiting
	time.Sleep(200 * time.Millisecond)

	select {
	case <-acquired:
		t.Fatal("second goroutine should not have acquired lock yet")
	default:
	}

	err = locker.ReleaseLock(ctx, key)
	require.NoError(t, err)

	select {
	case <-acquired:
	case <-time.After(5 * time.Second):
		t.Fatal("second goroutine should have acquired lock after release")
	}
}

func TestLockTimeout(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()
	key := "timeout-key"

	holdParams := NewLockParams(10*time.Second, 100*time.Millisecond, 1)
	err := locker.AcquireLock(ctx, key, holdParams)
	require.NoError(t, err)
	defer locker.ReleaseLock(ctx, key)

	failParams := NewLockParams(10*time.Second, 50*time.Millisecond, 2)
	err = locker.AcquireLock(ctx, key, failParams)
	assert.Error(t, err)
}

func TestLockWaitTime(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()
	key := "wait-key"

	params := NewLockParams(2*time.Second, 100*time.Millisecond, 30)
	err := locker.AcquireLock(ctx, key, params)
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		done <- locker.AcquireLock(ctx, key, params)
	}()

	time.Sleep(500 * time.Millisecond)
	err = locker.ReleaseLock(ctx, key)
	require.NoError(t, err)

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("second acquirer should have succeeded after release")
	}
}

func TestConcurrentLockers(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()
	key := "counter-key"
	n := 20

	params := NewLockParams(5*time.Second, 100*time.Millisecond, 100)

	var counter int64
	var wg sync.WaitGroup
	wg.Add(n)

	for range n {
		go func() {
			defer wg.Done()
			err := locker.RunWithLock(ctx, key, params, func() error {
				val := atomic.LoadInt64(&counter)
				time.Sleep(10 * time.Millisecond)
				atomic.StoreInt64(&counter, val+1)
				return nil
			})
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(n), counter)
}

func TestAcquireLocksPartialRollback(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()

	heldKey := "held-key"
	otherKey := "other-key"

	holdParams := NewLockParams(10*time.Second, 100*time.Millisecond, 1)
	err := locker.AcquireLock(ctx, heldKey, holdParams)
	require.NoError(t, err)

	failParams := NewLockParams(10*time.Second, 50*time.Millisecond, 2)
	err = locker.AcquireLocks(ctx, []string{otherKey, heldKey}, failParams)
	assert.Error(t, err)

	// otherKey should have been rolled back and be reacquirable
	err = locker.AcquireLock(ctx, otherKey, DefaultLockParams)
	assert.NoError(t, err)
}

func TestLockTTLExpiry(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()
	key := "ttl-key"

	shortTTL := NewLockParams(500*time.Millisecond, 100*time.Millisecond, 1)
	err := locker.AcquireLock(ctx, key, shortTTL)
	require.NoError(t, err)

	// Wait for TTL to expire
	time.Sleep(700 * time.Millisecond)

	// Should be able to acquire without explicit release
	err = locker.AcquireLock(ctx, key, DefaultLockParams)
	assert.NoError(t, err)
}

func TestContextCancellation(t *testing.T) {
	locker := setupLocker(t)
	key := "ctx-cancel-key"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := locker.AcquireLock(ctx, key, DefaultLockParams)
	assert.Error(t, err)
}

func TestRunWithLock(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()
	key := "run-key"

	executed := false
	err := locker.RunWithLock(ctx, key, DefaultLockParams, func() error {
		executed = true
		return nil
	})
	require.NoError(t, err)
	assert.True(t, executed)

	// Lock should be released — verify by reacquiring
	err = locker.AcquireLock(ctx, key, DefaultLockParams)
	assert.NoError(t, err)
}

func TestRunWithLockError(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()
	key := "run-err-key"

	expectedErr := assert.AnError
	err := locker.RunWithLock(ctx, key, DefaultLockParams, func() error {
		return expectedErr
	})
	assert.ErrorIs(t, err, expectedErr)

	// Lock should still be released
	err = locker.AcquireLock(ctx, key, DefaultLockParams)
	assert.NoError(t, err)
}

func TestRunWithLocks(t *testing.T) {
	locker := setupLocker(t)
	ctx := context.Background()
	keys := []string{"run-multi-1", "run-multi-2"}

	executed := false
	err := locker.RunWithLocks(ctx, keys, DefaultLockParams, func() error {
		executed = true
		return nil
	})
	require.NoError(t, err)
	assert.True(t, executed)

	// All locks should be released
	err = locker.AcquireLocks(ctx, keys, DefaultLockParams)
	assert.NoError(t, err)
}
