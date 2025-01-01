package common

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedisLockService_Acquire_Release(t *testing.T) {
	client := ConnectRedis(context.Background(), "redis://192.168.100.34:6379")
	service := &RedisLockService{client: client}

	lockName := "testLock1"
	expiresAt := DefaultExpiry

	log.Printf("Acquiring lock %s", lockName)

	lock, err := service.Acquire(context.Background(), lockName)
	assert.NoError(t, err)
	assert.Equal(t, lockName, lock.Name)
	assert.Equal(t, expiresAt, lock.ExpiresAt)

	retryErr := service.tryAcquireLock(context.Background(), lockName)
	assert.Error(t, retryErr) // The lock should not be available for other clients

	log.Printf("Releasing lock %s", lockName)
	err = service.Release(context.Background(), lock)
	assert.NoError(t, err)

	log.Printf("ensuring lock %s is released", lockName)
	err = service.tryAcquireLock(context.Background(), lockName)
	assert.NoError(t, err)
}

func TestRedisLockService_Release_DurationExpired(t *testing.T) {
	client := ConnectRedis(context.Background(), "redis://192.168.100.34:6379")
	service := &RedisLockService{client: client}

	lockName := "testLock2"
	expiresAt := 5 * time.Second // Short expiration time for testing

	log.Printf("Acquiring lock %s with a short expiration time", lockName)

	_, err := service.CustomeDurationAcquire(context.Background(), lockName, expiresAt, time.Second)
	assert.NoError(t, err)

	// Wait for the lock to expire
	time.Sleep(5 * time.Second)

	log.Printf("ensuring lock %s is released", lockName)
	err = service.tryAcquireLock(context.Background(), lockName)
	assert.NoError(t, err)
}

func TestRedisLockService_MultipleReads(t *testing.T) {
	client := ConnectRedis(context.Background(), "redis://192.168.100.34:6379")
	service := &RedisLockService{client: client}
	lockName := "testLock3"
	expiresAt := 7 * time.Second
	const numReaders = 5
	errors := make([]error, 0)
	tryAcquireLock := func() {
		_, err := service.CustomeDurationAcquire(context.Background(), lockName, expiresAt, time.Second)
		errors = append(errors, err)
		time.Sleep(100 * time.Millisecond)
	}
	for i := 0; i < numReaders; i++ {
		tryAcquireLock()
	}

	firstReaderFailed := false
	for _, err := range errors {
		if !firstReaderFailed {
			assert.NoError(t, err) // The first reader should not fail
			firstReaderFailed = true
		} else {
			assert.Error(t, err) // All other readers should fail
		}
	}
}

func TestRedisLockService_Release_MultipleTimes(t *testing.T) {
	client := ConnectRedis(context.Background(), "redis://192.168.100.34:6379")
	service := &RedisLockService{client: client}

	lockName := "testLock4"
	expiresAt := DefaultExpiry

	log.Printf("Acquiring lock %s", lockName)

	lock, err := service.Acquire(context.Background(), lockName)
	assert.NoError(t, err)
	assert.Equal(t, lockName, lock.Name)
	assert.Equal(t, expiresAt, lock.ExpiresAt)

	// Try to release the same lock multiple times
	for i := 0; i < 3; i++ {
		log.Printf("Releasing lock %s (Attempt %d)", lockName, i+1)
		err = service.Release(context.Background(), lock)
		assert.NoError(t, err) // No error should be returned when releasing the same lock multiple times
	}
}
