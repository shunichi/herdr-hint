package label

import (
	"testing"

	"github.com/shunichi/herdr-hint/internal/herdr"
)

func items(n int) []herdr.Item {
	its := make([]herdr.Item, n)
	for i := range its {
		its[i].TargetID = string(rune('A' + i%26))
	}
	return its
}

func TestAssignZero(t *testing.T) {
	its := items(0)
	Assign(its) // must not panic
	if len(its) != 0 {
		t.Fatal("expected empty")
	}
}

func TestAssignSingle26(t *testing.T) {
	its := items(26)
	Assign(its)
	if its[0].Label != "a" || its[25].Label != "z" {
		t.Fatalf("labels: first=%q last=%q", its[0].Label, its[25].Label)
	}
	if UsesDouble(its) {
		t.Fatal("26 items should be single-label")
	}
}

func TestAssignDouble27(t *testing.T) {
	its := items(27)
	Assign(its)
	if !UsesDouble(its) {
		t.Fatal("27 items should be double-label")
	}
	if its[0].Label != "aa" || its[25].Label != "az" || its[26].Label != "ba" {
		t.Fatalf("labels: [0]=%q [25]=%q [26]=%q", its[0].Label, its[25].Label, its[26].Label)
	}
}

func TestAssignAtCap676(t *testing.T) {
	its := items(Cap) // 676
	Assign(its)
	if its[0].Label != "aa" || its[Cap-1].Label != "zz" {
		t.Fatalf("labels: [0]=%q [675]=%q", its[0].Label, its[Cap-1].Label)
	}
	if Overflow(Cap) != 0 {
		t.Fatalf("Overflow(676) = %d, want 0", Overflow(Cap))
	}
}

func TestAssignOverflow677(t *testing.T) {
	its := items(Cap + 1) // 677
	Assign(its)
	if its[Cap-1].Label != "zz" {
		t.Fatalf("item[675] = %q, want zz", its[Cap-1].Label)
	}
	if its[Cap].Label != "" {
		t.Fatalf("item[676] should be unlabeled, got %q", its[Cap].Label)
	}
	if Overflow(Cap+1) != 1 {
		t.Fatalf("Overflow(677) = %d, want 1", Overflow(Cap+1))
	}
}

func TestResolve(t *testing.T) {
	its := items(3)
	Assign(its) // a, b, c
	if got := Resolve(its, "b"); got == nil || got.TargetID != its[1].TargetID {
		t.Fatalf("Resolve(b) = %+v", got)
	}
	if got := Resolve(its, "z"); got != nil {
		t.Fatalf("Resolve(z) should be nil, got %+v", got)
	}
	if got := Resolve(its, ""); got != nil {
		t.Fatalf("Resolve(empty) should be nil")
	}
}

func TestResolveIgnoresUnlabeled(t *testing.T) {
	its := items(Cap + 1)
	Assign(its)
	// item[676] has empty label; an empty input must not resolve to it.
	if got := Resolve(its, ""); got != nil {
		t.Fatalf("empty input resolved to %+v", got)
	}
}

func TestLabelValues(t *testing.T) {
	if singleLabel(0) != "a" || singleLabel(25) != "z" {
		t.Fatal("singleLabel wrong")
	}
	if doubleLabel(0) != "aa" || doubleLabel(26) != "ba" || doubleLabel(Cap-1) != "zz" {
		t.Fatalf("doubleLabel: [0]=%q [26]=%q [675]=%q", doubleLabel(0), doubleLabel(26), doubleLabel(Cap-1))
	}
}
