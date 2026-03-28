package storageutil

import (
	"context"
	"time"
)

type ObjectManager interface {
	Put(ctx context.Context, params PutObjectParams) (ObjectMeta, error)
	Delete(ctx context.Context, key string) error
	DeleteMany(ctx context.Context, keys []string) error
	List(ctx context.Context, prefix string, max int) ([]ObjectMeta, error)
	Get(ctx context.Context, key string) (Object, error)
	GetMeta(ctx context.Context, key string) (ObjectMeta, error)
}

type Presigner interface {
	PresignGet(ctx context.Context, key string, expiry time.Duration) (PresignedURL, error)
	PresignPut(ctx context.Context, key string, contentType string, expiry time.Duration) (PresignedURL, error)
}

type URLResolver interface {
	ResolveURL(ctx context.Context, key string) (string, error)
}

type CollectionManager interface {
	Create(ctx context.Context, key string) (Collection, error)
	List(ctx context.Context) ([]Collection, error)
	Delete(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (Collection, error)
	GetObjectManager(ctx context.Context, key string) (ObjectManager, error)
}
