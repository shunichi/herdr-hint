// Package label assigns Vimium-style hint labels to agent items and resolves a
// typed label back to its item. It is pure logic (no I/O, no TUI): the terminal
// input handling (Backspace, mismatch, resize, ...) lives in the UI (task 0004).
//
// Scheme (mirrors upstream assign_labels, agents-only): with <=26 items each
// gets a single letter a..z; with 27..676 items each gets a two-letter label
// aa..zz. The two-letter space caps at Cap (=676) labels; items beyond that get
// no label and are not selectable (see Overflow).
package label

import "github.com/shunichi/herdr-hint/internal/herdr"

// Cap is the maximum number of items that can receive a label (26*26, the size
// of the two-letter space).
const Cap = 26 * 26

// Assign sets Label on each item in place, per the package scheme. Items at
// index >= Cap get an empty Label (not selectable).
func Assign(items []herdr.Item) {
	double := UsesDouble(items)
	for i := range items {
		switch {
		case i >= Cap:
			items[i].Label = ""
		case double:
			items[i].Label = doubleLabel(i)
		default:
			items[i].Label = singleLabel(i)
		}
	}
}

// Overflow reports how many items exceed the labelable capacity (0 when none).
// The UI uses this to tell the user some agents are not selectable.
func Overflow(n int) int {
	if n > Cap {
		return n - Cap
	}
	return 0
}

// UsesDouble reports whether two-letter labels are in effect (>26 items).
func UsesDouble(items []herdr.Item) bool {
	return len(items) > 26
}

// Resolve returns a pointer to the item whose Label exactly equals input, or nil
// if none matches. Unlabeled items (Label == "") never match.
func Resolve(items []herdr.Item, input string) *herdr.Item {
	if input == "" {
		return nil
	}
	for i := range items {
		if items[i].Label != "" && items[i].Label == input {
			return &items[i]
		}
	}
	return nil
}

func singleLabel(i int) string {
	return string(rune('a' + i))
}

func doubleLabel(i int) string {
	return string(rune('a'+i/26)) + string(rune('a'+i%26))
}
