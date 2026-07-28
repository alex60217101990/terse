package protocol_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/protocol"
)

// FuzzDecodeInputBytes exercises the hook's decode surface with arbitrary bytes:
// DecodeInputBytes and the polymorphic ToolResponse (object, MCP content-block
// array, Agent object-with-array-content, Read nested file.content). The
// property is simply "never panic" — malformed input must return an error, and
// any successfully decoded shape must be safe to inspect via Text/HasOutput/
// EchoLen. This is the class of bug that hit the MCP content-block array path.
func FuzzDecodeInputBytes(f *testing.F) {
	seeds := []string{
		`{"session_id":"s","tool_name":"Read","tool_input":{"file_path":"/x"},"tool_response":{"file":{"content":"hi","filePath":"/x","startLine":1,"numLines":1,"totalLines":1}}}`,
		`{"tool_name":"mcp__srv__q","hook_event_name":"PostToolUse","tool_input":{},"tool_response":[{"type":"text","text":"a"},{"type":"text","text":"b"}]}`,
		`{"tool_name":"Agent","tool_response":{"status":"completed","content":[{"type":"text","text":"report"}]}}`,
		`{"tool_name":"Bash","tool_response":{"stdout":"out","stderr":"err"}}`,
		`{"tool_name":"X","tool_response":[]}`,
		`{"tool_response":{"content":123}}`,
		`[]`,
		`{`,
		``,
		// escaped-unicode string content (surrogate pair emoji, quotes, backslash, newline)
		`{"tool_response":{"content":"café 😀 \"quoted\" \\backslash\\ \nnewline"}}`,
		// MCP-shaped content-block array nested under an object's "content" field
		// (as opposed to the seed above where tool_response itself is the array).
		`{"tool_response":{"content":[{"type":"text","text":"mcp block one"},{"type":"text","text":"mcp block two"}]}}`,
		// nested content array with a non-text block mixed in
		`{"tool_response":{"status":"completed","content":[{"type":"image","text":""},{"type":"text","text":"nested report"}]}}`,
		// null content
		`{"tool_response":{"content":null}}`,
		// huge escaped string content, to exercise the contentField fast path
		// on a large plain-string shape (the case the RawMessage removal targets).
		`{"tool_response":{"content":"` + strings.Repeat(`escaped \"chunk\" with \\backslash\\ and \ttab `, 500) + `"}}`,
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		inp, err := protocol.DecodeInputBytes(data)
		if err != nil {
			return // rejecting malformed input is correct; it must not panic
		}
		if inp == nil || inp.ToolResponse == nil {
			return
		}
		_ = inp.ToolResponse.Text()
		_ = inp.ToolResponse.HasOutput()
		_ = inp.ToolResponse.EchoLen()
	})
}
