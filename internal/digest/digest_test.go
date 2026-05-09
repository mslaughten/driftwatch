package digest_test

import (
	"testing"

	"github.com/yourusername/driftwatch/internal/digest"
)

func TestHash_Consistent(t *testing.T) {
	a := digest.Hash([]byte("hello"))
	b := digest.Hash([]byte("hello"))
	if a != b {
		t.Fatalf("expected identical hashes, got %q and %q", a, b)
	}
}

func TestHash_DifferentContent(t *testing.T) {
	a := digest.Hash([]byte("hello"))
	b := digest.Hash([]byte("world"))
	if a == b {
		t.Fatal("expected different hashes for different content")
	}
}

func TestIsDuplicate_FirstCallNotDuplicate(t *testing.T) {
	r := digest.New()
	if r.IsDuplicate("/etc/app.conf", []byte("content")) {
		t.Fatal("first call should never be a duplicate")
	}
}

func TestIsDuplicate_SamePayloadIsDuplicate(t *testing.T) {
	r := digest.New()
	payload := []byte(`{"path":"/etc/app.conf"}`)
	r.IsDuplicate("/etc/app.conf", payload) // seed
	if !r.IsDuplicate("/etc/app.conf", payload) {
		t.Fatal("second call with same payload should be duplicate")
	}
}

func TestIsDuplicate_ChangedPayloadNotDuplicate(t *testing.T) {
	r := digest.New()
	r.IsDuplicate("/etc/app.conf", []byte("v1"))
	if r.IsDuplicate("/etc/app.conf", []byte("v2")) {
		t.Fatal("changed payload should not be a duplicate")
	}
}

func TestIsDuplicate_IndependentPaths(t *testing.T) {
	r := digest.New()
	payload := []byte("same")
	r.IsDuplicate("/a", payload)
	// /b has never been seen — must not be duplicate even though payload matches /a
	if r.IsDuplicate("/b", payload) {
		t.Fatal("different path should be treated independently")
	}
}

func TestReset_ForcesNextCallFresh(t *testing.T) {
	r := digest.New()
	payload := []byte("data")
	r.IsDuplicate("/etc/app.conf", payload) // seed
	r.Reset("/etc/app.conf")
	if r.IsDuplicate("/etc/app.conf", payload) {
		t.Fatal("after Reset, same payload should not be a duplicate")
	}
}

func TestLen_TracksEntries(t *testing.T) {
	r := digest.New()
	if r.Len() != 0 {
		t.Fatalf("expected 0, got %d", r.Len())
	}
	r.IsDuplicate("/a", []byte("x"))
	r.IsDuplicate("/b", []byte("y"))
	if r.Len() != 2 {
		t.Fatalf("expected 2, got %d", r.Len())
	}
	r.Reset("/a")
	if r.Len() != 1 {
		t.Fatalf("expected 1 after reset, got %d", r.Len())
	}
}
