package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSend_Success(t *testing.T) {
	var received Payload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := New(ts.URL, 10, 5)
	if err := n.Send("/etc/app/config.yaml"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.File != "/etc/app/config.yaml" {
		t.Errorf("expected file field, got %q", received.File)
	}
	if received.Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestSend_Non2xxResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := New(ts.URL, 10, 5)
	err := n.Send("/etc/app/config.yaml")
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestSend_InvalidURL(t *testing.T) {
	n := New("http://127.0.0.1:0/nope", 10, 5)
	if err := n.Send("file.yaml"); err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestSend_RateLimitDrops(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// max=2, rate=0 so no refill
	n := New(ts.URL, 2, 0)
	for i := 0; i < 5; i++ {
		_ = n.Send("file.yaml")
	}
	if calls != 2 {
		t.Errorf("expected 2 webhook calls (rate limited), got %d", calls)
	}
}
