package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/alex60217101990/qdf-hook/internal/daemon"
	"github.com/spf13/cobra"
)

// hookSpec is one hook entry qdf-hook installs into Claude Code settings.
type hookSpec struct {
	event   string // PreToolUse | PostToolUse | PreCompact | PostCompact | SessionStart
	matcher string
	// sub is the token isQdfHookCommand matches against (the first argument
	// after the qdf-hook binary, or after the hybrid command's "||"
	// fallback). For most events the installed command is literally
	// "<exe> <sub>"; PostToolUse and SessionStart build a longer command via
	// hookCommand but still key off this same token — see there.
	sub string
}

// qdfHooks is the full set installed by `qdf-hook init`. A single catch-all
// PostToolUse hook routes every tool (Read/Write/Bash/Glob/Grep/mcp__*/…)
// through `post`, which dispatches internally — so new tools are covered with
// no config change. The PostToolUse and SessionStart entries get their full
// command line from hookCommand rather than "<exe> <sub>" — see there.
var qdfHooks = []hookSpec{
	{"PreToolUse", "Read", "pretooluse"},
	{"PostToolUse", ".*", "post"},
	{"PreCompact", ".*", "precompact"},
	{"PostCompact", ".*", "postcompact"},
	{"SessionStart", "", "daemon"},
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

	// Migrate: the single catch-all `post` supersedes the old per-tool
	// PostToolUse hooks (read/bash/write/glob/grep). Drop them so upgrading
	// doesn't double-process every tool.
	pruned := pruneSuperseded(hb)
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

	// Nothing new to add and nothing stale to remove — leave the file untouched.
	if added == 0 && pruned == 0 {
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

	fmt.Printf("qdf-hook: installed/updated %d hook(s) into %s\n", added, path)
	fmt.Println("  Restart Claude Code (or start a new session) for the hooks to load.")
	return nil
}

// mergeHooks idempotently brings every qdfHooks entry in hb up to date, using
// exe as the command binary. Returns how many entries it changed (added or
// upgraded); other tools' entries are preserved, and an entry already at the
// current command is left untouched so re-running is a no-op.
//
// An existing qdf-hook command for the event whose text differs from the
// current one is rewritten IN PLACE rather than skipped — this upgrades an
// older install (e.g. a plain "<exe> post" catch-all from before the daemon)
// to the current hybrid form, and absorbs any drift in the socket path or nc
// flags, without duplicating the entry.
func mergeHooks(hb hooksBlock, exe string) int {
	changed := 0
	for _, h := range qdfHooks {
		want := hookCommand(h, exe)
		if upgradeExisting(hb[h.event], h.sub, want) {
			changed++ // rewrote an out-of-date qdf-hook command in place
			continue
		}
		if commandPresent(hb[h.event], h.sub) {
			continue // already present and current
		}
		hb[h.event] = append(hb[h.event], hookEntry{
			Matcher: h.matcher,
			Hooks:   []hookCmd{{Type: "command", Command: want}},
		})
		changed++
	}
	return changed
}

// upgradeExisting finds the first qdf-hook command for sub whose text differs
// from want and rewrites it in place, returning true. A command already equal
// to want is left alone (returns false — commandPresent then treats it as an
// up-to-date no-op).
func upgradeExisting(entries []hookEntry, sub, want string) bool {
	for ei := range entries {
		for hi := range entries[ei].Hooks {
			cmd := entries[ei].Hooks[hi].Command
			if isQdfHookCommand(cmd, sub) && cmd != want {
				entries[ei].Hooks[hi].Command = want
				return true
			}
		}
	}
	return false
}

// hookCommand builds the full command line for a hookSpec. Most events run
// the qdf-hook binary directly ("<exe> <sub>"); two are special-cased:
//
//   - PostToolUse ("post") and PreToolUse ("pretooluse") go through the
//     daemon/CLI hybrid so a warm qdf-hookd answers from RAM (a repeated
//     Read's mtime check never pays a fresh process + disk decode). On Unix
//     the tiny native client qdf-hookc — installed next to qdf-hook, resolved
//     via filepath.Dir(exe) — dials the daemon's unix socket; on any failure
//     the shell `||` falls back to `<exe> <sub>`, which itself dials the
//     socket then dispatches inline. Windows has no native client, so it
//     runs `<exe> <sub>` directly.
//   - SessionStart ("daemon") starts/refreshes the daemon so the socket is
//     already live before the first PostToolUse hook fires.
func hookCommand(h hookSpec, exe string) string {
	switch {
	case h.event == "PostToolUse" && h.sub == "post",
		h.event == "PreToolUse" && h.sub == "pretooluse":
		if runtime.GOOS == "windows" {
			return exe + " " + h.sub
		}
		client := filepath.Join(filepath.Dir(exe), "qdf-hookc")
		return fmt.Sprintf("%s %s 2>/dev/null || %s %s", client, daemon.SockPath(), exe, h.sub)
	case h.event == "SessionStart" && h.sub == "daemon":
		return exe + " daemon --ensure"
	default:
		return exe + " " + h.sub
	}
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

// supersededSubs are the old per-tool PostToolUse subcommands now covered by
// the single catch-all `post`.
var supersededSubs = []string{"read", "bash", "write", "glob", "grep"}

// pruneSuperseded removes PostToolUse entries that consist solely of superseded
// qdf-hook commands (from an older install). Returns how many were removed.
func pruneSuperseded(hb hooksBlock) int {
	entries := hb["PostToolUse"]
	kept := entries[:0]
	removed := 0
	for _, e := range entries {
		if entryAllSuperseded(e) {
			removed++
			continue
		}
		kept = append(kept, e)
	}
	hb["PostToolUse"] = kept
	return removed
}

// entryAllSuperseded reports whether every hook in e is a superseded qdf-hook
// command (so dropping the whole entry can't remove another tool's hook).
func entryAllSuperseded(e hookEntry) bool {
	if len(e.Hooks) == 0 {
		return false
	}
	for _, h := range e.Hooks {
		if !slices.ContainsFunc(supersededSubs, func(sub string) bool {
			return isQdfHookCommand(h.Command, sub)
		}) {
			return false
		}
	}
	return true
}

// isQdfHookCommand reports whether command c invokes `qdf-hook <sub>`: its
// binary (bare or absolute) has basename "qdf-hook" (a ".test" build counts too,
// so tests are deterministic) and its first argument is sub.
//
// Hybrid commands (see hookCommand) wrap this in `nc ... || <exe> <sub>`; the
// nc/socket half never looks like qdf-hook, so when a "||" is present only
// the fallback half after the LAST one is checked — this also means the
// nc/|| noise can't be mistaken for, or hide, a foreign tool's own command.
func isQdfHookCommand(c, sub string) bool {
	if i := strings.LastIndex(c, "||"); i >= 0 {
		c = c[i+len("||"):]
	}
	f := strings.Fields(c)
	if len(f) < 2 || f[1] != sub {
		return false
	}
	base := filepath.Base(f[0])
	return base == "qdf-hook" || strings.HasPrefix(base, "qdf-hook.")
}
