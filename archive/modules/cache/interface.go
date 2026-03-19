package cache

import (
	"context"
	"time"

	"github.com/nayefradwi/nayef_go_common/result"
)

const (
	LruCache = iota
	LfuCache
	FifoCache
	LifoCache
)

type ICacheStore interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any)
	SetEx(ctx context.Context, key string, value any, expiration time.Duration)
	GetWithCacheMiss(ctx context.Context, key string, miss func() (any, error)) (any, error)
	GetCachable(ctx context.Context, cachable ICachable) (any, error)
	SetCachable(ctx context.Context, cachable ICachable)
	SetCachableEx(ctx context.Context, cachable ICachable, expiration time.Duration)
	Delete(ctx context.Context, key string)
}

type ICachable interface {
	CacheKey() string
	GetValue() (any, error)
}

func CastCacheValue[T any](value any) (T, error) {
	v, ok := value.(T)
	if !ok {
		return v, result.BadRequestError("Failed to cast cache value")
	}

	return v, nil
}
