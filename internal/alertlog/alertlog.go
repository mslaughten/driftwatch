// Package alertlog provides an in-memory ring buffer for recording
// drift alert events, useful for surfacing recent alerts via the health endpoint.
package alertlog

import (
	"sync"
	"time"
)

// Entry represents a single recorded drift alert.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	FilePath  string    `json:"file_path"`
	WebhookURL string   `json:"webhook_url"`
	Success   bool      `json:"success"`
}

// Log is a thread-safe ring buffer of alert entries.
type Log struct {
	mu      sync.Mutex
	entries []Entry
	cap     int
}

// New creates a new Log with the given maximum capacity.
// When capacity is exceeded, the oldest entry is dropped.
func New(capacity int) *Log {
	if capacity <= 0 {
		capacity = 50
	}
	return &Log{
		entries: make([]Entry, 0, capacity),
		cap:     capacity,
	}
}

// Record appends a new alert entry to the log.
func (l *Log) Record(filePath, webhookURL string, success bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := Entry{
		Timestamp:  time.Now().UTC(),
		FilePath:   filePath,
		WebhookURL: webhookURL,
		Success:    success,
	}

	if len(l.entries) >= l.cap {
		l.entries = l.entries[1:]
	}
	l.entries = append(l.entries, entry)
}

// Recent returns a copy of all entries, oldest first.
func (l *Log) Recent() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	result := make([]Entry, len(l.entries))
	copy(result, l.entries)
	return result
}

// RecentFailures returns a copy of only the failed alert entries, oldest first.
func (l *Log) RecentFailures() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	var result []Entry
	for _, e := range l.entries {
		if !e.Success {
			result = append(result, e)
		}
	}
	return result
}

// Len returns the current number of entries.
func (l *Log) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}
