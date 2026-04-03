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
	Key         string
	ContentType string
	Metadata    map[string]string
	Body        io.Reader
	Opts        ProviderPutOptions
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
