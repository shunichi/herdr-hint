// Package ui builds the hint display from herdr items and drives the Bubble Tea
// program. Arrange/Lines are pure (grouping, deterministic ordering, rendering)
// and unit-tested without a terminal; the Model wires them to Bubble Tea.
package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shunichi/herdr-hint/internal/herdr"
	"github.com/shunichi/herdr-hint/internal/label"
)

// orphanGroup is the header for agents whose workspace_id is unknown/empty.
const orphanGroup = "?"

// Group is a workspace's agents in display order, with labels already assigned.
type Group struct {
	Label string
	Items []herdr.Item
}

// Arrange groups items by workspace and assigns labels across the flattened,
// deterministically-ordered sequence. Order (see docs/plan.md §3.2):
//   - workspaces by Number, then ID (total order);
//   - within a workspace by TargetID, then original index;
//   - agents with an unknown/empty workspace_id go last under "?";
//   - workspaces with no agents are omitted.
//
// It returns the groups and the overflow count (agents beyond the label cap).
func Arrange(items []herdr.Item, workspaces []herdr.Workspace) ([]Group, int) {
	wsOrder := append([]herdr.Workspace(nil), workspaces...)
	sort.SliceStable(wsOrder, func(i, j int) bool {
		if wsOrder[i].Number != wsOrder[j].Number {
			return wsOrder[i].Number < wsOrder[j].Number
		}
		return wsOrder[i].ID < wsOrder[j].ID
	})
	known := make(map[string]bool, len(wsOrder))
	labelOf := make(map[string]string, len(wsOrder))
	for _, w := range wsOrder {
		known[w.ID] = true
		labelOf[w.ID] = w.Label
	}

	type entry struct {
		it   herdr.Item
		orig int
	}
	byWs := map[string][]entry{}
	var orphans []entry
	for i, it := range items {
		if known[it.GroupID] {
			byWs[it.GroupID] = append(byWs[it.GroupID], entry{it, i})
		} else {
			orphans = append(orphans, entry{it, i})
		}
	}
	byTargetThenOrig := func(es []entry) func(a, b int) bool {
		return func(a, b int) bool {
			if es[a].it.TargetID != es[b].it.TargetID {
				return es[a].it.TargetID < es[b].it.TargetID
			}
			return es[a].orig < es[b].orig
		}
	}

	var flat []herdr.Item
	type span struct {
		label string
		n     int
	}
	var spans []span
	appendGroup := func(lbl string, es []entry) {
		if len(es) == 0 {
			return
		}
		for _, e := range es {
			flat = append(flat, e.it)
		}
		spans = append(spans, span{lbl, len(es)})
	}

	for _, w := range wsOrder {
		es := byWs[w.ID]
		sort.SliceStable(es, byTargetThenOrig(es))
		appendGroup(labelOf[w.ID], es)
	}
	// orphans: order by workspace_id, then TargetID, then original index.
	sort.SliceStable(orphans, func(a, b int) bool {
		if orphans[a].it.GroupID != orphans[b].it.GroupID {
			return orphans[a].it.GroupID < orphans[b].it.GroupID
		}
		if orphans[a].it.TargetID != orphans[b].it.TargetID {
			return orphans[a].it.TargetID < orphans[b].it.TargetID
		}
		return orphans[a].orig < orphans[b].orig
	})
	appendGroup(orphanGroup, orphans)

	overflow := label.Overflow(len(flat))
	label.Assign(flat)

	groups := make([]Group, 0, len(spans))
	pos := 0
	for _, sp := range spans {
		groups = append(groups, Group{Label: sp.label, Items: flat[pos : pos+sp.n]})
		pos += sp.n
	}
	return groups, overflow
}

// Flatten returns all items across groups in display order (for label resolve).
func Flatten(groups []Group) []herdr.Item {
	var flat []herdr.Item
	for _, g := range groups {
		flat = append(flat, g.Items...)
	}
	return flat
}

// Lines renders the groups into display lines (one string per row). A blank line
// separates groups; a footer notes any overflow. Labels use the §3.2 fallbacks.
func Lines(groups []Group, overflow int) []string {
	var typeW, statusW, ctxW int
	for _, g := range groups {
		for _, it := range g.Items {
			typeW = max(typeW, len(it.TypeLabel()))
			statusW = max(statusW, len(it.Status))
			ctxW = max(ctxW, len(it.ContextLabel()))
		}
	}
	var lines []string
	for gi, g := range groups {
		if gi > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, " "+g.Label)
		for _, it := range g.Items {
			marker := " "
			if it.Focused {
				marker = "*"
			}
			labelCell := "    " // unlabeled (overflow): keep column width
			if it.Label != "" {
				labelCell = fmt.Sprintf("[%s]", it.Label)
			}
			row := fmt.Sprintf("   %s %-4s %-*s  %-*s  %-*s  %s",
				marker, labelCell,
				typeW, it.TypeLabel(),
				statusW, it.Status,
				ctxW, it.ContextLabel(),
				it.TitleLabel())
			lines = append(lines, strings.TrimRight(row, " "))
		}
	}
	if overflow > 0 {
		lines = append(lines, "",
			fmt.Sprintf(" (%d more agent(s) not selectable — label limit %d)", overflow, label.Cap))
	}
	return lines
}
