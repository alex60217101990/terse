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
