// Package herdr talks to the herdr CLI (JSON over its subcommands) and converts
// the result into display items for the hint UI. Every external command goes
// through an injectable Runner so the parsing and conversion logic is testable
// without spawning processes.
package herdr

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// EmDash is the placeholder shown when a display field has no value. It is kept
// distinct from an empty raw field so callers can tell "information missing"
// apart from "the value is literally empty".
const EmDash = "—"

// Runner executes an external command and returns its stdout. Injecting it lets
// tests stub herdr/git responses. The default (New) runs real processes.
type Runner func(name string, args ...string) ([]byte, error)

func execRunner(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// Client issues herdr and git commands through a Runner.
type Client struct {
	bin string
	run Runner
}

// New returns a Client that invokes the given herdr binary with the real
// process runner.
func New(bin string) *Client { return &Client{bin: bin, run: execRunner} }

// NewWithRunner is New with an injected Runner, for tests.
func NewWithRunner(bin string, run Runner) *Client { return &Client{bin: bin, run: run} }

// Workspace is a herdr workspace (used to group and order agents).
type Workspace struct {
	ID     string
	Label  string
	Number int
}

// Agent is a herdr agent as reported by `herdr agent list`. Zero values mean the
// field was absent/empty in the JSON.
type Agent struct {
	PaneID        string // herdr 0.7.5+: `agent focus` target (terminal_id no longer accepted)
	Name          string
	Type          string // herdr "agent" field: claude / codex / ...
	Status        string
	Cwd           string
	Focused       bool
	TitleStripped string
	WorkspaceID   string
}

// Item is a display row for one agent. Raw fields keep "" for missing values;
// the *Label helpers apply the §3.2 fallbacks. Label is filled by the label
// package (task 0003); TargetID is what `herdr agent focus` receives.
type Item struct {
	Label       string
	TargetID    string // pane_id — the `herdr agent focus` target (herdr 0.7.5+)
	DisplayName string
	Type        string
	Status      string
	Focused     bool
	Context     string // "repo:branch"; "" = unavailable
	Group       string // workspace label; "" = unknown workspace
	GroupID     string // workspace_id (grouping + sort tie-break)
	Title       string // terminal_title_stripped; "" = none
}

// TypeLabel is the agent type for display: the raw value, or "unknown" if empty.
func (it Item) TypeLabel() string {
	if it.Type == "" {
		return "unknown"
	}
	return it.Type
}

// TitleLabel is the terminal title for display: the raw title, else the display
// name, else an em dash.
func (it Item) TitleLabel() string {
	if it.Title != "" {
		return it.Title
	}
	if it.DisplayName != "" {
		return it.DisplayName
	}
	return EmDash
}

// ContextLabel is the git context for display: "repo:branch" or an em dash.
func (it Item) ContextLabel() string {
	if it.Context == "" {
		return EmDash
	}
	return it.Context
}

type wsResponse struct {
	Result struct {
		Workspaces []struct {
			WorkspaceID string `json:"workspace_id"`
			Label       string `json:"label"`
			Number      int    `json:"number"`
		} `json:"workspaces"`
	} `json:"result"`
}

type agResponse struct {
	Result struct {
		Agents []struct {
			PaneID        string `json:"pane_id"`
			Name          string `json:"name"`
			Agent         string `json:"agent"`
			AgentStatus   string `json:"agent_status"`
			Cwd           string `json:"cwd"`
			Focused       bool   `json:"focused"`
			TitleStripped string `json:"terminal_title_stripped"`
			WorkspaceID   string `json:"workspace_id"`
		} `json:"agents"`
	} `json:"result"`
}

// ListWorkspaces runs `herdr workspace list` and parses the result.
func (c *Client) ListWorkspaces() ([]Workspace, error) {
	out, err := c.run(c.bin, "workspace", "list")
	if err != nil {
		return nil, fmt.Errorf("herdr workspace list: %w", err)
	}
	return parseWorkspaces(out)
}

// ListAgents runs `herdr agent list` and parses the result.
func (c *Client) ListAgents() ([]Agent, error) {
	out, err := c.run(c.bin, "agent", "list")
	if err != nil {
		return nil, fmt.Errorf("herdr agent list: %w", err)
	}
	return parseAgents(out)
}

// Focus runs `herdr agent focus <paneID>`. herdr 0.7.5+ accepts a pane id (or a
// unique agent name) as the target; terminal ids are no longer accepted.
func (c *Client) Focus(paneID string) error {
	if _, err := c.run(c.bin, "agent", "focus", paneID); err != nil {
		return fmt.Errorf("herdr agent focus %s: %w", paneID, err)
	}
	return nil
}

func parseWorkspaces(data []byte) ([]Workspace, error) {
	var r wsResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse workspace list: %w", err)
	}
	ws := make([]Workspace, len(r.Result.Workspaces))
	for i, w := range r.Result.Workspaces {
		ws[i] = Workspace{ID: w.WorkspaceID, Label: w.Label, Number: w.Number}
	}
	return ws, nil
}

func parseAgents(data []byte) ([]Agent, error) {
	var r agResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse agent list: %w", err)
	}
	agents := make([]Agent, len(r.Result.Agents))
	for i, a := range r.Result.Agents {
		agents[i] = Agent{
			PaneID:        a.PaneID,
			Name:          a.Name,
			Type:          a.Agent,
			Status:        a.AgentStatus,
			Cwd:           a.Cwd,
			Focused:       a.Focused,
			TitleStripped: a.TitleStripped,
			WorkspaceID:   a.WorkspaceID,
		}
	}
	return agents, nil
}

// ToItems converts agents into display items, resolving workspace labels and git
// context. Label assignment (0003) and ordering (0004) happen downstream.
func (c *Client) ToItems(agents []Agent, workspaces []Workspace) []Item {
	labels := make(map[string]string, len(workspaces))
	for _, w := range workspaces {
		labels[w.ID] = w.Label
	}
	items := make([]Item, len(agents))
	for i, a := range agents {
		items[i] = Item{
			TargetID:    a.PaneID,
			DisplayName: displayName(a),
			Type:        a.Type,
			Status:      a.Status,
			Focused:     a.Focused,
			Context:     c.GitContext(a.Cwd),
			Group:       labels[a.WorkspaceID],
			GroupID:     a.WorkspaceID,
			Title:       a.TitleStripped,
		}
	}
	return items
}

// displayName mirrors upstream parse_agents: name, else type, else pane id.
func displayName(a Agent) string {
	if a.Name != "" {
		return a.Name
	}
	if a.Type != "" {
		return a.Type
	}
	return a.PaneID
}

// GitContext returns "repo:branch" for cwd, or "" when cwd is not inside a git
// repo (or git is unavailable). Mirrors upstream git_context.
func (c *Client) GitContext(cwd string) string {
	if cwd == "" {
		return ""
	}
	toplevel, err := c.gitOut(cwd, "rev-parse", "--show-toplevel")
	if err != nil || toplevel == "" {
		return ""
	}
	commonDir, _ := c.gitOut(cwd, "rev-parse", "--git-common-dir")
	repo := repoName(toplevel, commonDir)
	branch, err := c.gitOut(cwd, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil || repo == "" || branch == "" {
		return ""
	}
	return repo + ":" + branch
}

func (c *Client) gitOut(cwd string, args ...string) (string, error) {
	full := append([]string{"-C", cwd}, args...)
	out, err := c.run("git", full...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// repoName mirrors upstream repo_name_from_paths: for an absolute git-common-dir
// (a worktree points at the main repo's .git) use its parent directory name so
// worktrees report the main repo; otherwise use the toplevel basename.
func repoName(toplevel, commonDir string) string {
	if strings.HasPrefix(commonDir, "/") {
		if parent := filepath.Base(filepath.Dir(commonDir)); parent != "" && parent != "." && parent != string(filepath.Separator) {
			return parent
		}
	}
	return filepath.Base(toplevel)
}
