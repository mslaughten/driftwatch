package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_ZeroCounters(t *testing.T) {
	m := New()
	snap := m.Snapshot()
	if snap.DriftDetections != 0 || snap.WebhooksSent != 0 ||
		snap.WebhookErrors != 0 || snap.RateLimitDrops != 0 {
		t.Fatal("expected all counters to be zero on init")
	}
}

func TestRecordDrift(t *testing.T) {
	m := New()
	m.RecordDrift()
	m.RecordDrift()
	if got := m.Snapshot().DriftDetections; got != 2 {
		t.Fatalf("expected 2 drift detections, got %d", got)
	}
}

func TestRecordWebhookSent(t *testing.T) {
	m := New()
	m.RecordWebhookSent()
	if got := m.Snapshot().WebhooksSent; got != 1 {
		t.Fatalf("expected 1 webhook sent, got %d", got)
	}
}

func TestRecordWebhookError(t *testing.T) {
	m := New()
	m.RecordWebhookError()
	if got := m.Snapshot().WebhookErrors; got != 1 {
		t.Fatalf("expected 1 webhook error, got %d", got)
	}
}

func TestRecordRateLimitDrop(t *testing.T) {
	m := New()
	m.RecordRateLimitDrop()
	m.RecordRateLimitDrop()
	m.RecordRateLimitDrop()
	if got := m.Snapshot().RateLimitDrops; got != 3 {
		t.Fatalf("expected 3 rate-limit drops, got %d", got)
	}
}

func TestSnapshot_UptimeNonNegative(t *testing.T) {
	m := New()
	if m.Snapshot().UptimeSeconds < 0 {
		t.Fatal("uptime should be non-negative")
	}
}

func TestHandler_ReturnsJSON(t *testing.T) {
	m := New()
	m.RecordDrift()
	m.RecordWebhookSent()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	m.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %s", ct)
	}

	var snap Snapshot
	if err := json.NewDecoder(rr.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if snap.DriftDetections != 1 {
		t.Fatalf("expected 1 drift detection in response, got %d", snap.DriftDetections)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	m := New()
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rr := httptest.NewRecorder()
	m.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
