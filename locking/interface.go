package locking

import "context"

type ILocker interface {
	AcquireLock(ctx context.Context, key string, params LockParams) error
	ReleaseLock(ctx context.Context, key string) error
	AcquireLocks(ctx context.Context, keys []string, params LockParams) error
	ReleaseLocks(ctx context.Context, keys []string) error
	RunWithLock(ctx context.Context, key string, params LockParams, f func() error) error
	RunWithLocks(ctx context.Context, keys []string, params LockParams, f func() error) error
}
