// Command herdr-hint renders Vimium-style hint labels for herdr agents and
// jumps to the selected one via `herdr agent focus`.
//
// This file is the scaffold (task 0001). The herdr client (0002), label
// assignment (0003), and the real Bubble Tea UI (0004) are added in later
// tasks; see docs/plan.md §4.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// herdrBin returns the herdr executable to invoke: $HERDR_BIN_PATH when set,
// otherwise "herdr" resolved from PATH. Kept as the single source of truth so
// later tasks (client / focus) call the same binary.
func herdrBin() string {
	if p := os.Getenv("HERDR_BIN_PATH"); p != "" {
		return p
	}
	return "herdr"
}

// model is a placeholder Bubble Tea model that quits immediately. It exists so
// the scaffold builds and exercises the Bubble Tea toolchain; the real hint UI
// replaces it in task 0004.
type model struct{}

func (m model) Init() tea.Cmd                       { return tea.Quit }
func (m model) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, tea.Quit }
func (m model) View() string                        { return "" }

func main() {
	_ = herdrBin()
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "herdr-hint:", err)
		os.Exit(1)
	}
}
