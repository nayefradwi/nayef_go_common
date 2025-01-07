package common

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	DefaultExpiry  = 1 * time.Minute
	DefaultTimeout = 5 * time.Second
)

type IDistributedLockingService interface {
	Acquire(ctx context.Context, name string) (Lock, error)
	CustomeDurationAcquire(ctx context.Context, name string, expiresAt, timeout time.Duration) (Lock, error)
	Release(ctx context.Context, lock Lock) error
	ReleaseMany(ctx context.Context, locks *[]Lock)
	RunWithLock(ctx context.Context, name string, f func() error) error
}

type Lock struct {
	Name      string
	ExpiresAt time.Duration
}

type RedisLockService struct {
	client *redis.Client
}

func CreateNewRedisLockService(client *redis.Client) IDistributedLockingService {
	return &RedisLockService{
		client,
	}
}

func (s *RedisLockService) Acquire(ctx context.Context, name string) (Lock, error) {
	return s.CustomeDurationAcquire(ctx, name, DefaultExpiry, DefaultTimeout)
}

func (s *RedisLockService) CustomeDurationAcquire(ctx context.Context, name string, expiresAt, timeout time.Duration) (Lock, error) {
	GetLogger().Debug("Acquiring lock", zap.String("name", name))
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ch := make(chan Lock)
	go s.acquire(ctx, ch, name, expiresAt)
	select {
	case <-ctx.Done():
		return Lock{}, errors.New("failed to Acquire lock")
	case lock := <-ch:
		return s.handleLockResult(lock)
	}
}

func (s *RedisLockService) acquire(ctx context.Context, ch chan Lock, name string, expiresAt time.Duration) {
	defer ctx.Done()
	err := s.waitForLock(ctx, name)
	if err != nil {
		GetLogger().Error("failed to wait for lock", zap.Error(err))
		ch <- Lock{}
		return
	}
	_, lockErr := s.client.SetEx(ctx, name, name, expiresAt).Result()
	if lockErr != nil {
		GetLogger().Error("failed to set lock", zap.Error(lockErr))
		ch <- Lock{}
	}
	ch <- Lock{Name: name, ExpiresAt: expiresAt}
}

func (s *RedisLockService) waitForLock(ctx context.Context, name string) error {
	for {
		err := s.tryAcquireLock(ctx, name)
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return errors.New("waiting for lock timed out")
		case <-time.After(1 * time.Second):
			continue
		}
	}
}

func (s *RedisLockService) tryAcquireLock(ctx context.Context, name string) error {
	value, err := s.client.Get(ctx, name).Result()
	if err == redis.Nil && value != name {
		return nil
	}
	return errors.New("failed to Acquire lock because it is already acquired")
}

func (s *RedisLockService) handleLockResult(lock Lock) (Lock, error) {
	if lock.Name == "" {
		return Lock{}, errors.New("lock is empty")
	}
	return lock, nil
}

func (s *RedisLockService) Release(ctx context.Context, lock Lock) error {
	GetLogger().Debug("Releasing lock", zap.String("name", lock.Name))
	_, err := s.client.Del(ctx, lock.Name).Result()
	if err != nil {
		GetLogger().Error("failed to Release lock", zap.Error(err))
		return errors.New("failed to Release lock")
	}
	return nil
}

func (s *RedisLockService) ReleaseMany(ctx context.Context, locks *[]Lock) {
	for _, lock := range *locks {
		s.Release(ctx, lock)
	}
}

func (s *RedisLockService) RunWithLock(ctx context.Context, name string, f func() error) error {
	lock, err := s.Acquire(ctx, name)
	if err != nil {
		return errors.New("Failed to acquire lock")
	}
	defer s.Release(ctx, lock)
	return f()
}
