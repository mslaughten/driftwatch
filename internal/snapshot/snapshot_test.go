package snapshot

import (
	"os"
	"path/filepath"
	"testing"
)

func tempStorePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "snapshot.json")
}

func TestNew_EmptyWhenMissing(t *testing.T) {
	s, err := New(tempStorePath(t))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(s.All()) != 0 {
		t.Errorf("expected empty store, got %d entries", len(s.All()))
	}
}

func TestSetAndGet(t *testing.T) {
	s, _ := New(tempStorePath(t))
	if err := s.Set("/etc/app/config.yaml", "abc123"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	e, ok := s.Get("/etc/app/config.yaml")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Hash != "abc123" {
		t.Errorf("expected hash abc123, got %s", e.Hash)
	}
	if e.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestPersistence(t *testing.T) {
	path := tempStorePath(t)
	s, _ := New(path)
	_ = s.Set("/etc/svc/cfg.toml", "deadbeef")

	// Reload from disk
	s2, err := New(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	e, ok := s2.Get("/etc/svc/cfg.toml")
	if !ok {
		t.Fatal("entry missing after reload")
	}
	if e.Hash != "deadbeef" {
		t.Errorf("expected deadbeef, got %s", e.Hash)
	}
}

func TestNew_CorruptFile(t *testing.T) {
	path := tempStorePath(t)
	_ = os.WriteFile(path, []byte("not json{"), 0o644)
	_, err := New(path)
	if err == nil {
		t.Error("expected error on corrupt snapshot file")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s, _ := New(tempStorePath(t))
	_ = s.Set("/a", "h1")
	_ = s.Set("/b", "h2")
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
	// Mutating the copy should not affect the store
	delete(all, "/a")
	if _, ok := s.Get("/a"); !ok {
		t.Error("store was mutated via All() copy")
	}
}
