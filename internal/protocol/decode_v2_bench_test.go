package protocol_test

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"strings"
	"sync"
	"testing"

	"github.com/alex60217101990/terse/internal/protocol"
)

// A PreToolUse Bash request as Claude Code sends it.
var hookReq = []byte(`{"session_id":"6f1c9a3e-2b7d-4c11-9a5e-0f7b2d8c1e34",` +
	`"transcript_path":"/Users/x/.claude/projects/-Users-x-work/6f1c9a3e.jsonl",` +
	`"cwd":"/Users/x/work/src/github.com/terse","permission_mode":"default",` +
	`"hook_event_name":"PreToolUse","tool_name":"Bash",` +
	`"tool_input":{"command":"GOWORK=off go test ./internal/hook/ -run TestSomething",` +
	`"description":"Run the hook tests"}}`)

// hookFields is what the pipeline actually reads out of a request.
type hookFields struct {
	SessionID     string          `json:"session_id"`
	ToolName      string          `json:"tool_name"`
	HookEventName string          `json:"hook_event_name"`
	AgentID       string          `json:"agent_id"`
	Transcript    string          `json:"transcript_path"`
	ToolInput     json.RawMessage `json:"tool_input"`
}

func BenchmarkDecode_V1Unmarshal(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var f hookFields
		if err := json.Unmarshal(hookReq, &f); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_V2Unmarshal(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var f hookFields
		if err := jsonv2.Unmarshal(hookReq, &f); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecode_JsontextScan pulls the same fields with the token reader,
// keeping every value as a sub-slice of the request instead of a copy.
func BenchmarkDecode_JsontextScan(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var f struct {
			sessionID, toolName, event, agentID, transcript, toolInput []byte
		}
		v := jsontext.Value(hookReq)
		if err := scanTop(v, func(key, val []byte) {
			switch string(key) {
			case "session_id":
				f.sessionID = val
			case "tool_name":
				f.toolName = val
			case "hook_event_name":
				f.event = val
			case "agent_id":
				f.agentID = val
			case "transcript_path":
				f.transcript = val
			case "tool_input":
				f.toolInput = val
			}
		}); err != nil {
			b.Fatal(err)
		}
		if len(f.toolInput) == 0 {
			b.Fatal("no tool_input")
		}
	}
}

// scanTop walks the top-level object of v, handing each member's key and raw
// value to fn without copying either.
func scanTop(v jsontext.Value, fn func(key, val []byte)) error {
	dec := jsontext.NewDecoder(strings.NewReader(""))
	dec.Reset(strings.NewReader(string(v)))
	if _, err := dec.ReadToken(); err != nil { // '{'
		return err
	}
	for {
		tok, err := dec.ReadToken()
		if err != nil {
			return err
		}
		if tok.Kind() == '}' {
			return nil
		}
		key := tok.String()
		val, err := dec.ReadValue()
		if err != nil {
			return err
		}
		fn([]byte(key), val)
	}
}

// decPool keeps one decoder alive per call, the way the daemon would.
var decPool = sync.Pool{New: func() any { return jsontext.NewDecoder(bytes.NewReader(nil)) }}

// BenchmarkDecode_JsontextPooled is the honest version of the token scan.
//
// jsontext.Value from ReadValue aliases the decoder's buffer and is only valid
// until the next read — the decoder overwrites its first byte as it advances —
// so a key must be compared before the value is read, and any value the caller
// keeps must be copied into its own buffer. Here that is only tool_input,
// appended into a pooled scratch: one memmove instead of a string allocation.
func BenchmarkDecode_JsontextPooled(b *testing.B) {
	b.ReportAllocs()
	rd := bytes.NewReader(nil)
	scratch := make([]byte, 0, 1024)
	for b.Loop() {
		dec, _ := decPool.Get().(*jsontext.Decoder)
		rd.Reset(hookReq)
		dec.Reset(rd)
		var session, tool, event string
		scratch = scratch[:0]
		if _, err := dec.ReadToken(); err != nil {
			b.Fatal(err)
		}
		for dec.PeekKind() != '}' {
			key, err := dec.ReadValue()
			if err != nil {
				b.Fatal(err)
			}
			var want int // 1 session, 2 tool, 3 event, 4 input
			switch string(key) {
			case `"session_id"`:
				want = 1
			case `"tool_name"`:
				want = 2
			case `"hook_event_name"`:
				want = 3
			case `"tool_input"`:
				want = 4
			}
			val, err := dec.ReadValue()
			if err != nil {
				b.Fatal(err)
			}
			switch want {
			case 1:
				session = string(val)
			case 2:
				tool = string(val)
			case 3:
				event = string(val)
			case 4:
				scratch = append(scratch, val...)
			}
		}
		decPool.Put(dec)
		if session == "" || tool == "" || event == "" || len(scratch) == 0 {
			b.Fatal("missing field")
		}
	}
}

func BenchmarkDecode_Protocol(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if _, err := protocol.DecodeInputBytes(hookReq); err != nil {
			b.Fatal(err)
		}
	}
}
