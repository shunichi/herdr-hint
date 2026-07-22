package filter

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shunichi/herdr-hint/internal/herdr"
)

func runeKey(r rune) tea.KeyMsg      { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }
func key(t tea.KeyType) tea.KeyMsg   { return tea.KeyMsg{Type: t} }
func upd(m Model, msg tea.Msg) Model { nm, _ := m.Update(msg); return nm.(Model) }
func cmdOf(m Model, msg tea.Msg) tea.Cmd { _, c := m.Update(msg); return c }

func TestFilterNarrowsAndSelects(t *testing.T) {
	m := NewModel(groups())
	if len(m.flat) != 4 {
		t.Fatalf("initial selectable = %d, want 4", len(m.flat))
	}
	// type "fam" -> only family-app2 (2 agents)
	for _, r := range "fam" {
		m = upd(m, runeKey(r))
	}
	if len(m.flat) != 2 {
		t.Fatalf("after 'fam' selectable = %d, want 2", len(m.flat))
	}
	if m.cursor != 0 {
		t.Fatalf("cursor should reset to 0 on filter, got %d", m.cursor)
	}
	// move to second agent and select
	m = upd(m, key(tea.KeyDown))
	if m.cursor != 1 {
		t.Fatalf("cursor after down = %d", m.cursor)
	}
	cmd := cmdOf(m, key(tea.KeyEnter))
	m = upd(m, key(tea.KeyEnter))
	if m.Selected() != "t2" {
		t.Fatalf("selected = %q, want t2", m.Selected())
	}
	if cmd == nil {
		t.Error("enter should quit")
	}
}

func TestCursorKeysAndClamp(t *testing.T) {
	m := NewModel(groups()) // 4 selectable
	// up at top stays 0
	m = upd(m, key(tea.KeyUp))
	if m.cursor != 0 {
		t.Fatalf("up at top = %d", m.cursor)
	}
	// Ctrl+N moves down, Ctrl+P moves up
	m = upd(m, key(tea.KeyCtrlN))
	if m.cursor != 1 {
		t.Fatalf("ctrl+n = %d", m.cursor)
	}
	m = upd(m, key(tea.KeyCtrlP))
	if m.cursor != 0 {
		t.Fatalf("ctrl+p = %d", m.cursor)
	}
	// down past end clamps to last
	for i := 0; i < 10; i++ {
		m = upd(m, key(tea.KeyDown))
	}
	if m.cursor != 3 {
		t.Fatalf("clamp bottom = %d, want 3", m.cursor)
	}
}

func TestBackspaceWidensFilter(t *testing.T) {
	m := NewModel(groups())
	for _, r := range "fam" {
		m = upd(m, runeKey(r))
	}
	if len(m.flat) != 2 {
		t.Fatalf("after fam = %d", len(m.flat))
	}
	m = upd(m, key(tea.KeyBackspace)) // "fa"
	m = upd(m, key(tea.KeyBackspace)) // "f"
	m = upd(m, key(tea.KeyBackspace)) // "" -> all
	if m.pattern != "" || len(m.flat) != 4 {
		t.Fatalf("after backspaces pattern=%q flat=%d", m.pattern, len(m.flat))
	}
}

func TestEscCancels(t *testing.T) {
	m := NewModel(groups())
	cmd := cmdOf(m, key(tea.KeyEsc))
	m = upd(m, key(tea.KeyEsc))
	if m.Selected() != "" {
		t.Fatalf("esc should not select, got %q", m.Selected())
	}
	if cmd == nil {
		t.Error("esc should quit")
	}
}

func TestEnterWithNoMatchDoesNotSelect(t *testing.T) {
	m := NewModel(groups())
	for _, r := range "zzzz" { // no project matches
		m = upd(m, runeKey(r))
	}
	if len(m.flat) != 0 {
		t.Fatalf("expected no matches, got %d", len(m.flat))
	}
	m = upd(m, key(tea.KeyEnter))
	if m.Selected() != "" {
		t.Fatalf("enter with no match should not select, got %q", m.Selected())
	}
}

func TestViewNoMatchMessage(t *testing.T) {
	m := NewModel([]Group{{Label: "abc", Items: []herdr.Item{{TargetID: "t1"}}}})
	m = upd(m, runeKey('z'))
	if !strings.Contains(m.View(), "no matching project") {
		t.Errorf("expected no-match message, got:\n%s", m.View())
	}
}

func TestViewShowsPromptAndCursor(t *testing.T) {
	m := NewModel(groups())
	m = upd(m, runeKey('f'))
	out := m.View()
	if !strings.HasPrefix(out, "project: f") {
		t.Errorf("prompt missing: %q", out)
	}
	if !strings.Contains(out, ">") {
		t.Errorf("cursor marker missing:\n%s", out)
	}
}
