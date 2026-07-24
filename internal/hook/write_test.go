package hook_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/hook"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

func makeWriteInput(t *testing.T, sessionID, path, content string) string {
	t.Helper()
	inp := map[string]any{
		"session_id":    sessionID,
		"tool_name":     "Write",
		"tool_input":    map[string]any{"file_path": path},
		"tool_response": map[string]any{"content": content},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

func TestWrite_SmallContent_Passthrough(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var out strings.Builder
	_ = hook.HandleWrite(strings.NewReader(makeWriteInput(t, "sess-w-1", "/f.go", "short")), &out)
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput != nil {
		t.Error("short content should passthrough")
	}
}

func TestWrite_LargeContent_Compressed(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	content := strings.Repeat("package main\nfunc foo() {}\n", 20) // 500+ bytes
	// The hook caches (and hashes) the ACTUAL file on disk, so the file must
	// exist; its bytes here equal `content`. The large tool_response echo is
	// what the compact marker replaces.
	path := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	_ = hook.HandleWrite(strings.NewReader(makeWriteInput(t, "sess-w-2", path, content)), &out)
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("a large echo of a cached file should be compressed")
	}
	compressed := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(compressed, "[WRITE") {
		t.Errorf("compressed output should start with [WRITE: %s", compressed)
	}
	// Hash prefix must match first 8 bytes of sha256(the file bytes).
	hash := sha256.Sum256([]byte(content))
	hashHex := fmt.Sprintf("%x", hash[:8])
	if !strings.Contains(compressed, hashHex) {
		t.Errorf("compressed output should contain hash %s: %s", hashHex, compressed)
	}
}

// TestWrite_CachesRealFileNotResponse pins the fix for caching the tool
// response (an Edit snippet) instead of the file's bytes. After a Write, a Read
// of the same file must resolve to §unchanged§, not a bogus delta between the
// snippet and the real file.
func TestWrite_CachesRealFileNotResponse(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "real.go")
	fileContent := strings.Repeat("package main\nfunc foo() {}\n", 20) // >256B
	if err := os.WriteFile(path, []byte(fileContent), 0o600); err != nil {
		t.Fatal(err)
	}

	// tool_response is a snippet that is NOT the file's content.
	snippet := "Applied 3 edits to real.go:\n" + strings.Repeat("  + added line\n", 20)
	var wout strings.Builder
	if err := hook.HandleWrite(strings.NewReader(makeWriteInput(t, "sess-w-real", path, snippet)), &wout); err != nil {
		t.Fatalf("HandleWrite: %v", err)
	}

	// Now Read the file with its ACTUAL content.
	rin := map[string]any{
		"session_id":    "sess-w-real",
		"tool_name":     "Read",
		"tool_input":    map[string]any{"file_path": path},
		"tool_response": map[string]any{"content": fileContent},
	}
	rb, _ := json.Marshal(rin)
	var rout strings.Builder
	if err := hook.HandleRead(strings.NewReader(string(rb)), &rout); err != nil {
		t.Fatalf("HandleRead: %v", err)
	}
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(rout.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("expected a compressed read output")
	}
	got := resp.HookSpecificOutput.UpdatedToolOutput
	if !strings.Contains(got, "§unchanged") {
		t.Errorf("Read after Write of the same file must be §unchanged§, got:\n%s", got)
	}
}

func TestWrite_CachesContent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	content := bytes.Repeat([]byte("x"), 300)
	path := filepath.Join(t.TempDir(), "a.go")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	err := hook.HandleWrite(strings.NewReader(makeWriteInput(t, "sess-w-3", path, string(content))), &out)
	if err != nil {
		t.Fatalf("HandleWrite returned error: %v", err)
	}
	// Output must be a valid HookOutput JSON (compressed, not empty).
	var resp protocol.HookOutput
	if jsonErr := json.Unmarshal([]byte(out.String()), &resp); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v", jsonErr)
	}
	if resp.HookSpecificOutput == nil {
		t.Fatal("content >256 bytes must produce a compressed output")
	}
	// The compressed marker must reference the expected path.
	if !strings.Contains(resp.HookSpecificOutput.UpdatedToolOutput, path) {
		t.Errorf("compressed output must contain the file path: %s",
			resp.HookSpecificOutput.UpdatedToolOutput)
	}
}

// TestWrite_EditShape_CachesAndCompresses uses the REAL Edit tool_response
// shape (no "content"/"stdout" — {filePath, oldString, newString,
// originalFile, structuredPatch}) verified from a live payload. handleWrite
// must still cache the on-disk file and replace the large echo with a marker,
// even though Text() is empty (echo size comes from the raw response length).
func TestWrite_EditShape_CachesAndCompresses(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path := filepath.Join(t.TempDir(), "edited.go")
	fileContent := strings.Repeat("package main\nfunc foo() {}\n", 20)
	if err := os.WriteFile(path, []byte(fileContent), 0o600); err != nil {
		t.Fatal(err)
	}
	// Edit-shaped response: no content field; originalFile makes it large.
	inp := map[string]any{
		"session_id": "sess-edit",
		"tool_name":  "Edit",
		"tool_input": map[string]any{"file_path": path},
		"tool_response": map[string]any{
			"filePath":        path,
			"oldString":       "func foo() {}",
			"newString":       "func foo() { return }",
			"originalFile":    fileContent,
			"structuredPatch": []any{},
		},
	}
	b, _ := json.Marshal(inp)
	var out strings.Builder
	if err := hook.HandleWrite(strings.NewReader(string(b)), &out); err != nil {
		t.Fatalf("HandleWrite: %v", err)
	}
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput == nil || !strings.Contains(resp.HookSpecificOutput.UpdatedToolOutput, "[WRITE") {
		t.Fatalf("Edit's large echo must be compressed to a [WRITE ...] marker, got: %v", out.String())
	}
	// And the file must be cached for delta priming: a later Read is §unchanged.
	rin := map[string]any{
		"session_id": "sess-edit", "tool_name": "Read",
		"tool_input":    map[string]any{"file_path": path},
		"tool_response": map[string]any{"file": map[string]any{"content": fileContent, "startLine": 1, "numLines": 40, "totalLines": 40}},
	}
	rb, _ := json.Marshal(rin)
	var rout strings.Builder
	_ = hook.HandleRead(strings.NewReader(string(rb)), &rout)
	var rr protocol.HookOutput
	_ = json.Unmarshal([]byte(rout.String()), &rr)
	if rr.HookSpecificOutput == nil || !strings.Contains(rr.HookSpecificOutput.UpdatedToolOutput, "§unchanged") {
		t.Errorf("Read after Edit must be §unchanged (delta primed by Edit), got: %v", rout.String())
	}
}
