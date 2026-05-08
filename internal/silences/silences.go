// Package silences provides time-bounded muting of drift alerts for specific paths.
package silences

import (
	"sync"
	"time"
)

// Silence represents a muted path with an expiry.
type Silence struct {
	Path      string    `json:"path"`
	ExpiresAt time.Time `json:"expires_at"`
	Reason    string    `json:"reason"`
}

// Store holds active silences.
type Store struct {
	mu       sync.RWMutex
	silences map[string]Silence
	now      func() time.Time
}

// New returns a new silence Store.
func New() *Store {
	return &Store{
		silences: make(map[string]Silence),
		now:      time.Now,
	}
}

// Add creates or overwrites a silence for path lasting duration d.
func (s *Store) Add(path, reason string, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.silences[path] = Silence{
		Path:      path,
		ExpiresAt: s.now().Add(d),
		Reason:    reason,
	}
}

// Remove lifts a silence immediately.
func (s *Store) Remove(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.silences, path)
}

// IsSilenced reports whether path is currently silenced.
func (s *Store) IsSilenced(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sl, ok := s.silences[path]
	if !ok {
		return false
	}
	if s.now().After(sl.ExpiresAt) {
		return false
	}
	return true
}

// Active returns all non-expired silences.
func (s *Store) Active() []Silence {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	out := make([]Silence, 0, len(s.silences))
	for _, sl := range s.silences {
		if now.Before(sl.ExpiresAt) {
			out = append(out, sl)
		}
	}
	return out
}

// Purge removes all expired silences.
func (s *Store) Purge() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	for path, sl := range s.silences {
		if now.After(sl.ExpiresAt) {
			delete(s.silences, path)
		}
	}
}
