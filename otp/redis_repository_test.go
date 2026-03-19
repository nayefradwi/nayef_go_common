package otp

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertAndGetOtp(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	otp := NewOtp("owner-1", "1234", "hashed1234", 0, 3, time.Now().Add(5*time.Minute), time.Now())

	err := repo.UpsertOtp(ctx, otp)
	require.NoError(t, err)

	got, err := repo.GetOtp(ctx, "owner-1")
	require.NoError(t, err)

	assert.Equal(t, otp.OwnerId, got.OwnerId)
	assert.Equal(t, otp.RetryCount, got.RetryCount)
	assert.Equal(t, otp.MaxRetries, got.MaxRetries)
	assert.WithinDuration(t, otp.ExpiresAt, got.ExpiresAt, time.Second)
	// Code and HashedCode are json:"-" so they won't round-trip
	assert.Empty(t, got.Code)
	assert.Empty(t, got.HashedCode)
}

func TestGetOtp_NotFound(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	_, err := repo.GetOtp(ctx, "nonexistent-owner")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestUpsertOtp_NilOtp(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	err := repo.UpsertOtp(ctx, nil)
	assert.Error(t, err)
}

func TestUpsertOtp_ExpiredOtp(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	otp := NewOtp("owner-expired", "1234", "hashed1234", 0, 3, time.Now().Add(-1*time.Minute), time.Now())

	err := repo.UpsertOtp(ctx, otp)
	assert.ErrorIs(t, err, ErrExpiredOTP)
}

func TestOtpTTLExpiry(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	otp := NewOtp("owner-ttl", "1234", "hashed1234", 0, 3, time.Now().Add(500*time.Millisecond), time.Now())

	err := repo.UpsertOtp(ctx, otp)
	require.NoError(t, err)

	time.Sleep(600 * time.Millisecond)

	_, err = repo.GetOtp(ctx, "owner-ttl")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestUpsertOtp_OverwritesExisting(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	otp1 := NewOtp("owner-overwrite", "1111", "hash1", 0, 3, time.Now().Add(5*time.Minute), time.Now())
	err := repo.UpsertOtp(ctx, otp1)
	require.NoError(t, err)

	otp2 := NewOtp("owner-overwrite", "2222", "hash2", 2, 5, time.Now().Add(10*time.Minute), time.Now())
	err = repo.UpsertOtp(ctx, otp2)
	require.NoError(t, err)

	got, err := repo.GetOtp(ctx, "owner-overwrite")
	require.NoError(t, err)

	assert.Equal(t, 2, got.RetryCount)
	assert.Equal(t, 5, got.MaxRetries)
}
