package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DriftEvent represents a config drift notification payload.
type DriftEvent struct {
	Service  string `json:"service"`
	FilePath string `json:"file_path"`
	OldHash  string `json:"old_hash"`
	NewHash  string `json:"new_hash"`
	Detected string `json:"detected_at"`
}

// Notifier sends drift events to a webhook endpoint.
type Notifier struct {
	webhookURL string
	client     *http.Client
}

// New creates a new Notifier with the given webhook URL.
func New(webhookURL string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send serializes and POSTs a DriftEvent to the configured webhook.
func (n *Notifier) Send(service, filePath, oldHash, newHash string) error {
	event := DriftEvent{
		Service:  service,
		FilePath: filePath,
		OldHash:  oldHash,
		NewHash:  newHash,
		Detected: time.Now().UTC().Format(time.RFC3339),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("notifier: failed to marshal event: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("notifier: webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notifier: webhook returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
