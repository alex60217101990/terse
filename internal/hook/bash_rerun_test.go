package hook_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/hook"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

func mkBashRerun(out string) string {
	inp := map[string]any{
		"session_id": "rerun",
		"tool_name":  "Bash",
		"tool_input": map[string]any{"command": "./run.sh", "working_directory": "/p"},
		"tool_response": map[string]any{"content": out},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

func TestBash_RerunDelta(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	var base strings.Builder
	for i := range 40 {
		base.WriteString("processing record ")
		base.WriteString(strconv.Itoa(i))
		base.WriteString(" — status ready\n")
	}
	baseOut := base.String()

	// First run: nothing cached yet → passthrough, stored for next time.
	var o1 strings.Builder
	if err := hook.HandleBash(strings.NewReader(mkBashRerun(baseOut)), &o1); err != nil {
		t.Fatalf("run1: %v", err)
	}

	// Second run: same command, one extra line → should be a rerun-delta.
	changed := baseOut + "processing record 40 — status FAILED\n"
	var o2 strings.Builder
	if err := hook.HandleBash(strings.NewReader(mkBashRerun(changed)), &o2); err != nil {
		t.Fatalf("run2: %v", err)
	}
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(o2.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("re-run with a small change should be compressed to a delta")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(got, "§rerun-delta") {
		t.Errorf("expected §rerun-delta, got:\n%s", got)
	}
	if len(got) >= len(changed) {
		t.Errorf("delta (%d) must be smaller than full output (%d)", len(got), len(changed))
	}
	if !strings.Contains(got, "FAILED") {
		t.Errorf("delta must include the changed line, got:\n%s", got)
	}
}
