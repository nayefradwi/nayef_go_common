package storageutil

import (
	"io"
	"time"
)

type ObjectMeta struct {
	Key         string
	Size        int64
	ContentType string
	Metadata    map[string]string
}

type PutObjectParams struct {
	ObjectMeta
	Body io.Reader
}

func NewPutObjectParams(
	key string,
	body io.Reader,
	size int64,
	contentType string,
	metadata map[string]string,
) PutObjectParams {
	return PutObjectParams{
		ObjectMeta: ObjectMeta{
			Key:         key,
			Size:        size,
			ContentType: contentType,
			Metadata:    metadata,
		},
		Body: body,
	}
}

type Collection struct {
	Key string
}

type Object struct {
	ObjectMeta
	Body io.ReadCloser
}

type PresignedURL struct {
	URL       string
	Method    string
	ExpiresAt time.Time
}
