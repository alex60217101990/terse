package hook_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/hook"
	"github.com/alex60217101990/terse/internal/protocol"
)

// catNRead renders count lines of the cat -n output Claude Code returns for a
// Read, starting at line start.
func catNRead(start, count int) string {
	var b strings.Builder
	for i := range count {
		fmt.Fprintf(&b, "%d\t\tresult, err := handler%02d(ctx, request)\n", start+i, i)
	}
	return b.String()
}

func readReq(t *testing.T, sessionID, path, content string, startLine, totalLines int) string {
	t.Helper()
	file := map[string]any{"content": content, "filePath": path}
	if startLine > 0 {
		file["startLine"] = startLine
		file["numLines"] = strings.Count(content, "\n")
		file["totalLines"] = totalLines
	}
	b, _ := json.Marshal(map[string]any{
		"session_id":    sessionID,
		"tool_name":     "Read",
		"tool_input":    map[string]any{"file_path": path},
		"tool_response": map[string]any{"file": file},
	})
	return string(b)
}

func runRead(t *testing.T, req string) protocol.HookOutput {
	t.Helper()
	var out strings.Builder
	if err := hook.HandleRead(strings.NewReader(req), &out); err != nil {
		t.Fatalf("HandleRead: %v", err)
	}
	var resp protocol.HookOutput
	// An empty response is the pass-through contract: a hook that has
	// nothing to say writes nothing, so no hook_success record is made.
	if raw := out.String(); raw != "" {
		if err := json.Unmarshal([]byte(raw), &resp); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}
	}
	return resp
}

// A first read used to pass through untouched — 1.45M tokens of one project's
// archive, at zero saving, most of a sixth of it line numbers the model can
// derive. It must now come back thinned.
func TestRead_FirstRead_GutterThinned(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	content := catNRead(1, 40)
	resp := runRead(t, readReq(t, "sess-g-1", "/srv/app/handler.go", content, 0, 0))
	if resp.HookSpecificOutput == nil {
		t.Fatal("a 40-line first read must be thinned, not passed through")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if len(got) >= len(content) {
		t.Fatalf("thinned output is not smaller: %d >= %d", len(got), len(content))
	}
	if !strings.HasPrefix(got, "1\t") {
		t.Errorf("the first line must keep its number:\n%.80s", got)
	}
	if !strings.Contains(got, "\n10\t") || !strings.Contains(got, "\n30\t") {
		t.Errorf("anchors at 10 and 30 are missing:\n%s", got)
	}
	if strings.Contains(got, "\n11\t") {
		t.Errorf("line 11 should have lost its number:\n%s", got)
	}
}

// Thinning happens on the way out, after the content is hashed and cached. If
// it leaked into the cache, the next identical read would diff the thinned form
// against the raw one and emit a bogus delta.
func TestRead_ThinningDoesNotPoisonTheDeltaCache(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	content := catNRead(1, 40)
	req := readReq(t, "sess-g-2", "/srv/app/handler.go", content, 0, 0)

	if resp := runRead(t, req); resp.HookSpecificOutput == nil {
		t.Fatal("first read must be thinned")
	}
	resp := runRead(t, req)
	if resp.HookSpecificOutput == nil {
		t.Fatal("second identical read must compress")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(got, "§unchanged") {
		t.Errorf("re-reading identical content must be §unchanged, not a delta:\n%s", got)
	}
}

// Windowed reads were returning without recording anything at all, so they were
// invisible in the stats and untouched by every transform. They are 170k tokens
// of the same archive.
func TestRead_WindowedRead_GutterThinned(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	content := catNRead(137, 40)
	resp := runRead(t, readReq(t, "sess-g-3", "/srv/app/handler.go", content, 137, 900))
	if resp.HookSpecificOutput == nil {
		t.Fatal("a windowed read must be thinned")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.HasPrefix(got, "137\t") {
		t.Errorf("the window's first line must keep its absolute number:\n%.80s", got)
	}
	if !strings.Contains(got, "\n140\t") {
		t.Errorf("anchor at absolute line 140 is missing:\n%s", got)
	}
}

// Never-worse: a payload the transform cannot help must come back untouched.
func TestRead_UngutteredContent_PassesThrough(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	content := strings.Repeat("plain content with no line numbers at all\n", 40)
	resp := runRead(t, readReq(t, "sess-g-4", "/srv/app/notes.txt", content, 0, 0))
	if resp.HookSpecificOutput != nil {
		t.Fatalf("content without a gutter must pass through, got:\n%s",
			resp.HookSpecificOutput.UpdatedToolOutput)
	}
}
