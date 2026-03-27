package storageutil

import "errors"

var (
	ErrCollectionManagerNotFound             = errors.New("collection manager not found")
	ErrNoPermissionToAccessCollectionManager = errors.New("no permission to access collection manager")
)
