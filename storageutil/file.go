package storageutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
)

type FileCollectionManager struct {
	RootDir string
}

func NewFileCollectionManager(root string) (CollectionManager, error) {
	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf(
			"failed to create collection manager with key: %s, %w %w",
			root,
			ErrCollectionManagerNotFound,
			err,
		)
	}

	mode := info.Mode().Perm()
	if mode&0700 != 0700 {
		return nil, fmt.Errorf("no permissions to access: %s, %w", root, ErrNoPermissionToAccessCollectionManager)
	}

	return &FileCollectionManager{RootDir: root}, nil
}

func (f *FileCollectionManager) Create(ctx context.Context, key string) (Collection, error) {
	path := path.Join(f.RootDir, key)
	return Collection{Key: path}, os.MkdirAll(path, 0755)
}

func (f *FileCollectionManager) Delete(ctx context.Context, key string) error {
	path := path.Join(f.RootDir, key)
	return os.RemoveAll(path)
}

func (f *FileCollectionManager) Get(ctx context.Context, key string) (Collection, error) {
	path := path.Join(f.RootDir, key)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Collection{Key: path}, err
	}

	return Collection{Key: path}, nil
}

func (f *FileCollectionManager) GetObjectManager(ctx context.Context, key string) (ObjectManager, error) {
	path := path.Join(f.RootDir, key)
	collection := Collection{Key: path}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	return NewFileObjectManager(collection), nil
}

func (f *FileCollectionManager) List(ctx context.Context) ([]Collection, error) {
	panic("not implemented")
}

type FileObjectManager struct {
	Collection Collection
}

func NewFileObjectManager(collection Collection) ObjectManager {
	return &FileObjectManager{Collection: collection}
}

func (f *FileObjectManager) Delete(ctx context.Context, key string) error {
	path := path.Join(f.Collection.Key, key)
	return os.Remove(path)
}

func (f *FileObjectManager) DeleteMany(ctx context.Context, keys []string) error {
	for _, key := range keys {
		path := path.Join(f.Collection.Key, key)
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileObjectManager) Get(ctx context.Context, key string) (Object, error) {
	panic("unimplemented")
}

func (f *FileObjectManager) List(ctx context.Context, prefix string, max int) []Object {
	path := path.Join(f.Collection.Key, prefix)
	objects := make([]Object, max)
	entries, err := os.ReadDir(path)
	if err != nil {
		return objects
	}

	addedObjects := 0
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || info.IsDir() {
			continue
		}

		objects[addedObjects] = f.toObject(info)
		addedObjects++
	}

	return objects
}

func (f *FileObjectManager) Put(ctx context.Context, key string, body io.Reader) (Object, error) {
	panic("unimplemented")
}

func (f *FileObjectManager) toObject(info os.FileInfo) Object {
	return Object{
		Key:         info.Name(),
		Body:        nil,
		Size:        info.Size(),
		ContentType: "",
		Metadata:    map[string]string{},
	}
}
