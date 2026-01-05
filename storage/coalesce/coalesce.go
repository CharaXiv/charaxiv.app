// Package coalesce implements write coalescing storage.
// Writes are buffered in SQLite and flushed to JSON files on read.
package coalesce

import (
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

// Store handles write coalescing with SQLite buffer and JSON file persistence.
type Store struct {
	db      *sql.DB
	dataDir string
	mu      sync.Mutex
}

// Config for creating a new Store
type Config struct {
	DBPath  string // SQLite database path
	DataDir string // Directory for JSON files
}

// New creates a new coalescing store
func New(cfg Config) (*Store, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
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
		dataDir: cfg.DataDir,
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
	start := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load from disk
	loadStart := time.Now()
	data, err := s.loadFromDisk(characterID)
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

		if err := s.saveToDisk(characterID, data); err != nil {
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

// loadFromDisk loads a character's JSON file
func (s *Store) loadFromDisk(characterID string) (map[string]any, error) {
	path := filepath.Join(s.dataDir, characterID+".json")
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

// saveToDisk writes a character's JSON file
func (s *Store) saveToDisk(characterID string, data map[string]any) error {
	path := filepath.Join(s.dataDir, characterID+".json")

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
