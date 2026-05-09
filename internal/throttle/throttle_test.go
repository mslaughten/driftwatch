package throttle

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_FirstCallAllowed(t *testing.T) {
	th := New(5 * time.Second)
	if !th.Allow("/etc/app/config.yaml") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_WithinCooldownBlocked(t *testing.T) {
	base := time.Now()
	th := New(10 * time.Second)
	th.now = fixedNow(base)
	th.Allow("/etc/app/config.yaml")

	th.now = fixedNow(base.Add(5 * time.Second))
	if th.Allow("/etc/app/config.yaml") {
		t.Fatal("expected call within cooldown to be blocked")
	}
}

func TestAllow_AfterCooldownAllowed(t *testing.T) {
	base := time.Now()
	th := New(10 * time.Second)
	th.now = fixedNow(base)
	th.Allow("/etc/app/config.yaml")

	th.now = fixedNow(base.Add(11 * time.Second))
	if !th.Allow("/etc/app/config.yaml") {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestReset_AllowsImmediately(t *testing.T) {
	base := time.Now()
	th := New(30 * time.Second)
	th.now = fixedNow(base)
	th.Allow("/etc/app/config.yaml")

	th.Reset("/etc/app/config.yaml")
	th.now = fixedNow(base.Add(1 * time.Second))
	if !th.Allow("/etc/app/config.yaml") {
		t.Fatal("expected reset path to be allowed immediately")
	}
}

func TestSnapshot_ReturnsCopy(t *testing.T) {
	th := New(5 * time.Second)
	th.Allow("/a")
	th.Allow("/b")

	snap := th.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	delete(snap, "/a")
	if len(th.Snapshot()) != 2 {
		t.Fatal("snapshot mutation affected internal state")
	}
}

func TestHandler_GET_ReturnsJSON(t *testing.T) {
	th := New(5 * time.Second)
	th.Allow("/etc/nginx/nginx.conf")

	req := httptest.NewRequest(http.MethodGet, "/throttle", nil)
	rec := httptest.NewRecorder()
	Handler(th)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []throttleEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(entries) != 1 || entries[0].Path != "/etc/nginx/nginx.conf" {
		t.Fatalf("unexpected entries: %+v", entries)
	}
}

func TestHandler_DELETE_ResetPath(t *testing.T) {
	th := New(30 * time.Second)
	th.Allow("/etc/app.conf")

	req := httptest.NewRequest(http.MethodDelete, "/throttle?path=/etc/app.conf", nil)
	rec := httptest.NewRecorder()
	Handler(th)(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if len(th.Snapshot()) != 0 {
		t.Fatal("expected path to be removed after reset")
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	th := New(5 * time.Second)
	req := httptest.NewRequest(http.MethodPost, "/throttle", nil)
	rec := httptest.NewRecorder()
	Handler(th)(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
