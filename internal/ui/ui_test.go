package ui

import (
	"strings"
	"testing"

	"github.com/shunichi/herdr-hint/internal/herdr"
)

func TestArrangeOrderingAndGrouping(t *testing.T) {
	ws := []herdr.Workspace{
		{ID: "wB", Label: "beta", Number: 2},
		{ID: "wA", Label: "alpha", Number: 1},
		{ID: "wC", Label: "gamma", Number: 3}, // no agents -> omitted
	}
	items := []herdr.Item{
		{TargetID: "t3", GroupID: "wB"},
		{TargetID: "t1", GroupID: "wA"},
		{TargetID: "t2", GroupID: "wA"},
		{TargetID: "t9", GroupID: "wZ"}, // unknown workspace -> orphan
	}
	groups, overflow := Arrange(items, ws)
	if overflow != 0 {
		t.Fatalf("overflow = %d", overflow)
	}
	if len(groups) != 3 {
		t.Fatalf("groups = %d (gamma should be omitted)", len(groups))
	}
	// order: alpha (number 1), beta (number 2), orphan "?"
	if groups[0].Label != "alpha" || groups[1].Label != "beta" || groups[2].Label != orphanGroup {
		t.Fatalf("group order: %q %q %q", groups[0].Label, groups[1].Label, groups[2].Label)
	}
	// within alpha: t1 then t2 (TargetID order), labels a,b in flatten order
	if groups[0].Items[0].TargetID != "t1" || groups[0].Items[0].Label != "a" {
		t.Fatalf("alpha[0] = %+v", groups[0].Items[0])
	}
	if groups[0].Items[1].TargetID != "t2" || groups[0].Items[1].Label != "b" {
		t.Fatalf("alpha[1] = %+v", groups[0].Items[1])
	}
	if groups[1].Items[0].TargetID != "t3" || groups[1].Items[0].Label != "c" {
		t.Fatalf("beta[0] = %+v", groups[1].Items[0])
	}
	if groups[2].Items[0].TargetID != "t9" || groups[2].Items[0].Label != "d" {
		t.Fatalf("orphan[0] = %+v", groups[2].Items[0])
	}
}

func TestArrangeEmpty(t *testing.T) {
	groups, overflow := Arrange(nil, nil)
	if len(groups) != 0 || overflow != 0 {
		t.Fatalf("empty: groups=%d overflow=%d", len(groups), overflow)
	}
}

func TestArrangeWorkspaceNumberTiebreakByID(t *testing.T) {
	ws := []herdr.Workspace{
		{ID: "wB", Label: "b", Number: 1},
		{ID: "wA", Label: "a", Number: 1}, // same number -> tie-break by ID: wA before wB
	}
	items := []herdr.Item{{TargetID: "x", GroupID: "wB"}, {TargetID: "y", GroupID: "wA"}}
	groups, _ := Arrange(items, ws)
	if groups[0].Label != "a" || groups[1].Label != "b" {
		t.Fatalf("tie-break order: %q %q", groups[0].Label, groups[1].Label)
	}
}

func TestLinesContent(t *testing.T) {
	groups := []Group{{
		Label: "alpha",
		Items: []herdr.Item{
			{Label: "a", Type: "claude", Status: "working", Context: "repo:main", Title: "Fix bug", Focused: true},
			{Label: "b", Type: "codex", Status: "idle", Context: "", Title: ""}, // fallbacks: context/title -> em dash / display
		},
	}}
	out := strings.Join(Lines(groups, 0), "\n")
	if !strings.Contains(out, " alpha") {
		t.Error("missing group header")
	}
	if !strings.Contains(out, "[a]") || !strings.Contains(out, "claude") ||
		!strings.Contains(out, "working") || !strings.Contains(out, "repo:main") || !strings.Contains(out, "Fix bug") {
		t.Errorf("row a missing fields:\n%s", out)
	}
	// focused marker present on row a
	rowA := lineWith(Lines(groups, 0), "[a]")
	if !strings.Contains(rowA, "*") {
		t.Errorf("row a should have focus marker: %q", rowA)
	}
	// row b: empty context -> em dash
	if !strings.Contains(out, herdr.EmDash) {
		t.Errorf("expected em dash fallback:\n%s", out)
	}
}

func TestLinesOverflowFooter(t *testing.T) {
	groups := []Group{{Label: "w", Items: []herdr.Item{{Label: "a", Type: "claude", Status: "idle"}}}}
	out := strings.Join(Lines(groups, 5), "\n")
	if !strings.Contains(out, "5 more") {
		t.Errorf("overflow footer missing:\n%s", out)
	}
}

func lineWith(lines []string, sub string) string {
	for _, l := range lines {
		if strings.Contains(l, sub) {
			return l
		}
	}
	return ""
}
