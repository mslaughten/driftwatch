package statuspage

import (
	"testing"
	"time"
)

func TestNew_RegistersPaths(t *testing.T) {
	p := New([]string{"/etc/app/config.yaml", "/etc/app/db.yaml"})
	statuses := p.Snapshot()
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
}

func TestRecordCheck_UpdatesDrifted(t *testing.T) {
	p := New([]string{"/etc/app/config.yaml"})

	p.RecordCheck("/etc/app/config.yaml", true)

	statuses := p.Snapshot()
	if !statuses[0].Drifted {
		t.Error("expected Drifted to be true")
	}
}

func TestRecordCheck_SetsLastCheck(t *testing.T) {
	p := New([]string{"/etc/app/config.yaml"})
	before := time.Now()
	p.RecordCheck("/etc/app/config.yaml", false)

	statuses := p.Snapshot()
	if statuses[0].LastCheck.Before(before) {
		t.Error("expected LastCheck to be set to current time")
	}
}

func TestRecordCheck_SetsLastDriftOnlyWhenDrifted(t *testing.T) {
	p := New([]string{"/etc/app/config.yaml"})
	p.RecordCheck("/etc/app/config.yaml", false)

	statuses := p.Snapshot()
	if !statuses[0].LastDrift.IsZero() {
		t.Error("expected LastDrift to be zero when not drifted")
	}

	p.RecordCheck("/etc/app/config.yaml", true)
	statuses = p.Snapshot()
	if statuses[0].LastDrift.IsZero() {
		t.Error("expected LastDrift to be set when drifted")
	}
}

func TestAnyDrifted_FalseWhenClean(t *testing.T) {
	p := New([]string{"/a", "/b"})
	p.RecordCheck("/a", false)
	p.RecordCheck("/b", false)

	if p.AnyDrifted() {
		t.Error("expected AnyDrifted to be false")
	}
}

func TestAnyDrifted_TrueWhenOneDrifted(t *testing.T) {
	p := New([]string{"/a", "/b"})
	p.RecordCheck("/a", false)
	p.RecordCheck("/b", true)

	if !p.AnyDrifted() {
		t.Error("expected AnyDrifted to be true")
	}
}

func TestRecordCheck_UnknownPathRegistered(t *testing.T) {
	p := New(nil)
	p.RecordCheck("/new/path", false)

	statuses := p.Snapshot()
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if statuses[0].Path != "/new/path" {
		t.Errorf("unexpected path: %s", statuses[0].Path)
	}
}
