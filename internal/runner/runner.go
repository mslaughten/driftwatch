package runner

import (
	"log"
	"time"

	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/notifier"
	"github.com/driftwatch/internal/watcher"
)

// Runner orchestrates the watch loop, polling for file changes
// and dispatching webhook notifications on drift detection.
type Runner struct {
	cfg     *config.Config
	watcher *watcher.Watcher
	notify  *notifier.Notifier
	stop    chan struct{}
}

// New creates a Runner from the provided config.
func New(cfg *config.Config) *Runner {
	return &Runner{
		cfg:    cfg,
		watcher: watcher.New(cfg),
		notify:  notifier.New(cfg.WebhookURL),
		stop:   make(chan struct{}),
	}
}

// Start begins the polling loop. It blocks until Stop is called.
func (r *Runner) Start() {
	interval := time.Duration(r.cfg.IntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("driftwatch started: polling every %s", interval)

	for {
		select {
		case <-ticker.C:
			r.poll()
		case <-r.stop:
			log.Println("driftwatch stopped")
			return
		}
	}
}

// Stop signals the runner to cease polling.
func (r *Runner) Stop() {
	close(r.stop)
}

func (r *Runner) poll() {
	changed, err := r.watcher.CheckAll()
	if err != nil {
		log.Printf("watcher error: %v", err)
		return
	}
	for _, path := range changed {
		log.Printf("drift detected: %s", path)
		if err := r.notify.Send(path); err != nil {
			log.Printf("notification failed for %s: %v", path, err)
		}
	}
}
