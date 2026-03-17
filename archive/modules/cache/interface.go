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
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{})
	SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration)
	GetWithCacheMiss(ctx context.Context, key string, miss func() (interface{}, error)) (interface{}, error)
	GetCachable(ctx context.Context, cachable ICachable) (interface{}, error)
	SetCachable(ctx context.Context, cachable ICachable)
	SetCachableEx(ctx context.Context, cachable ICachable, expiration time.Duration)
	Delete(ctx context.Context, key string)
}

type ICachable interface {
	CacheKey() string
	GetValue() (interface{}, error)
}

func CastCacheValue[T any](value interface{}) (T, error) {
	v, ok := value.(T)
	if !ok {
		return v, result.BadRequestError("Failed to cast cache value")
	}

	return v, nil
}
