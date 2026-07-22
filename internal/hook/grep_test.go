package hook_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/hook"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

func makeGrepInput(t *testing.T, content string) string {
	t.Helper()
	inp := map[string]any{
		"session_id":    "grep-sess",
		"tool_name":     "Grep",
		"tool_input":    map[string]any{"pattern": "TODO"},
		"tool_response": map[string]any{"content": content},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

func runGrep(t *testing.T, content string) *protocol.HookOutput {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	var out strings.Builder
	if err := hook.HandleGrep(strings.NewReader(makeGrepInput(t, content)), &out); err != nil {
		t.Fatalf("HandleGrep: %v", err)
	}
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	return &resp
}

func TestGrep_ContentMode_GroupsByFile(t *testing.T) {
	var b strings.Builder
	for i := range 40 {
		b.WriteString("internal/hook/read.go:")
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString(":\t// TODO handle this case\n")
	}
	for i := range 40 {
		b.WriteString("internal/cache/store.go:")
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString(":\t// TODO revisit\n")
	}
	resp := runGrep(t, b.String())
	if resp.HookSpecificOutput == nil {
		t.Fatal("expected grouped grep summary")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(got, "internal/hook/read.go (40 matches)") {
		t.Errorf("expected per-file header with count, got:\n%s", got)
	}
	if !strings.Contains(got, "... +32 more") { // 40 - grepFileCap(8)
		t.Errorf("expected per-file cap elision, got:\n%s", got)
	}
	if !strings.Contains(got, "[grep: 80 matches in 2 files]") {
		t.Errorf("expected totals trailer, got:\n%s", got)
	}
	if len(got) >= len(b.String()) {
		t.Errorf("summary must be shorter: %d >= %d", len(got), len(b.String()))
	}
}

func TestGrep_FilesWithMatches_DelegatesToTree(t *testing.T) {
	var b strings.Builder
	for i := range 40 {
		b.WriteString("internal/pkg/file")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(".go\n")
	}
	resp := runGrep(t, b.String())
	if resp.HookSpecificOutput == nil {
		t.Fatal("expected tree summary for files_with_matches")
	}
	if !strings.Contains(resp.HookSpecificOutput.UpdatedToolOutput, "files") {
		t.Errorf("expected tree-style output, got:\n%s", resp.HookSpecificOutput.UpdatedToolOutput)
	}
}

func TestGrep_ShortContent_Passthrough(t *testing.T) {
	resp := runGrep(t, "a.go:1:hit")
	if resp.HookSpecificOutput != nil {
		t.Error("short content should pass through")
	}
}

func BenchmarkGrep_Grouped(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	var sb strings.Builder
	for i := range 500 {
		sb.WriteString("internal/pkg")
		sb.WriteString(strconv.Itoa(i % 10))
		sb.WriteString("/file.go:")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(":\tmatch text here\n")
	}
	raw := makeGrepInputBench(sb.String())
	b.ResetTimer()
	for b.Loop() {
		var out strings.Builder
		_ = hook.HandleGrep(strings.NewReader(raw), &out)
	}
}

func makeGrepInputBench(content string) string {
	inp := map[string]any{
		"session_id":    "grep-bench",
		"tool_name":     "Grep",
		"tool_input":    map[string]any{"pattern": "x"},
		"tool_response": map[string]any{"content": content},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}
