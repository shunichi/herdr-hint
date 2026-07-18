package herdr

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// stubRunner records calls and returns canned output/error keyed by the first
// two args (command + subcommand or "git -C").
type stubRunner struct {
	calls   [][]string
	respond func(name string, args []string) ([]byte, error)
}

func (s *stubRunner) run(name string, args ...string) ([]byte, error) {
	s.calls = append(s.calls, append([]string{name}, args...))
	return s.respond(name, args)
}

const wsJSON = `{"id":"cli:workspace:list","result":{"type":"workspace_list","workspaces":[
  {"workspace_id":"wA","label":"herdr","number":1},
  {"workspace_id":"wB","label":"ga-pms","number":2}
]}}`

const agJSON = `{"id":"cli:agent:list","result":{"type":"agent_list","agents":[
  {"terminal_id":"t1","name":"my-agent","agent":"claude","agent_status":"working","cwd":"/repo/a","focused":true,"terminal_title_stripped":"Fix bug","workspace_id":"wA"},
  {"terminal_id":"t2","agent":"codex","agent_status":"idle","cwd":"/repo/b","focused":false,"workspace_id":"wA"}
]}}`

func TestParseWorkspaces(t *testing.T) {
	ws, err := parseWorkspaces([]byte(wsJSON))
	if err != nil {
		t.Fatal(err)
	}
	want := []Workspace{{ID: "wA", Label: "herdr", Number: 1}, {ID: "wB", Label: "ga-pms", Number: 2}}
	if !reflect.DeepEqual(ws, want) {
		t.Fatalf("got %+v, want %+v", ws, want)
	}
}

func TestParseAgents(t *testing.T) {
	ag, err := parseAgents([]byte(agJSON))
	if err != nil {
		t.Fatal(err)
	}
	if len(ag) != 2 {
		t.Fatalf("len = %d, want 2", len(ag))
	}
	if ag[0].TerminalID != "t1" || ag[0].Type != "claude" || ag[0].TitleStripped != "Fix bug" || !ag[0].Focused {
		t.Fatalf("agent[0] wrong: %+v", ag[0])
	}
	// Missing fields (name/cwd/title) must decode to empty, not error.
	if ag[1].Name != "" || ag[1].TitleStripped != "" || ag[1].Type != "codex" {
		t.Fatalf("agent[1] wrong: %+v", ag[1])
	}
}

func TestParseMalformedJSON(t *testing.T) {
	if _, err := parseAgents([]byte("{not json")); err == nil {
		t.Fatal("expected error for malformed agent JSON")
	}
	if _, err := parseWorkspaces([]byte("{not json")); err == nil {
		t.Fatal("expected error for malformed workspace JSON")
	}
}

func TestListNonZeroExit(t *testing.T) {
	s := &stubRunner{respond: func(name string, args []string) ([]byte, error) {
		return nil, errors.New("exit status 1")
	}}
	c := NewWithRunner("herdr", s.run)
	if _, err := c.ListAgents(); err == nil {
		t.Fatal("ListAgents should error on non-zero exit")
	}
	if _, err := c.ListWorkspaces(); err == nil {
		t.Fatal("ListWorkspaces should error on non-zero exit")
	}
}

func TestFocusPassesTerminalID(t *testing.T) {
	s := &stubRunner{respond: func(name string, args []string) ([]byte, error) { return nil, nil }}
	c := NewWithRunner("herdr", s.run)
	if err := c.Focus("t42"); err != nil {
		t.Fatal(err)
	}
	want := []string{"herdr", "agent", "focus", "t42"}
	if len(s.calls) != 1 || !reflect.DeepEqual(s.calls[0], want) {
		t.Fatalf("focus call = %v, want %v", s.calls, want)
	}
}

func TestDisplayNamePrecedence(t *testing.T) {
	cases := []struct {
		a    Agent
		want string
	}{
		{Agent{Name: "n", Type: "claude", TerminalID: "t"}, "n"},
		{Agent{Type: "codex", TerminalID: "t"}, "codex"},
		{Agent{TerminalID: "t"}, "t"},
	}
	for _, c := range cases {
		if got := displayName(c.a); got != c.want {
			t.Errorf("displayName(%+v) = %q, want %q", c.a, got, c.want)
		}
	}
}

func TestRepoName(t *testing.T) {
	// normal repo: relative common dir -> toplevel basename
	if got := repoName("/home/user/ga-pms", ".git"); got != "ga-pms" {
		t.Errorf("normal: got %q", got)
	}
	// worktree: absolute common dir -> its parent name (main repo)
	if got := repoName("/tmp/worktrees/research", "/home/user/ga-pms/.git"); got != "ga-pms" {
		t.Errorf("worktree: got %q", got)
	}
	// no common dir -> toplevel basename
	if got := repoName("/home/user/my-repo", ""); got != "my-repo" {
		t.Errorf("no-common: got %q", got)
	}
}

// gitStub answers the three rev-parse queries git_context issues.
func gitStub(top, common, branch string, fail bool) Runner {
	return func(name string, args ...string) ([]byte, error) {
		if name != "git" || fail {
			return nil, errors.New("not a git repo")
		}
		last := args[len(args)-1]
		switch {
		case last == "--show-toplevel":
			return []byte(top + "\n"), nil
		case last == "--git-common-dir":
			return []byte(common + "\n"), nil
		case last == "HEAD": // rev-parse --abbrev-ref HEAD
			return []byte(branch + "\n"), nil
		}
		return nil, errors.New("unexpected git args")
	}
}

func TestGitContext(t *testing.T) {
	cases := []struct {
		name              string
		top, common, brch string
		fail              bool
		cwd               string
		want              string
	}{
		{"normal repo", "/home/u/ga-pms", ".git", "main", false, "/home/u/ga-pms", "ga-pms:main"},
		{"worktree", "/tmp/wt/x", "/home/u/ga-pms/.git", "feature", false, "/tmp/wt/x", "ga-pms:feature"},
		{"detached HEAD", "/home/u/repo", ".git", "HEAD", false, "/home/u/repo", "repo:HEAD"},
		{"non-git cwd", "", "", "", true, "/tmp/plain", ""},
		{"empty cwd", "", "", "", false, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewWithRunner("herdr", gitStub(tc.top, tc.common, tc.brch, tc.fail))
			if got := c.GitContext(tc.cwd); got != tc.want {
				t.Fatalf("GitContext(%q) = %q, want %q", tc.cwd, got, tc.want)
			}
		})
	}
}

func TestToItems(t *testing.T) {
	agents, _ := parseAgents([]byte(agJSON))
	ws, _ := parseWorkspaces([]byte(wsJSON))
	// git context stub: /repo/a -> a:main, others none
	run := func(name string, args ...string) ([]byte, error) {
		if name == "git" {
			cwd := args[1] // -C <cwd>
			if cwd == "/repo/a" {
				last := args[len(args)-1]
				switch last {
				case "--show-toplevel":
					return []byte("/repo/a\n"), nil
				case "--git-common-dir":
					return []byte(".git\n"), nil
				case "HEAD":
					return []byte("main\n"), nil
				}
			}
			return nil, errors.New("no repo")
		}
		return nil, nil
	}
	c := NewWithRunner("herdr", run)
	items := c.ToItems(agents, ws)
	if len(items) != 2 {
		t.Fatalf("len = %d", len(items))
	}
	if items[0].TargetID != "t1" || items[0].DisplayName != "my-agent" || items[0].Type != "claude" ||
		items[0].Group != "herdr" || items[0].GroupID != "wA" || items[0].Context != "a:main" ||
		items[0].Title != "Fix bug" {
		t.Fatalf("item[0] wrong: %+v", items[0])
	}
	// second agent: no name -> display type; no git -> empty context; no title
	if items[1].DisplayName != "codex" || items[1].Context != "" || items[1].Title != "" {
		t.Fatalf("item[1] wrong: %+v", items[1])
	}
}

func TestLabelFallbacks(t *testing.T) {
	// missing type -> unknown; missing title -> display name; missing context -> em dash
	it := Item{DisplayName: "codex", Type: "", Title: "", Context: ""}
	if it.TypeLabel() != "unknown" {
		t.Errorf("TypeLabel = %q", it.TypeLabel())
	}
	if it.TitleLabel() != "codex" {
		t.Errorf("TitleLabel = %q", it.TitleLabel())
	}
	if it.ContextLabel() != EmDash {
		t.Errorf("ContextLabel = %q", it.ContextLabel())
	}
	// title empty AND no display name -> em dash
	if (Item{}).TitleLabel() != EmDash {
		t.Errorf("empty item TitleLabel should be em dash")
	}
	// present values pass through
	full := Item{DisplayName: "x", Type: "claude", Title: "Fix bug", Context: "repo:main"}
	if full.TypeLabel() != "claude" || full.TitleLabel() != "Fix bug" || full.ContextLabel() != "repo:main" {
		t.Errorf("passthrough wrong: %+v", full)
	}
}

func TestListWorkspacesParsesRunnerOutput(t *testing.T) {
	run := func(name string, args ...string) ([]byte, error) {
		if strings.Join(args, " ") == "workspace list" {
			return []byte(wsJSON), nil
		}
		return nil, errors.New("unexpected")
	}
	c := NewWithRunner("herdr", run)
	ws, err := c.ListWorkspaces()
	if err != nil || len(ws) != 2 || ws[0].Label != "herdr" {
		t.Fatalf("ListWorkspaces = %+v, err %v", ws, err)
	}
}
