package throttle

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"
)

type throttleEntry struct {
	Path      string    `json:"path"`
	LastAlert time.Time `json:"last_alert"`
	Cooldown  string    `json:"cooldown"`
}

// Handler returns an http.HandlerFunc that exposes the current throttle
// state as JSON. Supports GET to list all entries and DELETE /{path} to
// reset a specific path's cooldown.
func Handler(t *Throttle) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			snap := t.Snapshot()
			entries := make([]throttleEntry, 0, len(snap))
			for path, last := range snap {
				entries = append(entries, throttleEntry{
					Path:      path,
					LastAlert: last,
					Cooldown:  t.cooldown.String(),
				})
			}
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Path < entries[j].Path
			})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entries)

		case http.MethodDelete:
			path := r.URL.Query().Get("path")
			if path == "" {
				http.Error(w, "missing path query param", http.StatusBadRequest)
				return
			}
			t.Reset(path)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
