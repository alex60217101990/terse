package protocol_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

func TestDecodeHookInput(t *testing.T) {
	raw := `{
		"session_id": "sess-abc123",
		"tool_name": "Read",
		"tool_input": {"file_path": "/tmp/foo.go", "limit": 200},
		"tool_response": {"content": "package main\n"}
	}`
	inp, err := protocol.DecodeInput(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("DecodeInput: %v", err)
	}
	if inp.SessionID != "sess-abc123" {
		t.Errorf("SessionID = %q, want sess-abc123", inp.SessionID)
	}
	if inp.ToolName != "Read" {
		t.Errorf("ToolName = %q, want Read", inp.ToolName)
	}
	if inp.ToolResponse == nil || inp.ToolResponse.Content != "package main\n" {
		t.Errorf("ToolResponse.Content unexpected: %+v", inp.ToolResponse)
	}
}

func TestEncodeHookOutput(t *testing.T) {
	out := &protocol.HookOutput{
		HookSpecificOutput: &protocol.HookSpecificOutput{
			UpdatedToolOutput: "§unchanged:abc§",
		},
	}
	var b strings.Builder
	if err := protocol.EncodeOutput(&b, out); err != nil {
		t.Fatalf("EncodeOutput: %v", err)
	}
	got := b.String()
	if !strings.Contains(got, "updatedToolOutput") {
		t.Errorf("output missing updatedToolOutput: %s", got)
	}
	if !strings.Contains(got, "§unchanged:abc§") {
		t.Errorf("output missing content: %s", got)
	}
}

// TestText_ReadFileContent locks in the live Read-tool shape: the file text
// is nested under tool_response.file.content (top-level "content" is empty),
// verified against a real PostToolUse payload. Text() must resolve it.
func TestText_ReadFileContent(t *testing.T) {
	raw := `{"session_id":"s","tool_name":"Read","hook_event_name":"PostToolUse",` +
		`"tool_input":{"file_path":"/x"},` +
		`"tool_response":{"type":"text","file":{"content":"package main\n","filePath":"/x","startLine":1,"numLines":1,"totalLines":1}}}`
	inp, err := protocol.DecodeInputBytes([]byte(raw))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got := inp.ToolResponse.Text(); got != "package main\n" {
		t.Errorf("Text() must resolve file.content, got %q", got)
	}
	if !inp.ToolResponse.HasOutput() {
		t.Error("HasOutput must be true when file.content is present")
	}
}
