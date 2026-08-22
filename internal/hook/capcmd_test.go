package hook

import (
	"strings"
	"testing"
)

func TestCappable(t *testing.T) {
	cases := []struct {
		cmd  string
		want bool
	}{
		{"ls -la", true},
		{"go build ./...", true},
		{"cd /tmp && ls", true},          // && is safe under brace grouping
		{"make a || make b", true},       // so is ||
		{"cd /tmp; ls", true},            // and ;
		{"ls | head -20", true},          // a pipeline stays inside the group
		{"cat f > out.txt", false},       // routes its own stdout
		{"go build ./... 2>&1", true},    // a merge routes nothing out
		{"go build 2>&1 > o.txt", false}, // a merge plus a real redirect still routes
		{"sort < in.txt", false},         // routes its own stdin
		{"cat <<'EOF'\nx\nEOF", false},   // heredoc
		{"sleep 5 &", false},             // backgrounds; capture would race
		{"echo $(date)", false},          // substitution consumed by the caller
		{"echo `date`", false},           // same, backtick form
		{"sleep 5\\&&", false},           // \& is a literal; the second & backgrounds
		{"echo a\\||wc -l", true},        // \| is a literal; the real pipe is fine
		{"", false},                      // nothing to wrap
	}
	for _, c := range cases {
		if got := cappable(c.cmd); got != c.want {
			t.Errorf("cappable(%q) = %v, want %v", c.cmd, got, c.want)
		}
	}
}

// TestCappable_Interactive covers spec §6's interactive/streaming deny list.
// The list is narrow on purpose: a shell tool never types into a REPL, so only
// the bare, argument-less form of a session program is refused.
func TestCappable_Interactive(t *testing.T) {
	cases := []struct {
		cmd  string
		want bool
	}{
		{"vim foo", false},               // drives a terminal whatever follows
		{"top -l 1", false},              // same
		{"python", false},                // a bare REPL
		{"python3", false},               // same
		{"psql; echo done", false},       // the session ends its segment
		{"/usr/bin/vi notes.txt", false}, // matched on the basename
		{"GOWORK=off vim foo", false},    // env assignments are skipped
		{"cd /tmp && vim x", false},      // a separator re-arms the scan
		{"cargo watch -x test", false},   // the watch program
		{"git log|vim -", false},         // a pipe stage is a command of its own
		{"cat f | less", true},           // less reads as cat once stdout is a file
		{"git log | head -20", true},     // an ordinary pipeline
		{"go test --watch ./...", false}, // the long flag form
		{"ssh host uptime", true},        // ssh is not on the list: no TTY either way
		{"python3 -c 'print(1)'", true},  // a script run, not a session
		{"less README.md", true},         // reads as cat once stdout is a file
		{"vimdiff-report --list", true},  // prefix of a denied name only
		{"git commit -m 'ssh'", true},    // denied name only in program position
		{"grep -w pattern file", true},   // -w is a word-match flag far more often
	}
	for _, c := range cases {
		if got := cappable(c.cmd); got != c.want {
			t.Errorf("cappable(%q) = %v, want %v", c.cmd, got, c.want)
		}
	}
}

func TestWrapCommand_Shape(t *testing.T) {
	got := string(wrapCommand(nil, "ls -la", "/tmp/cap/x.out", "x", 1600))
	for _, want := range []string{
		// The group ends in a newline, not " ; ", so a trailing "# comment" on
		// cmd can't swallow the closing brace.
		"{ ls -la\n} > '/tmp/cap/x.out' 2>&1",
		"exit \"$__qrc\"",
		"qdf-hook expand x",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("wrapper missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "( ls -la )") {
		t.Error("wrapper must use brace grouping, not a subshell: cd must survive")
	}
}

func TestWrapCommand_RefusesUnsafeCapturePath(t *testing.T) {
	// A quote inside capturePath would close the shell's single-quoting early
	// and hand the rest of the wrapper to the shell as syntax, so wrapCommand
	// must refuse to emit anything rather than produce a broken command.
	if got := wrapCommand(nil, "ls -la", "/tmp/o'brien/x.out", "x", 1600); got != nil {
		t.Errorf("wrapCommand with a quote in capturePath = %q, want nil", got)
	}
}

func TestPreToolUse_Bash_RewritesWhenCappable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	in := `{"session_id":"s","hook_event_name":"PreToolUse","tool_name":"Bash",` +
		`"tool_input":{"command":"ls -la"}}`
	var out strings.Builder
	if err := HandlePreToolUse(strings.NewReader(in), &out); err != nil {
		t.Fatalf("HandlePreToolUse: %v", err)
	}
	if !strings.Contains(out.String(), `"updatedInput"`) {
		t.Errorf("a cappable Bash command must be rewritten, got: %s", out.String())
	}
}

func TestPreToolUse_Bash_LeavesRedirectingCommandAlone(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	in := `{"session_id":"s","hook_event_name":"PreToolUse","tool_name":"Bash",` +
		`"tool_input":{"command":"ls > out.txt"}}`
	var out strings.Builder
	if err := HandlePreToolUse(strings.NewReader(in), &out); err != nil {
		t.Fatalf("HandlePreToolUse: %v", err)
	}
	// Silence, not "{}": any output at all makes Claude Code record a
	// hook_success attachment that then rides the prefix all session.
	if got := out.String(); got != "" {
		t.Errorf("a skipped command must produce no output at all, got: %s", got)
	}
}
