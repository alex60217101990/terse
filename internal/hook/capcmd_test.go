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
		{"cd /tmp && ls", true},        // && is safe under brace grouping
		{"make a || make b", true},     // so is ||
		{"cd /tmp; ls", true},          // and ;
		{"ls | head -20", false},       // agent already bounded its output
		{"cat f > out.txt", false},     // routes its own stdout
		{"sort < in.txt", false},       // routes its own stdin
		{"cat <<'EOF'\nx\nEOF", false}, // heredoc
		{"sleep 5 &", false},           // backgrounds; capture would race
		{"echo $(date)", false},        // substitution consumed by the caller
		{"echo `date`", false},         // same, backtick form
		{"sleep 5\\&&", false},         // \& is a literal; the second & backgrounds
		{"echo a\\||wc -l", false},     // \| is a literal; the second | is a real pipe
		{"", false},                    // nothing to wrap
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
