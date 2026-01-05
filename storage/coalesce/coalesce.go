// Package coalesce implements write coalescing storage.
// Writes are buffered in SQLite and flushed to a backend on read.
package coalesce

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Store handles write coalescing with SQLite buffer and pluggable backend persistence.
type Store struct {
	db      *sql.DB
	backend Backend
	mu      sync.Mutex
}

// Config for creating a new Store
type Config struct {
	DBPath  string  // SQLite database path for write buffer
	Backend Backend // Backend for persistent storage
}

// New creates a new coalescing store
func New(cfg Config) (*Store, error) {
	if cfg.Backend == nil {
		return nil, fmt.Errorf("backend is required")
	}

	// Ensure directory exists for SQLite buffer
	if dir := filepath.Dir(cfg.DBPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create buffer dir: %w", err)
		}
	}

	// Open SQLite
	db, err := sql.Open("sqlite", cfg.DBPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Create buffer table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS write_buffer (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			character_id TEXT NOT NULL,
			path TEXT NOT NULL,
			value TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_buffer_char ON write_buffer(character_id);
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}

	return &Store{
		db:      db,
		backend: cfg.Backend,
	}, nil
}

// Close closes the database
func (s *Store) Close() error {
	return s.db.Close()
}

// Write buffers a write operation
func (s *Store) Write(characterID, path string, value any) error {
	start := time.Now()
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}

	_, err = s.db.Exec(
		`INSERT INTO write_buffer (character_id, path, value) VALUES (?, ?, ?)`,
		characterID, path, string(valueJSON),
	)
	if err != nil {
		return fmt.Errorf("insert buffer: %w", err)
	}

	slog.Debug("coalesce.write", "characterID", characterID, "path", path, "duration", time.Since(start))
	return nil
}

// Read returns the current state for a character, flushing any pending writes
func (s *Store) Read(characterID string) (map[string]any, error) {
	return s.ReadWithContext(context.Background(), characterID)
}

// ReadWithContext returns the current state for a character, flushing any pending writes
func (s *Store) ReadWithContext(ctx context.Context, characterID string) (map[string]any, error) {
	start := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load from backend
	loadStart := time.Now()
	data, err := s.backend.Load(ctx, characterID)
	if err != nil {
		return nil, err
	}
	if data == nil {
		data = make(map[string]any)
	}
	loadDuration := time.Since(loadStart)

	// Get pending writes
	pending, err := s.getPendingWrites(characterID)
	if err != nil {
		return nil, err
	}

	// Apply and flush if there are pending writes
	var flushDuration time.Duration
	if len(pending) > 0 {
		flushStart := time.Now()
		for _, p := range pending {
			setPath(data, p.path, p.value)
		}

		if err := s.backend.Save(ctx, characterID, data); err != nil {
			return nil, err
		}

		if err := s.clearBuffer(characterID); err != nil {
			return nil, err
		}
		flushDuration = time.Since(flushStart)
	}

	slog.Debug("coalesce.read",
		"characterID", characterID,
		"pendingWrites", len(pending),
		"loadDuration", loadDuration,
		"flushDuration", flushDuration,
		"duration", time.Since(start),
	)
	return data, nil
}

type pendingWrite struct {
	path  string
	value any
}

func (s *Store) getPendingWrites(characterID string) ([]pendingWrite, error) {
	rows, err := s.db.Query(
		`SELECT path, value FROM write_buffer WHERE character_id = ? ORDER BY id`,
		characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("query buffer: %w", err)
	}
	defer rows.Close()

	var pending []pendingWrite
	for rows.Next() {
		var path, valueJSON string
		if err := rows.Scan(&path, &valueJSON); err != nil {
			return nil, fmt.Errorf("scan buffer: %w", err)
		}
		var value any
		if err := json.Unmarshal([]byte(valueJSON), &value); err != nil {
			return nil, fmt.Errorf("unmarshal value: %w", err)
		}
		pending = append(pending, pendingWrite{path: path, value: value})
	}

	return pending, nil
}

func (s *Store) clearBuffer(characterID string) error {
	_, err := s.db.Exec(`DELETE FROM write_buffer WHERE character_id = ?`, characterID)
	if err != nil {
		return fmt.Errorf("clear buffer: %w", err)
	}
	return nil
}

// setPath sets a value at a dot-separated path in a nested map
// e.g., setPath(data, "skills.回避.job", 5)
func setPath(data map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return
	}

	// Navigate to parent
	current := data
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]
		if current[key] == nil {
			current[key] = make(map[string]any)
		}
		if next, ok := current[key].(map[string]any); ok {
			current = next
		} else {
			// Can't traverse, create new map
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
		}
	}

	// Set final value
	current[parts[len(parts)-1]] = value
}
