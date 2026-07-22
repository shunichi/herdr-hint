package filter

import (
	"strings"
	"testing"

	"github.com/shunichi/herdr-hint/internal/herdr"
)

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		pat, text string
		ok        bool
	}{
		{"", "anything", true},
		{"fam", "family-app2", true},
		{"fa2", "family-app2", true},   // subsequence with gaps
		{"FAM", "family-app2", true},   // case-insensitive
		{"xyz", "family-app2", false},  // missing chars
		{"app2x", "family-app2", false},
	}
	for _, c := range cases {
		if _, ok := fuzzyMatch(c.pat, c.text); ok != c.ok {
			t.Errorf("fuzzyMatch(%q,%q) ok=%v, want %v", c.pat, c.text, ok, c.ok)
		}
	}
}

func TestFuzzyScorePrefersFewerGaps(t *testing.T) {
	tight, _ := fuzzyMatch("ab", "abxx")
	loose, _ := fuzzyMatch("ab", "axxb")
	if !(tight < loose) {
		t.Errorf("tighter match should score lower: tight=%d loose=%d", tight, loose)
	}
}

func groups() []Group {
	return []Group{
		{Label: "family-app2", Items: []herdr.Item{
			{TargetID: "t1", Type: "claude", Status: "working"},
			{TargetID: "t2", Type: "codex", Status: "idle"},
		}},
		{Label: "workforce", Items: []herdr.Item{{TargetID: "t3", Type: "claude", Status: "idle"}}},
		{Label: "?", Items: []herdr.Item{{TargetID: "t9", Type: "codex", Status: "idle"}}},
	}
}

func TestFilterEmptyReturnsAll(t *testing.T) {
	g := Filter(groups(), "")
	if len(g) != 3 {
		t.Fatalf("empty pattern should return all, got %d", len(g))
	}
}

func TestFilterByProject(t *testing.T) {
	g := Filter(groups(), "fam")
	if len(g) != 1 || g[0].Label != "family-app2" {
		t.Fatalf("filter 'fam' = %+v", g)
	}
	// selectable agents of the matched project are kept
	if flat := Selectable(g); len(flat) != 2 {
		t.Fatalf("selectable = %d, want 2", len(flat))
	}
}

func TestFilterExcludesOrphanWhenPatterned(t *testing.T) {
	// "?" group has no project name; a non-empty pattern must never keep it.
	for _, g := range Filter(groups(), "o") { // 'o' is in workforce, not "?"
		if g.Label == "?" {
			t.Fatal("orphan group should be excluded when filtering")
		}
	}
}

func TestSelectableOrder(t *testing.T) {
	flat := Selectable(groups())
	want := []string{"t1", "t2", "t3", "t9"}
	if len(flat) != 4 {
		t.Fatalf("len = %d", len(flat))
	}
	for i, w := range want {
		if flat[i].TargetID != w {
			t.Fatalf("flat[%d] = %q, want %q", i, flat[i].TargetID, w)
		}
	}
}

func TestListLinesCursor(t *testing.T) {
	g := groups()
	lines, cursorLine := ListLines(g, 2) // third selectable = t3 (workforce)
	if cursorLine < 0 {
		t.Fatal("cursorLine should be set")
	}
	if !strings.HasPrefix(strings.TrimSpace(lines[cursorLine]), ">") {
		t.Errorf("cursor line should start with '>': %q", lines[cursorLine])
	}
	// headers present
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, " family-app2") || !strings.Contains(joined, " workforce") {
		t.Errorf("group headers missing:\n%s", joined)
	}
}

func TestListLinesNoSelectable(t *testing.T) {
	_, cursorLine := ListLines(nil, 0)
	if cursorLine != -1 {
		t.Errorf("empty groups cursorLine = %d, want -1", cursorLine)
	}
}
