package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/daemon"
)

func TestIsQdfHookCommand(t *testing.T) {
	cases := []struct {
		cmd, sub string
		want     bool
	}{
		{"/Users/x/.local/bin/qdf-hook precompact", "precompact", true},
		{"qdf-hook read", "read", true},
		{"/tmp/qdf-hook.test precompact", "precompact", true},         // go test binary
		{"/opt/homebrew/bin/sqz hook precompact", "precompact", false}, // must NOT match
		{"atuin hook claude-code", "read", false},
		{"qdf-hook read", "bash", false},
		{"qdf-hook", "read", false}, // no subcommand arg
	}
	for _, c := range cases {
		if got := isQdfHookCommand(c.cmd, c.sub); got != c.want {
			t.Errorf("isQdfHookCommand(%q, %q) = %v, want %v", c.cmd, c.sub, got, c.want)
		}
	}
}

func TestMergeHooks_DoesNotFalseMatchOtherTools(t *testing.T) {
	// A PreCompact hook from another tool ending in "precompact" must not block
	// qdf-hook's own precompact from being installed.
	hb := hooksBlock{
		"PreCompact": []hookEntry{
			{Hooks: []hookCmd{{Type: "command", Command: "/opt/homebrew/bin/sqz hook precompact"}}},
		},
	}
	added := mergeHooks(hb, "qdf-hook")
	if added != len(qdfHooks) {
		t.Errorf("expected all %d hooks added, got %d", len(qdfHooks), added)
	}

	var sawSqz, sawQdf bool
	for _, e := range hb["PreCompact"] {
		for _, h := range e.Hooks {
			if h.Command == "/opt/homebrew/bin/sqz hook precompact" {
				sawSqz = true
			}
			if isQdfHookCommand(h.Command, "precompact") {
				sawQdf = true
			}
		}
	}
	if !sawSqz {
		t.Error("sqz precompact hook was dropped")
	}
	if !sawQdf {
		t.Error("qdf-hook precompact was not installed")
	}

	// Idempotency: a second merge adds nothing.
	if again := mergeHooks(hb, "qdf-hook"); again != 0 {
		t.Errorf("second merge should add 0, added %d", again)
	}
}

func TestIsQdfHookCommand_HybridAndEnsure(t *testing.T) {
	cases := []struct {
		cmd, sub string
		want     bool
	}{
		// Hybrid PostToolUse: match on the fallback half after "||".
		{"nc -N -U /Users/x/.qdf-hook/d.sock 2>/dev/null || /Users/x/.local/bin/qdf-hook post", "post", true},
		{"nc -N -U /Users/x/.qdf-hook/d.sock 2>/dev/null || qdf-hook post", "post", true},
		// SessionStart ensure hook.
		{"/Users/x/.local/bin/qdf-hook daemon --ensure", "daemon", true},
		{"qdf-hook daemon --ensure", "daemon", true},
		// A foreign tool's own hybrid-shaped fallback must not match.
		{"nc -N -U /tmp/d.sock 2>/dev/null || /opt/homebrew/bin/sqz hook post", "post", false},
		// nc/|| noise alone (no qdf-hook fallback) must not match.
		{"nc -N -U /tmp/d.sock 2>/dev/null || true", "post", false},
	}
	for _, c := range cases {
		if got := isQdfHookCommand(c.cmd, c.sub); got != c.want {
			t.Errorf("isQdfHookCommand(%q, %q) = %v, want %v", c.cmd, c.sub, got, c.want)
		}
	}
}

func TestPruneSuperseded_RemovesOldPerToolHookNotHybrid(t *testing.T) {
	hb := hooksBlock{
		"PostToolUse": []hookEntry{
			{Matcher: "Read", Hooks: []hookCmd{{Type: "command", Command: "/Users/x/.local/bin/qdf-hook read"}}},
			{Matcher: ".*", Hooks: []hookCmd{{Type: "command", Command: "nc -N -U /tmp/d.sock 2>/dev/null || /Users/x/.local/bin/qdf-hook post"}}},
		},
	}
	removed := pruneSuperseded(hb)
	if removed != 1 {
		t.Fatalf("expected 1 stale per-tool entry pruned, got %d", removed)
	}
	if len(hb["PostToolUse"]) != 1 {
		t.Fatalf("expected 1 entry remaining, got %d", len(hb["PostToolUse"]))
	}
	if !isQdfHookCommand(hb["PostToolUse"][0].Hooks[0].Command, "post") {
		t.Errorf("remaining entry should be the hybrid post hook, got %q", hb["PostToolUse"][0].Hooks[0].Command)
	}
}

// TestRunInit_HybridAndSessionStart is the Task 5 end-to-end check: after
// init, PostToolUse carries the nc/exe-fallback hybrid command and
// SessionStart carries the daemon --ensure command, and a second init run
// adds neither again.
func TestRunInit_HybridAndSessionStart(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := runInit(false, "", false); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	readHooks := func() hooksBlock {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
		if err != nil {
			t.Fatalf("read settings: %v", err)
		}
		top := map[string]json.RawMessage{}
		if err := json.Unmarshal(data, &top); err != nil {
			t.Fatalf("unmarshal settings: %v", err)
		}
		hb := hooksBlock{}
		if err := json.Unmarshal(top["hooks"], &hb); err != nil {
			t.Fatalf("unmarshal hooks: %v", err)
		}
		return hb
	}

	exe := execPath()
	sock := daemon.SockPath()

	countMatching := func(hb hooksBlock, event string, want func(string) bool) int {
		n := 0
		for _, e := range hb[event] {
			for _, h := range e.Hooks {
				if want(h.Command) {
					n++
				}
			}
		}
		return n
	}

	isHybridPost := func(cmd string) bool { return isQdfHookCommand(cmd, "post") }
	isEnsure := func(cmd string) bool { return isQdfHookCommand(cmd, "daemon") }

	hb := readHooks()

	// (a) exactly one hybrid PostToolUse hook, with the mandatory shape.
	if n := countMatching(hb, "PostToolUse", isHybridPost); n != 1 {
		t.Fatalf("expected exactly 1 hybrid PostToolUse hook, got %d", n)
	}
	var postCmd string
	for _, e := range hb["PostToolUse"] {
		for _, h := range e.Hooks {
			if isHybridPost(h.Command) {
				postCmd = h.Command
			}
		}
	}
	if !strings.Contains(postCmd, "nc -N -U") {
		t.Errorf("PostToolUse command missing %q: %q", "nc -N -U", postCmd)
	}
	if !strings.Contains(postCmd, "|| ") {
		t.Errorf("PostToolUse command missing %q: %q", "|| ", postCmd)
	}
	if !strings.Contains(postCmd, sock) {
		t.Errorf("PostToolUse command missing sock path %q: %q", sock, postCmd)
	}
	if !strings.Contains(postCmd, exe+" post") {
		t.Errorf("PostToolUse command missing %q: %q", exe+" post", postCmd)
	}

	// (b) exactly one SessionStart daemon --ensure hook.
	if n := countMatching(hb, "SessionStart", isEnsure); n != 1 {
		t.Fatalf("expected exactly 1 SessionStart ensure hook, got %d", n)
	}
	for _, e := range hb["SessionStart"] {
		for _, h := range e.Hooks {
			if isEnsure(h.Command) && !strings.Contains(h.Command, "daemon --ensure") {
				t.Errorf("SessionStart ensure command missing %q: %q", "daemon --ensure", h.Command)
			}
		}
	}

	// (c) re-running init is idempotent: counts stay at 1, no duplicates.
	if err := runInit(false, "", false); err != nil {
		t.Fatalf("second runInit: %v", err)
	}
	hb2 := readHooks()
	if n := countMatching(hb2, "PostToolUse", isHybridPost); n != 1 {
		t.Errorf("expected still exactly 1 hybrid PostToolUse hook after second init, got %d", n)
	}
	if n := countMatching(hb2, "SessionStart", isEnsure); n != 1 {
		t.Errorf("expected still exactly 1 SessionStart ensure hook after second init, got %d", n)
	}
}
