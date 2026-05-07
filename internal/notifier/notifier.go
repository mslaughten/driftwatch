// Package notifier sends drift-alert payloads to a configured webhook URL.
package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/driftwatch/internal/ratelimit"
)

// Payload is the JSON body posted to the webhook.
type Payload struct {
	File      string    `json:"file"`
	ChangedAt time.Time `json:"changed_at"`
	Message   string    `json:"message"`
}

// Notifier posts drift alerts to a webhook endpoint.
type Notifier struct {
	webhookURL string
	client     *http.Client
	limiter    *ratelimit.Limiter
}

// New creates a Notifier for the given webhook URL.
// max and ratePerSec configure the built-in rate limiter.
func New(webhookURL string, max float64, ratePerSec float64) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
		limiter:    ratelimit.New(max, ratePerSec),
	}
}

// Send posts a drift alert for the given file. It returns an error if
// the rate limit is exceeded, the request fails, or the server responds
// with a non-2xx status code.
func (n *Notifier) Send(file string) error {
	if !n.limiter.Allow() {
		return fmt.Errorf("rate limit exceeded, alert for %q dropped", file)
	}

	p := Payload{
		File:      file,
		ChangedAt: time.Now().UTC(),
		Message:   fmt.Sprintf("config drift detected in %s", file),
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("post webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}
	return nil
}
