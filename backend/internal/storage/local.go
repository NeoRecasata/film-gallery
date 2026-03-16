package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("creating storage directory: %w", err)
	}
	return &LocalStorage{basePath: basePath}, nil
}

func (s *LocalStorage) Put(_ context.Context, key string, reader io.Reader) error {
	fullPath := filepath.Join(s.basePath, key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("creating directories: %w", err)
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}

func (s *LocalStorage) Get(_ context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, key)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	return f, nil
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	fullPath := filepath.Join(s.basePath, key)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("deleting file: %w", err)
	}
	return nil
}

func (s *LocalStorage) URL(_ context.Context, key string) (string, error) {
	return fmt.Sprintf("/media/%s", key), nil
}
