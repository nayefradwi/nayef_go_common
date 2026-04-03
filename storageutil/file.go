package storageutil

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"time"
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

	if err != nil {
		return nil, err
	}

	mode := info.Mode().Perm()
	if mode&0700 != 0700 {
		return nil, fmt.Errorf("no permissions to access: %s, %w", root, ErrNoPermissionToAccessCollectionManager)
	}

	return &FileCollectionManager{RootDir: root}, nil
}

func (f *FileCollectionManager) Create(ctx context.Context, key string) (Collection, error) {
	dirPath := path.Join(f.RootDir, key)
	return Collection{Key: dirPath}, os.MkdirAll(dirPath, 0755)
}

func (f *FileCollectionManager) Delete(ctx context.Context, key string) error {
	dirPath := path.Join(f.RootDir, key)
	return os.RemoveAll(dirPath)
}

func (f *FileCollectionManager) Get(ctx context.Context, key string) (Collection, error) {
	dirPath := path.Join(f.RootDir, key)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return Collection{}, err
	}

	return Collection{Key: dirPath}, nil
}

func (f *FileCollectionManager) GetObjectManager(ctx context.Context, key string) (ObjectManager, error) {
	dirPath := path.Join(f.RootDir, key)
	collection := Collection{Key: dirPath}
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, err
	}

	return NewFileObjectManager(collection), nil
}

func (f *FileCollectionManager) List(ctx context.Context) ([]Collection, error) {
	entries, err := os.ReadDir(f.RootDir)
	if err != nil {
		return nil, err
	}

	collections := make([]Collection, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		collections = append(collections, Collection{Key: path.Join(f.RootDir, entry.Name())})
	}

	return collections, nil
}

type FileObjectManager struct {
	Collection Collection
}

func NewFileObjectManagerFromPath(root string, key string) (ObjectManager, error) {
	collection, err := NewFileCollectionManager(root)
	if err != nil {
		return nil, err
	}

	return collection.GetObjectManager(context.Background(), key)
}

func NewFileObjectManager(collection Collection) ObjectManager {
	return &FileObjectManager{Collection: collection}
}

func (f *FileObjectManager) Delete(ctx context.Context, key string) error {
	filePath := path.Join(f.Collection.Key, key)
	return os.Remove(filePath)
}

func (f *FileObjectManager) DeleteMany(ctx context.Context, keys []string) error {
	for _, key := range keys {
		filePath := path.Join(f.Collection.Key, key)
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileObjectManager) Get(ctx context.Context, key string) (Object, error) {
	filePath := path.Join(f.Collection.Key, key)
	file, err := os.Open(filePath)
	if err != nil && os.IsNotExist(err) {
		return Object{}, fmt.Errorf("%w: %s", ErrObjectNotFound, key)
	}

	if err != nil {
		return Object{}, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return Object{}, err
	}

	return Object{ObjectMeta: f.toObjectMeta(info), Body: file}, nil
}

func (f *FileObjectManager) GetMeta(ctx context.Context, key string) (ObjectMeta, error) {
	filePath := path.Join(f.Collection.Key, key)
	info, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return ObjectMeta{}, fmt.Errorf("%w: %s", ErrObjectNotFound, key)
	}

	if err != nil {
		return ObjectMeta{}, err
	}

	return f.toObjectMeta(info), nil
}

func (f *FileObjectManager) List(ctx context.Context, prefix string, max int) ([]ObjectMeta, error) {
	dirPath := path.Join(f.Collection.Key, prefix)
	objects := make([]ObjectMeta, 0, max)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if len(objects) >= max {
			break
		}
		info, err := entry.Info()
		if err != nil || info.IsDir() {
			continue
		}

		objects = append(objects, f.toObjectMeta(info))
	}

	return objects, nil
}

func (f *FileObjectManager) Put(ctx context.Context, params PutObjectParams) (ObjectMeta, error) {
	filePath := path.Join(f.Collection.Key, params.Key)
	file, err := os.Create(filePath)
	if err != nil {
		return ObjectMeta{}, err
	}

	defer file.Close()

	if _, err := io.Copy(file, params.Body); err != nil {
		return ObjectMeta{}, err
	}

	info, err := file.Stat()
	if err != nil {
		return ObjectMeta{}, err
	}

	return f.toObjectMeta(info), nil
}

func (f *FileObjectManager) toObjectMeta(info os.FileInfo) ObjectMeta {
	return ObjectMeta{
		Key:         path.Join(f.Collection.Key, info.Name()),
		Size:        info.Size(),
		ContentType: mime.TypeByExtension(filepath.Ext(info.Name())),
		Metadata: map[string]string{
			"mod_time": info.ModTime().Format(time.RFC3339),
			"mode":     info.Mode().String(),
		},
	}
}
