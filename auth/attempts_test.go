package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countingStore struct {
	counts  map[string]int
	resetAt time.Time
	err     error
}

func (s *countingStore) Hit(_ context.Context, key string, _ time.Duration) (Attempt, error) {
	s.counts[key]++
	return Attempt{Count: s.counts[key], ResetAt: s.resetAt}, s.err
}
func (s *countingStore) Reset(context.Context, string) error { return nil }
func (s *countingStore) DeleteExpired(context.Context) error { return nil }

func TestLimiter_Boundary(t *testing.T) {
	store := &countingStore{counts: map[string]int{}, resetAt: time.Now().Add(time.Minute)}
	l, err := NewLimiter(store, 3, time.Minute)
	require.NoError(t, err)

	for i := 1; i <= 3; i++ {
		ok, _, err := l.Allow(context.Background(), "k")
		require.NoError(t, err)
		assert.True(t, ok, "hit %d", i)
	}

	ok, retryAfter, err := l.Allow(context.Background(), "k")
	require.NoError(t, err)
	assert.False(t, ok)
	assert.InDelta(t, time.Minute, retryAfter, float64(time.Second))

	ok, _, _ = l.Allow(context.Background(), "other")
	assert.True(t, ok, "keys are counted separately")
}

func TestLimiter_StoreErrorDenies(t *testing.T) {
	l, err := NewLimiter(&countingStore{counts: map[string]int{}, err: errors.New("down")}, 3, time.Minute)
	require.NoError(t, err)
	ok, _, err := l.Allow(context.Background(), "k")
	assert.Error(t, err)
	assert.False(t, ok)
}

func TestNewLimiter_RejectsBadConfig(t *testing.T) {
	store := &countingStore{}
	_, err := NewLimiter(store, 0, time.Minute)
	assert.Error(t, err)
	_, err = NewLimiter(store, 1, 0)
	assert.Error(t, err)
	_, err = NewLimiter(nil, 1, time.Minute)
	assert.Error(t, err)
}
