package auditlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew_DefaultCapacity(t *testing.T) {
	l := New(0)
	if l.capacity != defaultCapacity {
		t.Fatalf("expected capacity %d, got %d", defaultCapacity, l.capacity)
	}
}

func TestRecord_AddsEntry(t *testing.T) {
	l := New(10)
	l.Record(KindDrift, "/etc/app.conf", "hash changed")
	entries := l.Recent(10)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Kind != KindDrift {
		t.Errorf("expected kind %q, got %q", KindDrift, entries[0].Kind)
	}
	if entries[0].Path != "/etc/app.conf" {
		t.Errorf("unexpected path: %s", entries[0].Path)
	}
}

func TestRecord_TimestampSet(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	l := New(5)
	l.Record(KindWebhookOK, "", "sent")
	after := time.Now().UTC().Add(time.Second)
	entries := l.Recent(1)
	if entries[0].Timestamp.Before(before) || entries[0].Timestamp.After(after) {
		t.Errorf("timestamp out of range: %v", entries[0].Timestamp)
	}
}

func TestRecord_RingBufferEvictsOldest(t *testing.T) {
	l := New(3)
	l.Record(KindDrift, "a", "")
	l.Record(KindDrift, "b", "")
	l.Record(KindDrift, "c", "")
	l.Record(KindDrift, "d", "") // evicts "a"

	entries := l.Recent(10)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Path != "b" {
		t.Errorf("expected oldest to be 'b', got %q", entries[0].Path)
	}
	if entries[2].Path != "d" {
		t.Errorf("expected newest to be 'd', got %q", entries[2].Path)
	}
}

func TestRecent_EmptyLog(t *testing.T) {
	l := New(10)
	if got := l.Recent(5); got != nil {
		t.Errorf("expected nil for empty log, got %v", got)
	}
}

func TestHandler_ReturnsJSON(t *testing.T) {
	l := New(10)
	l.Record(KindSilenced, "/etc/nginx.conf", "silenced by operator")

	req := httptest.NewRequest(http.MethodGet, "/auditlog", nil)
	rec := httptest.NewRecorder()
	Handler(l)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	l := New(10)
	req := httptest.NewRequest(http.MethodPost, "/auditlog", nil)
	rec := httptest.NewRecorder()
	Handler(l)(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_LimitQueryParam(t *testing.T) {
	l := New(20)
	for i := 0; i < 10; i++ {
		l.Record(KindDrift, "/path", "changed")
	}
	req := httptest.NewRequest(http.MethodGet, "/auditlog?limit=3", nil)
	rec := httptest.NewRecorder()
	Handler(l)(rec, req)

	var entries []Entry
	_ = json.NewDecoder(rec.Body).Decode(&entries)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries with limit=3, got %d", len(entries))
	}
}
