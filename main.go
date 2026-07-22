// Command herdr-hint helps jump to a herdr agent's pane. Two pickers share the
// herdr client and display model:
//
//	herdr-hint          hint picker: Vimium-style labels (entrypoint "jump")
//	herdr-hint filter   incremental filter: fuzzy-narrow by project, cursor-select
//
// Both query herdr over its CLI (JSON) and — after the TUI exits and the
// terminal is restored — focus the chosen agent via `herdr agent focus`.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shunichi/herdr-hint/internal/filter"
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
	mode := ""
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	var err error
	switch mode {
	case "filter":
		err = runFilter()
	case "", "jump", "hint":
		err = runHint()
	default:
		err = fmt.Errorf("unknown subcommand %q (use \"filter\" or none for hint)", mode)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "herdr-hint:", err)
		os.Exit(1)
	}
}

// fetch queries herdr and returns the client plus display items. A CLI failure /
// malformed JSON surfaces as an error; a successful empty list is fine.
func fetch() (*herdr.Client, []herdr.Item, []herdr.Workspace, error) {
	client := herdr.New(herdrBin())
	workspaces, err := client.ListWorkspaces()
	if err != nil {
		return nil, nil, nil, err
	}
	agents, err := client.ListAgents()
	if err != nil {
		return nil, nil, nil, err
	}
	return client, client.ToItems(agents, workspaces), workspaces, nil
}

func runHint() error {
	client, items, workspaces, err := fetch()
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	groups, overflow := ui.Arrange(items, workspaces)
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

func runFilter() error {
	client, items, workspaces, err := fetch()
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	res, err := tea.NewProgram(filter.NewModel(filter.Arrange(items, workspaces))).Run()
	if err != nil {
		return err
	}
	final, ok := res.(filter.Model)
	if !ok {
		return fmt.Errorf("unexpected model type %T", res)
	}
	if target := final.Selected(); target != "" {
		return client.Focus(target)
	}
	return nil
}
