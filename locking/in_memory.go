package locking

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/nayefradwi/nayef_go_common/errors"
)

type InMemoryLocker struct {
	mu    sync.Mutex
	locks map[string]*lockEntry
}

func NewInMemoryLocker() ILocker {
	return &InMemoryLocker{
		locks: map[string]*lockEntry{},
	}
}

func (i *InMemoryLocker) refEntry(key string) *lockEntry {
	i.mu.Lock()
	defer i.mu.Unlock()
	entry, ok := i.locks[key]
	if !ok {
		entry = &lockEntry{ch: make(chan struct{}, 1)}
		i.locks[key] = entry
	}
	entry.waiters++
	return entry
}

func (i *InMemoryLocker) unrefEntry(key string, entry *lockEntry) {
	i.mu.Lock()
	defer i.mu.Unlock()
	entry.waiters--
	i.maybeDeleteLocked(key, entry)
}

func (i *InMemoryLocker) maybeDeleteLocked(key string, entry *lockEntry) {
	if entry.waiters == 0 && len(entry.ch) == 0 {
		if current, ok := i.locks[key]; ok && current == entry {
			delete(i.locks, key)
		}
	}
}

func (i *InMemoryLocker) releaseIfToken(key string, entry *lockEntry, token uint64) {
	i.mu.Lock()
	defer i.mu.Unlock()
	entry.releaseIfToken(token)
	i.maybeDeleteLocked(key, entry)
}

func (i *InMemoryLocker) scheduleTTL(key string, entry *lockEntry, ttl time.Duration) {
	i.mu.Lock()
	defer i.mu.Unlock()
	entry.token++
	myToken := entry.token
	if ttl <= 0 {
		return
	}
	entry.timer = time.AfterFunc(ttl, func() {
		i.releaseIfToken(key, entry, myToken)
	})
}

func (i *InMemoryLocker) AcquireLock(ctx context.Context, key string, params LockParams) error {
	entry := i.refEntry(key)
	defer i.unrefEntry(key, entry)

	timer := newRetryTimer(params.WaitTime)
	defer timer.stop()

	for attempt := 0; attempt <= params.MaxRetries; attempt++ {
		select {
		case entry.ch <- struct{}{}:
			i.scheduleTTL(key, entry, params.TimeToLive)
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.arm():
		}
	}

	slog.ErrorContext(ctx, "failed to acquire lock", "key", key)
	return errors.InternalError("failed to acquire lock")
}

func (i *InMemoryLocker) AcquireLocks(ctx context.Context, keys []string, params LockParams) error {
	for idx, key := range keys {
		if err := i.AcquireLock(ctx, key, params); err != nil {
			i.ReleaseLocks(ctx, keys[:idx])
			return err
		}
	}

	return nil
}

func (i *InMemoryLocker) ReleaseLock(ctx context.Context, key string) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	entry, ok := i.locks[key]
	if !ok {
		slog.WarnContext(ctx, "attempted to release unknown lock", "key", key)
		return nil
	}

	entry.releaseToken()
	i.maybeDeleteLocked(key, entry)
	return nil
}

func (i *InMemoryLocker) ReleaseLocks(ctx context.Context, keys []string) error {
	for _, key := range keys {
		i.ReleaseLock(ctx, key)
	}
	return nil
}

func (i *InMemoryLocker) RunWithLock(ctx context.Context, key string, params LockParams, f func() error) error {
	if err := i.AcquireLock(ctx, key, params); err != nil {
		return err
	}

	defer i.ReleaseLock(ctx, key)
	return f()
}

func (i *InMemoryLocker) RunWithLocks(ctx context.Context, keys []string, params LockParams, f func() error) error {
	if err := i.AcquireLocks(ctx, keys, params); err != nil {
		return err
	}

	defer i.ReleaseLocks(ctx, keys)
	return f()
}
