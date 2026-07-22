package hook_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/hook"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

// makeReadInput marshals a PostToolUse Read hook payload as JSON.
func makeReadInput(t testing.TB, sessionID, path, content string) string {
	t.Helper()
	inp := map[string]any{
		"session_id":    sessionID,
		"tool_name":     "Read",
		"tool_input":    map[string]any{"file_path": path},
		"tool_response": map[string]any{"content": content},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

func TestReadHook_FirstRead_PassThrough(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	content, err := os.ReadFile("../cache/testdata/encoder_go_v1.txt")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	raw := makeReadInput(t, "sess-1", "/project/encoder.go", string(content))

	var out strings.Builder
	if err := hook.HandleRead(strings.NewReader(raw), &out); err != nil {
		t.Fatalf("HandleRead: %v", err)
	}

	var resp protocol.HookOutput
	if err := json.Unmarshal([]byte(out.String()), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v / raw: %s", err, out.String())
	}
	// First read: pass through (no updatedToolOutput OR updatedToolOutput == original).
	if resp.HookSpecificOutput != nil {
		s := resp.HookSpecificOutput.UpdatedToolOutput
		if !strings.Contains(s, "package encoder") {
			t.Errorf("first read should pass full content, got: %.80s", s)
		}
	}
}

func TestReadHook_SecondRead_Unchanged(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	content, err := os.ReadFile("../cache/testdata/encoder_go_v1.txt")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	raw := makeReadInput(t, "sess-2", "/project/encoder.go", string(content))

	// First read
	var out1 strings.Builder
	_ = hook.HandleRead(strings.NewReader(raw), &out1)

	// Second read — same content
	var out2 strings.Builder
	if err := hook.HandleRead(strings.NewReader(raw), &out2); err != nil {
		t.Fatalf("HandleRead second: %v", err)
	}

	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out2.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("second read should produce updatedToolOutput")
	}
	s := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(s, "§unchanged") {
		t.Errorf("second identical read should produce §unchanged marker, got: %.200s", s)
	}
}

func TestReadHook_Changed_ShowsDelta(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	v1, err := os.ReadFile("../cache/testdata/encoder_go_v1.txt")
	if err != nil {
		t.Fatalf("read testdata v1: %v", err)
	}
	v2, err := os.ReadFile("../cache/testdata/encoder_go_v2.txt")
	if err != nil {
		t.Fatalf("read testdata v2: %v", err)
	}

	const sid = "sess-3"
	const path = "/project/encoder.go"

	// First read: v1
	_ = hook.HandleRead(strings.NewReader(makeReadInput(t, sid, path, string(v1))), &strings.Builder{})

	// Second read: v2 (changed)
	var out strings.Builder
	_ = hook.HandleRead(strings.NewReader(makeReadInput(t, sid, path, string(v2))), &out)

	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("changed file read should produce updatedToolOutput")
	}
	s := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(s, "§delta") {
		t.Errorf("changed file should produce §delta marker, got: %.200s", s)
	}
}
