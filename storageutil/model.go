package storageutil

import "io"

type Object struct {
	Key         string
	Body        io.Reader
	Size        int64
	ContentType string
	Metadata    map[string]string
}

type Collection struct {
	Key string
}
