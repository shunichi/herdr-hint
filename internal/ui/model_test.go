package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shunichi/herdr-hint/internal/herdr"
)

// buildModel arranges n agents in one workspace and returns the model.
func buildModel(t *testing.T, n int) Model {
	t.Helper()
	items := make([]herdr.Item, n)
	for i := range items {
		// TargetID must sort in creation order for predictable labels.
		items[i] = herdr.Item{TargetID: string(rune('A'+i/26)) + string(rune('a'+i%26)), GroupID: "w1", Type: "claude", Status: "idle"}
	}
	ws := []herdr.Workspace{{ID: "w1", Label: "work", Number: 1}}
	groups, overflow := Arrange(items, ws)
	return NewModel(groups, overflow, 0)
}

func key(t tea.KeyType) tea.KeyMsg      { return tea.KeyMsg{Type: t} }
func runeKey(r rune) tea.KeyMsg         { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }
func upd(m Model, msg tea.Msg) Model    { nm, _ := m.Update(msg); return nm.(Model) }
func updCmd(m Model, msg tea.Msg) tea.Cmd { _, c := m.Update(msg); return c }

func TestSingleCharSelect(t *testing.T) {
	m := buildModel(t, 3) // labels a,b,c
	if m.double {
		t.Fatal("3 items should be single-char")
	}
	cmd := updCmd(m, runeKey('b'))
	m = upd(m, runeKey('b'))
	if m.Selected() != m.items[1].TargetID {
		t.Fatalf("selected = %q, want %q", m.Selected(), m.items[1].TargetID)
	}
	if cmd == nil {
		t.Error("select should return a quit command")
	}
}

func TestSingleCharNoMatch(t *testing.T) {
	m := buildModel(t, 3)
	m = upd(m, runeKey('z'))
	if m.Selected() != "" {
		t.Fatalf("no-match should not select, got %q", m.Selected())
	}
}

func TestTwoCharSelect(t *testing.T) {
	m := buildModel(t, 27) // labels aa..ba
	if !m.double {
		t.Fatal("27 items should be two-char")
	}
	m = upd(m, runeKey('a')) // pending "a"
	if m.pending != "a" || m.Selected() != "" {
		t.Fatalf("after first char: pending=%q sel=%q", m.pending, m.Selected())
	}
	m = upd(m, runeKey('b')) // "ab" -> second item
	if m.Selected() != m.items[1].TargetID {
		t.Fatalf("selected = %q, want %q", m.Selected(), m.items[1].TargetID)
	}
}

func TestTwoCharMismatchClearsBuffer(t *testing.T) {
	m := buildModel(t, 27)
	m = upd(m, runeKey('z')) // pending "z"
	m = upd(m, runeKey('z')) // "zz" not present -> clear buffer, no select
	if m.Selected() != "" {
		t.Fatalf("mismatch should not select, got %q", m.Selected())
	}
	if m.pending != "" {
		t.Fatalf("buffer should be cleared, pending=%q", m.pending)
	}
	// next key starts fresh as a new first char
	m = upd(m, runeKey('a'))
	if m.pending != "a" {
		t.Fatalf("fresh first char: pending=%q", m.pending)
	}
}

func TestBackspaceClearsPending(t *testing.T) {
	m := buildModel(t, 27)
	m = upd(m, runeKey('a'))
	m = upd(m, key(tea.KeyBackspace))
	if m.pending != "" {
		t.Fatalf("pending after backspace = %q", m.pending)
	}
}

func TestEscCancels(t *testing.T) {
	m := buildModel(t, 3)
	cmd := updCmd(m, key(tea.KeyEsc))
	m = upd(m, key(tea.KeyEsc))
	if m.Selected() != "" {
		t.Fatalf("esc should not select, got %q", m.Selected())
	}
	if cmd == nil {
		t.Error("esc should quit")
	}
}

func TestNonAlphaIgnored(t *testing.T) {
	m := buildModel(t, 3)
	m = upd(m, runeKey('1'))
	m = upd(m, runeKey('!'))
	if m.Selected() != "" || m.pending != "" {
		t.Fatalf("non-alpha should be ignored: sel=%q pending=%q", m.Selected(), m.pending)
	}
}

func TestScrollClamp(t *testing.T) {
	m := Model{lines: make([]string, 50), height: 10} // visible = 9
	m = upd(m, key(tea.KeyCtrlD))
	if m.offset != 4 { // half of 9 = 4
		t.Fatalf("after ctrl+d offset = %d, want 4", m.offset)
	}
	// scroll far down: clamp to maxOffset = 50-9 = 41
	for i := 0; i < 20; i++ {
		m = upd(m, key(tea.KeyCtrlD))
	}
	if m.offset != 41 {
		t.Fatalf("offset should clamp to 41, got %d", m.offset)
	}
	// scroll back up past 0 -> clamp to 0
	for i := 0; i < 20; i++ {
		m = upd(m, key(tea.KeyCtrlU))
	}
	if m.offset != 0 {
		t.Fatalf("offset should clamp to 0, got %d", m.offset)
	}
}

func TestWindowSizeClampsOffset(t *testing.T) {
	m := Model{lines: make([]string, 20), height: 5, offset: 18}
	m = upd(m, tea.WindowSizeMsg{Width: 80, Height: 5}) // visible=4, maxOffset=16
	if m.offset != 16 {
		t.Fatalf("offset after resize = %d, want 16", m.offset)
	}
}

func TestViewShowsVisibleWindow(t *testing.T) {
	m := Model{lines: []string{"l0", "l1", "l2", "l3", "l4"}, height: 3, offset: 1} // visible=2
	got := m.View()
	want := "l1\nl2\n"
	if got != want {
		t.Fatalf("View() = %q, want %q", got, want)
	}
}
