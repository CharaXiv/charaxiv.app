package coalesce

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

// GCSBackend stores character data as JSON objects in Google Cloud Storage.
type GCSBackend struct {
	client *storage.Client
	bucket string
	prefix string // e.g., "characters/" for objects like "characters/demo.json"
}

// GCSBackendConfig holds configuration for the GCS backend.
type GCSBackendConfig struct {
	Bucket string
	Prefix string // Optional prefix for object keys (default: "characters/")
}

// NewGCSBackend creates a new GCS-based backend.
func NewGCSBackend(ctx context.Context, cfg GCSBackendConfig) (*GCSBackend, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create GCS client: %w", err)
	}

	prefix := cfg.Prefix
	if prefix == "" {
		prefix = "characters/"
	}

	return &GCSBackend{
		client: client,
		bucket: cfg.Bucket,
		prefix: prefix,
	}, nil
}

// Close closes the GCS client.
func (b *GCSBackend) Close() error {
	return b.client.Close()
}

// objectKey returns the full GCS object key for a character.
func (b *GCSBackend) objectKey(characterID string) string {
	return b.prefix + characterID + ".json"
}

// Load reads character data from GCS.
func (b *GCSBackend) Load(ctx context.Context, characterID string) (map[string]any, error) {
	obj := b.client.Bucket(b.bucket).Object(b.objectKey(characterID))
	r, err := obj.NewReader(ctx)
	if errors.Is(err, storage.ErrObjectNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open GCS object: %w", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read GCS object: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return result, nil
}

// Save writes character data to GCS.
func (b *GCSBackend) Save(ctx context.Context, characterID string, data map[string]any) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	obj := b.client.Bucket(b.bucket).Object(b.objectKey(characterID))
	w := obj.NewWriter(ctx)
	w.ContentType = "application/json"

	if _, err := io.Copy(w, bytes.NewReader(jsonData)); err != nil {
		w.Close()
		return fmt.Errorf("write to GCS: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close GCS writer: %w", err)
	}

	return nil
}

var _ Backend = (*GCSBackend)(nil)
