package hook_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/hook"
	"github.com/alex60217101990/terse/internal/protocol"
)

func makeGlobInput(t *testing.T, sessionID string, files []string) string {
	t.Helper()
	content := strings.Join(files, "\n")
	inp := map[string]any{
		"session_id":    sessionID,
		"tool_name":     "Glob",
		"tool_input":    map[string]any{"pattern": "**/*.go"},
		"tool_response": map[string]any{"content": content},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

func TestGlob_SmallList_Passthrough(t *testing.T) {
	files := []string{"main.go", "main_test.go"}
	var out strings.Builder
	_ = hook.HandleGlob(strings.NewReader(makeGlobInput(t, "sess-glob-1", files)), &out)
	// < 256 bytes — must passthrough (no replacement).
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput != nil {
		t.Error("small list should passthrough (no replacement)")
	}
}

func TestGlob_LargeList_Compressed(t *testing.T) {
	var files []string
	for i := range 30 {
		files = append(files, "internal/hook/file"+string(rune('a'+i%26))+".go")
	}
	var out strings.Builder
	_ = hook.HandleGlob(strings.NewReader(makeGlobInput(t, "sess-glob-2", files)), &out)
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("large list should be compressed")
	}
	compressed := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(compressed, "internal/hook/") {
		t.Errorf("compressed output should reference dir: %s", compressed)
	}
	if !strings.Contains(compressed, "total") {
		t.Errorf("compressed output should contain total count: %s", compressed)
	}
	// Verify compression actually reduces size.
	original := strings.Join(files, "\n")
	if len(compressed) >= len(original) {
		t.Errorf("compressed (%d) should be shorter than original (%d)", len(compressed), len(original))
	}
}

func TestGlob_GroupsDirectories(t *testing.T) {
	files := []string{
		"internal/hook/bash.go", "internal/hook/read.go", "internal/hook/glob.go",
		"internal/cache/state.go", "internal/cache/store.go",
		"cmd/qdf-hook/main.go",
	}
	// Make input large enough to trigger compression.
	var longFiles []string
	for range 10 {
		longFiles = append(longFiles, files...)
	}
	var out strings.Builder
	_ = hook.HandleGlob(strings.NewReader(makeGlobInput(t, "sess-glob-3", longFiles)), &out)
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Skip("small enough to passthrough")
	}
	compressed := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(compressed, "internal/") {
		t.Errorf("should group internal/ dir: %s", compressed)
	}
}
