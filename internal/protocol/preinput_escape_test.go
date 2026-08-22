package protocol_test

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/alex60217101990/terse/internal/protocol"
)

// reference is what EncodePreInput produced before it was hand-rolled: the same
// document through encoding/json with HTML escaping off.
func reference(t *testing.T, cmd string) string {
	t.Helper()
	type bashInput struct {
		Command string `json:"command"`
	}
	type preOut struct {
		HookSpecificOutput struct {
			HookEventName      string    `json:"hookEventName"`
			PermissionDecision string    `json:"permissionDecision"`
			UpdatedInput       bashInput `json:"updatedInput"`
		} `json:"hookSpecificOutput"`
	}
	var out preOut
	out.HookSpecificOutput.HookEventName = "PreToolUse"
	out.HookSpecificOutput.PermissionDecision = "allow"
	out.HookSpecificOutput.UpdatedInput.Command = cmd
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		t.Fatalf("reference encode: %v", err)
	}
	return b.String()
}

// TestEncodePreInput_MatchesEncodingJSON pins the hand-rolled encoder against
// the standard library it replaced. A command reaches the shell through this
// string, so a missed escape is a broken command at best.
func TestEncodePreInput_MatchesEncodingJSON(t *testing.T) {
	cases := []string{
		"",
		"ls -la",
		`echo "hi"`,
		`grep '\d+' file`,
		"echo a\nb",
		"echo a\tb\rc",
		"printf '\x00\x01\x1f'",
		"echo <html> & 'quotes'",
		"echo ünïcødé 日本語 🎉",
		"echo \u2028 \u2029",
		"echo " + string([]byte{0xff, 0xfe}),
		"echo " + strings.Repeat("x", 5000),
		"if : > '/tmp/c'; then { ls\n} > '/tmp/c' 2>&1\nfi",
	}
	for _, cmd := range cases {
		var got strings.Builder
		if err := protocol.EncodePreInput(&got, cmd); err != nil {
			t.Fatalf("EncodePreInput(%q): %v", cmd, err)
		}
		if want := reference(t, cmd); got.String() != want {
			t.Errorf("EncodePreInput(%q):\n got %q\nwant %q", cmd, got.String(), want)
		}
	}
}

// FuzzEncodePreInput keeps the two encoders equal on inputs no one thought of,
// and checks the result still parses back to the original command.
func FuzzEncodePreInput(f *testing.F) {
	for _, s := range []string{"ls", "echo \"x\"", "a\nb", "\xff", "\u2028"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, cmd string) {
		var got strings.Builder
		if err := protocol.EncodePreInput(&got, cmd); err != nil {
			t.Fatalf("EncodePreInput: %v", err)
		}
		if want := reference(t, cmd); got.String() != want {
			t.Fatalf("mismatch for %q:\n got %q\nwant %q", cmd, got.String(), want)
		}
		var back struct {
			HookSpecificOutput struct {
				UpdatedInput struct {
					Command string `json:"command"`
				} `json:"updatedInput"`
			} `json:"hookSpecificOutput"`
		}
		if err := json.Unmarshal([]byte(got.String()), &back); err != nil {
			t.Fatalf("re-parse %q: %v", got.String(), err)
		}
		// Invalid UTF-8 cannot survive JSON: both encoders replace each bad
		// byte, and the equality against encoding/json above already pins that
		// they do it identically. Only valid input must round-trip exactly.
		if round := back.HookSpecificOutput.UpdatedInput.Command; utf8.ValidString(cmd) && round != cmd {
			t.Fatalf("round-trip changed the command: %q -> %q", cmd, round)
		}
	})
}
