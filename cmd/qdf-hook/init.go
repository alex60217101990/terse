package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// hookSpec is one hook entry qdf-hook installs into Claude Code settings.
type hookSpec struct {
	event   string // PreToolUse | PostToolUse | PreCompact | PostCompact
	matcher string
	sub     string // qdf-hook subcommand
}

// qdfHooks is the full set installed by `qdf-hook init`.
var qdfHooks = []hookSpec{
	{"PreToolUse", "Read", "pretooluse"},
	{"PostToolUse", "Read", "read"},
	{"PostToolUse", "Bash", "bash"},
	{"PostToolUse", "Write|Edit|MultiEdit", "write"},
	{"PostToolUse", "Glob", "glob"},
	{"PostToolUse", "Grep", "grep"},
	{"PreCompact", ".*", "precompact"},
	{"PostCompact", ".*", "postcompact"},
}

// Typed views of the settings.json hook schema — no interface{} anywhere.
type hookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type hookEntry struct {
	Matcher string    `json:"matcher,omitempty"`
	Hooks   []hookCmd `json:"hooks"`
}

// hooksBlock is the "hooks" object: event name -> ordered entries.
type hooksBlock map[string][]hookEntry

func cmdInit() *cobra.Command {
	var project bool
	var dir string
	var printOnly bool
	c := &cobra.Command{
		Use:   "init",
		Short: "Install qdf-hook into Claude Code settings.json (one-shot, idempotent)",
		Long: `init wires every qdf-hook hook into your Claude Code settings.json.

By default it edits the global ~/.claude/settings.json. Use --project to edit
.claude/settings.json in the current directory instead. It is idempotent:
existing hooks (qdf-hook's or anyone else's) are preserved, and re-running
never duplicates entries. Restart Claude Code afterwards for the hooks to load.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(project, dir, printOnly)
		},
	}
	c.Flags().BoolVar(&project, "project", false, "edit ./.claude/settings.json instead of the global one")
	c.Flags().StringVar(&dir, "dir", "", "project directory for --project (default: cwd)")
	c.Flags().BoolVar(&printOnly, "print", false, "print the merged settings.json to stdout, write nothing")
	return c
}

// settingsPath resolves the target settings.json path.
func settingsPath(project bool, dir string) (string, error) {
	if project {
		if dir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return "", err
			}
			dir = cwd
		}
		return filepath.Join(dir, ".claude", "settings.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

// execPath returns the running binary's absolute path so installed hooks
// resolve regardless of PATH; it falls back to the bare name.
func execPath() string {
	if exe, err := os.Executable(); err == nil && exe != "" {
		return exe
	}
	return "qdf-hook"
}

func runInit(project bool, dir string, printOnly bool) error {
	path, err := settingsPath(project, dir)
	if err != nil {
		return err
	}

	// Top level is kept as raw JSON per key so unknown settings (model, other
	// tools' config, …) survive untouched.
	top := map[string]json.RawMessage{}
	if data, rerr := os.ReadFile(path); rerr == nil {
		if jerr := json.Unmarshal(data, &top); jerr != nil {
			return fmt.Errorf("existing %s is not valid JSON: %w", path, jerr)
		}
	}

	hb := hooksBlock{}
	if raw, ok := top["hooks"]; ok && len(raw) > 0 {
		if jerr := json.Unmarshal(raw, &hb); jerr != nil {
			return fmt.Errorf("existing hooks block is not valid: %w", jerr)
		}
	}

	added := mergeHooks(hb, execPath())

	hooksRaw, err := json.Marshal(hb)
	if err != nil {
		return err
	}
	top["hooks"] = hooksRaw

	// Marshal compactly, then indent the whole document (json.Indent reaches
	// into the embedded hooks JSON that a plain MarshalIndent would leave flat).
	compact, err := json.Marshal(top)
	if err != nil {
		return err
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, compact, "", "  "); err != nil {
		return err
	}
	pretty.WriteByte('\n')
	out := pretty.Bytes()

	if printOnly {
		fmt.Print(string(out))
		return nil
	}

	// Nothing new to add — leave the file (and its formatting) untouched, no
	// backup churn.
	if added == 0 {
		fmt.Printf("qdf-hook already installed in %s (nothing to do)\n", path)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if data, rerr := os.ReadFile(path); rerr == nil {
		_ = os.WriteFile(path+".bak", data, 0o600)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}

	fmt.Printf("qdf-hook: installed %d hook(s) into %s\n", added, path)
	fmt.Println("  Restart Claude Code (or start a new session) for the hooks to load.")
	return nil
}

// mergeHooks idempotently adds every qdfHooks entry to hb, using exe as the
// command binary. Returns how many were newly added; existing entries (from any
// tool) are preserved, and a qdf-hook subcommand already wired for its event is
// skipped so re-running is a no-op.
func mergeHooks(hb hooksBlock, exe string) int {
	added := 0
	for _, h := range qdfHooks {
		if commandPresent(hb[h.event], h.sub) {
			continue
		}
		hb[h.event] = append(hb[h.event], hookEntry{
			Matcher: h.matcher,
			Hooks:   []hookCmd{{Type: "command", Command: exe + " " + h.sub}},
		})
		added++
	}
	return added
}

// commandPresent reports whether any entry already invokes THIS tool's given
// subcommand. It matches only qdf-hook's own command (binary basename
// "qdf-hook"), so another tool ending in the same word — e.g. sqz's
// "hook precompact" — is not mistaken for qdf-hook's "precompact".
func commandPresent(entries []hookEntry, sub string) bool {
	for _, e := range entries {
		for _, h := range e.Hooks {
			if isQdfHookCommand(h.Command, sub) {
				return true
			}
		}
	}
	return false
}

// isQdfHookCommand reports whether command c invokes `qdf-hook <sub>`: its
// binary (bare or absolute) has basename "qdf-hook" (a ".test" build counts too,
// so tests are deterministic) and its first argument is sub.
func isQdfHookCommand(c, sub string) bool {
	f := strings.Fields(c)
	if len(f) < 2 || f[1] != sub {
		return false
	}
	base := filepath.Base(f[0])
	return base == "qdf-hook" || strings.HasPrefix(base, "qdf-hook.")
}
