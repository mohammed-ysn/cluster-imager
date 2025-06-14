package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for object storage
type Storage interface {
	// Upload uploads data to storage and returns the key
	Upload(ctx context.Context, key string, data io.Reader, contentType string) error
	
	// Download downloads data from storage
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	
	// Delete deletes an object from storage
	Delete(ctx context.Context, key string) error
	
	// GetURL returns a URL for accessing the object
	// The URL may be signed and have an expiration time
	GetURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	
	// Exists checks if an object exists
	Exists(ctx context.Context, key string) (bool, error)
}

// Metadata represents object metadata
type Metadata struct {
	ContentType string
	Size        int64
	ETag        string
	LastModified time.Time
}

// Config represents storage configuration
type Config struct {
	Type       string // "minio", "s3", "local"
	Endpoint   string
	Bucket     string
	AccessKey  string
	SecretKey  string
	Region     string
	UseSSL     bool
	LocalPath  string // for local storage
}