// Package suppressions provides a mechanism to temporarily silence drift alerts
// for specific file paths, useful during planned maintenance windows.
package suppressions

import (
	"sync"
	"time"
)

// Suppression represents a single suppression rule for a file path.
type Suppression struct {
	Path      string
	ExpiresAt time.Time
}

// Store holds active suppressions keyed by file path.
type Store struct {
	mu          sync.RWMutex
	suppressions map[string]time.Time
	now         func() time.Time
}

// New creates a new suppression Store.
func New() *Store {
	return &Store{
		suppressions: make(map[string]time.Time),
		now:         time.Now,
	}
}

// Suppress silences alerts for the given path until the duration elapses.
func (s *Store) Suppress(path string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.suppressions[path] = s.now().Add(duration)
}

// IsSuppressed reports whether the given path currently has an active suppression.
func (s *Store) IsSuppressed(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	expiry, ok := s.suppressions[path]
	if !ok {
		return false
	}
	return s.now().Before(expiry)
}

// Remove explicitly lifts a suppression for the given path.
func (s *Store) Remove(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.suppressions, path)
}

// Active returns a snapshot of all currently active suppressions.
func (s *Store) Active() []Suppression {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	result := make([]Suppression, 0, len(s.suppressions))
	for path, expiry := range s.suppressions {
		if now.Before(expiry) {
			result = append(result, Suppression{Path: path, ExpiresAt: expiry})
		}
	}
	return result
}
