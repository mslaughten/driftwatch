package alertlog

import (
	"encoding/json"
	"net/http"
)

// Handler returns an http.Handler that serves recent alert log entries as JSON.
func (l *Log) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		entries := l.Recent()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"count":   len(entries),
			"entries": entries,
		}); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}
