// Package healthcheck provides an HTTP health endpoint for driftwatch.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// Status holds the current health state of the daemon.
type Status struct {
	Healthy    bool      `json:"healthy"`
	Uptime     string    `json:"uptime"`
	WatchedFiles int     `json:"watched_files"`
	LastCheck  time.Time `json:"last_check"`
}

// Handler serves health status over HTTP.
type Handler struct {
	startTime    time.Time
	watchedFiles atomic.Int32
	lastCheck    atomic.Value // stores time.Time
}

// New creates a new Handler.
func New() *Handler {
	h := &Handler{
		startTime: time.Now(),
	}
	h.lastCheck.Store(time.Time{})
	return h
}

// SetWatchedFiles updates the count of actively watched files.
func (h *Handler) SetWatchedFiles(n int) {
	h.watchedFiles.Store(int32(n))
}

// RecordCheck records the time of the most recent drift check.
func (h *Handler) RecordCheck() {
	h.lastCheck.Store(time.Now())
}

// ServeHTTP handles GET /healthz requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lc, _ := h.lastCheck.Load().(time.Time)
	s := Status{
		Healthy:      true,
		Uptime:       time.Since(h.startTime).Round(time.Second).String(),
		WatchedFiles: int(h.watchedFiles.Load()),
		LastCheck:    lc,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(s)
}
