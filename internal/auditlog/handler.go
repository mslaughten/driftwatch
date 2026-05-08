package auditlog

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const defaultQueryLimit = 50

// Handler returns an http.HandlerFunc that serves recent audit log entries
// as JSON. Accepts an optional ?limit=N query parameter.
func Handler(l *Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		limit := defaultQueryLimit
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if v, err := strconv.Atoi(raw); err == nil && v > 0 {
				limit = v
			}
		}

		entries := l.Recent(limit)
		if entries == nil {
			entries = []Entry{}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(entries)
	}
}
