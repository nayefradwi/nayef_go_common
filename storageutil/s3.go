package storageutil

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3CollectionManager struct {
	s3Client *s3.Client
}

func NewS3CollectionManager(s3Client *s3.Client) CollectionManager {
	return &S3CollectionManager{s3Client: s3Client}
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
	_, err := s.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(key)})
	if err != nil {
		return nil, fmt.Errorf("bucket not found %w %w", ErrCollectionNotFound, err)
	}

	return
}

func (s *S3CollectionManager) List(ctx context.Context) ([]Collection, error) {
	panic("unimplemented")
}

type S3ObjectManager struct {
	s3Client *s3.Client
	bucket   string
}

func NewS3ObjectManager(ctx context.Context, bucket string, client *s3.Client) (ObjectManager, error) {
	cm := NewS3CollectionManager(client)
	return cm.GetObjectManager(ctx, bucket)
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
	input := s.toPutObjectInput(params)
	out, err := s.s3Client.PutObject(ctx, &input)
	if err != nil {
		return ObjectMeta{}, fmt.Errorf("failed to upload object in s3: %w", err)
	}

	return ObjectMeta{
		Key:         params.Key,
		ContentType: params.ContentType,
		Size:        *out.Size,
		Metadata:    params.Metadata,
	}, nil
}

func (s *S3ObjectManager) toPutObjectInput(params PutObjectParams) s3.PutObjectInput {
	if s3Opts, ok := params.Opts.(S3PutOptions); ok {
		return s3Opts.PutObjectInput
	}

	return s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(params.Key),
		Body:        params.Body,
		ContentType: aws.String(params.ContentType),
		Metadata:    params.Metadata,
	}
}
