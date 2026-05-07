package statuspage

import (
	"encoding/json"
	"net/http"
	"sort"
)

type response struct {
	Healthy  bool            `json:"healthy"`
	Services []ServiceStatus `json:"services"`
}

// Handler returns an http.Handler that serves the current status page as JSON.
func (p *Page) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		services := p.Snapshot()
		sort.Slice(services, func(i, j int) bool {
			return services[i].Path < services[j].Path
		})

		resp := response{
			Healthy:  !p.AnyDrifted(),
			Services: services,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	})
}
