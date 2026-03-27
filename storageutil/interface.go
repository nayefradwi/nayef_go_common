package storageutil

import (
	"context"
	"io"
)

type ObjectManager interface {
	Put(ctx context.Context, key string, body io.Reader) (Object, error)
	Delete(ctx context.Context, key string) error
	DeleteMany(ctx context.Context, keys []string) error
	List(ctx context.Context, prefix string, max int) []Object
	Get(ctx context.Context, key string) (Object, error)
}

type CollectionManager interface {
	Create(ctx context.Context, key string) (Collection, error)
	List(ctx context.Context) ([]Collection, error)
	Delete(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (Collection, error)
	GetObjectManager(ctx context.Context, key string) (ObjectManager, error)
}
