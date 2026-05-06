package watcher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/user/driftwatch/internal/config"
)

// FileState holds the last known hash of a watched file.
type FileState struct {
	Path    string
	LastHash string
}

// ChangeEvent represents a detected config file change.
type ChangeEvent struct {
	Path    string
	OldHash string
	NewHash string
	DetectedAt time.Time
}

// Watcher monitors a set of file paths for content changes.
type Watcher struct {
	cfg      *config.Config
	states   map[string]*FileState
	mu       sync.Mutex
	Changes  chan ChangeEvent
	stopCh   chan struct{}
}

// New creates a new Watcher from the given config.
func New(cfg *config.Config) *Watcher {
	return &Watcher{
		cfg:    cfg,
		states: make(map[string]*FileState),
		Changes: make(chan ChangeEvent, 16),
		stopCh:  make(chan struct{}),
	}
}

// Start begins polling watched paths at the configured interval.
func (w *Watcher) Start() {
	interval := time.Duration(w.cfg.PollIntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Seed initial hashes so first run doesn't false-positive.
	w.seedStates()

	for {
		select {
		case <-ticker.C:
			w.poll()
		case <-w.stopCh:
			return
		}
	}
}

// Stop signals the watcher to cease polling.
func (w *Watcher) Stop() {
	close(w.stopCh)
}

func (w *Watcher) seedStates() {
	for _, path := range w.cfg.WatchPaths {
		hash, _ := hashFile(path)
		w.states[path] = &FileState{Path: path, LastHash: hash}
	}
}

func (w *Watcher) poll() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, path := range w.cfg.WatchPaths {
		current, err := hashFile(path)
		if err != nil {
			continue
		}
		prev, known := w.states[path]
		if !known || prev.LastHash != current {
			oldHash := ""
			if known {
				oldHash = prev.LastHash
			}
			w.Changes <- ChangeEvent{
				Path:       path,
				OldHash:    oldHash,
				NewHash:    current,
				DetectedAt: time.Now().UTC(),
			}
			w.states[path] = &FileState{Path: path, LastHash: current}
		}
	}
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
