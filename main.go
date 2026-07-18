// Command herdr-hint renders Vimium-style hint labels for herdr agents and
// jumps to the selected one via `herdr agent focus`. It queries herdr over its
// CLI (JSON), groups agents by workspace, shows a hint picker, and — after the
// TUI exits and the terminal is restored — focuses the chosen agent.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shunichi/herdr-hint/internal/herdr"
	"github.com/shunichi/herdr-hint/internal/ui"
)

// herdrBin returns the herdr executable to invoke: $HERDR_BIN_PATH when set,
// otherwise "herdr" resolved from PATH.
func herdrBin() string {
	if p := os.Getenv("HERDR_BIN_PATH"); p != "" {
		return p
	}
	return "herdr"
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "herdr-hint:", err)
		os.Exit(1)
	}
}

func run() error {
	client := herdr.New(herdrBin())

	// Workspaces are only for grouping/labels; tolerate their absence.
	workspaces, _ := client.ListWorkspaces()

	agents, err := client.ListAgents()
	if err != nil {
		return err
	}
	items := client.ToItems(agents, workspaces)
	if len(items) == 0 {
		return nil // nothing to jump to
	}

	groups, overflow := ui.Arrange(items, workspaces)

	// Run the picker. Focus happens only after the program has exited and the
	// terminal is restored, so we never invoke herdr while in raw mode.
	res, err := tea.NewProgram(ui.NewModel(groups, overflow, 0)).Run()
	if err != nil {
		return err
	}
	final, ok := res.(ui.Model)
	if !ok {
		return fmt.Errorf("unexpected model type %T", res)
	}
	if target := final.Selected(); target != "" {
		return client.Focus(target)
	}
	return nil
}
