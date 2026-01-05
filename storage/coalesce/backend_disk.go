package coalesce

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DiskBackend stores character data as JSON files on local disk.
type DiskBackend struct {
	dataDir string
}

// NewDiskBackend creates a new disk-based backend.
func NewDiskBackend(dataDir string) (*DiskBackend, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	return &DiskBackend{dataDir: dataDir}, nil
}

// Load reads character data from a JSON file.
func (b *DiskBackend) Load(_ context.Context, characterID string) (map[string]any, error) {
	path := filepath.Join(b.dataDir, characterID+".json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return result, nil
}

// Save writes character data to a JSON file atomically.
func (b *DiskBackend) Save(_ context.Context, characterID string, data map[string]any) error {
	path := filepath.Join(b.dataDir, characterID+".json")

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	// Atomic write via temp file
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, jsonData, 0644); err != nil {
		return fmt.Errorf("write temp: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}

var _ Backend = (*DiskBackend)(nil)
