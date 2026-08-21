package protocol_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/protocol"
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

// TestDecode_MCPArrayToolResponse covers the MCP / content-block array shape
// Claude Code sends for MCP and completed Agent tools:
// "tool_response": [{"type":"text","text":"..."}]. It must decode (not error)
// and Text() must return the concatenated text so the compression pipeline can
// see and dedup/summarize it, instead of the whole request failing to parse.
func TestDecode_MCPArrayToolResponse(t *testing.T) {
	raw := `{
		"session_id": "s1",
		"tool_name": "mcp__server__query",
		"hook_event_name": "PostToolUse",
		"tool_input": {},
		"tool_response": [
			{"type": "text", "text": "first block"},
			{"type": "text", "text": "second block"}
		]
	}`
	inp, err := protocol.DecodeInput(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("array tool_response must decode, got error: %v", err)
	}
	if inp.ToolResponse == nil {
		t.Fatal("ToolResponse is nil for array shape")
	}
	got := inp.ToolResponse.Text()
	if !strings.Contains(got, "first block") || !strings.Contains(got, "second block") {
		t.Fatalf("Text() must contain both blocks, got %q", got)
	}
	if !inp.ToolResponse.HasOutput() {
		t.Error("HasOutput() must be true for a non-empty array response")
	}
}

// TestDecode_AgentContentArray covers a completed Agent's shape: tool_response
// is an OBJECT whose "content" is a content-block array rather than a string.
// It must decode and Text() must expose the block text.
func TestDecode_AgentContentArray(t *testing.T) {
	raw := `{
		"session_id": "s1",
		"tool_name": "Agent",
		"hook_event_name": "PostToolUse",
		"tool_input": {},
		"tool_response": {
			"status": "completed",
			"content": [{"type": "text", "text": "agent final report body"}]
		}
	}`
	inp, err := protocol.DecodeInput(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("object with array content must decode, got error: %v", err)
	}
	if inp.ToolResponse == nil || !strings.Contains(inp.ToolResponse.Text(), "agent final report body") {
		t.Fatalf("Text() must expose the block text, got %q", func() string {
			if inp.ToolResponse == nil {
				return "<nil>"
			}
			return inp.ToolResponse.Text()
		}())
	}
}

// TestDecode_PlainStringContent guards the common object+string-content path
// still works unchanged after the flexible-content refactor.
func TestDecode_PlainStringContent(t *testing.T) {
	raw := `{"session_id":"s","tool_name":"Bash","tool_input":{},"tool_response":{"content":"hello world"}}`
	inp, err := protocol.DecodeInput(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if inp.ToolResponse.Text() != "hello world" {
		t.Fatalf("Text() = %q, want %q", inp.ToolResponse.Text(), "hello world")
	}
}

// TestReplace_CarriesHookEventName locks the PostToolUse response contract:
// Claude Code validates hookSpecificOutput and requires hookEventName. Without
// it the whole object is rejected, the compressed replacement is dropped, and
// a validation error is injected into the context instead — costing far more
// than the compression saved.
func TestReplace_CarriesHookEventName(t *testing.T) {
	var b strings.Builder
	if err := protocol.EncodeOutput(&b, protocol.Replace("§ref:abc§")); err != nil {
		t.Fatalf("EncodeOutput: %v", err)
	}
	got := b.String()
	if !strings.Contains(got, `"hookEventName":"PostToolUse"`) {
		t.Errorf("Replace output missing hookEventName: %s", got)
	}
	if !strings.Contains(got, `"updatedToolOutput":"§ref:abc§"`) {
		t.Errorf("Replace output missing updatedToolOutput: %s", got)
	}
}

// TestPassthrough_WritesNothing pins the cheapest hook response there is:
// silence. Claude Code records a hook_success attachment for every hook that
// writes anything at all — even "{}" — and that record then rides the prefix for
// the rest of the session. Measured across 2,070 local transcripts: 26,517 such
// empty records cost 6.01% of the entire token bill. A hook that writes nothing
// produces no record, and means exactly what "{}" meant: do nothing.
//
// This matches what the daemon already does when dispatch fails — it closes the
// connection with no reply and calls that a safe passthrough.
func TestPassthrough_WritesNothing(t *testing.T) {
	var b strings.Builder
	if err := protocol.EncodeOutput(&b, protocol.Passthrough()); err != nil {
		t.Fatalf("EncodeOutput: %v", err)
	}
	if got := b.String(); got != "" {
		t.Errorf("Passthrough must write nothing, wrote %q", got)
	}
}

// TestEncodeOutput_NilIsSilent covers the nil handle the daemon can reach.
func TestEncodeOutput_NilIsSilent(t *testing.T) {
	var b strings.Builder
	if err := protocol.EncodeOutput(&b, nil); err != nil {
		t.Fatalf("EncodeOutput(nil): %v", err)
	}
	if got := b.String(); got != "" {
		t.Errorf("nil output must write nothing, wrote %q", got)
	}
}
