package alertlog

import (
	"testing"
	"time"
)

func TestNew_DefaultCapacity(t *testing.T) {
	l := New(0)
	if l.cap != 50 {
		t.Errorf("expected default capacity 50, got %d", l.cap)
	}
}

func TestRecord_AddsEntry(t *testing.T) {
	l := New(10)
	l.Record("/etc/app/config.yaml", "http://example.com/hook", true)

	if l.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", l.Len())
	}

	entries := l.Recent()
	if entries[0].FilePath != "/etc/app/config.yaml" {
		t.Errorf("unexpected file path: %s", entries[0].FilePath)
	}
	if !entries[0].Success {
		t.Error("expected success=true")
	}
}

func TestRecord_TimestampSet(t *testing.T) {
	before := time.Now().UTC()
	l := New(10)
	l.Record("/tmp/test.conf", "http://hook", false)
	after := time.Now().UTC()

	entry := l.Recent()[0]
	if entry.Timestamp.Before(before) || entry.Timestamp.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", entry.Timestamp, before, after)
	}
}

func TestRecord_RingBufferEvictsOldest(t *testing.T) {
	l := New(3)
	l.Record("file1", "url", true)
	l.Record("file2", "url", true)
	l.Record("file3", "url", true)
	l.Record("file4", "url", true) // should evict file1

	if l.Len() != 3 {
		t.Fatalf("expected 3 entries after overflow, got %d", l.Len())
	}

	entries := l.Recent()
	if entries[0].FilePath != "file2" {
		t.Errorf("expected oldest entry to be file2, got %s", entries[0].FilePath)
	}
	if entries[2].FilePath != "file4" {
		t.Errorf("expected newest entry to be file4, got %s", entries[2].FilePath)
	}
}

func TestRecent_ReturnsCopy(t *testing.T) {
	l := New(10)
	l.Record("fileA", "url", true)

	a := l.Recent()
	a[0].FilePath = "mutated"

	b := l.Recent()
	if b[0].FilePath == "mutated" {
		t.Error("Recent() should return a copy, not a reference to internal state")
	}
}

func TestLen_Empty(t *testing.T) {
	l := New(5)
	if l.Len() != 0 {
		t.Errorf("expected 0, got %d", l.Len())
	}
}
