package auth

import (
	"context"
	"time"

	. "github.com/nayefradwi/nayef_go_common/errors"
)

type Attempt struct {
	Count   int
	ResetAt time.Time
}

type AttemptStore interface {
	Hit(ctx context.Context, key string, window time.Duration) (Attempt, error)
	Reset(ctx context.Context, key string) error
	DeleteExpired(ctx context.Context) error
}

type Limiter struct {
	store  AttemptStore
	limit  int
	window time.Duration
}

func NewLimiter(store AttemptStore, limit int, window time.Duration) (Limiter, error) {
	if store == nil || limit < 1 || window <= 0 {
		return Limiter{}, InternalError("limiter needs a store, a limit of at least 1 and a positive window")
	}
	return Limiter{store: store, limit: limit, window: window}, nil
}

func (l Limiter) Allow(ctx context.Context, key string) (bool, time.Duration, error) {
	a, err := l.store.Hit(ctx, key, l.window)
	if err != nil {
		return false, 0, err
	}
	if a.Count <= l.limit {
		return true, 0, nil
	}
	return false, max(time.Until(a.ResetAt), 0), nil
}
