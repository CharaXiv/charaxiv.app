package coalesce

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	// Setup temp directory
	tmpDir, err := os.MkdirTemp("", "coalesce-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store with short flush interval for testing
	store, err := New(Config{
		DBPath:        filepath.Join(tmpDir, "buffer.db"),
		DataDir:       filepath.Join(tmpDir, "data"),
		FlushInterval: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	charID := "char-123"

	// Write some data
	if err := store.Write(charID, "name", "テストキャラ"); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(charID, "skills.回避.job", 5); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(charID, "skills.回避.hobby", 10); err != nil {
		t.Fatal(err)
	}

	// Read immediately (from cache)
	data, err := store.Read(charID)
	if err != nil {
		t.Fatal(err)
	}

	if data["name"] != "テストキャラ" {
		t.Errorf("expected name=テストキャラ, got %v", data["name"])
	}

	skills := data["skills"].(map[string]any)
	kaihi := skills["回避"].(map[string]any)
	if kaihi["job"] != float64(5) { // JSON numbers are float64
		t.Errorf("expected skills.回避.job=5, got %v", kaihi["job"])
	}

	// Wait for flush
	time.Sleep(200 * time.Millisecond)

	// Verify file was written
	filePath := filepath.Join(tmpDir, "data", charID+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("expected JSON file to be written")
	}

	// Create new store instance (simulates restart)
	store.Close()
	store2, err := New(Config{
		DBPath:        filepath.Join(tmpDir, "buffer.db"),
		DataDir:       filepath.Join(tmpDir, "data"),
		FlushInterval: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store2.Close()

	// Read from new store (should load from disk)
	data2, err := store2.Read(charID)
	if err != nil {
		t.Fatal(err)
	}

	if data2["name"] != "テストキャラ" {
		t.Errorf("after restart: expected name=テストキャラ, got %v", data2["name"])
	}
}

func TestSetPath(t *testing.T) {
	data := make(map[string]any)

	setPath(data, "a.b.c", 123)
	setPath(data, "a.b.d", "hello")
	setPath(data, "x", true)

	if data["x"] != true {
		t.Errorf("expected x=true")
	}

	a := data["a"].(map[string]any)
	b := a["b"].(map[string]any)
	if b["c"] != 123 {
		t.Errorf("expected a.b.c=123, got %v", b["c"])
	}
	if b["d"] != "hello" {
		t.Errorf("expected a.b.d=hello, got %v", b["d"])
	}
}
