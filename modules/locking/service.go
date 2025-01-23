package locking

import (
	"context"
	"sync"
	"time"

	"github.com/nayefradwi/nayef_go_common/core"
)

type InMemoryLocker struct {
	locks map[string]*sync.Mutex
}

func NewInMemoryLocker() *InMemoryLocker {
	return &InMemoryLocker{
		locks: make(map[string]*sync.Mutex),
	}
}

func (l *InMemoryLocker) AquireLock(ctx context.Context, key string, params LockParams) error {
	lock, ok := l.locks[key]
	if !ok {
		lock = &sync.Mutex{}
		l.locks[key] = lock
	}

	return l.tryToAquireLock(ctx, lock, params)
}

func (l *InMemoryLocker) AcquireLocks(ctx context.Context, keys []string, params LockParams) error {
	for _, key := range keys {
		err := l.AquireLock(ctx, key, params)
		if err != nil {
			l.ReleaseLocks(ctx, keys)
			return err
		}
	}

	return nil
}

func (l *InMemoryLocker) ReleaseLock(ctx context.Context, key string) {
	lock, ok := l.locks[key]
	if !ok {
		return
	}

	lock.Unlock()
}

func (l *InMemoryLocker) ReleaseLocks(ctx context.Context, keys []string) {
	for _, key := range keys {
		l.ReleaseLock(ctx, key)
	}
}

func (l *InMemoryLocker) RunWithLock(ctx context.Context, key string, params LockParams, f func() error) error {
	err := l.AquireLock(ctx, key, params)
	if err != nil {
		return err
	}

	defer l.ReleaseLock(ctx, key)
	return f()
}

func (l *InMemoryLocker) RunWithLocks(ctx context.Context, keys []string, params LockParams, f func() error) error {
	err := l.AcquireLocks(ctx, keys, params)
	if err != nil {
		return err
	}

	defer l.ReleaseLocks(ctx, keys)
	return f()
}

func (l *InMemoryLocker) tryToAquireLock(context context.Context, lock *sync.Mutex, params LockParams) error {
	time.Sleep(params.InitialWaitTime)
	defer l.releaseAfter(lock, params.TimeToLive)

	for i := 0; i < params.MaxRetries; i++ {
		select {
		case <-context.Done():
			return core.BadRequestError("Context cancelled")
		default:
			if lock.TryLock() {
				return nil
			}
			time.Sleep(params.WaitTime)
		}
	}

	return core.BadRequestError("Failed to aquire lock")
}

func (l *InMemoryLocker) releaseAfter(lock *sync.Mutex, timeToLive time.Duration) {
	go func() {
		time.Sleep(timeToLive)
		lock.Unlock()
	}()
}
