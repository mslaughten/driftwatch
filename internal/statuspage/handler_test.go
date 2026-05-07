package statuspage

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_ReturnsJSON(t *testing.T) {
	p := New([]string{"/etc/app/config.yaml"})
	p.RecordCheck("/etc/app/config.yaml", false)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	p.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestHandler_HealthyWhenNoDrift(t *testing.T) {
	p := New([]string{"/etc/app/config.yaml"})
	p.RecordCheck("/etc/app/config.yaml", false)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	p.Handler().ServeHTTP(rec, req)

	var resp response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Healthy {
		t.Error("expected healthy=true when no drift")
	}
}

func TestHandler_UnhealthyWhenDrifted(t *testing.T) {
	p := New([]string{"/etc/app/config.yaml"})
	p.RecordCheck("/etc/app/config.yaml", true)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	p.Handler().ServeHTTP(rec, req)

	var resp response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Healthy {
		t.Error("expected healthy=false when drifted")
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	p := New(nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/status", nil)
	p.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_ServicesAreSorted(t *testing.T) {
	p := New([]string{"/z/path", "/a/path", "/m/path"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	p.Handler().ServeHTTP(rec, req)

	var resp response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Services) != 3 {
		t.Fatalf("expected 3 services, got %d", len(resp.Services))
	}
	if resp.Services[0].Path != "/a/path" {
		t.Errorf("expected first service to be /a/path, got %s", resp.Services[0].Path)
	}
}
