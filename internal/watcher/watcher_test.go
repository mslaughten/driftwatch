package watcher

import (
	"os"
	"testing"
	"time"

	"github.com/user/driftwatch/internal/config"
)

func tempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "driftwatch-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func makeConfig(paths []string) *config.Config {
	return &config.Config{
		WatchPaths:          paths,
		WebhookURL:          "http://localhost/webhook",
		PollIntervalSeconds: 1,
	}
}

func TestHashFile_Consistent(t *testing.T) {
	path := tempFile(t, "hello world")
	h1, err := hashFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h2, err := hashFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h1 != h2 {
		t.Errorf("expected identical hashes, got %s and %s", h1, h2)
	}
}

func TestHashFile_DifferentContent(t *testing.T) {
	p1 := tempFile(t, "content A")
	p2 := tempFile(t, "content B")
	h1, _ := hashFile(p1)
	h2, _ := hashFile(p2)
	if h1 == h2 {
		t.Error("expected different hashes for different content")
	}
}

func TestHashFile_Missing(t *testing.T) {
	_, err := hashFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestWatcher_DetectsChange(t *testing.T) {
	path := tempFile(t, "initial content")
	cfg := makeConfig([]string{path})

	w := New(cfg)
	w.seedStates()

	// Modify the file.
	if err := os.WriteFile(path, []byte("modified content"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	w.poll()

	select {
	case event := <-w.Changes:
		if event.Path != path {
			t.Errorf("expected path %s, got %s", path, event.Path)
		}
		if event.OldHash == event.NewHash {
			t.Error("expected old and new hashes to differ")
		}
		if event.DetectedAt.IsZero() {
			t.Error("expected non-zero DetectedAt timestamp")
		}
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for change event")
	}
}

func TestWatcher_NoChangeNoEvent(t *testing.T) {
	path := tempFile(t, "stable content")
	cfg := makeConfig([]string{path})

	w := New(cfg)
	w.seedStates()
	w.poll()

	select {
	case event := <-w.Changes:
		t.Errorf("unexpected change event for unchanged file: %+v", event)
	default:
		// expected: no event
	}
}
