package protocol

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"unicode/utf8"
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
	var inp HookInput
	if err := json.NewDecoder(r).Decode(&inp); err != nil {
		return nil, err
	}
	return &inp, nil
}

// DecodeInputBytes decodes one HookInput from a fully-buffered request via
// json.Unmarshal, avoiding the json.Decoder buffering that DecodeInput's
// io.Reader path pays. For callers (the daemon) that already hold the whole
// request slice.
func DecodeInputBytes(data []byte) (*HookInput, error) {
	var inp HookInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil, err
	}
	return &inp, nil
}

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

const hexDigits = "0123456789abcdef"

// AppendJSONString appends s to dst as the *body* of a JSON string — the
// surrounding quotes belong to the caller's literal.
//
// It matches encoding/json with SetEscapeHTML(false) byte for byte, which the
// differential test pins: the two-character escapes where they exist, \u00xx for
// the other control bytes, \ufffd for a byte that is not valid UTF-8, and
// \u2028/\u2029 for the two line terminators JavaScript would otherwise treat as
// newlines. Everything else, non-ASCII included, is copied through.
func AppendJSONString(dst []byte, s string) []byte {
	start := 0
	for i := 0; i < len(s); {
		c := s[i]
		if c < utf8.RuneSelf {
			if c >= 0x20 && c != '"' && c != '\\' {
				i++
				continue
			}
			dst = append(dst, s[start:i]...)
			switch c {
			case '"', '\\':
				dst = append(dst, '\\', c)
			case '\n':
				dst = append(dst, '\\', 'n')
			case '\r':
				dst = append(dst, '\\', 'r')
			case '\t':
				dst = append(dst, '\\', 't')
			case '\b':
				dst = append(dst, '\\', 'b')
			case '\f':
				dst = append(dst, '\\', 'f')
			default:
				dst = append(dst, '\\', 'u', '0', '0', hexDigits[c>>4], hexDigits[c&0xf])
			}
			i++
			start = i
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			dst = append(dst, s[start:i]...)
			dst = append(dst, "\ufffd"...) // the replacement character itself, as encoding/json writes it
			i += size
			start = i
			continue
		}
		if r == '\u2028' || r == '\u2029' {
			dst = append(dst, s[start:i]...)
			dst = append(dst, '\\', 'u', '2', '0', '2', hexDigits[r&0xf])
			i += size
			start = i
			continue
		}
		i += size
	}
	return append(dst, s[start:]...)
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
