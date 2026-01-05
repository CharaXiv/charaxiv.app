package coalesce

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore(t *testing.T) {
	// Setup temp directory
	tmpDir, err := os.MkdirTemp("", "coalesce-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	backend, err := NewDiskBackend(dataDir)
	if err != nil {
		t.Fatal(err)
	}

	store, err := New(Config{
		DBPath:  filepath.Join(tmpDir, "buffer.db"),
		Backend: backend,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	charID := "char-123"

	// Write some data (buffered, not yet on disk)
	if err := store.Write(charID, "name", "テストキャラ"); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(charID, "skills.回避.job", 5); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(charID, "skills.回避.hobby", 10); err != nil {
		t.Fatal(err)
	}

	// Verify file doesn't exist yet
	filePath := filepath.Join(dataDir, charID+".json")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("expected JSON file to NOT exist before read")
	}

	// Read triggers flush
	data, err := store.Read(charID)
	if err != nil {
		t.Fatal(err)
	}

	// Verify data
	if data["name"] != "テストキャラ" {
		t.Errorf("expected name=テストキャラ, got %v", data["name"])
	}

	skills := data["skills"].(map[string]any)
	kaihi := skills["回避"].(map[string]any)
	if kaihi["job"] != float64(5) {
		t.Errorf("expected skills.回避.job=5, got %v", kaihi["job"])
	}

	// Verify file now exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("expected JSON file to exist after read")
	}

	// Verify buffer is cleared (second read should not have pending writes)
	store.Close()

	backend2, err := NewDiskBackend(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	store2, err := New(Config{
		DBPath:  filepath.Join(tmpDir, "buffer.db"),
		Backend: backend2,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store2.Close()

	// Read from new store
	data2, err := store2.Read(charID)
	if err != nil {
		t.Fatal(err)
	}

	if data2["name"] != "テストキャラ" {
		t.Errorf("after restart: expected name=テストキャラ, got %v", data2["name"])
	}
}

func TestWriteCoalescing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "coalesce-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewDiskBackend(filepath.Join(tmpDir, "data"))
	if err != nil {
		t.Fatal(err)
	}

	store, err := New(Config{
		DBPath:  filepath.Join(tmpDir, "buffer.db"),
		Backend: backend,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	charID := "char-456"

	// Multiple writes to same path - last one wins
	store.Write(charID, "skills.回避.job", 1)
	store.Write(charID, "skills.回避.job", 2)
	store.Write(charID, "skills.回避.job", 3)

	data, _ := store.Read(charID)
	skills := data["skills"].(map[string]any)
	kaihi := skills["回避"].(map[string]any)

	if kaihi["job"] != float64(3) {
		t.Errorf("expected coalesced value 3, got %v", kaihi["job"])
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
