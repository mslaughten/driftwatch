// Package diffview provides a simple line-level diff between two file snapshots.
package diffview

import (
	"fmt"
	"strings"
)

// Line represents a single line in a diff output.
type Line struct {
	Op      string // "+" added, "-" removed, " " unchanged
	Content string
}

// Diff holds the result of comparing two text snapshots.
type Diff struct {
	Path    string
	Lines   []Line
	Changed bool
}

// String returns a human-readable unified-style diff.
func (d *Diff) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "--- %s (before)\n", d.Path)
	fmt.Fprintf(&sb, "+++ %s (after)\n", d.Path)
	for _, l := range d.Lines {
		fmt.Fprintf(&sb, "%s%s\n", l.Op, l.Content)
	}
	return sb.String()
}

// Compare performs a naive line-level diff between before and after text.
// It returns a Diff describing which lines were added or removed.
func Compare(path, before, after string) *Diff {
	old := splitLines(before)
	new := splitLines(after)

	lines := lcs(old, new)
	changed := before != after

	return &Diff{Path: path, Lines: lines, Changed: changed}
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimRight(s, "\n"), "\n")
}

// lcs builds diff lines using a simple LCS-based approach.
func lcs(old, new []string) []Line {
	m, n := len(old), len(new)
	// dp[i][j] = length of LCS of old[:i] and new[:j]
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if old[i-1] == new[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	var result []Line
	i, j := m, n
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && old[i-1] == new[j-1]:
			result = append([]Line{{Op: " ", Content: old[i-1]}}, result...)
			i--
			j--
		case j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]):
			result = append([]Line{{Op: "+", Content: new[j-1]}}, result...)
			j--
		default:
			result = append([]Line{{Op: "-", Content: old[i-1]}}, result...)
			i--
		}
	}
	return result
}
