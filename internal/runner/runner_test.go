package runner

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/driftwatch/internal/config"
)

func makeConfig(t *testing.T, webhookURL string) *config.Config {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "watched-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	_ = f.Close()
	return &config.Config{
		WatchPaths:      []string{f.Name()},
		WebhookURL:      webhookURL,
		IntervalSeconds: 1,
	}
}

func TestRunner_StopsCleanly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := makeConfig(t, server.URL)
	r := New(cfg)

	done := make(chan struct{})
	go func() {
		r.Start()
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	r.Stop()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("runner did not stop in time")
	}
}

func TestRunner_DetectsDrift(t *testing.T) {
	notified := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notified <- r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := makeConfig(t, server.URL)
	r := New(cfg)

	// Seed initial hashes
	_, _ = r.watcher.CheckAll()

	// Mutate the watched file to trigger drift
	if err := os.WriteFile(cfg.WatchPaths[0], []byte("changed"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	r.poll()

	select {
	case <-notified:
		// webhook was called
	case <-time.After(2 * time.Second):
		t.Fatal("expected webhook notification, got none")
	}
}
