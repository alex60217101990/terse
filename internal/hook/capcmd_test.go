package hook

import "testing"

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
		{"ls | head -20", false},         // agent already bounded its output
		{"cat f > out.txt", false},       // routes its own stdout
		{"sort < in.txt", false},         // routes its own stdin
		{"cat <<'EOF'\nx\nEOF", false},   // heredoc
		{"sleep 5 &", false},             // backgrounds; capture would race
		{"echo $(date)", false},          // substitution consumed by the caller
		{"echo `date`", false},           // same, backtick form
		{"", false},                      // nothing to wrap
	}
	for _, c := range cases {
		if got := cappable(c.cmd); got != c.want {
			t.Errorf("cappable(%q) = %v, want %v", c.cmd, got, c.want)
		}
	}
}
