package silences

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_GET_Empty(t *testing.T) {
	store := New()
	h := Handler(store)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/silences", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out []Silence
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty list, got %d items", len(out))
	}
}

func TestHandler_POST_Add(t *testing.T) {
	store := New()
	h := Handler(store)
	body := `{"path":"/etc/app.yaml","reason":"deploy","duration":"30m"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/silences", bytes.NewBufferString(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if !store.IsSilenced("/etc/app.yaml") {
		t.Error("expected path to be silenced after POST")
	}
}

func TestHandler_POST_MissingPath(t *testing.T) {
	store := New()
	h := Handler(store)
	body := `{"reason":"test","duration":"10m"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/silences", bytes.NewBufferString(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_DELETE_Remove(t *testing.T) {
	store := New()
	store.Add("/etc/x", "r", 999999)
	h := Handler(store)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/silences?path=%2Fetc%2Fx", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if store.IsSilenced("/etc/x") {
		t.Error("expected silence to be removed")
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	store := New()
	h := Handler(store)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/silences", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
