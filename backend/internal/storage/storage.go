package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/NeoRecasata/film-gallery/backend/internal/config"
)

type Storage interface {
	Put(ctx context.Context, key string, reader io.Reader) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	URL(ctx context.Context, key string) (string, error)
}

func New(cfg *config.Config) (Storage, error) {
	switch cfg.StorageType {
	case "local":
		return NewLocalStorage(cfg.StorageLocalPath)
	case "s3":
		return NewS3Storage(cfg)
	default:
		return nil, fmt.Errorf("unknown storage type: %s (supported: local, s3)", cfg.StorageType)
	}
}
