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

	"github.com/alex60217101990/terse/internal/hook"
	"github.com/alex60217101990/terse/internal/protocol"
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

// A visible echo that genuinely costs more tokens than the marker is replaced.
// The echo here is prose, which tokenizes at roughly a token per word, so it
// clears the marker's cost (a 32-char hex hash is most of that cost).
func TestWrite_ExpensiveEcho_Compressed(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	echo := strings.Repeat("wrote the configuration section for the billing service\n", 12)
	path := filepath.Join(t.TempDir(), "a.go")
	if err := os.WriteFile(path, []byte(echo), 0o600); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err := hook.HandleWrite(strings.NewReader(makeWriteInput(t, "sess-w-3", path, echo)), &out); err != nil {
		t.Fatalf("HandleWrite returned error: %v", err)
	}
	var resp protocol.HookOutput
	if jsonErr := json.Unmarshal([]byte(out.String()), &resp); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v", jsonErr)
	}
	if resp.HookSpecificOutput == nil {
		t.Fatal("an echo far more expensive than the marker must be compressed")
	}
	if !strings.Contains(resp.HookSpecificOutput.UpdatedToolOutput, path) {
		t.Errorf("compressed output must contain the file path: %s",
			resp.HookSpecificOutput.UpdatedToolOutput)
	}
}

// The never-worse invariant, in the unit that matters. 300 repeated bytes are
// large enough to have passed the old byte-length gate, and cheap enough in
// tokens that the marker — most of it a 32-char hex hash — costs more. Emitting
// it would make the payload bigger, which is what the old gate did on 1251 real
// Write/Edit calls.
func TestWrite_CheapEcho_PassesThroughAndStillCaches(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	content := bytes.Repeat([]byte("x"), 300)
	path := filepath.Join(t.TempDir(), "a.go")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err := hook.HandleWrite(strings.NewReader(makeWriteInput(t, "sess-w-4", path, string(content))), &out); err != nil {
		t.Fatalf("HandleWrite returned error: %v", err)
	}
	var resp protocol.HookOutput
	if jsonErr := json.Unmarshal([]byte(out.String()), &resp); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v", jsonErr)
	}
	if resp.HookSpecificOutput != nil {
		t.Fatalf("a marker costlier than the echo must not be emitted, got: %s",
			resp.HookSpecificOutput.UpdatedToolOutput)
	}
	assertReadIsUnchanged(t, "sess-w-4", path, string(content))
}

// TestWrite_EditShape_CachesWithoutCompressing uses the REAL Edit
// tool_response shape (no "content"/"stdout" — {filePath, oldString,
// newString, originalFile, structuredPatch}) verified from a live payload.
//
// That shape is large in bytes — originalFile embeds the whole pre-edit file —
// but the model is never shown it. Claude Code renders one short sentence, and
// the sentence is what the tokens are spent on. This process cannot see it, so
// it cannot show a marker to be cheaper, and must not gamble: caching is the
// point of this hook, and caching happens either way.
func TestWrite_EditShape_CachesWithoutCompressing(t *testing.T) {
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
	if resp.HookSpecificOutput != nil {
		t.Fatalf("Edit's echo is not visible to the hook — it must pass through, got: %s",
			resp.HookSpecificOutput.UpdatedToolOutput)
	}
	// And the file must still be cached for delta priming, which is the whole
	// reason this hook runs on Edit at all.
	assertReadIsUnchanged(t, "sess-edit", path, fileContent)
}

// assertReadIsUnchanged checks that a Read of path in this session resolves to
// the §unchanged marker — i.e. the write/edit primed the delta cache.
func assertReadIsUnchanged(t *testing.T, sessionID, path, content string) {
	t.Helper()
	lines := strings.Count(content, "\n")
	rin := map[string]any{
		"session_id": sessionID, "tool_name": "Read",
		"tool_input": map[string]any{"file_path": path},
		"tool_response": map[string]any{"file": map[string]any{
			"content": content, "startLine": 1, "numLines": lines, "totalLines": lines,
		}},
	}
	rb, _ := json.Marshal(rin)
	var rout strings.Builder
	if err := hook.HandleRead(strings.NewReader(string(rb)), &rout); err != nil {
		t.Fatalf("HandleRead: %v", err)
	}
	var rr protocol.HookOutput
	_ = json.Unmarshal([]byte(rout.String()), &rr)
	if rr.HookSpecificOutput == nil || !strings.Contains(rr.HookSpecificOutput.UpdatedToolOutput, "§unchanged") {
		t.Errorf("Read after the write must be §unchanged (delta primed), got: %v", rout.String())
	}
}
