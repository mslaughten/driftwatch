// Package throttle provides per-path alert throttling to prevent
// notification storms when a config file changes repeatedly.
package throttle

import (
	"sync"
	"time"
)

// Throttle tracks the last alert time per watched path and suppresses
// duplicate alerts that arrive within a configurable cooldown window.
type Throttle struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[string]time.Time
	now      func() time.Time
}

// New creates a Throttle with the given cooldown duration.
func New(cooldown time.Duration) *Throttle {
	return &Throttle{
		cooldown: cooldown,
		last:     make(map[string]time.Time),
		now:      time.Now,
	}
}

// Allow returns true if an alert for path should be sent, and records
// the current time as the last alert time. Returns false if the path
// is still within the cooldown window.
func (t *Throttle) Allow(path string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	if last, ok := t.last[path]; ok {
		if now.Sub(last) < t.cooldown {
			return false
		}
	}
	t.last[path] = now
	return true
}

// Reset clears the throttle state for a specific path, allowing the
// next alert to pass through immediately.
func (t *Throttle) Reset(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.last, path)
}

// Snapshot returns a copy of the current last-alert times keyed by path.
func (t *Throttle) Snapshot() map[string]time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make(map[string]time.Time, len(t.last))
	for k, v := range t.last {
		out[k] = v
	}
	return out
}
