package cache

import (
	"context"
	"time"

	"github.com/nayefradwi/nayef_go_common/core"
)

type LifoCacheStore struct {
	params        InMemoryCacheParams
	cacheKeyStack []string
	cache         map[string]interface{}
}

func NewLifoCacheStore(params InMemoryCacheParams) ICacheStore {
	return &LifoCacheStore{
		params:        params,
		cacheKeyStack: make([]string, 0),
		cache:         make(map[string]interface{}),
	}
}

func (l *LifoCacheStore) Get(ctx context.Context, key string) (interface{}, error) {
	if value, ok := l.cache[key]; ok {
		return value, nil
	}

	return nil, core.BadRequestError("Key not found")
}

func (l *LifoCacheStore) Set(ctx context.Context, key string, value interface{}) {
	if len(l.cacheKeyStack) >= l.params.MaxSize {
		delete(l.cache, l.cacheKeyStack[len(l.cacheKeyStack)-1])
		l.cacheKeyStack = l.cacheKeyStack[:len(l.cacheKeyStack)-1]
	}

	l.cache[key] = value
	l.cacheKeyStack = append([]string{key}, l.cacheKeyStack...)
}

func (l *LifoCacheStore) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) {
	go l.DeleteAfter(ctx, key, expiration)
	l.Set(ctx, key, value)
}

func (l *LifoCacheStore) GetWithCacheMiss(ctx context.Context, key string, miss func() (interface{}, error)) (interface{}, error) {
	if value, ok := l.cache[key]; ok {
		return value, nil
	}

	value, err := miss()
	if err != nil {
		return nil, err
	}

	l.Set(ctx, key, value)
	return value, nil
}

func (l *LifoCacheStore) GetCachable(ctx context.Context, cachable ICachable) (interface{}, error) {
	return l.Get(ctx, cachable.CacheKey())
}

func (l *LifoCacheStore) SetCachable(ctx context.Context, cachable ICachable) {
	value, err := cachable.GetValue()
	if err != nil {
		return
	}

	l.Set(ctx, cachable.CacheKey(), value)
}

func (l *LifoCacheStore) SetCachableEx(ctx context.Context, cachable ICachable, expiration time.Duration) {
	value, err := cachable.GetValue()
	if err != nil {
		return
	}

	l.SetEx(ctx, cachable.CacheKey(), value, expiration)
}

func (l *LifoCacheStore) Delete(ctx context.Context, key string) {
	delete(l.cache, key)
	for i, k := range l.cacheKeyStack {
		if k == key {
			l.cacheKeyStack = append(l.cacheKeyStack[:i], l.cacheKeyStack[i+1:]...)
		}
	}
}

func (f *LifoCacheStore) DeleteAfter(ctx context.Context, key string, expiration time.Duration) {
	timeOutCtx, cancel := context.WithTimeout(ctx, expiration)
	defer cancel()
	<-timeOutCtx.Done()
	f.Delete(ctx, key)
}
