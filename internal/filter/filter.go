// Package filter implements the incremental filter picker: type part of a
// project (workspace) name to fuzzy-narrow the agent list, move a cursor with
// the arrow keys / Ctrl+N / Ctrl+P, and press Enter to focus that agent.
//
// It reuses the hint view's deterministic grouping (internal/ui.Arrange) so the
// two commands order workspaces/agents identically; the hint packages are left
// untouched. Arrange/Filter/fuzzyMatch/ListLines are pure and unit-tested.
package filter

import (
	"fmt"
	"strings"

	"github.com/shunichi/herdr-hint/internal/herdr"
	"github.com/shunichi/herdr-hint/internal/ui"
)

// orphanGroup is the header ui.Arrange uses for agents with an unknown workspace.
const orphanGroup = "?"

// Group is a workspace's agents in display order (no hint labels).
type Group struct {
	Label string
	Items []herdr.Item
}

// Arrange groups items by workspace in the same deterministic order as the hint
// view, dropping the hint labels (unused here).
func Arrange(items []herdr.Item, workspaces []herdr.Workspace) []Group {
	ug, _ := ui.Arrange(items, workspaces)
	groups := make([]Group, len(ug))
	for i, g := range ug {
		groups[i] = Group{Label: g.Label, Items: g.Items}
	}
	return groups
}

// Filter keeps the groups whose project (workspace) name fuzzy-matches pattern.
// An empty pattern returns every group; a non-empty pattern drops the orphan
// ("?") group since it has no project name to match. Workspace order is
// preserved (stable) so the list does not jump around as the user types.
func Filter(groups []Group, pattern string) []Group {
	if strings.TrimSpace(pattern) == "" {
		return groups
	}
	var out []Group
	for _, g := range groups {
		if g.Label == orphanGroup {
			continue
		}
		if _, ok := fuzzyMatch(pattern, g.Label); ok {
			out = append(out, g)
		}
	}
	return out
}

// Selectable flattens the groups into the agents a cursor can land on, in
// display order.
func Selectable(groups []Group) []herdr.Item {
	var flat []herdr.Item
	for _, g := range groups {
		flat = append(flat, g.Items...)
	}
	return flat
}

// fuzzyMatch reports whether pattern is a subsequence of text (case-insensitive)
// and a score where lower means fewer gaps (better). Rune-based so non-ASCII
// project names work.
func fuzzyMatch(pattern, text string) (int, bool) {
	if pattern == "" {
		return 0, true
	}
	pr := []rune(strings.ToLower(pattern))
	tr := []rune(strings.ToLower(text))
	score, ti, last := 0, 0, -1
	for _, pc := range pr {
		found := false
		for ti < len(tr) {
			if tr[ti] == pc {
				if last >= 0 {
					score += ti - last - 1
				}
				last = ti
				ti++
				found = true
				break
			}
			ti++
		}
		if !found {
			return 0, false
		}
	}
	return score, true
}

// ListLines renders the filtered groups (headers + agent rows) and returns the
// lines plus the line index of the cursor-th selectable agent (-1 if there are
// no selectable agents). The cursor row is marked with ">".
func ListLines(groups []Group, cursor int) (lines []string, cursorLine int) {
	var typeW, statusW, ctxW int
	for _, g := range groups {
		for _, it := range g.Items {
			typeW = max(typeW, len(it.TypeLabel()))
			statusW = max(statusW, len(it.Status))
			ctxW = max(ctxW, len(it.ContextLabel()))
		}
	}
	cursorLine = -1
	sel := 0
	for gi, g := range groups {
		if gi > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, " "+g.Label)
		for _, it := range g.Items {
			marker := " "
			if sel == cursor {
				marker = ">"
				cursorLine = len(lines)
			}
			row := fmt.Sprintf("  %s %-*s  %-*s  %-*s  %s",
				marker, typeW, it.TypeLabel(), statusW, it.Status, ctxW, it.ContextLabel(), it.TitleLabel())
			lines = append(lines, strings.TrimRight(row, " "))
			sel++
		}
	}
	return lines, cursorLine
}
