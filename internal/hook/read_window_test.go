package hook_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/hook"
	"github.com/alex60217101990/terse/internal/hookcore"
	"github.com/alex60217101990/terse/internal/protocol"
)

// dispatchReadEvent builds a PostToolUse Read event (tool_input file_path/
// offset/limit; tool_response.file content/startLine/numLines/totalLines),
// routes it through hook.Dispatch against store, and returns the decoded
// HookOutput.
func dispatchReadEvent(t *testing.T, store hookcore.StateStore, path, content string, offset, limit, startLine, numLines, totalLines int) protocol.HookOutput {
	t.Helper()
	evt := map[string]any{
		"session_id": "sess-window",
		"tool_name":  "Read",
		"tool_input": map[string]any{
			"file_path": path,
			"offset":    offset,
			"limit":     limit,
		},
		"tool_response": map[string]any{
			"file": map[string]any{
				"content":    content,
				"filePath":   path,
				"startLine":  startLine,
				"numLines":   numLines,
				"totalLines": totalLines,
			},
		},
	}
	raw, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	var out strings.Builder
	if err := hook.Dispatch(store, strings.NewReader(string(raw)), &out); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	var resp protocol.HookOutput
	if err := json.Unmarshal([]byte(out.String()), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v / raw: %s", err, out.String())
	}
	return resp
}

// TestRead_WindowUnchanged_Deduped: a windowed re-read of an UNCHANGED cached
// file collapses to a marker; a change on disk passes the window through.
func TestRead_WindowUnchanged_Deduped(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := t.TempDir()
	path := filepath.Join(dir, "big.txt")
	var content strings.Builder
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&content, "line %03d of the big file with padding text\n", i)
	}
	full := content.String()
	if err := os.WriteFile(path, []byte(full), 0o644); err != nil {
		t.Fatal(err)
	}
	store := hookcore.NewDiskStore()

	// 1) full read — populates the cache.
	dispatchReadEvent(t, store, path, full, 0, 0, 1, 100, 100)

	// 2) windowed read of lines 40–59 (identical bytes).
	lines := strings.SplitAfter(full, "\n")
	window := strings.Join(lines[39:59], "")
	out := dispatchReadEvent(t, store, path, window, 40, 20, 40, 20, 100)
	if out.HookSpecificOutput == nil {
		t.Fatal("unchanged window should dedup")
	}
	msg := out.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(msg, "§unchanged-window:") || !strings.Contains(msg, "lines 40–59") {
		t.Errorf("marker malformed: %s", msg)
	}
	if len(msg) >= len(window) {
		t.Errorf("never-worse violated: %d >= %d", len(msg), len(window))
	}

	// 3) file changes on disk → windowed read passes through.
	if err := os.WriteFile(path, []byte(full+"tail\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out = dispatchReadEvent(t, store, path, window, 40, 20, 40, 20, 101)
	if out.HookSpecificOutput != nil {
		t.Error("changed file must pass window through")
	}
}

// TestRead_Window_NothingCached: a windowed read with no cache entry passes
// through (no full read ever populated the slot).
func TestRead_Window_NothingCached(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "big.txt")
	var content strings.Builder
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&content, "line %03d of the big file with padding text\n", i)
	}
	full := content.String()
	if err := os.WriteFile(path, []byte(full), 0o644); err != nil {
		t.Fatal(err)
	}
	store := hookcore.NewDiskStore()

	lines := strings.SplitAfter(full, "\n")
	window := strings.Join(lines[39:59], "")
	out := dispatchReadEvent(t, store, path, window, 40, 20, 40, 20, 100)
	if out.HookSpecificOutput != nil {
		t.Error("window with nothing cached must pass through")
	}
}

// TestRead_Window_BytesDiffer: the file's mtime/ctime match the cache but the
// window bytes differ from the cached slice → passthrough (byte guard).
func TestRead_Window_BytesDiffer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "big.txt")
	var content strings.Builder
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&content, "line %03d of the big file with padding text\n", i)
	}
	full := content.String()
	if err := os.WriteFile(path, []byte(full), 0o644); err != nil {
		t.Fatal(err)
	}
	store := hookcore.NewDiskStore()

	// full read caches the real content + metadata.
	dispatchReadEvent(t, store, path, full, 0, 0, 1, 100, 100)

	// window claims lines 40–59 but supplies different bytes (line-numbered);
	// mtime/ctime still match so only the byte comparison rejects it.
	lines := strings.SplitAfter(full, "\n")
	realWindow := strings.Join(lines[39:59], "")
	tampered := strings.ReplaceAll(realWindow, "padding", "TAMPER!")
	out := dispatchReadEvent(t, store, path, tampered, 40, 20, 40, 20, 100)
	if out.HookSpecificOutput != nil {
		t.Errorf("window with differing bytes must pass through, got: %s",
			out.HookSpecificOutput.UpdatedToolOutput)
	}
}

// TestRead_Window_LastLineNoTrailingNewline: a window covering the final line
// of a file with no trailing newline still dedups (sliceLines EOF branch).
func TestRead_Window_LastLineNoTrailingNewline(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "notrail.txt")
	var content strings.Builder
	for i := 1; i <= 30; i++ {
		fmt.Fprintf(&content, "line %03d of a file with plenty of padding text here", i)
		if i != 30 {
			content.WriteByte('\n')
		}
	}
	full := content.String() // no trailing newline on the last line
	if err := os.WriteFile(path, []byte(full), 0o644); err != nil {
		t.Fatal(err)
	}
	store := hookcore.NewDiskStore()

	dispatchReadEvent(t, store, path, full, 0, 0, 1, 30, 30)

	// window covers lines 21–30 (the last line has no trailing '\n').
	lines := strings.SplitAfter(full, "\n")
	window := strings.Join(lines[20:30], "")
	out := dispatchReadEvent(t, store, path, window, 21, 10, 21, 10, 30)
	if out.HookSpecificOutput == nil {
		t.Fatal("window over last line (no trailing newline) should dedup")
	}
	if !strings.Contains(out.HookSpecificOutput.UpdatedToolOutput, "lines 21–30") {
		t.Errorf("marker malformed: %s", out.HookSpecificOutput.UpdatedToolOutput)
	}
}

// TestRead_Window_StartLinePastEOF: StartLine beyond the cached file's line
// count passes through (sliceLines ok=false).
func TestRead_Window_StartLinePastEOF(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "small.txt")
	full := "alpha\nbeta\ngamma\n"
	if err := os.WriteFile(path, []byte(full), 0o644); err != nil {
		t.Fatal(err)
	}
	store := hookcore.NewDiskStore()

	dispatchReadEvent(t, store, path, full, 0, 0, 1, 3, 3)

	// Ask for lines starting at 50 — well past EOF.
	out := dispatchReadEvent(t, store, path, "whatever\n", 50, 5, 50, 5, 3)
	if out.HookSpecificOutput != nil {
		t.Error("StartLine past EOF must pass through")
	}
}

// TestRead_Window_NeverWorseTinyWindow: a 1-line window whose marker would be
// longer than the window passes through (never-worse).
func TestRead_Window_NeverWorseTinyWindow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.txt")
	full := "a\nb\nc\nd\ne\n"
	if err := os.WriteFile(path, []byte(full), 0o644); err != nil {
		t.Fatal(err)
	}
	store := hookcore.NewDiskStore()

	dispatchReadEvent(t, store, path, full, 0, 0, 1, 5, 5)

	// window = line 3 only ("c\n"); the marker is far longer → passthrough.
	out := dispatchReadEvent(t, store, path, "c\n", 3, 1, 3, 1, 5)
	if out.HookSpecificOutput != nil {
		t.Errorf("tiny window must pass through (never-worse), got: %s",
			out.HookSpecificOutput.UpdatedToolOutput)
	}
}
