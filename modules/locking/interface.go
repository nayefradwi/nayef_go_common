package locking

import "context"

type ILockingService interface {
	AquireLock(ctx context.Context, key string, params LockParams)
	ReleaseLock(ctx context.Context, key string)
	AcquireLocks(ctx context.Context, keys []string, params LockParams)
	ReleaseLocks(ctx context.Context, keys []string)
	RunWithLock(ctx context.Context, key string, params LockParams, f func() error)
	RunWithLocks(ctx context.Context, keys []string, params LockParams, f func() error)
}
