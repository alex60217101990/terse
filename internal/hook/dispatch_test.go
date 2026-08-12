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
	if !strings.Contains(got, "[repeat: ") {
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

// A Bash grep summary elides per-file match lines past grepFileCap, so the raw
// output must be recoverable: the summary carries a `qdf-hook expand <hash>`
// footer and that hash round-trips through the ref store back to the original.
func TestDispatch_BashGrep_Recoverable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
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
	marker := "qdf-hook expand "
	idx := strings.Index(got, marker)
	if idx < 0 {
		t.Fatalf("expected a recovery pointer, got:\n%s", got)
	}
	hash := strings.FieldsFunc(got[idx+len(marker):], func(r rune) bool {
		return !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f'))
	})[0]
	if recovered, ok := cache.RefGet(hash); !ok || recovered != content {
		t.Fatalf("raw grep output not recoverable via expand %s (ok=%v)", hash, ok)
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

// TestDispatch_BashJSONObject_SummarizedAndRecoverable exercises the
// tryJSONObject branch wired in right after tryJSON in dispatch.go: a large
// single JSON object (config/API dump, not an array) must be reduced to a key
// schema and made recoverable via the standard withRecovery footer.
func TestDispatch_BashJSONObject_SummarizedAndRecoverable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var b strings.Builder
	b.WriteString(`{"service":"widget-api","version":"2.3.1","replicas":4,"debug":false,`)
	b.WriteString(`"description":"` + strings.Repeat("long descriptive text ", 40) + `",`)
	b.WriteString(`"items":[`)
	for i := range 30 {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"id":%d,"state":"ready","zone":"eu-%d"}`, i, i%3)
	}
	b.WriteString(`],"limits":{"cpu":"2","mem":"4Gi"}}`)
	content := b.String()

	store := hookcore.NewDiskStore()
	var o strings.Builder
	if err := hook.Dispatch(store, strings.NewReader(dispatchInput(t, "Bash", content)), &o); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(o.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("large JSON object should be summarized")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(got, "[JSON OBJECT") {
		t.Errorf("expected JSON object summary, got:\n%s", got)
	}
	marker := "qdf-hook expand "
	idx := strings.Index(got, marker)
	if idx < 0 {
		t.Fatalf("expected a recovery pointer, got:\n%s", got)
	}
	hash := strings.FieldsFunc(got[idx+len(marker):], func(r rune) bool {
		return !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f'))
	})[0]
	if recovered, ok := cache.RefGet(hash); !ok || recovered != content {
		t.Fatalf("raw object not recoverable via expand %s (ok=%v)", hash, ok)
	}
}

func TestDispatch_BashGitDiff_FoldedAndRecoverable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var b strings.Builder
	b.WriteString("diff --git a/x.go b/x.go\nindex 1..2 100644\n--- a/x.go\n+++ b/x.go\n@@ -1,60 +1,61 @@\n")
	for range 55 {
		b.WriteString(" \tunchanged padding line with some width to it\n")
	}
	b.WriteString("-old\n+new\n")
	content := b.String()
	resp := runDispatch(t, "Bash", content)
	if resp.HookSpecificOutput == nil {
		t.Fatal("large diff should compress")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(got, "⋯") || !strings.Contains(got, "qdf-hook expand ") {
		t.Errorf("expected fold marker + recovery footer:\n%s", got)
	}
	if !strings.Contains(got, "-old") || !strings.Contains(got, "+new") {
		t.Errorf("changed lines must survive verbatim:\n%s", got)
	}
}

func TestDispatch_BashGoPanic_Summarized(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var b strings.Builder
	b.WriteString("panic: runtime error: index out of range [7] with length 3\n\ngoroutine 1 [running]:\n")
	for i := range 40 {
		fmt.Fprintf(&b, "example.com/pkg.func%d(0x%x)\n\t/src/pkg/file%d.go:%d +0x%x\n", i, i, i, 10+i, i)
	}
	resp := runDispatch(t, "Bash", b.String())
	if resp.HookSpecificOutput == nil {
		t.Fatal("long panic should compress")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(got, "panic: runtime error") || !strings.Contains(got, "⋯") {
		t.Errorf("message + fold marker required:\n%s", got)
	}
}

func TestDispatch_BashDockerPs_TableSummarized(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var b strings.Builder
	b.WriteString("CONTAINER ID   IMAGE          COMMAND     STATUS        PORTS\n")
	for i := range 30 {
		fmt.Fprintf(&b, "%012x   web-%02d:latest  \"/run.sh\"   Up %d hours    80/tcp\n", i, i, i%24)
	}
	resp := runDispatch(t, "Bash", b.String())
	if resp.HookSpecificOutput == nil {
		t.Fatal("30-row table should compress")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(got, "[TABLE 30 rows") || !strings.Contains(got, "qdf-hook expand ") {
		t.Errorf("expected table summary + recovery:\n%s", got)
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

// Path-dense output from a path-heavy action gets its shared directory
// folded via detect.FoldPathPrefix. buildGlobTree already collapses Glob
// output into per-directory counts (no repeated per-line paths survive to
// fold), so this exercises the grep "grouped" path instead: 6 files under
// one long shared directory, 3 matches each, produces 6 group-header lines
// that repeat the long prefix — exactly what FoldPathPrefix targets.
func TestDispatch_GrepLongPaths_PrefixFolded(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var b strings.Builder
	for i := range 6 {
		file := fmt.Sprintf("/Users/dev/work/src/github.com/acme/widget-service/internal/detect/file%02d.go", i)
		for j := 1; j <= 3; j++ {
			fmt.Fprintf(&b, "%s:%d:some match text %d\n", file, j, j)
		}
	}
	resp := runDispatch(t, "Grep", b.String())
	if resp.HookSpecificOutput == nil {
		t.Fatal("grep should compress")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(got, "[^=") {
		t.Errorf("expected prefix fold:\n%s", got)
	}
}

// A file dumped through Bash with a header line in front of it — `wc -l f &&
// cat -n f`, or a loop echoing a name before each file — is a numbered listing
// no summarizer claims and the strict Read-path check refuses. The runs get
// thinned anyway, and every body line survives byte-for-byte.
func TestDispatch_BashCatNWithHeader_GutterThinned(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var b strings.Builder
	b.WriteString("===== internal/detect/readgutter.go =====\n")
	for i := 1; i <= 40; i++ {
		fmt.Fprintf(&b, "%d\t\tif err := step%02d(ctx); err != nil {\n", i, i)
	}
	in := b.String()

	resp := runDispatch(t, "Bash", in)
	if resp.HookSpecificOutput == nil {
		t.Fatal("a numbered dump behind a header should compress")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(got, "===== internal/detect/readgutter.go =====") {
		t.Errorf("the header must survive:\n%s", got)
	}
	for _, anchor := range []string{"\n1\t", "\n10\t", "\n40\t"} {
		if !strings.Contains(got, anchor) {
			t.Errorf("anchor %q must survive:\n%s", anchor, got)
		}
	}
	if strings.Contains(got, "\n13\t") {
		t.Errorf("non-anchor line numbers must be gone:\n%s", got)
	}
	// The number takes its tab with it, so a thinned line keeps the file's own
	// indentation and nothing else.
	for i := 1; i <= 40; i++ {
		body := fmt.Sprintf("\tif err := step%02d(ctx); err != nil {", i)
		if !strings.Contains(got, body) {
			t.Fatalf("body of line %d was corrupted:\n%s", i, got)
		}
	}
}

// QDF_OFF is the A/B switch cmd/qdf-cost flips: the hook still runs, and still
// answers, but it must hand back the output untouched. A payload the pipeline
// would otherwise claim is the only honest test of that.
func TestDispatch_QDFOff_PassesEverythingThrough(t *testing.T) {
	t.Setenv("QDF_OFF", "1")
	var b strings.Builder
	b.WriteString("===== internal/detect/readgutter.go =====\n")
	for i := 1; i <= 40; i++ {
		fmt.Fprintf(&b, "%d\t\tif err := step%02d(ctx); err != nil {\n", i, i)
	}

	resp := runDispatch(t, "Bash", b.String())
	if resp.HookSpecificOutput != nil {
		t.Errorf("QDF_OFF must not rewrite the output:\n%s",
			resp.HookSpecificOutput.UpdatedToolOutput)
	}
}

// A PreToolUse event needs a permission decision, not an empty object. The
// QDF_OFF short-circuit used to answer both events the same way, which handed
// Claude Code `{}` where it expects hookSpecificOutput.permissionDecision.
func TestDispatch_QDFOff_PreToolUse_AnswersAllow(t *testing.T) {
	t.Setenv("QDF_OFF", "1")
	t.Setenv("HOME", t.TempDir())

	req := map[string]any{
		"session_id":      "off-pre",
		"hook_event_name": "PreToolUse",
		"tool_name":       "Read",
		"tool_input":      map[string]any{"file_path": "/nonexistent/file.go"},
	}
	body, _ := json.Marshal(req)

	var out strings.Builder
	if err := hook.Dispatch(hookcore.NewDiskStore(), strings.NewReader(string(body)), &out); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, `"permissionDecision":"allow"`) {
		t.Errorf("PreToolUse under QDF_OFF must allow, got: %s", got)
	}
	if !strings.Contains(got, `"hookEventName":"PreToolUse"`) {
		t.Errorf("response must name the event it answers, got: %s", got)
	}
}
