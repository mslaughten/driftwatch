// Package rollback provides snapshot-based config rollback suggestions
// when drift is detected on a watched path.
package rollback

import (
	"fmt"
	"sync"
	"time"
)

// Entry holds a saved snapshot of a file's content at a point in time.
type Entry struct {
	Path      string
	Content   []byte
	SavedAt   time.Time
	Hash      string
}

// Store keeps the last known-good content for each watched path.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		entries: make(map[string]Entry),
	}
}

// Save records content as the last known-good state for path.
func (s *Store) Save(path, hash string, content []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[path] = Entry{
		Path:    path,
		Content: append([]byte(nil), content...),
		SavedAt: time.Now().UTC(),
		Hash:    hash,
	}
}

// Get returns the stored entry for path, or an error if none exists.
func (s *Store) Get(path string) (Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[path]
	if !ok {
		return Entry{}, fmt.Errorf("rollback: no snapshot for %q", path)
	}
	return e, nil
}

// Remove deletes the stored entry for path.
func (s *Store) Remove(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, path)
}

// All returns a copy of every stored entry, keyed by path.
func (s *Store) All() map[string]Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]Entry, len(s.entries))
	for k, v := range s.entries {
		out[k] = v
	}
	return out
}
