package storageutil

import "github.com/aws/aws-sdk-go-v2/service/s3"

type S3PutOptions struct {
	s3.PutObjectInput
}

func (s S3PutOptions) Opts() {}
