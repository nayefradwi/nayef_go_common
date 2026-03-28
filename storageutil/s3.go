package storageutil

import (
	"context"
)

type S3CollectionManager struct{}

func NewS3CollectionManager() CollectionManager {
	return &S3CollectionManager{}
}

func (s *S3CollectionManager) Create(ctx context.Context, key string) (Collection, error) {
	panic("unimplemented")
}

func (s *S3CollectionManager) Delete(ctx context.Context, key string) error {
	panic("unimplemented")
}

func (s *S3CollectionManager) Get(ctx context.Context, key string) (Collection, error) {
	panic("unimplemented")
}

func (s *S3CollectionManager) GetObjectManager(ctx context.Context, key string) (ObjectManager, error) {
	panic("unimplemented")
}

func (s *S3CollectionManager) List(ctx context.Context) ([]Collection, error) {
	panic("unimplemented")
}

type S3ObjectManager struct{}

func NewS3ObjectManager() ObjectManager {
	return &S3ObjectManager{}
}

func (s *S3ObjectManager) Delete(ctx context.Context, key string) error {
	panic("unimplemented")
}

func (s *S3ObjectManager) DeleteMany(ctx context.Context, keys []string) error {
	panic("unimplemented")
}

func (s *S3ObjectManager) Get(ctx context.Context, key string) (Object, error) {
	panic("unimplemented")
}

func (s *S3ObjectManager) GetMeta(ctx context.Context, key string) (ObjectMeta, error) {
	panic("unimplemented")
}

func (s *S3ObjectManager) List(ctx context.Context, prefix string, max int) ([]ObjectMeta, error) {
	panic("unimplemented")
}

func (s *S3ObjectManager) Put(ctx context.Context, params PutObjectParams) (ObjectMeta, error) {
	panic("unimplemented")
}
