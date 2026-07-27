package hook_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/hook"
	"github.com/alex60217101990/terse/internal/hookcore"
	"github.com/alex60217101990/terse/internal/protocol"
)

func dispatchInput(t *testing.T, tool, content string) string {
	t.Helper()
	inp := map[string]any{
		"session_id":    "disp",
		"tool_name":     tool,
		"tool_input":    map[string]any{"q": "x"},
		"tool_response": map[string]any{"content": content},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

func runDispatch(t *testing.T, tool, content string) *protocol.HookOutput {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	var out strings.Builder
	if err := hook.Dispatch(hookcore.NewDiskStore(), strings.NewReader(dispatchInput(t, tool, content)), &out); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	return &resp
}

// An arbitrary (MCP) tool returning a JSON array gets the columnar summary for
// free — no per-tool hardcoding.
func TestDispatch_MCPToolJSON_Summarized(t *testing.T) {
	var b strings.Builder
	b.WriteByte('[')
	for i := range 40 {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`{"id":`)
		b.WriteString(strconv.Itoa(i))
		b.WriteString(`,"status":"ok","name":"item"}`)
	}
	b.WriteByte(']')
	resp := runDispatch(t, "mcp__graph__query", b.String())
	if resp.HookSpecificOutput == nil {
		t.Fatal("MCP JSON-array output should be summarized")
	}
	if !strings.Contains(resp.HookSpecificOutput.UpdatedToolOutput, "COLUMNAR") {
		t.Errorf("expected columnar summary, got:\n%s", resp.HookSpecificOutput.UpdatedToolOutput)
	}
}

// An MCP batch that re-dumps the same section under several query headers
// (non-adjacent duplicates) gets folded on the first call — no §ref cache hit
// needed, and run-length collapse alone can't see non-adjacent repeats.
func TestDispatch_MCPRepeatedBlocks_Folded(t *testing.T) {
	block := "### grant-client-interface-methods\n" +
		strings.Repeat("func (c *client) PostAPIEntityV1Grants(ctx, body) (*Resp, error)\n", 6)
	other := "### other-query\nsome unrelated grep output line\nanother unrelated line\n"
	content := "## query one\n\n" + block + "\n\n" + other + "\n\n## query two\n\n" + block + "\n"

	resp := runDispatch(t, "mcp__plugin_context-mode__ctx_batch_execute", content)
	if resp.HookSpecificOutput == nil {
		t.Fatal("MCP output with duplicate blocks should be compressed, got passthrough")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if len(got) >= len(content) {
		t.Fatalf("expected shrink: %d >= %d", len(got), len(content))
	}
	if !strings.Contains(got, "↑ repeat:") {
		t.Errorf("expected a fold back-reference marker, got:\n%s", got)
	}
}

// A skip-listed tool is always passed through verbatim.
func TestDispatch_SkipList_Passthrough(t *testing.T) {
	big := strings.Repeat("todo item content line\n", 40)
	resp := runDispatch(t, "TodoWrite", big)
	if resp.HookSpecificOutput != nil {
		t.Error("TodoWrite must pass through verbatim")
	}
}

// Repeated identical output from any tool collapses to a §ref on the 2nd call.
func TestDispatch_GenericRefDedup(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	out := strings.Repeat("unstructured mcp output line here\n", 40)
	in := dispatchInput(t, "mcp__x__y", out)

	store := hookcore.NewDiskStore()
	var o1 strings.Builder
	_ = hook.Dispatch(store, strings.NewReader(in), &o1)
	var o2 strings.Builder
	_ = hook.Dispatch(store, strings.NewReader(in), &o2)

	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(o2.String()), &resp)
	if resp.HookSpecificOutput == nil || !strings.Contains(resp.HookSpecificOutput.UpdatedToolOutput, "§ref:") {
		t.Errorf("2nd identical MCP output should be a §ref, got: %v", o2.String())
	}
}
