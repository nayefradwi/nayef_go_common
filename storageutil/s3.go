package storageutil

import (
	"context"
	"io"
)

type S3CollectionManager struct{}

func NewS3CollectionManager() CollectionManager {
	return &S3CollectionManager{}
}

// Create implements [CollectionManager].
func (s *S3CollectionManager) Create(ctx context.Context, key string) (Collection, error) {
	panic("unimplemented")
}

// Delete implements [CollectionManager].
func (s *S3CollectionManager) Delete(ctx context.Context, key string) error {
	panic("unimplemented")
}

// Get implements [CollectionManager].
func (s *S3CollectionManager) Get(ctx context.Context, key string) (Collection, error) {
	panic("unimplemented")
}

// GetObjectManager implements [CollectionManager].
func (s *S3CollectionManager) GetObjectManager(ctx context.Context, key string) (ObjectManager, error) {
	panic("unimplemented")
}

// List implements [CollectionManager].
func (s *S3CollectionManager) List(ctx context.Context) ([]Collection, error) {
	panic("unimplemented")
}

type S3ObjectManager struct{}

func NewS3ObjectManager() ObjectManager {
	return &S3ObjectManager{}
}

// Delete implements [ObjectManager].
func (s *S3ObjectManager) Delete(ctx context.Context, key string) error {
	panic("unimplemented")
}

// DeleteMany implements [ObjectManager].
func (s *S3ObjectManager) DeleteMany(ctx context.Context, keys []string) error {
	panic("unimplemented")
}

// Get implements [ObjectManager].
func (s *S3ObjectManager) Get(ctx context.Context, key string) (Object, error) {
	panic("unimplemented")
}

// List implements [ObjectManager].
func (s *S3ObjectManager) List(ctx context.Context, prefix string, max int) []Object {
	panic("unimplemented")
}

// Put implements [ObjectManager].
func (s *S3ObjectManager) Put(ctx context.Context, key string, body io.Reader) (Object, error) {
	panic("unimplemented")
}
