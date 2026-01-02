package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// ErrSignedURLNotSupported is returned when signed URLs are not supported.
var ErrSignedURLNotSupported = errors.New("signed URLs not supported")

// ErrNotFound is returned when an object doesn't exist.
var ErrNotFound = errors.New("object not found")

// MemoryStorage is an in-memory implementation of Storage for testing.
type MemoryStorage struct {
	mu      sync.RWMutex
	objects map[string][]byte
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		objects: make(map[string][]byte),
	}
}

func (m *MemoryStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	buf, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	m.mu.Lock()
	m.objects[key] = buf
	m.mu.Unlock()

	return nil
}

func (m *MemoryStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	m.mu.RLock()
	data, ok := m.objects[key]
	m.mu.RUnlock()

	if !ok {
		return nil, ErrNotFound
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *MemoryStorage) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	delete(m.objects, key)
	m.mu.Unlock()

	return nil
}

func (m *MemoryStorage) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	_, ok := m.objects[key]
	m.mu.RUnlock()

	return ok, nil
}

func (m *MemoryStorage) SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "", ErrSignedURLNotSupported
}

func (m *MemoryStorage) SignedUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error) {
	return "", ErrSignedURLNotSupported
}

func (m *MemoryStorage) PublicURL(key string) string {
	return fmt.Sprintf("memory://%s", key)
}

// Keys returns all keys in storage (useful for testing).
func (m *MemoryStorage) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.objects))
	for k := range m.objects {
		keys = append(keys, k)
	}
	return keys
}

// Clear removes all objects (useful for testing).
func (m *MemoryStorage) Clear() {
	m.mu.Lock()
	m.objects = make(map[string][]byte)
	m.mu.Unlock()
}
