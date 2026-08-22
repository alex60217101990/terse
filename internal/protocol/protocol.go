package protocol

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	"io"
	"strings"
	"sync"

	"github.com/alex60217101990/terse/internal/bytesconv"
)

// HookInput is the JSON Claude Code sends to any PostToolUse hook on stdin.
type HookInput struct {
	ToolResponse *ToolResponse `json:"tool_response,omitempty"`
	SessionID    string        `json:"session_id"`
	ToolName     string        `json:"tool_name"`
	// ToolUseID is never read anywhere in the codebase, so we skip decoding it
	// (json:"-") to avoid the per-event string alloc + copy. The field is kept
	// for forward-compat: restoring the tag is all it takes to start reading it.
	ToolUseID     string `json:"-"`
	HookEventName string `json:"hook_event_name,omitempty"`
	// AgentID and TranscriptPath identify the CONTEXT a tool call came from,
	// which session_id does not: Claude Code reports a subagent's calls under
	// the PARENT's session_id, but a subagent has its own context that is
	// discarded when it finishes — nothing it reads reaches the parent.
	//
	// AgentID is the reliable discriminator: it is documented as present only
	// on subagent tool calls. TranscriptPath is a fallback for builds that do
	// not send it; subagent transcripts live under
	// <project>/<session-id>/subagents/. hook.ContextKey combines the two into
	// the state-store key so a subagent's reads cannot deny the parent's.
	AgentID        string          `json:"agent_id,omitempty"`
	TranscriptPath string          `json:"transcript_path,omitempty"`
	ToolInput      json.RawMessage `json:"tool_input"`
}

// ToolResponse holds the raw tool output Claude Code produced. Different tools
// expose their output under different keys: Bash uses "stdout"/"stderr"; some
// tools a top-level "content"/"output"; and the Read tool nests the file text
// under a "file" object ({content, filePath, startLine, numLines, totalLines})
// — verified against a live PostToolUse payload. Text() returns whichever is
// present.
type ToolResponse struct {
	File *FileResponse `json:"file"` // Read tool: file text + window metadata

	Content string `json:"content"`
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
	Output  string `json:"output"` // some tools use a generic "output" key

	rawLen int // byte length of the raw tool_response JSON (echo size); see EchoLen
}

// UnmarshalJSON records the raw byte length of the tool_response object before
// decoding its fields. Some tools (Edit/MultiEdit) put no plain-text echo in
// any field this struct reads, yet their response is large (Edit embeds the
// whole originalFile) — EchoLen lets handleWrite size that echo for its
// never-worse compression decision, which Text()==0 could not.
func (t *ToolResponse) UnmarshalJSON(data []byte) error {
	t.rawLen = len(data)

	// MCP tools and completed Agent tasks don't use the plain object shape the
	// fields below expect: MCP sends tool_response as a content-block ARRAY
	// ([{"type":"text","text":"..."}, ...]) and a completed Agent sends an
	// object whose "content" is that same block array (not a string). Decoding
	// either into ToolResponse straight used to error and abort the whole hook
	// (empty output, zero compression for every MCP/Agent result). Fold both
	// into Content so Text() exposes them to the pipeline like any other output.
	if head := bytes.TrimLeft(data, " \t\r\n"); len(head) > 0 && head[0] == '[' {
		t.Content = blocksText(data)
		return nil
	}

	// Object shape. Decode "content" via contentField so it can be either a
	// string (the usual case) or a content-block array (completed Agent)
	// without routing through an intermediate json.RawMessage copy.
	var r struct {
		File    *FileResponse `json:"file"`
		Content contentField  `json:"content"`
		Stdout  string        `json:"stdout"`
		Stderr  string        `json:"stderr"`
		Output  string        `json:"output"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	t.File = r.File
	t.Stdout = r.Stdout
	t.Stderr = r.Stderr
	t.Output = r.Output
	t.Content = string(r.Content)
	return nil
}

// contentField decodes a tool_response "content" value in place, matching
// the former contentText helper's semantics exactly (plain JSON string; MCP
// / Agent content-block array; null; anything else -> ""), but operating
// directly on the raw per-field bytes encoding/json hands to UnmarshalJSON.
// For the dominant plain-string shape this avoids the json.RawMessage copy
// of the (escaped) content bytes that contentText's json.Unmarshal(raw,
// &r) previously produced before contentText re-Unmarshaled them.
type contentField string

// UnmarshalJSON implements json.Unmarshaler for contentField.
func (f *contentField) UnmarshalJSON(data []byte) error {
	// Fast path: plain JSON string, the dominant shape on non-Read traffic.
	// Decode straight into the field — one Unmarshal call, no RawMessage
	// copy, no intermediate string variable.
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, (*string)(f)); err != nil {
			*f = ""
		}
		return nil
	}
	head := bytes.TrimLeft(data, " \t\r\n")
	if len(head) == 0 {
		*f = ""
		return nil
	}
	switch head[0] {
	case '"':
		// Leading whitespace before the opening quote (should not occur via
		// encoding/json's own decoder, but kept for parity with the old
		// contentText which trimmed before switching on head[0]).
		if err := json.Unmarshal(head, (*string)(f)); err != nil {
			*f = ""
		}
	case '[':
		*f = contentField(blocksText(data))
	default:
		*f = ""
	}
	return nil
}

// blocksText concatenates the "text" of every block in a content-block array
// ([{"type":"text","text":"..."}, ...]), skipping non-text blocks.
func blocksText(raw []byte) string {
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) != nil {
		return ""
	}
	var sb strings.Builder
	for _, b := range blocks {
		if b.Text == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(b.Text)
	}
	return sb.String()
}

// EchoLen is the byte size of the raw tool_response — a proxy for how much the
// unmodified tool output would cost, used to gate/measure compression.
func (t *ToolResponse) EchoLen() int { return t.rawLen }

// FileResponse is the Read tool's "file" object. Content is the raw file text
// (not line-numbered); StartLine/NumLines/TotalLines describe the returned
// window (a partial read has StartLine>1 or NumLines<TotalLines).
type FileResponse struct {
	Content    string `json:"content"`
	FilePath   string `json:"filePath"`
	StartLine  int    `json:"startLine"`
	NumLines   int    `json:"numLines"`
	TotalLines int    `json:"totalLines"`
}

// Text returns the tool's textual output, preferring a top-level "content",
// then the Read "file.content", then a generic "output", then combined
// stdout+stderr (the Bash shape).
func (t *ToolResponse) Text() string {
	switch {
	case t.Content != "":
		return t.Content
	case t.File != nil && t.File.Content != "":
		return t.File.Content
	case t.Output != "":
		return t.Output
	case t.Stdout != "" && t.Stderr != "":
		return t.Stdout + "\n" + t.Stderr
	case t.Stderr != "":
		return t.Stderr
	default:
		return t.Stdout
	}
}

// HasOutput reports whether any output field is populated (used to distinguish
// a missing tool_response from an empty one).
func (t *ToolResponse) HasOutput() bool {
	return t.Content != "" || t.Output != "" || t.Stdout != "" || t.Stderr != "" ||
		(t.File != nil && t.File.Content != "")
}

// ReadInput is the parsed tool_input for the Read tool.
type ReadInput struct {
	FilePath string `json:"file_path"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// BashInput is the parsed tool_input for the Bash tool.
type BashInput struct {
	Command string `json:"command"`
	Cwd     string `json:"working_directory,omitempty"`
}

// EventPostToolUse is the hookEventName Claude Code expects on every
// PostToolUse hookSpecificOutput object.
const EventPostToolUse = "PostToolUse"

// HookOutput is the JSON written to stdout. Nil HookSpecificOutput = pass-through.
type HookOutput struct {
	HookSpecificOutput *HookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

// HookSpecificOutput carries the replacement for what Claude sees.
//
// HookEventName is mandatory: Claude Code validates hookSpecificOutput and
// rejects the whole object when it is absent. A rejected object drops the
// replacement (the original, uncompressed output reaches the model) and injects
// a validation error into the context on top — the opposite of the intent.
type HookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	UpdatedToolOutput string `json:"updatedToolOutput,omitempty"`
}

// DecodeInput reads exactly one JSON object from r.
func DecodeInput(r io.Reader) (*HookInput, error) {
	// Read the whole request first: it is one small JSON object, and the byte
	// path below hands out values that alias this buffer, which a streaming
	// decoder's rotating buffer could not promise. The buffer is not pooled for
	// the same reason — the HookInput outlives this call.
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return DecodeInputBytes(data)
}

// DecodeInputBytes decodes one HookInput from a fully-buffered request.
//
// The scan is zero-copy where it can be: tool_input and tool_response are
// sliced straight out of data instead of being copied, and only the handful of
// string fields the pipeline keeps are allocated. data must stay valid and
// unmodified for as long as the returned HookInput is used.
func DecodeInputBytes(data []byte) (*HookInput, error) {
	inp := &HookInput{}
	// The five string fields are gathered raw and materialized together, so a
	// request costs one string allocation instead of one per field.
	var raw [numStrFields][]byte
	err := ScanObject(data, func(key, val []byte) error {
		switch bytesconv.B2S(key) {
		case "session_id":
			raw[fieldSessionID] = val
		case "tool_name":
			raw[fieldToolName] = val
		case "hook_event_name":
			raw[fieldEventName] = val
		case "agent_id":
			raw[fieldAgentID] = val
		case "transcript_path":
			raw[fieldTranscript] = val
		case "tool_input":
			inp.ToolInput = val
		case "tool_response":
			if len(val) == 0 || bytesconv.B2S(val) == "null" {
				return nil
			}
			tr := &ToolResponse{}
			if err := tr.UnmarshalJSON(val); err != nil {
				return err
			}
			inp.ToolResponse = tr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var out [numStrFields]string
	joinStrings(&out, &raw)
	inp.SessionID = out[fieldSessionID]
	inp.ToolName = out[fieldToolName]
	inp.HookEventName = out[fieldEventName]
	inp.AgentID = out[fieldAgentID]
	inp.TranscriptPath = out[fieldTranscript]
	return inp, nil
}

// The string fields of a request, in the order they are packed.
const (
	fieldSessionID = iota
	fieldToolName
	fieldEventName
	fieldAgentID
	fieldTranscript
	numStrFields
)

// EncodeOutput writes HookOutput as JSON to w followed by a newline, and writes
// NOTHING when there is nothing to say.
//
// Silence is not cosmetic. Claude Code records a hook_success attachment for
// every hook that writes anything at all, "{}" included, and that record then
// rides the conversation prefix and is re-billed on every later turn. Measured
// across 2,070 local transcripts: 26,517 empty records cost 6.01% of the entire
// token bill. Writing nothing produces no record and carries exactly the meaning
// "{}" carried — take no action — which is also what the daemon already falls
// back to when dispatch fails.
func EncodeOutput(w io.Writer, out *HookOutput) error {
	if out == nil || out.HookSpecificOutput == nil {
		return nil
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(out)
}

// EncodePre writes a PreToolUse hook response to w.
// decision is "allow" or "deny". reason is only used when decision is "deny".
func EncodePre(w io.Writer, decision, reason string) error {
	type preOut struct {
		HookSpecificOutput struct {
			HookEventName            string `json:"hookEventName"`
			PermissionDecision       string `json:"permissionDecision"`
			PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
		} `json:"hookSpecificOutput"`
	}
	var out preOut
	out.HookSpecificOutput.HookEventName = "PreToolUse"
	out.HookSpecificOutput.PermissionDecision = decision
	out.HookSpecificOutput.PermissionDecisionReason = reason
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(out)
}

// EncodePreInput writes a PreToolUse response that allows the call and replaces
// the Bash tool's command with cmd.
//
// updatedInput is the only field Claude Code honors for input rewriting; the
// transcript still records the model's ORIGINAL command, so the replacement text
// costs no context tokens however long it is. Only the command's OUTPUT changes.
//
// permissionDecision "allow" is required, not decoration. Measured live against
// Claude Code on 2026-08-22: without it the rewritten command is rejected before
// it runs with "Contains compound_statement" — the permission analyzer refuses to
// evaluate the wrapper's if/then/else shape — and the tool call fails outright.
// With it, the same rewrite capped a 23,893-byte output to 1,689 bytes.
//
// What "allow" does NOT do, also measured: it does not override the user's
// rules. A deny rule on the model's ORIGINAL command still blocked the call
// through the rewrite, an ask rule still forced a prompt, and a deny rule
// matching only the wrapper's own text never fired — so rules are evaluated
// against what the model wrote, and the rewrite is invisible to them. The one
// thing it does skip is the prompt for a command no rule covers.
func EncodePreInput(w io.Writer, cmd string) error {
	bp, _ := preInputBuf.Get().(*[]byte)
	if bp == nil {
		fresh := make([]byte, 0, 2048)
		bp = &fresh
	}
	buf := *bp
	buf = append(buf[:0], PreInputHead...)
	buf = AppendJSONString(buf, cmd)
	if buf == nil {
		preInputBuf.Put(bp)
		return errUnencodable
	}
	buf = append(buf, PreInputTail...)
	_, err := w.Write(buf)
	*bp = buf
	preInputBuf.Put(bp)
	return err
}

// The response is a fixed shape around one string, so it is written by hand.
// encoding/json cost 60% of this hook's CPU and every one of its allocations,
// measured, for a document whose only variable part is the command.
const (
	// PreInputHead and PreInputTail bracket the rewritten command. They are
	// exported so a caller that already builds the command in a buffer can emit
	// the whole response in one pass instead of handing the text back to be
	// copied and rescanned.
	PreInputHead = `{"hookSpecificOutput":{"hookEventName":"PreToolUse",` +
		`"permissionDecision":"allow","updatedInput":{"command":"`
	PreInputTail = "\"}}}\n"
)

var preInputBuf = sync.Pool{New: func() any { b := make([]byte, 0, 2048); return &b }}

// AppendJSONString appends s to dst as the *body* of a JSON string — the
// surrounding quotes belong to the caller's literal.
//
// The escaping is the standard library's, through a jsontext encoder carrying
// the two options that reproduce encoding/json exactly: EscapeForJS restores
// the U+2028/U+2029 escapes, AllowInvalidUTF8 the replacement character for a
// byte that is not valid UTF-8. The differential test pins that equality.
//
// The encoder writes a whole JSON string, so its quotes and trailing newline
// are shifted off. Measured against a hand-rolled escaper over the real
// distribution of Bash commands (p50 160 B, mean 362, p90 793): 8% slower at
// the median, 9% faster at the mean, 18% at p90 and 24% at p99, at the same
// zero allocations.
//
// Returns nil if the encoder refuses the string, which leaves the caller to
// skip whatever it was building rather than emit a broken document.
func AppendJSONString(dst []byte, s string) []byte {
	return appendEscaped(dst, s, utf8Lenient)
}

// AppendJSONStringStrict is AppendJSONString for a string whose bytes must
// survive unchanged. It refuses (returns nil) a string that is not valid
// UTF-8, where the lenient form would substitute U+FFFD.
//
// The difference matters when the escaped string is a command: substituting a
// byte would hand the shell something other than what the model wrote, and a
// command that cannot be represented is better run unwrapped than altered.
func AppendJSONStringStrict(dst []byte, s string) []byte {
	return appendEscaped(dst, s, utf8Strict)
}

func appendEscaped(dst []byte, s string, utf8Opt jsontext.Options) []byte {
	e, _ := escaperPool.Get().(*escaper)
	if e == nil {
		e = newEscaper()
	}
	e.out.b = dst
	e.enc.Reset(&e.out, escapeOpts, utf8Opt)
	err := e.enc.WriteToken(jsontext.String(s))
	b := e.out.b
	e.out.b = nil
	escaperPool.Put(e)
	if err != nil {
		return nil
	}
	start := len(dst)
	copy(b[start:], b[start+1:len(b)-2]) // drop the opening quote
	return b[:len(b)-3]                  // and the closing quote plus newline
}

// escaper is a jsontext encoder bound to a slice appender, both pooled: the
// encoder is reset per call and the appender writes straight into the caller's
// buffer, so escaping a command allocates nothing.
type escaper struct {
	enc *jsontext.Encoder
	out sliceWriter
}

type sliceWriter struct{ b []byte }

func (w *sliceWriter) Write(p []byte) (int, error) {
	w.b = append(w.b, p...)
	return len(p), nil
}

var (
	escapeOpts  = jsontext.EscapeForJS(true)
	utf8Lenient = jsontext.AllowInvalidUTF8(true)
	utf8Strict  = jsontext.AllowInvalidUTF8(false)
	escaperPool = sync.Pool{New: func() any { return newEscaper() }}
)

func newEscaper() *escaper {
	e := &escaper{}
	e.enc = jsontext.NewEncoder(&e.out, escapeOpts, utf8Lenient)
	return e
}

// Passthrough returns an empty HookOutput that tells Claude Code to use the original output.
func Passthrough() *HookOutput { return &HookOutput{} }

// Replace returns a HookOutput that substitutes the tool output with s.
func Replace(s string) *HookOutput {
	return &HookOutput{HookSpecificOutput: &HookSpecificOutput{
		HookEventName:     EventPostToolUse,
		UpdatedToolOutput: s,
	}}
}
