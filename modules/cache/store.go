package cache

type InMemoryCacheParams struct {
	MaxSize         int
	CachingStrategy int
}

func NewInMemoryCacheStore(params InMemoryCacheParams) ICacheStore {
	switch params.CachingStrategy {
	case FifoCache:
		return NewFifoCacheStore(params)
	default:

	}
}
