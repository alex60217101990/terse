package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

// hookCommand returns the absolute command string for a subcommand, using the
// running binary's own path so the hook resolves regardless of PATH.
func hookCommand(sub string) string {
	exe, err := os.Executable()
	if err != nil || exe == "" {
		return "qdf-hook " + sub
	}
	return exe + " " + sub
}

func runInit(project bool, dir string, printOnly bool) error {
	path, err := settingsPath(project, dir)
	if err != nil {
		return err
	}

	// Load existing settings (preserve everything), or start fresh.
	settings := map[string]any{}
	if data, rerr := os.ReadFile(path); rerr == nil {
		if jerr := json.Unmarshal(data, &settings); jerr != nil {
			return fmt.Errorf("existing %s is not valid JSON: %w", path, jerr)
		}
	}

	added := mergeHooks(settings)

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	if printOnly {
		fmt.Print(string(out))
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Back up an existing file before overwriting, then write atomically.
	if _, serr := os.Stat(path); serr == nil {
		if data, rerr := os.ReadFile(path); rerr == nil {
			_ = os.WriteFile(path+".bak", data, 0o600)
		}
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}

	if added == 0 {
		fmt.Printf("qdf-hook already installed in %s (nothing to do)\n", path)
	} else {
		fmt.Printf("qdf-hook: installed %d hook(s) into %s\n", added, path)
		if _, serr := os.Stat(path + ".bak"); serr == nil {
			fmt.Printf("  (previous settings backed up to %s.bak)\n", path)
		}
		fmt.Println("  Restart Claude Code (or start a new session) for the hooks to load.")
	}
	return nil
}

// mergeHooks idempotently adds every qdfHooks entry to settings["hooks"],
// returning how many were newly added. Existing entries (any command) are
// preserved; a qdf-hook subcommand already present for its event is skipped.
func mergeHooks(settings map[string]any) int {
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		hooks = map[string]any{}
		settings["hooks"] = hooks
	}

	added := 0
	for _, h := range qdfHooks {
		cmdStr := hookCommand(h.sub)

		arr, _ := hooks[h.event].([]any)
		if commandPresent(arr, h.sub) {
			continue // idempotent: this qdf-hook subcommand is already wired
		}
		entry := map[string]any{
			"matcher": h.matcher,
			"hooks": []any{
				map[string]any{"type": "command", "command": cmdStr},
			},
		}
		hooks[h.event] = append(arr, entry)
		added++
	}
	return added
}

// commandPresent reports whether any hook command in the event array already
// invokes the given qdf-hook subcommand (matched by the " <sub>" suffix so it
// works whether the command is bare "qdf-hook read" or an absolute path).
func commandPresent(arr []any, sub string) bool {
	suffix := " " + sub
	for _, e := range arr {
		em, ok := e.(map[string]any)
		if !ok {
			continue
		}
		hs, ok := em["hooks"].([]any)
		if !ok {
			continue
		}
		for _, hh := range hs {
			hm, ok := hh.(map[string]any)
			if !ok {
				continue
			}
			if c, ok := hm["command"].(string); ok {
				if c == "qdf-hook "+sub || hasSuffixWord(c, suffix) {
					return true
				}
			}
		}
	}
	return false
}

// hasSuffixWord reports whether s ends with suffix and suffix is preceded by a
// path/word boundary — i.e. the command's final token is the subcommand.
func hasSuffixWord(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}
