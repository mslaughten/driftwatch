package suppressions

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSuppress_IsSuppressed(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)

	s.Suppress("/etc/app/config.yaml", 10*time.Minute)

	if !s.IsSuppressed("/etc/app/config.yaml") {
		t.Fatal("expected path to be suppressed")
	}
}

func TestSuppress_Expired(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)
	s.Suppress("/etc/app/config.yaml", 5*time.Minute)

	// advance clock past expiry
	s.now = fixedNow(base.Add(10 * time.Minute))

	if s.IsSuppressed("/etc/app/config.yaml") {
		t.Fatal("expected suppression to have expired")
	}
}

func TestSuppress_UnknownPath(t *testing.T) {
	s := New()
	if s.IsSuppressed("/etc/unknown.yaml") {
		t.Fatal("expected unknown path to not be suppressed")
	}
}

func TestRemove_LiftsSuppression(t *testing.T) {
	s := New()
	s.Suppress("/etc/app/config.yaml", 10*time.Minute)
	s.Remove("/etc/app/config.yaml")

	if s.IsSuppressed("/etc/app/config.yaml") {
		t.Fatal("expected suppression to be removed")
	}
}

func TestActive_ReturnsOnlyValid(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)

	s.Suppress("/etc/a.yaml", 10*time.Minute)
	s.Suppress("/etc/b.yaml", 1*time.Nanosecond)

	// advance past /etc/b.yaml expiry only
	s.now = fixedNow(base.Add(5 * time.Minute))

	active := s.Active()
	if len(active) != 1 {
		t.Fatalf("expected 1 active suppression, got %d", len(active))
	}
	if active[0].Path != "/etc/a.yaml" {
		t.Errorf("expected /etc/a.yaml, got %s", active[0].Path)
	}
}

func TestActive_EmptyWhenNone(t *testing.T) {
	s := New()
	if len(s.Active()) != 0 {
		t.Fatal("expected no active suppressions")
	}
}
