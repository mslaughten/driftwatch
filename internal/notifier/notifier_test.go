package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSend_Success(t *testing.T) {
	var received DriftEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := New(server.URL)
	err := n.Send("auth-service", "/etc/auth/config.yaml", "abc123", "def456")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if received.Service != "auth-service" {
		t.Errorf("expected service 'auth-service', got %q", received.Service)
	}
	if received.OldHash != "abc123" {
		t.Errorf("expected old hash 'abc123', got %q", received.OldHash)
	}
	if received.NewHash != "def456" {
		t.Errorf("expected new hash 'def456', got %q", received.NewHash)
	}
	if received.Detected == "" {
		t.Error("expected detected_at to be set")
	}
}

func TestSend_Non2xxResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := New(server.URL)
	err := n.Send("svc", "/etc/svc/config.yaml", "aaa", "bbb")
	if err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
}

func TestSend_InvalidURL(t *testing.T) {
	n := New("http://127.0.0.1:0/no-listener")
	err := n.Send("svc", "/etc/svc/config.yaml", "aaa", "bbb")
	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}
