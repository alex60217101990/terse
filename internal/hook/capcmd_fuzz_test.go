package hook

import (
	"encoding/json"
	"strings"
	"testing"
)

// FuzzCappable checks the scan survives arbitrary command text. It walks bytes
// with lookahead and an escape skip, which is exactly the shape that panics on
// the input nobody thought of.
func FuzzCappable(f *testing.F) {
	for _, s := range []string{
		"", "\\", "&", "&&", "|", "||", ";", "2>&1", "a\\", "cmd &>", "x |& y",
		"vim", "python3 -c 'x'", "cd /tmp && ls", "echo $(date)", "cat <<'EOF'\nx\nEOF",
	} {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, cmd string) {
		_ = cappable(cmd)
	})
}

// FuzzWrappedJSON pins the two properties the rewrite has to hold for every
// command the model can send: the response is valid JSON, and the command it
// carries is exactly what the shell-truth builder would have produced. A drift
// between them means the agent runs something other than what the tests check.
func FuzzWrappedJSON(f *testing.F) {
	for _, s := range []string{
		"ls -la", `echo "q"`, "a\nb", "\x00\x1f", "echo \u2028", "ünïcødé 🎉",
		string([]byte{0xff, 0xfe}), strings.Repeat("x", 3000),
	} {
		f.Add(s)
	}
	const path = "/Users/x/.qdf-hook/captures/abc.out"
	const id = "0123456789abcdef0123456789abcdef"
	escapedPath := string(jsonEscape(path))

	f.Fuzz(func(t *testing.T, cmd string) {
		resp := appendWrappedJSON(nil, cmd, escapedPath, id, capBytes)
		if resp == nil {
			return // refused: the caller runs the command untouched
		}
		var out struct {
			HookSpecificOutput struct {
				UpdatedInput struct {
					Command string `json:"command"`
				} `json:"updatedInput"`
			} `json:"hookSpecificOutput"`
		}
		if err := json.Unmarshal(resp, &out); err != nil {
			t.Fatalf("response is not valid JSON for %q: %v\n%s", cmd, err, resp)
		}
		want := string(wrapCommand(nil, cmd, path, id, capBytes))
		if got := out.HookSpecificOutput.UpdatedInput.Command; got != want {
			t.Fatalf("fused emitter drifted for %q:\n got %q\nwant %q", cmd, got, want)
		}
	})
}
