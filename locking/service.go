package locking

import (
	"context"
	"log/slog"
	"sync"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/redisutil"
	"github.com/redis/go-redis/v9"
)

type DistributedLocker struct {
	rs    *redsync.Redsync
	mu    sync.Mutex
	locks map[string]*redsync.Mutex
}

func NewDistributedLocker(rs *redsync.Redsync) ILocker {
	return &DistributedLocker{
		rs:    rs,
		locks: make(map[string]*redsync.Mutex),
	}
}

func NewDistributedLockerFromClient(client *redis.Client) ILocker {
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)
	return NewDistributedLocker(rs)
}

func NewDistributedLockerFromConnection(ctx context.Context, connection string) ILocker {
	client := redisutil.ConnectToRedis(ctx, connection)
	return NewDistributedLockerFromClient(client)
}

func (l *DistributedLocker) AcquireLock(
	ctx context.Context,
	key string,
	params LockParams,
) error {
	mutex := l.rs.NewMutex(
		key,
		redsync.WithTries(params.MaxRetries),
		redsync.WithExpiry(params.TimeToLive),
		redsync.WithRetryDelay(params.WaitTime),
	)

	if err := mutex.LockContext(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to acquire lock", "key", key, "error", err)
		return errors.InternalError("failed to acquire lock")
	}

	l.mu.Lock()
	l.locks[key] = mutex
	l.mu.Unlock()

	return nil
}

func (l *DistributedLocker) AcquireLocks(
	ctx context.Context,
	keys []string,
	params LockParams,
) error {
	for i, key := range keys {
		err := l.AcquireLock(ctx, key, params)
		if err != nil {
			l.safeReleaseLocks(ctx, keys[:i])
			return err
		}
	}

	return nil
}

func (l *DistributedLocker) ReleaseLock(ctx context.Context, key string) error {
	l.mu.Lock()
	mutex, ok := l.locks[key]
	if !ok {
		l.mu.Unlock()
		slog.WarnContext(ctx, "attempted to release unknown lock", "key", key)
		return nil
	}
	delete(l.locks, key)
	l.mu.Unlock()

	if _, err := mutex.UnlockContext(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to release lock", "key", key, "error", err)
		return errors.InternalError("failed to release lock")
	}

	return nil
}

func (l *DistributedLocker) ReleaseLocks(ctx context.Context, keys []string) error {
	var firstErr error
	for _, key := range keys {
		if err := l.ReleaseLock(ctx, key); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (l *DistributedLocker) RunWithLock(
	ctx context.Context,
	key string,
	params LockParams,
	f func() error,
) error {
	err := l.AcquireLock(ctx, key, params)
	if err != nil {
		return err
	}

	defer l.safeReleaseLock(ctx, key)
	return f()
}

func (l *DistributedLocker) safeReleaseLock(ctx context.Context, key string) {
	if releaseErr := l.ReleaseLock(ctx, key); releaseErr != nil {
		slog.ErrorContext(ctx, "failed to release lock in RunWithLock", "key", key, "error", releaseErr)
	}
}

func (l *DistributedLocker) RunWithLocks(
	ctx context.Context,
	keys []string,
	params LockParams,
	f func() error,
) error {
	err := l.AcquireLocks(ctx, keys, params)
	if err != nil {
		return err
	}

	defer l.safeReleaseLocks(ctx, keys)
	return f()
}

func (l *DistributedLocker) safeReleaseLocks(ctx context.Context, keys []string) {
	if releaseErr := l.ReleaseLocks(ctx, keys); releaseErr != nil {
		slog.ErrorContext(ctx, "failed to release locks in RunWithLocks", "error", releaseErr)
	}
}
