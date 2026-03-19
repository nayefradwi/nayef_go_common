package locking

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/redisutil"
	"github.com/redis/go-redis/v9"
)

type DistributedLocker struct {
	rs *redsync.Redsync
}

func NewDistributedLocker(rs *redsync.Redsync) ILocker {
	return &DistributedLocker{
		rs: rs,
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

	time.Sleep(params.InitialWaitTime)
	if err := mutex.Lock(); err != nil {
		return errors.BadRequestError("failed to acquire lock")
	}

	return nil
}

func (l *DistributedLocker) AcquireLocks(
	ctx context.Context,
	keys []string,
	params LockParams,
) error {
	for _, key := range keys {
		err := l.AcquireLock(ctx, key, params)
		if err != nil {
			l.ReleaseLocks(ctx, keys)
			return err
		}
	}

	return nil
}

func (l *DistributedLocker) ReleaseLock(ctx context.Context, key string) {
	mutex := l.rs.NewMutex(key)
	mutex.Unlock()
}

func (l *DistributedLocker) ReleaseLocks(ctx context.Context, keys []string) {
	for _, key := range keys {
		l.ReleaseLock(ctx, key)
	}
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

	defer l.ReleaseLock(ctx, key)
	return f()
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

	defer l.ReleaseLocks(ctx, keys)
	return f()
}
