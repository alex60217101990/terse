package protocol

import (
	"encoding/json"
	"io"
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
// expose their output under different keys: Read/Glob/Grep/Write use "content",
// while Bash uses "stdout"/"stderr". Text() returns whichever is present.
type ToolResponse struct {
	Content string `json:"content"`
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
	Output  string `json:"output"` // some tools use a generic "output" key
}

// Text returns the tool's textual output, preferring "content", then a generic
// "output", then combined stdout+stderr (the Bash shape).
func (t *ToolResponse) Text() string {
	switch {
	case t.Content != "":
		return t.Content
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
	return t.Content != "" || t.Output != "" || t.Stdout != "" || t.Stderr != ""
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
