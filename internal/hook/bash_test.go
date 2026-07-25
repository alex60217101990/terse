package hook_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/hook"
	"github.com/alex60217101990/terse/internal/protocol"
)

func makeBashInput(t *testing.T, sessionID, cmd, output string) string {
	t.Helper()
	inp := map[string]any{
		"session_id":    sessionID,
		"tool_name":     "Bash",
		"tool_input":    map[string]any{"command": cmd},
		"tool_response": map[string]any{"content": output},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

func TestBashHook_JSONArray_Summarized(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	jsonData, _ := os.ReadFile("../../testdata/json_array_1k.json")

	raw := makeBashInput(t, "sess-bash-1", "cat data.json", string(jsonData))
	var out strings.Builder
	if err := hook.HandleBash(strings.NewReader(raw), &out); err != nil {
		t.Fatalf("HandleBash: %v", err)
	}
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("JSON array should produce updatedToolOutput")
	}
	s := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(s, "COLUMNAR SUMMARY") {
		t.Errorf("summary should contain COLUMNAR SUMMARY: %s", s)
	}
	if len(s) >= len(jsonData) {
		t.Errorf("summary should be shorter than source: %d >= %d", len(s), len(jsonData))
	}
}

func TestBashHook_GoTestPass_Summarized(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	data, _ := os.ReadFile("../../testdata/gotest_pass.txt")

	raw := makeBashInput(t, "sess-bash-2", "go test -v ./...", string(data))
	var out strings.Builder
	_ = hook.HandleBash(strings.NewReader(raw), &out)

	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("go test output should produce updatedToolOutput")
	}
	s := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(s, "PASS") {
		t.Errorf("summary should contain PASS: %s", s)
	}
}

func TestBashHook_GoTestFail_PreservesFailures(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	data, _ := os.ReadFile("../../testdata/gotest_fail.txt")

	raw := makeBashInput(t, "sess-bash-3", "go test -v ./...", string(data))
	var out strings.Builder
	_ = hook.HandleBash(strings.NewReader(raw), &out)

	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("go test fail output should produce updatedToolOutput")
	}
	s := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(s, "bar_test.go:42") {
		t.Errorf("failure location should be preserved in summary: %s", s)
	}
}

func TestBashHook_PlainText_Passthrough(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	raw := makeBashInput(t, "sess-bash-4", "echo hello", "hello world\n")
	var out strings.Builder
	_ = hook.HandleBash(strings.NewReader(raw), &out)

	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	// Plain text: either pass-through (nil) or unchanged output.
	if resp.HookSpecificOutput != nil {
		if resp.HookSpecificOutput.UpdatedToolOutput != "hello world\n" {
			t.Errorf("plain text should pass through unchanged")
		}
	}
}
