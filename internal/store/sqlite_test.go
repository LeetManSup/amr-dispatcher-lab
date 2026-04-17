package store

import (
	"path/filepath"
	"testing"
)

func TestNewSQLiteStoreCreatesParentDirectory(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "nested", "data", "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("create sqlite store: %v", err)
	}
	defer func() { _ = store.Close() }()
}
