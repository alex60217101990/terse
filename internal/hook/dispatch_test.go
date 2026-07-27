package hook_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/analytics"
	"github.com/alex60217101990/terse/internal/cache"
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
	block := "### widget-client-interface-methods\n" +
		strings.Repeat("func (c *client) PostAPIWidgetV1Items(ctx, body) (*Resp, error)\n", 6)
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

// grep/ripgrep "file:line:text" output run via Bash compresses through the
// generic try-chain (buildGrepSummary), the same as the Grep tool.
func TestDispatch_BashGrep_Compressed(t *testing.T) {
	// 6 files with 20 matches each: past grepFileCap (8), so most match lines
	// elide to "... +N more" — the realistic big-win shape.
	var b strings.Builder
	for i := range 120 {
		fmt.Fprintf(&b, "internal/pkg/file%d.go:%d:\tfound a matching identifier here on this line\n", i%6, i)
	}
	content := b.String()
	resp := runDispatch(t, "Bash", content)
	if resp.HookSpecificOutput == nil {
		t.Fatal("Bash grep output should be compressed, got passthrough")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if len(got) >= len(content) {
		t.Fatalf("expected shrink: %d >= %d", len(got), len(content))
	}
	if !strings.Contains(got, "[grep:") {
		t.Errorf("expected grep-grouped summary, got:\n%s", got)
	}
}

// Columnar JSON summary off the Read path must (1) not carry the old misleading
// "Read with offset/limit" / "[READ" wording, and (2) register the raw array so
// it is recoverable via `qdf-hook expand <hash>`.
func TestDispatch_ColumnarRecoverable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var b strings.Builder
	b.WriteByte('[')
	for i := range 60 {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"id":%d,"status":"ok","name":"item-%d"}`, i, i)
	}
	b.WriteByte(']')
	content := b.String()

	store := hookcore.NewDiskStore()
	var o strings.Builder
	if err := hook.Dispatch(store, strings.NewReader(dispatchInput(t, "Bash", content)), &o); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(o.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("JSON array should be summarized")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(got, "COLUMNAR SUMMARY") {
		t.Errorf("expected columnar summary, got:\n%s", got)
	}
	if strings.Contains(got, "Use Read with offset") || strings.Contains(got, "[READ ") {
		t.Errorf("misleading Read-path wording must be gone, got:\n%s", got)
	}
	// Extract the recovery hash and confirm the raw array round-trips.
	marker := "qdf-hook expand "
	idx := strings.Index(got, marker)
	if idx < 0 {
		t.Fatalf("expected a recovery pointer, got:\n%s", got)
	}
	hash := strings.FieldsFunc(got[idx+len(marker):], func(r rune) bool {
		return !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f'))
	})[0]
	if recovered, ok := cache.RefGet(hash); !ok || recovered != content {
		t.Fatalf("raw array not recoverable via expand %s (ok=%v)", hash, ok)
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

// Skip-listed tools must still record an analytics event (action="skip") so
// `stats` reflects their invocations instead of being blind to them.
func TestDispatch_SkipList_RecordsAnalytics(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	big := strings.Repeat("todo item content line\n", 40)
	var out strings.Builder
	if err := hook.Dispatch(hookcore.NewDiskStore(), strings.NewReader(dispatchInput(t, "TodoWrite", big)), &out); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	evs, err := analytics.LoadEvents(0)
	if err != nil {
		t.Fatalf("LoadEvents: %v", err)
	}
	found := false
	for _, e := range evs {
		if e.Hook == "TodoWrite" && e.Action == "skip" {
			found = true
			if e.BytesIn != len(big) || e.BytesOut != len(big) {
				t.Errorf("skip event must record neutral bytes in=%d out=%d", e.BytesIn, e.BytesOut)
			}
		}
	}
	if !found {
		t.Fatal("no skip analytics event recorded")
	}
}
