// Package metrics provides lightweight in-process counters for
// tracking drift detections, webhook deliveries, and errors.
package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds atomic counters for key driftwatch events.
type Metrics struct {
	DriftDetections  atomic.Int64
	WebhooksSent     atomic.Int64
	WebhookErrors    atomic.Int64
	RateLimitDrops   atomic.Int64

	mu        sync.RWMutex
	startedAt time.Time
}

// New returns a new Metrics instance with the start time recorded.
func New() *Metrics {
	return &Metrics{
		startedAt: time.Now(),
	}
}

// RecordDrift increments the drift detection counter.
func (m *Metrics) RecordDrift() {
	m.DriftDetections.Add(1)
}

// RecordWebhookSent increments the successful webhook delivery counter.
func (m *Metrics) RecordWebhookSent() {
	m.WebhooksSent.Add(1)
}

// RecordWebhookError increments the webhook error counter.
func (m *Metrics) RecordWebhookError() {
	m.WebhookErrors.Add(1)
}

// RecordRateLimitDrop increments the rate-limit drop counter.
func (m *Metrics) RecordRateLimitDrop() {
	m.RateLimitDrops.Add(1)
}

// Snapshot returns a point-in-time copy of all counters.
func (m *Metrics) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return Snapshot{
		DriftDetections: m.DriftDetections.Load(),
		WebhooksSent:    m.WebhooksSent.Load(),
		WebhookErrors:   m.WebhookErrors.Load(),
		RateLimitDrops:  m.RateLimitDrops.Load(),
		UptimeSeconds:   int64(time.Since(m.startedAt).Seconds()),
	}
}

// Snapshot is a value copy of metrics at a point in time.
type Snapshot struct {
	DriftDetections int64 `json:"drift_detections"`
	WebhooksSent    int64 `json:"webhooks_sent"`
	WebhookErrors   int64 `json:"webhook_errors"`
	RateLimitDrops  int64 `json:"rate_limit_drops"`
	UptimeSeconds   int64 `json:"uptime_seconds"`
}
