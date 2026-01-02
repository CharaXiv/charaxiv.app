package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
)

// GCSClient wraps the GCS client for storage operations.
type GCSClient struct {
	client *storage.Client
	bucket string
}

// GCSConfig holds the configuration for connecting to GCS.
type GCSConfig struct {
	Bucket string
}

// NewGCSClient creates a new GCS client.
// Uses Application Default Credentials (ADC).
func NewGCSClient(ctx context.Context, cfg GCSConfig) (*GCSClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSClient{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Close closes the GCS client.
func (g *GCSClient) Close() error {
	return g.client.Close()
}

// Upload stores data in GCS.
func (g *GCSClient) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	obj := g.client.Bucket(g.bucket).Object(key)
	w := obj.NewWriter(ctx)
	w.ContentType = contentType

	if _, err := io.Copy(w, data); err != nil {
		w.Close()
		return fmt.Errorf("failed to upload to GCS: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	return nil
}

// Download retrieves data from GCS.
func (g *GCSClient) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	obj := g.client.Bucket(g.bucket).Object(key)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to download from GCS: %w", err)
	}
	return r, nil
}

// Delete removes an object from GCS.
func (g *GCSClient) Delete(ctx context.Context, key string) error {
	obj := g.client.Bucket(g.bucket).Object(key)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete from GCS: %w", err)
	}
	return nil
}

// SignedURL generates a signed URL for temporary access.
func (g *GCSClient) SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	url, err := g.client.Bucket(g.bucket).SignedURL(key, &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(expiry),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}
	return url, nil
}

// SignedUploadURL generates a signed URL for direct uploads.
func (g *GCSClient) SignedUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error) {
	url, err := g.client.Bucket(g.bucket).SignedURL(key, &storage.SignedURLOptions{
		Method:      "PUT",
		Expires:     time.Now().Add(expiry),
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signed upload URL: %w", err)
	}
	return url, nil
}

// Exists checks if an object exists in GCS.
func (g *GCSClient) Exists(ctx context.Context, key string) (bool, error) {
	obj := g.client.Bucket(g.bucket).Object(key)
	_, err := obj.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

// PublicURL returns the public URL for an object (if bucket is public).
func (g *GCSClient) PublicURL(key string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucket, key)
}
