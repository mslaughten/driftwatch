package rollback

import (
	"encoding/json"
	"net/http"
	"strings"
)

type responseEntry struct {
	Path    string `json:"path"`
	Hash    string `json:"hash"`
	SavedAt string `json:"saved_at"`
}

// Handler exposes rollback snapshot data over HTTP.
// GET  /rollback        — list all stored snapshots (metadata only)
// GET  /rollback?path=  — return content for a specific path
func Handler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		path := strings.TrimSpace(r.URL.Query().Get("path"))
		if path != "" {
			e, err := s.Get(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.Write(e.Content) //nolint:errcheck
			return
		}

		all := s.All()
		out := make([]responseEntry, 0, len(all))
		for _, e := range all {
			out = append(out, responseEntry{
				Path:    e.Path,
				Hash:    e.Hash,
				SavedAt: e.SavedAt.Format("2006-01-02T15:04:05Z"),
			})
		}
		json.NewEncoder(w).Encode(out) //nolint:errcheck
	}
}
