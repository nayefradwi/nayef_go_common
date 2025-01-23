package locking

import "context"

type ILocker interface {
	AquireLock(ctx context.Context, key string, params LockParams) error
	ReleaseLock(ctx context.Context, key string)
	AcquireLocks(ctx context.Context, keys []string, params LockParams) error
	ReleaseLocks(ctx context.Context, keys []string)
	RunWithLock(ctx context.Context, key string, params LockParams, f func() error) error
	RunWithLocks(ctx context.Context, keys []string, params LockParams, f func() error) error
}
