package filter

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"github.com/shunichi/herdr-hint/internal/herdr"
)

// Model is the Bubble Tea model for the incremental filter picker. Typing edits
// the project-name filter; Up/Down and Ctrl+P/Ctrl+N move the cursor; Enter
// focuses the cursor's agent (recorded in sel and applied by main after exit);
// Esc/Ctrl+C cancels.
type Model struct {
	all     []Group
	pattern string
	groups  []Group      // filtered
	flat    []herdr.Item // selectable agents in filtered order
	cursor  int
	sel     string
	width   int
	height  int
}

// NewModel builds a Model from all arranged groups (unfiltered).
func NewModel(groups []Group) Model {
	m := Model{all: groups}
	(&m).refilter()
	return m
}

// Selected returns the chosen terminal_id, or "" if cancelled / no match.
func (m Model) Selected() string { return m.sel }

func (m *Model) refilter() {
	m.groups = Filter(m.all, m.pattern)
	m.flat = Selectable(m.groups)
	if m.cursor >= len(m.flat) {
		m.cursor = len(m.flat) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) moveCursor(d int) {
	if len(m.flat) == 0 {
		return
	}
	m.cursor += d
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor > len(m.flat)-1 {
		m.cursor = len(m.flat) - 1
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit // cancel: sel stays ""
		case tea.KeyEnter:
			if len(m.flat) > 0 {
				m.sel = m.flat[m.cursor].TargetID
			}
			return m, tea.Quit
		case tea.KeyUp, tea.KeyCtrlP:
			(&m).moveCursor(-1)
			return m, nil
		case tea.KeyDown, tea.KeyCtrlN:
			(&m).moveCursor(1)
			return m, nil
		case tea.KeyBackspace:
			if r := []rune(m.pattern); len(r) > 0 {
				m.pattern = string(r[:len(r)-1])
				m.cursor = 0
				(&m).refilter()
			}
			return m, nil
		case tea.KeyRunes:
			m.pattern += string(msg.Runes)
			m.cursor = 0
			(&m).refilter()
			return m, nil
		}
	}
	return m, nil
}

func (m Model) View() string {
	prompt := "project: " + m.pattern
	if len(m.flat) == 0 {
		return prompt + "\n\n (no matching project)\n"
	}
	lines, cursorLine := ListLines(m.groups, m.cursor)
	window := m.windowAround(lines, cursorLine)
	if m.width > 0 {
		clipped := make([]string, len(window))
		for i, r := range window {
			clipped[i] = runewidth.Truncate(r, m.width, "")
		}
		window = clipped
	}
	return prompt + "\n" + strings.Join(window, "\n") + "\n"
}

// windowAround returns the slice of lines to display, scrolled so cursorLine is
// visible. The prompt occupies one row, so the list gets height-1 rows.
func (m Model) windowAround(lines []string, cursorLine int) []string {
	visible := 0
	if m.height > 0 {
		if visible = m.height - 1; visible < 1 {
			visible = 1
		}
	}
	if visible <= 0 || len(lines) <= visible {
		return lines
	}
	off := 0
	if cursorLine >= visible {
		off = cursorLine - visible + 1
	}
	if max := len(lines) - visible; off > max {
		off = max
	}
	if off < 0 {
		off = 0
	}
	return lines[off : off+visible]
}
