package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shunichi/herdr-hint/internal/herdr"
	"github.com/shunichi/herdr-hint/internal/label"
)

// Model is the Bubble Tea model for the hint picker. It renders precomputed
// lines with a scroll offset and resolves typed labels to an agent. On select
// it stores the target terminal_id and quits; main runs `agent focus` after the
// program exits and the terminal is restored (see docs/plan.md §3.2).
type Model struct {
	lines   []string
	items   []herdr.Item // flattened, labeled — for resolve
	double  bool
	height  int
	offset  int
	pending string // first char in two-letter mode
	sel     string // selected terminal_id ("" = none)
}

// NewModel builds a Model from arranged groups. height 0 means "not yet known"
// (renders all lines until the first WindowSizeMsg).
func NewModel(groups []Group, overflow, height int) Model {
	flat := Flatten(groups)
	return Model{
		lines:  Lines(groups, overflow),
		items:  flat,
		double: label.UsesDouble(flat),
		height: height,
	}
}

// Selected returns the chosen terminal_id, or "" if the user cancelled.
func (m Model) Selected() string { return m.sel }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		(&m).clamp()
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit // cancel: sel stays ""
		case tea.KeyCtrlD:
			(&m).scroll(m.half())
			return m, nil
		case tea.KeyCtrlU:
			(&m).scroll(-m.half())
			return m, nil
		case tea.KeyBackspace:
			m.pending = "" // abandon the in-progress two-letter input
			return m, nil
		case tea.KeyRunes:
			return m.handleRunes(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleRunes(runes []rune) (tea.Model, tea.Cmd) {
	if len(runes) != 1 {
		return m, nil
	}
	ch := runes[0]
	if ch < 'a' || ch > 'z' {
		return m, nil // ignore non a-z
	}
	if m.double {
		if m.pending == "" {
			m.pending = string(ch)
			return m, nil
		}
		input := m.pending + string(ch)
		m.pending = "" // mismatch clears the buffer; next key is a fresh first char
		if it := label.Resolve(m.items, input); it != nil {
			m.sel = it.TargetID
			return m, tea.Quit
		}
		return m, nil
	}
	if it := label.Resolve(m.items, string(ch)); it != nil {
		m.sel = it.TargetID
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) View() string {
	if len(m.lines) == 0 {
		return ""
	}
	v := m.visible()
	end := m.offset + v
	if end > len(m.lines) {
		end = len(m.lines)
	}
	return strings.Join(m.lines[m.offset:end], "\n") + "\n"
}

// visible is the number of content lines that fit (one row reserved for prompt).
func (m Model) visible() int {
	if m.height > 1 {
		return m.height - 1
	}
	return len(m.lines)
}

func (m *Model) half() int {
	if h := m.visible() / 2; h >= 1 {
		return h
	}
	return 1
}

func (m *Model) maxOffset() int {
	if mo := len(m.lines) - m.visible(); mo > 0 {
		return mo
	}
	return 0
}

func (m *Model) scroll(delta int) {
	m.offset += delta
	m.clamp()
}

func (m *Model) clamp() {
	if m.offset < 0 {
		m.offset = 0
	}
	if mo := m.maxOffset(); m.offset > mo {
		m.offset = mo
	}
}
