package cache

type InMemoryCacheParams struct {
	MaxSize         int
	CachingStrategy int
}

func NewInMemoryCacheStore(params InMemoryCacheParams) ICacheStore {
	switch params.CachingStrategy {
	case FifoCache:
		return NewFifoCacheStore(params)
	case LifoCache:
		return NewLifoCacheStore(params)
	default:
		// TODO: change to LruCache
		return NewFifoCacheStore(params)
	}
}
