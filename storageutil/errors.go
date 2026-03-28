package storageutil

import "errors"

var (
	ErrCollectionManagerNotFound             = errors.New("collection manager not found")
	ErrNoPermissionToAccessCollectionManager = errors.New("no permission to access collection manager")
	ErrObjectNotFound                        = errors.New("object not found")
	ErrContentTypeNotAllowed                 = errors.New("content type not allowed")
	ErrUploadSizeExceeded                    = errors.New("upload size exceeded")
	ErrPresignFailed                         = errors.New("failed to generate presigned url")
)
