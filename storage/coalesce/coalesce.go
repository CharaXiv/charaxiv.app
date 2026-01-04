// Package coalesce implements write coalescing storage.
// Writes are buffered in SQLite and periodically flushed to JSON files on disk.
package coalesce

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store handles write coalescing with SQLite buffer and JSON file persistence.
type Store struct {
	db       *sql.DB
	dataDir  string
	mu       sync.RWMutex
	cache    map[string]map[string]any // characterID -> JSON data
	stopCh   chan struct{}
	flushInt time.Duration
	closed   bool
}

// Config for creating a new Store
type Config struct {
	DBPath        string        // SQLite database path
	DataDir       string        // Directory for JSON files
	FlushInterval time.Duration // How often to flush (default 3s)
}

// New creates a new coalescing store
func New(cfg Config) (*Store, error) {
	if cfg.FlushInterval == 0 {
		cfg.FlushInterval = 3 * time.Second
	}

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	// Open SQLite
	db, err := sql.Open("sqlite3", cfg.DBPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Create buffer table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS write_buffer (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			character_id TEXT NOT NULL,
			path TEXT NOT NULL,
			value TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_buffer_char ON write_buffer(character_id);
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}

	s := &Store{
		db:       db,
		dataDir:  cfg.DataDir,
		cache:    make(map[string]map[string]any),
		stopCh:   make(chan struct{}),
		flushInt: cfg.FlushInterval,
	}

	// Start background flusher
	go s.flusher()

	return s, nil
}

// Close stops the flusher and closes the database
func (s *Store) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	close(s.stopCh)
	// Give flusher time to stop
	time.Sleep(10 * time.Millisecond)
	// Final flush
	s.Flush()
	return s.db.Close()
}

// Write buffers a write operation
func (s *Store) Write(characterID, path string, value any) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}

	_, err = s.db.Exec(
		`INSERT INTO write_buffer (character_id, path, value, created_at) VALUES (?, ?, ?, ?)`,
		characterID, path, string(valueJSON), time.Now().UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("insert buffer: %w", err)
	}

	// Update in-memory cache immediately
	s.mu.Lock()
	if s.cache[characterID] == nil {
		// Load from disk if not cached
		data, _ := s.loadFromDisk(characterID)
		if data == nil {
			data = make(map[string]any)
		}
		s.cache[characterID] = data
	}
	setPath(s.cache[characterID], path, value)
	s.mu.Unlock()

	return nil
}

// Read returns the current state for a character (cache + pending writes)
func (s *Store) Read(characterID string) (map[string]any, error) {
	s.mu.RLock()
	if data, ok := s.cache[characterID]; ok {
		// Return copy to avoid mutation
		s.mu.RUnlock()
		return copyMap(data), nil
	}
	s.mu.RUnlock()

	// Load from disk
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if data, ok := s.cache[characterID]; ok {
		return copyMap(data), nil
	}

	data, err := s.loadFromDisk(characterID)
	if err != nil {
		return nil, err
	}
	if data == nil {
		data = make(map[string]any)
	}

	// Apply any pending writes from buffer
	rows, err := s.db.Query(
		`SELECT path, value FROM write_buffer WHERE character_id = ? ORDER BY id`,
		characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("query buffer: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var path, valueJSON string
		if err := rows.Scan(&path, &valueJSON); err != nil {
			return nil, fmt.Errorf("scan buffer: %w", err)
		}
		var value any
		if err := json.Unmarshal([]byte(valueJSON), &value); err != nil {
			return nil, fmt.Errorf("unmarshal value: %w", err)
		}
		setPath(data, path, value)
	}

	s.cache[characterID] = data
	return copyMap(data), nil
}

// Flush writes all pending changes to disk
func (s *Store) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get all character IDs with pending writes
	rows, err := s.db.Query(`SELECT DISTINCT character_id FROM write_buffer`)
	if err != nil {
		return fmt.Errorf("query characters: %w", err)
	}

	var charIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return fmt.Errorf("scan character: %w", err)
		}
		charIDs = append(charIDs, id)
	}
	rows.Close()

	// Flush each character
	for _, charID := range charIDs {
		data := s.cache[charID]
		if data == nil {
			// Load and apply pending writes
			data, _ = s.loadFromDisk(charID)
			if data == nil {
				data = make(map[string]any)
			}
			// Apply pending writes
			bufRows, err := s.db.Query(
				`SELECT path, value FROM write_buffer WHERE character_id = ? ORDER BY id`,
				charID,
			)
			if err != nil {
				return fmt.Errorf("query buffer for %s: %w", charID, err)
			}
			for bufRows.Next() {
				var path, valueJSON string
				if err := bufRows.Scan(&path, &valueJSON); err != nil {
					bufRows.Close()
					return fmt.Errorf("scan buffer: %w", err)
				}
				var value any
				json.Unmarshal([]byte(valueJSON), &value)
				setPath(data, path, value)
			}
			bufRows.Close()
		}

		// Write to disk
		if err := s.saveToDisk(charID, data); err != nil {
			return fmt.Errorf("save %s: %w", charID, err)
		}

		// Clear buffer for this character
		_, err := s.db.Exec(`DELETE FROM write_buffer WHERE character_id = ?`, charID)
		if err != nil {
			return fmt.Errorf("clear buffer for %s: %w", charID, err)
		}
	}

	return nil
}

// flusher runs the background flush loop
func (s *Store) flusher() {
	ticker := time.NewTicker(s.flushInt)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.Flush(); err != nil {
				// Log error but continue
				fmt.Printf("flush error: %v\n", err)
			}
		case <-s.stopCh:
			return
		}
	}
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

// copyMap creates a deep copy of a map
func copyMap(m map[string]any) map[string]any {
	data, _ := json.Marshal(m)
	var result map[string]any
	json.Unmarshal(data, &result)
	return result
}
