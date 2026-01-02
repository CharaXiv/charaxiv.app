package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for object storage operations.
type Storage interface {
	// Upload stores data with the given key and content type.
	Upload(ctx context.Context, key string, data io.Reader, contentType string) error

	// Download retrieves data by key.
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes an object by key.
	Delete(ctx context.Context, key string) error

	// Exists checks if an object exists.
	Exists(ctx context.Context, key string) (bool, error)

	// SignedURL generates a temporary URL for reading an object.
	// Returns ErrSignedURLNotSupported if the implementation doesn't support it.
	SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// SignedUploadURL generates a temporary URL for uploading an object.
	// Returns ErrSignedURLNotSupported if the implementation doesn't support it.
	SignedUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error)

	// PublicURL returns the public URL for an object (if publicly accessible).
	PublicURL(key string) string
}

// Ensure implementations satisfy the interface.
var (
	_ Storage = (*GCSClient)(nil)
	_ Storage = (*MemoryStorage)(nil)
)
