package alertlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_ReturnsJSON(t *testing.T) {
	l := New(10)
	l.Record("/etc/app/config.yaml", "abc123", "def456")

	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	rec := httptest.NewRecorder()

	l.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["count"].(float64) != 1 {
		t.Errorf("expected count 1, got %v", body["count"])
	}
}

func TestHandler_EmptyLog(t *testing.T) {
	l := New(10)

	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	rec := httptest.NewRecorder()

	l.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["count"].(float64) != 0 {
		t.Errorf("expected count 0, got %v", body["count"])
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	l := New(10)

	req := httptest.NewRequest(http.MethodPost, "/alerts", nil)
	rec := httptest.NewRecorder()

	l.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_ContentTypeHeader(t *testing.T) {
	l := New(5)
	l.Record("/etc/svc/settings.yaml", "aaa", "bbb")
	_ = time.Now() // ensure time package used

	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	rec := httptest.NewRecorder()

	l.Handler().ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
