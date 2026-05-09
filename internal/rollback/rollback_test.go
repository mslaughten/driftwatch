package rollback_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"driftwatch/internal/rollback"
)

func TestSave_And_Get(t *testing.T) {
	s := rollback.New()
	s.Save("/etc/app.yaml", "abc123", []byte("key: value"))

	e, err := s.Get("/etc/app.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Hash != "abc123" {
		t.Errorf("expected hash abc123, got %s", e.Hash)
	}
	if string(e.Content) != "key: value" {
		t.Errorf("unexpected content: %s", e.Content)
	}
}

func TestGet_Missing(t *testing.T) {
	s := rollback.New()
	_, err := s.Get("/nonexistent")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestRemove_DeletesEntry(t *testing.T) {
	s := rollback.New()
	s.Save("/etc/app.yaml", "abc123", []byte("data"))
	s.Remove("/etc/app.yaml")
	_, err := s.Get("/etc/app.yaml")
	if err == nil {
		t.Fatal("expected error after removal")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := rollback.New()
	s.Save("/a", "h1", []byte("a"))
	s.Save("/b", "h2", []byte("b"))
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestHandler_ListAll(t *testing.T) {
	s := rollback.New()
	s.Save("/etc/app.yaml", "deadbeef", []byte("content"))

	req := httptest.NewRequest(http.MethodGet, "/rollback", nil)
	w := httptest.NewRecorder()
	rollback.Handler(s)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var entries []map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0]["hash"] != "deadbeef" {
		t.Errorf("unexpected hash: %s", entries[0]["hash"])
	}
}

func TestHandler_GetContent(t *testing.T) {
	s := rollback.New()
	s.Save("/etc/app.yaml", "h1", []byte("key: value"))

	req := httptest.NewRequest(http.MethodGet, "/rollback?path=/etc/app.yaml", nil)
	w := httptest.NewRecorder()
	rollback.Handler(s)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "key: value" {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	s := rollback.New()
	req := httptest.NewRequest(http.MethodPost, "/rollback", nil)
	w := httptest.NewRecorder()
	rollback.Handler(s)(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
