package cache

import (
	"context"
	"time"

	"github.com/nayefradwi/nayef_go_common/core"
)

type FifoCacheStore struct {
	params        InMemoryCacheParams
	cacheKeyQueue []string
	cache         map[string]interface{}
}

func NewFifoCacheStore(params InMemoryCacheParams) ICacheStore {
	return &FifoCacheStore{
		params:        params,
		cacheKeyQueue: make([]string, 0),
		cache:         make(map[string]interface{}),
	}
}

func (f *FifoCacheStore) Get(ctx context.Context, key string) (interface{}, error) {
	if value, ok := f.cache[key]; ok {
		return value, nil
	}

	return nil, core.BadRequestError("Key not found")
}

func (f *FifoCacheStore) Set(ctx context.Context, key string, value interface{}) {
	if len(f.cacheKeyQueue) >= f.params.MaxSize {
		delete(f.cache, f.cacheKeyQueue[0])
		f.cacheKeyQueue = f.cacheKeyQueue[1:]
	}

	f.cache[key] = value
	f.cacheKeyQueue = append(f.cacheKeyQueue, key)
}

func (f *FifoCacheStore) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) {
	go f.DeleteAfter(ctx, key, expiration)
	f.Set(ctx, key, value)
}

func (f *FifoCacheStore) GetWithCacheMiss(ctx context.Context, key string, miss func() (interface{}, error)) (interface{}, error) {
	if value, ok := f.cache[key]; ok {
		return value, nil
	}

	value, err := miss()
	if err != nil {
		return nil, err
	}

	f.Set(ctx, key, value)

	return value, nil
}

func (f *FifoCacheStore) GetCachable(ctx context.Context, cachable ICachable) (interface{}, error) {
	return f.Get(ctx, cachable.CacheKey())
}

func (f *FifoCacheStore) SetCachable(ctx context.Context, cachable ICachable) {
	key := cachable.CacheKey()
	value, err := cachable.GetValue()
	if err != nil {
		return
	}

	f.Set(ctx, key, value)
}

func (f *FifoCacheStore) SetCachableEx(ctx context.Context, cachable ICachable, expiration time.Duration) {
	key := cachable.CacheKey()
	value, err := cachable.GetValue()
	if err != nil {
		return
	}

	f.SetEx(ctx, key, value, expiration)
}

func (f *FifoCacheStore) Delete(ctx context.Context, key string) {
	delete(f.cache, key)
	for i, k := range f.cacheKeyQueue {
		if k == key {
			f.cacheKeyQueue = append(f.cacheKeyQueue[:i], f.cacheKeyQueue[i+1:]...)
			break
		}
	}
}

func (f *FifoCacheStore) DeleteAfter(ctx context.Context, key string, expiration time.Duration) {
	timeOutCtx, cancel := context.WithTimeout(ctx, expiration)
	defer cancel()
	<-timeOutCtx.Done()
	f.Delete(ctx, key)
}
