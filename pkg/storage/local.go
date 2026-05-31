package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalStorage implements Storage using the local filesystem.
type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &LocalStorage{basePath: basePath}, nil
}

func (s *LocalStorage) Upload(_ context.Context, key string, data io.Reader, _ string) error {
	path := s.path(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func (s *LocalStorage) Download(_ context.Context, key string) (io.ReadCloser, error) {
	f, err := os.Open(s.path(key))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s", key)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return f, nil
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	if err := os.Remove(s.path(key)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetURL returns a local file path as the URL — no expiry for local storage.
func (s *LocalStorage) GetURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "file://" + s.path(key), nil
}

func (s *LocalStorage) Exists(_ context.Context, key string) (bool, error) {
	_, err := os.Stat(s.path(key))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *LocalStorage) path(key string) string {
	return filepath.Join(s.basePath, key)
}
