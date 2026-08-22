package hook

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/protocol"
)

// jsonEscape is the caller-side contract of appendWrappedJSON: the path it
// receives escaped must be the escaped form of the path it receives raw.
func jsonEscape(s string) []byte { return protocol.AppendJSONString(nil, s) }

// decodeWrapped pulls the rewritten command back out of a PreToolUse response.
func decodeWrapped(t *testing.T, resp []byte) string {
	t.Helper()
	var out struct {
		HookSpecificOutput struct {
			HookEventName      string `json:"hookEventName"`
			PermissionDecision string `json:"permissionDecision"`
			UpdatedInput       struct {
				Command string `json:"command"`
			} `json:"updatedInput"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(resp, &out); err != nil {
		t.Fatalf("response is not valid JSON: %v\n%s", err, resp)
	}
	if got := out.HookSpecificOutput.HookEventName; got != "PreToolUse" {
		t.Errorf("hookEventName = %q", got)
	}
	if got := out.HookSpecificOutput.PermissionDecision; got != "allow" {
		t.Errorf("permissionDecision = %q", got)
	}
	return out.HookSpecificOutput.UpdatedInput.Command
}

// TestAppendWrappedJSON_MatchesWrapCommand is the invariant behind the fused
// emitter: it writes the response and the wrapper in one pass, from the same
// segment table, so decoding it must give back exactly what the shell-truth
// builder produces. If the two ever drift, the shell tests keep passing while
// the agent runs something else — this is the test that catches that.
func TestAppendWrappedJSON_MatchesWrapCommand(t *testing.T) {
	const path = "/Users/x/.qdf-hook/captures/abc.out"
	const id = "abc"
	cases := []string{
		"ls -la",
		`echo "quoted" 'single'`,
		"printf 'a\\nb'",
		"cat <<'EOF'\nline\nEOF",
		"go test ./... 2>&1 | tail -5",
		"echo ünïcødé 🎉",
		"echo " + strings.Repeat("x", 4000),
		"echo tab\tand\nnewline",
	}
	for _, cmd := range cases {
		want := string(wrapCommand(nil, cmd, path, id, capBytes))
		resp := appendWrappedJSON(nil, cmd, string(jsonEscape(path)), id, capBytes)
		if got := decodeWrapped(t, resp); got != want {
			t.Errorf("fused emitter drifted for %q:\n got %q\nwant %q", cmd, got, want)
		}
	}
}

// TestPreToolUse_Bash_RefusesUnsafeHome keeps the shell-quoting guard where the
// untrusted value enters: a quote in $HOME would break out of the wrapper's
// quoting, so the command must be left alone entirely.
func TestPreToolUse_Bash_RefusesUnsafeHome(t *testing.T) {
	t.Setenv("HOME", t.TempDir()+"/o'brien")
	in := `{"session_id":"s","hook_event_name":"PreToolUse","tool_name":"Bash",` +
		`"tool_input":{"command":"ls -la"}}`
	var out strings.Builder
	if err := HandlePreToolUse(strings.NewReader(in), &out); err != nil {
		t.Fatalf("HandlePreToolUse: %v", err)
	}
	if got := out.String(); got != "" {
		t.Errorf("a home with a quote must be left alone, got: %s", got)
	}
}
