package diffview

import (
	"strings"
	"testing"
)

func TestCompare_Unchanged(t *testing.T) {
	d := Compare("/etc/app.yaml", "a: 1\nb: 2\n", "a: 1\nb: 2\n")
	if d.Changed {
		t.Fatal("expected Changed=false for identical content")
	}
	for _, l := range d.Lines {
		if l.Op != " " {
			t.Errorf("expected all lines unchanged, got op=%q", l.Op)
		}
	}
}

func TestCompare_AddedLine(t *testing.T) {
	d := Compare("/etc/app.yaml", "a: 1\n", "a: 1\nb: 2\n")
	if !d.Changed {
		t.Fatal("expected Changed=true")
	}
	var added int
	for _, l := range d.Lines {
		if l.Op == "+" {
			added++
		}
	}
	if added != 1 {
		t.Errorf("expected 1 added line, got %d", added)
	}
}

func TestCompare_RemovedLine(t *testing.T) {
	d := Compare("/etc/app.yaml", "a: 1\nb: 2\n", "a: 1\n")
	if !d.Changed {
		t.Fatal("expected Changed=true")
	}
	var removed int
	for _, l := range d.Lines {
		if l.Op == "-" {
			removed++
		}
	}
	if removed != 1 {
		t.Errorf("expected 1 removed line, got %d", removed)
	}
}

func TestCompare_EmptyBefore(t *testing.T) {
	d := Compare("/etc/app.yaml", "", "line1\nline2\n")
	if !d.Changed {
		t.Fatal("expected Changed=true")
	}
	for _, l := range d.Lines {
		if l.Op != "+" {
			t.Errorf("expected all lines added, got op=%q", l.Op)
		}
	}
}

func TestCompare_EmptyAfter(t *testing.T) {
	d := Compare("/etc/app.yaml", "line1\nline2\n", "")
	if !d.Changed {
		t.Fatal("expected Changed=true")
	}
	for _, l := range d.Lines {
		if l.Op != "-" {
			t.Errorf("expected all lines removed, got op=%q", l.Op)
		}
	}
}

func TestDiff_String_ContainsPath(t *testing.T) {
	d := Compare("/etc/app.yaml", "old\n", "new\n")
	s := d.String()
	if !strings.Contains(s, "/etc/app.yaml") {
		t.Errorf("expected path in diff string, got:\n%s", s)
	}
}

func TestDiff_String_ContainsOps(t *testing.T) {
	d := Compare("/etc/app.yaml", "old\n", "new\n")
	s := d.String()
	if !strings.Contains(s, "-old") {
		t.Errorf("expected removed line in output, got:\n%s", s)
	}
	if !strings.Contains(s, "+new") {
		t.Errorf("expected added line in output, got:\n%s", s)
	}
}
