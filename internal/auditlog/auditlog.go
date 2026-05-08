// Package auditlog records configuration drift events and suppression actions
// for operator review and compliance purposes.
package auditlog

import (
	"sync"
	"time"
)

// EventKind categorises an audit event.
type EventKind string

const (
	KindDrift      EventKind = "drift"
	KindSuppressed EventKind = "suppressed"
	KindSilenced   EventKind = "silenced"
	KindWebhookOK  EventKind = "webhook_ok"
	KindWebhookErr EventKind = "webhook_error"

	defaultCapacity = 200
)

// Entry is a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Kind      EventKind `json:"kind"`
	Path      string    `json:"path,omitempty"`
	Message   string    `json:"message,omitempty"`
}

// Log is a fixed-capacity ring-buffer of audit entries.
type Log struct {
	mu       sync.Mutex
	buf      []Entry
	head     int
	size     int
	capacity int
}

// New creates a Log with the given capacity (defaults to defaultCapacity if <= 0).
func New(capacity int) *Log {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	return &Log{
		buf:      make([]Entry, capacity),
		capacity: capacity,
	}
}

// Record appends an entry to the log, evicting the oldest if full.
func (l *Log) Record(kind EventKind, path, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	e := Entry{
		Timestamp: time.Now().UTC(),
		Kind:      kind,
		Path:      path,
		Message:   message,
	}
	l.buf[l.head] = e
	l.head = (l.head + 1) % l.capacity
	if l.size < l.capacity {
		l.size++
	}
}

// Recent returns up to n most-recent entries, oldest first.
func (l *Log) Recent(n int) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if n <= 0 || l.size == 0 {
		return nil
	}
	if n > l.size {
		n = l.size
	}
	out := make([]Entry, n)
	start := (l.head - n + l.capacity) % l.capacity
	for i := 0; i < n; i++ {
		out[i] = l.buf[(start+i)%l.capacity]
	}
	return out
}
