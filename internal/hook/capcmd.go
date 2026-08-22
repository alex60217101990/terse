package hook

import (
	"bytes"
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

// cappable reports whether cmd may be wrapped in a capture.
//
// Almost everything qualifies. A brace group is not a subshell, so a wrapped
// command keeps the caller's shell: cd, exports and assignments persist, exit
// codes propagate, pipelines keep their SIGPIPE behavior, heredocs feed the
// group, and a command's own redirections win over the group's because they are
// applied later (all verified against bash, sh and zsh).
//
// Two things are refused. A bare "&" backgrounds the command, and the wrapper
// would then read a capture the job is still writing. A bare interactive
// session (§6, denied below) has nothing to capture in the first place.
//
// One forward pass, early exit on the first disqualifier, no allocation and no
// regexp: this runs before every Bash call the agent makes.
func cappable(cmd string) bool {
	b := bytesconv.S2B(cmd)
	if len(b) == 0 {
		return false
	}
	for i := 0; i < len(b); i++ {
		switch b[i] {
		case '\\':
			i++ // the next byte is escaped: it is data, not shell syntax
		case '&':
			// Every other form of "&" is not a background: "&&" is control
			// flow, ">&"/"<&" duplicate a descriptor, "|&" pipes stderr too,
			// and "&>" redirects both streams.
			switch {
			case i+1 < len(b) && (b[i+1] == '&' || b[i+1] == '>'):
				i++
			case i > 0 && (b[i-1] == '>' || b[i-1] == '<' || b[i-1] == '|'):
			default:
				return false
			}
		}
	}
	return !denied(b)
}

// denied reports whether the command is on spec §6's interactive/streaming
// deny list: a program that drives a terminal, or a watcher that never returns.
//
// Wrapping redirects the program's stdout into a file, which is wrong twice
// over for these — a program the caller drives loses the terminal it expects,
// and a watcher that never returns loses the partial output the agent would
// otherwise see when the call times out.
//
// The list is deliberately narrow. A shell tool never types into a REPL, so
// "python3 -c ..." or "ssh host uptime" is an ordinary non-interactive command
// and capping it is exactly the point; only the bare, argument-less form opens
// a session. Measured on 24,323 corpus Bash calls, denying those with their
// arguments cost 1.5 points of coverage for nothing.
//
// The scan walks words in place: no split, no allocation, and the switches are
// over compile-time constants so comparing a subslice never copies it.
func denied(b []byte) bool {
	prog := true  // the next word sits in program position
	repl := false // a session program has been seen with nothing to run yet
	for i := 0; i < len(b); {
		for i < len(b) && (b[i] == ' ' || b[i] == '\t' || b[i] == '\n') {
			i++
		}
		if i < len(b) && b[i] == '|' {
			// A pipe stage is a command of its own: "cat f | less" has to be
			// caught as surely as "less f".
			i++
			if repl {
				return true
			}
			prog, repl = true, false
			continue
		}
		start := i
		for i < len(b) && b[i] != ' ' && b[i] != '\t' && b[i] != '\n' && b[i] != '|' {
			i++
		}
		w := b[start:i]
		if len(w) == 0 {
			break
		}
		// A separator closes the current command: a session program that
		// reached it was invoked bare, and the next word is a program again.
		sep := bytesconv.B2S(w) == "&&" || bytesconv.B2S(w) == "||"
		for len(w) > 0 && w[len(w)-1] == ';' {
			w, sep = w[:len(w)-1], true
		}
		switch bytesconv.B2S(w) {
		case "watch", "--watch":
			return true // a watcher never returns; its output would sit in the file
		}
		switch {
		case len(w) == 0 || sep && len(w) == 2: // a bare separator carries no program
		case !prog:
			repl = false // the session program was handed something to run
		case bytes.IndexByte(w, '=') >= 0: // an env assignment: the program is still ahead
		default:
			prog = false
			if j := bytes.LastIndexByte(w, '/'); j >= 0 {
				w = w[j+1:] // match on the basename, so /usr/bin/vi counts as vi
			}
			switch bytesconv.B2S(w) {
			case "vi", "vim", "nvim", "emacs", "nano",
				"top", "htop", "tmux", "screen",
				"telnet", "ftp", "sftp":
				return true // drives a terminal whatever its arguments
			case "python", "python3", "node", "irb", "psql", "mysql", "redis-cli", "sqlite3",
				"gdb", "lldb":
				repl = true // a session only when nothing follows it
			}
		}
		if sep {
			if repl {
				return true
			}
			prog, repl = true, false
		}
	}
	return repl
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
