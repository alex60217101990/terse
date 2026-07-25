package protocol

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// HookInput is the JSON Claude Code sends to any PostToolUse hook on stdin.
type HookInput struct {
	ToolResponse  *ToolResponse   `json:"tool_response,omitempty"`
	SessionID     string          `json:"session_id"`
	ToolName      string          `json:"tool_name"`
	ToolUseID     string          `json:"tool_use_id,omitempty"`
	HookEventName string          `json:"hook_event_name,omitempty"`
	ToolInput     json.RawMessage `json:"tool_input"`
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

	// Object shape. Decode "content" as a raw message so it can be either a
	// string (the usual case) or a content-block array (completed Agent).
	var r struct {
		File    *FileResponse   `json:"file"`
		Content json.RawMessage `json:"content"`
		Stdout  string          `json:"stdout"`
		Stderr  string          `json:"stderr"`
		Output  string          `json:"output"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	t.File = r.File
	t.Stdout = r.Stdout
	t.Stderr = r.Stderr
	t.Output = r.Output
	t.Content = contentText(r.Content)
	return nil
}

// contentText decodes a tool_response "content" value that may be a JSON
// string (the common case) or a content-block array (completed Agent),
// returning the plain text either way. Empty for anything else.
func contentText(raw json.RawMessage) string {
	head := bytes.TrimLeft(raw, " \t\r\n")
	if len(head) == 0 {
		return ""
	}
	switch head[0] {
	case '"':
		var s string
		if json.Unmarshal(raw, &s) == nil {
			return s
		}
	case '[':
		return blocksText(raw)
	}
	return ""
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

// HookOutput is the JSON written to stdout. Nil HookSpecificOutput = pass-through.
type HookOutput struct {
	HookSpecificOutput *HookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

// HookSpecificOutput carries the replacement for what Claude sees.
type HookSpecificOutput struct {
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

// EncodeOutput writes HookOutput as JSON to w followed by a newline.
func EncodeOutput(w io.Writer, out *HookOutput) error {
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

// Passthrough returns an empty HookOutput that tells Claude Code to use the original output.
func Passthrough() *HookOutput { return &HookOutput{} }

// Replace returns a HookOutput that substitutes the tool output with s.
func Replace(s string) *HookOutput {
	return &HookOutput{HookSpecificOutput: &HookSpecificOutput{UpdatedToolOutput: s}}
}
