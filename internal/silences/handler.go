package silences

import (
	"encoding/json"
	"net/http"
	"time"
)

type addRequest struct {
	Path     string `json:"path"`
	Reason   string `json:"reason"`
	Duration string `json:"duration"`
}

// Handler returns an http.Handler that exposes GET /silences and
// POST /silences (add) and DELETE /silences?path=... endpoints.
func Handler(store *Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/silences", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			active := store.Active()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(active) //nolint:errcheck

		case http.MethodPost:
			var req addRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if req.Path == "" {
				http.Error(w, "path is required", http.StatusBadRequest)
				return
			}
			d, err := time.ParseDuration(req.Duration)
			if err != nil || d <= 0 {
				http.Error(w, "invalid duration", http.StatusBadRequest)
				return
			}
			store.Add(req.Path, req.Reason, d)
			w.WriteHeader(http.StatusCreated)

		case http.MethodDelete:
			path := r.URL.Query().Get("path")
			if path == "" {
				http.Error(w, "path query param required", http.StatusBadRequest)
				return
			}
			store.Remove(path)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
