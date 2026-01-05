package coalesce

import (
	"context"
)

// Backend defines the interface for character data persistence.
// Implementations handle the actual storage of character JSON data.
type Backend interface {
	// Load retrieves character data. Returns nil, nil if not found.
	Load(ctx context.Context, characterID string) (map[string]any, error)

	// Save persists character data.
	Save(ctx context.Context, characterID string, data map[string]any) error
}
