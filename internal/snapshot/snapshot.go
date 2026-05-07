package snapshot

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry holds the last known hash and timestamp for a watched file.
type Entry struct {
	Hash      string    `json:"hash"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store persists file hashes to a JSON file on disk.
type Store struct {
	mu       sync.RWMutex
	path     string
	entries  map[string]Entry
}

// New loads an existing snapshot file or returns an empty store.
func New(path string) (*Store, error) {
	s := &Store{
		path:    path,
		entries: make(map[string]Entry),
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.entries); err != nil {
		return nil, err
	}
	return s, nil
}

// Get returns the stored entry for a file path, and whether it exists.
func (s *Store) Get(filePath string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[filePath]
	return e, ok
}

// Set updates the entry for a file path and persists the store.
func (s *Store) Set(filePath, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[filePath] = Entry{Hash: hash, UpdatedAt: time.Now().UTC()}
	return s.flush()
}

// flush writes the current entries to disk (must be called with lock held).
func (s *Store) flush() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// All returns a copy of all stored entries.
func (s *Store) All() map[string]Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copy := make(map[string]Entry, len(s.entries))
	for k, v := range s.entries {
		copy[k] = v
	}
	return copy
}
