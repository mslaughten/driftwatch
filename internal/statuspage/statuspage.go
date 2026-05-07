package statuspage

import (
	"sync"
	"time"
)

// ServiceStatus represents the current drift status of a watched service.
type ServiceStatus struct {
	Path      string    `json:"path"`
	Drifted   bool      `json:"drifted"`
	LastCheck time.Time `json:"last_check"`
	LastDrift time.Time `json:"last_drift,omitempty"`
}

// Page tracks per-path service statuses.
type Page struct {
	mu       sync.RWMutex
	statuses map[string]*ServiceStatus
}

// New creates a new Page with the given watch paths pre-registered.
func New(paths []string) *Page {
	p := &Page{
		statuses: make(map[string]*ServiceStatus, len(paths)),
	}
	for _, path := range paths {
		p.statuses[path] = &ServiceStatus{Path: path}
	}
	return p
}

// RecordCheck updates the last check time for a path.
func (p *Page) RecordCheck(path string, drifted bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	s, ok := p.statuses[path]
	if !ok {
		s = &ServiceStatus{Path: path}
		p.statuses[path] = s
	}

	now := time.Now()
	s.LastCheck = now
	s.Drifted = drifted
	if drifted {
		s.LastDrift = now
	}
}

// Snapshot returns a copy of all current service statuses.
func (p *Page) Snapshot() []ServiceStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]ServiceStatus, 0, len(p.statuses))
	for _, s := range p.statuses {
		out = append(out, *s)
	}
	return out
}

// AnyDrifted returns true if at least one watched path is currently drifted.
func (p *Page) AnyDrifted() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, s := range p.statuses {
		if s.Drifted {
			return true
		}
	}
	return false
}
