package silences

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAdd_IsSilenced(t *testing.T) {
	base := time.Now()
	s := New()
	s.now = fixedNow(base)
	s.Add("/etc/app/config.yaml", "maintenance", 10*time.Minute)
	if !s.IsSilenced("/etc/app/config.yaml") {
		t.Fatal("expected path to be silenced")
	}
}

func TestIsSilenced_Expired(t *testing.T) {
	base := time.Now()
	s := New()
	s.now = fixedNow(base)
	s.Add("/etc/app/config.yaml", "test", 5*time.Minute)
	s.now = fixedNow(base.Add(10 * time.Minute))
	if s.IsSilenced("/etc/app/config.yaml") {
		t.Fatal("expected silence to have expired")
	}
}

func TestIsSilenced_UnknownPath(t *testing.T) {
	s := New()
	if s.IsSilenced("/unknown/path") {
		t.Fatal("expected unknown path to not be silenced")
	}
}

func TestRemove_LiftsSilence(t *testing.T) {
	s := New()
	s.Add("/etc/x", "reason", time.Hour)
	s.Remove("/etc/x")
	if s.IsSilenced("/etc/x") {
		t.Fatal("expected silence to be removed")
	}
}

func TestActive_ReturnsOnlyNonExpired(t *testing.T) {
	base := time.Now()
	s := New()
	s.now = fixedNow(base)
	s.Add("/a", "r", time.Hour)
	s.Add("/b", "r", time.Minute)
	s.now = fixedNow(base.Add(5 * time.Minute))
	active := s.Active()
	if len(active) != 1 {
		t.Fatalf("expected 1 active silence, got %d", len(active))
	}
	if active[0].Path != "/a" {
		t.Errorf("expected /a, got %s", active[0].Path)
	}
}

func TestPurge_RemovesExpired(t *testing.T) {
	base := time.Now()
	s := New()
	s.now = fixedNow(base)
	s.Add("/a", "r", time.Minute)
	s.Add("/b", "r", time.Hour)
	s.now = fixedNow(base.Add(5 * time.Minute))
	s.Purge()
	if s.IsSilenced("/a") {
		t.Error("expected /a to be purged")
	}
	if !s.IsSilenced("/b") {
		t.Error("expected /b to still be silenced")
	}
}
