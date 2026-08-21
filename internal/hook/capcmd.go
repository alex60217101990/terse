package hook

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/alex60217101990/terse/internal/bytesconv"
	"github.com/alex60217101990/terse/internal/cache"
	"github.com/alex60217101990/terse/internal/protocol"
)

// classDisqualify marks bytes that make a command unsafe or pointless to wrap.
// A table lookup keeps the scan branchless per byte; the two-character forms
// (&&, ||, $() are resolved by a single lookahead at the marked byte.
const classDisqualify uint8 = 1

var cmdClass [256]uint8

func init() {
	for _, c := range []byte{'|', '>', '<', '&', '`', '$'} {
		cmdClass[c] |= classDisqualify
	}
}

// cappable reports whether cmd may be wrapped in a capture.
//
// One forward pass, early exit on the first disqualifier, no allocation and no
// regexp: this runs before every Bash call the agent makes.
func cappable(cmd string) bool {
	b := bytesconv.S2B(cmd)
	if len(b) == 0 {
		return false
	}
	for i := 0; i < len(b); i++ {
		c := b[i]
		if c == '\\' {
			i++ // the next byte is escaped: it is data, not shell syntax
			continue
		}
		if cmdClass[c] == 0 {
			continue
		}
		switch c {
		case '&', '|':
			// Doubled forms are ordinary control flow and stay in the current
			// shell under brace grouping. A single one backgrounds or pipes.
			if i+1 < len(b) && b[i+1] == c {
				i++
				continue
			}
			return false
		case '$':
			if i+1 < len(b) && b[i+1] == '(' {
				return false
			}
			// A plain variable reference is fine.
		default: // '>', '<', '`'
			return false
		}
	}
	return true
}

// shellSafeArg reports whether s can be embedded inside a single-quoted shell
// word without escaping. A quote would close the quoting and hand the rest of
// the wrapper to the shell as syntax; a newline or control byte would split or
// corrupt the command. Refusing is always safe here — the command simply runs
// unwrapped.
func shellSafeArg(s string) bool {
	for i := range len(s) {
		if c := s[i]; c == '\'' || c < 0x20 {
			return false
		}
	}
	return true
}

// wrapCommand appends to dst a rewrite of cmd that captures the command's full
// output and prints a bounded view of it.
//
// Under capBytes the capture is printed verbatim, so the transcript is
// byte-identical to an unwrapped run. Over it, both ends are kept — a file's
// header and a build's first error live at the top, summaries and failure counts
// at the bottom — with one elision line between them carrying the recovery
// handle.
//
// The command group ends with a newline, not a semicolon, before its closing
// brace: cmd may carry a trailing "# comment", and a semicolon on that same
// line would fall inside the comment and vanish along with the brace that was
// supposed to close the group.
//
// The capture path is probed for writability before cmd ever runs. An
// unwritable path (a missing directory, a read-only filesystem) falls back to
// running cmd completely unwrapped, so a bad path never costs the agent its
// command's real output or exit code.
//
// The bookkeeping that decides what to print — exit code, capture size, both
// ends versus the full capture — runs inside its own subshell. That keeps two
// promises at once: the subshell's own "exit" reports cmd's real exit code as
// the wrapper's exit code without terminating the caller's persistent shell,
// and the bookkeeping variables it uses are local to that subshell and never
// touch the caller's shell, so there is nothing left over to clean up there.
//
// capturePath and id are embedded inside single-quoted shell words; either
// containing a quote or a control byte would let the rest of the wrapper be
// reinterpreted as shell syntax, so wrapCommand refuses and returns nil rather
// than emit a broken command. A nil result means cmd must be run untouched.
//
// dst is appended to, never reallocated by the caller's convention: pass a
// pooled buffer with enough capacity and this does no allocation at all.
func wrapCommand(dst []byte, cmd, capturePath, id string, capBytes int) []byte {
	if !shellSafeArg(capturePath) || !shellSafeArg(id) {
		return nil
	}

	half := int64(capBytes / 2)

	dst = append(dst, "if : 2>/dev/null > '"...)
	dst = append(dst, capturePath...)
	dst = append(dst, "'; then { "...)
	dst = append(dst, cmd...)
	dst = append(dst, "\n} > '"...)
	dst = append(dst, capturePath...)
	dst = append(dst, "' 2>&1\n(__qrc=$?; __qn=$(wc -c < '"...)
	dst = append(dst, capturePath...)
	dst = append(dst, "'); if [ \"$__qn\" -le "...)
	dst = strconv.AppendInt(dst, int64(capBytes), 10)
	dst = append(dst, " ]; then cat '"...)
	dst = append(dst, capturePath...)
	dst = append(dst, "'; else head -c "...)
	dst = strconv.AppendInt(dst, half, 10)
	dst = append(dst, " '"...)
	dst = append(dst, capturePath...)
	dst = append(dst, "'; printf '\\n... %s bytes elided, full output: qdf-hook expand "...)
	dst = append(dst, id...)
	dst = append(dst, "\\n' \"$__qn\"; tail -c "...)
	dst = strconv.AppendInt(dst, half, 10)
	dst = append(dst, " '"...)
	dst = append(dst, capturePath...)
	dst = append(dst, "'; fi; exit \"$__qrc\")\nelse { "...)
	dst = append(dst, cmd...)
	dst = append(dst, "\n}\nfi"...)
	return dst
}

// wrapOverhead is the fixed byte cost of wrapCommand's scaffolding, used to size
// a buffer exactly once. cmd is embedded twice (the writable and fallback
// branches) and capturePath six times, so a caller sizing a buffer should add
// those in on top of this constant, not fold them into it. Deliberately
// generous: a few spare bytes cost nothing, a reallocation costs an
// allocation on every Bash call.
const wrapOverhead = 420

// capBytes is the output budget, in bytes. 1600 is 400 tokens at the
// conservative 4-bytes-per-token factor: measured on 2,070 local transcripts a
// 400-token cap recovers 2.42% of the bill and still pays for itself even if
// 37% of capped calls need the full output back.
const capBytes = 1600

var wrapBuf = sync.Pool{New: func() any { b := make([]byte, 0, 1024); return &b }}

// captureSeq disambiguates same-session, same-command captures. The daemon
// keeps handling requests in one long-lived process, so two runs of the exact
// same command text in the same session are common (a retry, a loop body) —
// hashing session+cmd alone would give them the same id and the same capture
// file. The second run would then either clobber the first mid-write or, run
// concurrently, race on it, and a recovery hint already printed for the first
// run would silently start pointing at the second run's content. Folding in a
// counter that only ever increases makes every invocation's id unique without
// a syscall or an allocation of its own — atomic.Uint64.Add is one bus-locked
// instruction.
var captureSeq atomic.Uint64

// decodeCappableCommand extracts the Bash command from raw tool_input and
// reports whether it both decoded and passed the cappable scan. Malformed or
// uncappable input just means the command runs untouched, not an error to
// hand back up — so this returns a bool, never an error, on purpose.
func decodeCappableCommand(raw json.RawMessage) (string, bool) {
	var ti protocol.BashInput
	if json.Unmarshal(raw, &ti) != nil {
		return "", false
	}
	return ti.Command, cappable(ti.Command)
}

// captureDirReady reports whether the capture directory exists and is
// writable, creating it if needed. A failure here just means no capture is
// available for this command, not an error worth propagating, so this
// reports a bool rather than returning the underlying error.
func captureDirReady() bool {
	return os.MkdirAll(cache.CaptureDir(), 0o700) == nil
}

// handleBashPreToolUse rewrites a Bash command so the shell bounds its own
// output. Commands the scanner rejects are passed through in silence: writing
// anything at all — "{}" included — would make Claude Code record a
// hook_success attachment that then rides the prefix for the rest of the session.
func handleBashPreToolUse(inp *protocol.HookInput, w io.Writer) error {
	cmd, ok := decodeCappableCommand(inp.ToolInput)
	if !ok {
		return nil
	}
	seq := strconv.FormatUint(captureSeq.Add(1), 10)
	id := cache.RefHashOf(inp.SessionID + cmd + seq)
	path := cache.CapturePath(id)
	// The path is single-quoted into a shell command. The id is hex and the
	// directory is fixed, but $HOME is not ours — a quote in it would break out
	// of the quoting. Refuse to wrap rather than emit a malformed command.
	if strings.ContainsAny(path, "'\n") {
		return nil
	}
	if !captureDirReady() {
		return nil // capture unavailable: run the command untouched
	}
	// No defer: this is the pre-execution hot path for every Bash call.
	bp, ok := wrapBuf.Get().(*[]byte)
	if !ok {
		fresh := make([]byte, 0, 1024)
		bp = &fresh
	}
	need := len(cmd) + 2*len(path) + wrapOverhead
	if cap(*bp) < need {
		*bp = make([]byte, 0, need)
	}
	b := wrapCommand((*bp)[:0], cmd, path, id, capBytes)
	if b == nil {
		// wrapCommand refused (an unsafe path or id): run cmd untouched. The nil
		// return discards whatever was in the buffer, so hand it back empty.
		wrapBuf.Put(bp)
		return nil
	}
	*bp = b
	err := protocol.EncodePreInput(w, bytesconv.B2S(b))
	wrapBuf.Put(bp)
	return err
}
