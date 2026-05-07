package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/healthcheck"
)

func TestServeHTTP_ReturnsOK(t *testing.T) {
	h := healthcheck.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestServeHTTP_JSONBody(t *testing.T) {
	h := healthcheck.New()
	h.SetWatchedFiles(3)
	h.RecordCheck()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	var s healthcheck.Status
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !s.Healthy {
		t.Error("expected healthy=true")
	}
	if s.WatchedFiles != 3 {
		t.Errorf("expected watched_files=3, got %d", s.WatchedFiles)
	}
	if s.LastCheck.IsZero() {
		t.Error("expected last_check to be set")
	}
}

func TestServeHTTP_UptimeNonEmpty(t *testing.T) {
	h := healthcheck.New()
	time.Sleep(10 * time.Millisecond)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	var s healthcheck.Status
	_ = json.NewDecoder(rec.Body).Decode(&s)
	if s.Uptime == "" {
		t.Error("expected non-empty uptime")
	}
}

func TestRecordCheck_UpdatesLastCheck(t *testing.T) {
	h := healthcheck.New()
	before := time.Now()
	h.RecordCheck()
	after := time.Now()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	var s healthcheck.Status
	_ = json.NewDecoder(rec.Body).Decode(&s)
	if s.LastCheck.Before(before) || s.LastCheck.After(after) {
		t.Errorf("last_check %v not in expected range [%v, %v]", s.LastCheck, before, after)
	}
}
