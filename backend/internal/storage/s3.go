package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/NeoRecasata/film-gallery/backend/internal/config"
)

type S3Storage struct {
	client   *s3.Client
	bucket   string
	isPublic bool
	endpoint string
	presign  *s3.PresignClient
}

func NewS3Storage(cfg *config.Config) (*S3Storage, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.StorageS3Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.StorageS3AccessKey,
				cfg.StorageS3SecretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.StorageS3Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.StorageS3Endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Storage{
		client:   client,
		bucket:   cfg.StorageS3Bucket,
		isPublic: cfg.StorageS3Public,
		endpoint: cfg.StorageS3Endpoint,
		presign:  s3.NewPresignClient(client),
	}, nil
}

func (s *S3Storage) Put(ctx context.Context, key string, reader io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return fmt.Errorf("uploading to S3: %w", err)
	}
	return nil
}

func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("getting from S3: %w", err)
	}
	return output.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting from S3: %w", err)
	}
	return nil
}

func (s *S3Storage) URL(ctx context.Context, key string) (string, error) {
	if s.isPublic {
		if s.endpoint != "" {
			return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key), nil
		}
		return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key), nil
	}

	// Presigned URL with 1-hour expiry
	presigned, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(1*time.Hour))
	if err != nil {
		return "", fmt.Errorf("generating presigned URL: %w", err)
	}
	return presigned.URL, nil
}
