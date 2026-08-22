package hook

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/buger/jsonparser"

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
// session — an editor, a terminal-driven program, a REPL with nothing to run,
// a watcher — has nothing to capture in the first place; that list is narrow
// because a shell tool never types into a REPL, so "python3 -c ..." is an
// ordinary command and only the argument-less form opens a session.
//
// One forward pass decides both, because both questions are about the same
// bytes: the background check reads them as operators, the session check reads
// the word in each command position. No allocation, no regexp — this runs
// before every Bash call the agent makes.
func cappable(cmd string) bool {
	b := bytesconv.S2B(cmd)
	if len(b) == 0 {
		return false
	}
	var (
		prog  = true  // the next word starts a command
		repl  = false // a session program is open with nothing to run yet
		hard  = false // a program that drives a terminal whatever follows it
		start = -1    // current word start, -1 between words
	)
	// endWord classifies the word that just closed and resets the tracker.
	endWord := func(i int) {
		if start < 0 {
			return
		}
		w := b[start:i]
		start = -1
		switch {
		case !prog:
			// "cargo watch" and friends put the watcher in argument position.
			hard = hard || bytesconv.B2S(w) == "watch"
			repl = false // the session program was handed something to run
		case bytes.IndexByte(w, '=') >= 0: // env assignment: the program is still ahead
		default:
			prog = false
			if j := bytes.LastIndexByte(w, '/'); j >= 0 {
				w = w[j+1:] // match on the basename, so /usr/bin/vi counts as vi
			}
			switch bytesconv.B2S(w) {
			case "vi", "vim", "nvim", "emacs", "nano",
				"top", "htop", "tmux", "screen", "watch",
				"telnet", "ftp", "sftp":
				hard = true // drives a terminal whatever its arguments
			case "python", "python3", "node", "irb", "psql", "mysql", "redis-cli", "sqlite3",
				"gdb", "lldb":
				repl = true // a session only when nothing follows it
			}
		}
	}
	for i := 0; i < len(b); i++ {
		switch c := b[i]; {
		case c == '\\':
			if start < 0 {
				start = i
			}
			i++ // the next byte is escaped: it is data, not shell syntax

		case c == ' ' || c == '\t' || c == '\n':
			endWord(i)

		case c == '|' || c == ';':
			// A separator closes the command: a session program that reached it
			// was invoked bare, and the next word starts a command again.
			endWord(i)
			if repl {
				return false
			}
			prog, repl = true, false
			if c == '|' && i+1 < len(b) && (b[i+1] == '|' || b[i+1] == '&') {
				i++ // "||" is control flow, "|&" pipes stderr too
			}

		case c == '&':
			switch {
			case i+1 < len(b) && b[i+1] == '&':
				endWord(i)
				if repl {
					return false
				}
				prog, repl, i = true, false, i+1
			case i+1 < len(b) && b[i+1] == '>':
				if start < 0 { // "&>" redirects both streams
					start = i
				}
			case i > 0 && (b[i-1] == '>' || b[i-1] == '<'):
				// ">&"/"<&" duplicate a descriptor, they do not background.
			default:
				return false // a bare "&" backgrounds: the capture would be raced
			}

		default:
			if start < 0 {
				start = i
			}
		}
	}
	endWord(len(b))
	// "--watch" sits in argument position, where the command-position scan does
	// not look, and a watcher's output would sit in the capture file until a
	// timeout the agent never sees past. Rare enough to cost one scan.
	return !hard && !repl && !bytes.Contains(b, []byte("--watch"))
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

// The wrapper, literal by literal, with the variable slots between them. One
// table, two emitters: wrapCommand writes shell text, appendWrappedJSON writes
// the same text already JSON-escaped straight into the response. Keeping a
// single source of truth is what makes the second emitter safe — the invariant
// test decodes its output and compares it with the first.
type wrapPart uint8

const (
	partPath wrapPart = iota
	partCmd
	partID
	partCap
	partHalf
)

var wrapLits = [...]string{
	"if : 2>/dev/null > '",
	"'; then { ",
	"\n} > '",
	"' 2>&1\n(__qrc=$?; __qn=$(wc -c < '",
	"'); if [ \"$__qn\" -le ",
	" ]; then cat '",
	"'; else head -c ",
	" '",
	"'; printf '\\n... %s bytes elided, full output: qdf-hook expand ",
	"\\n' \"$__qn\"; tail -c ",
	" '",
	"'; fi; exit \"$__qrc\")\nelse { ",
	"\n}\nfi",
}

var wrapSlots = [...]wrapPart{
	partPath, partCmd, partPath, partPath, partCap,
	partPath, partHalf, partPath, partID, partHalf,
	partPath, partCmd,
}

// wrapLitsJSON is wrapLits with the JSON escaping applied once, at startup, so
// the response encoder never rescans the 400-odd bytes of scaffolding it wrote
// itself. Measured, that scan was two thirds of the hook's CPU.
var wrapLitsJSON [len(wrapLits)]string

func init() {
	for i, lit := range wrapLits {
		wrapLitsJSON[i] = string(protocol.AppendJSONString(nil, lit))
	}
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
	for i, lit := range wrapLits {
		dst = append(dst, lit...)
		if i >= len(wrapSlots) {
			break
		}
		switch wrapSlots[i] {
		case partPath:
			dst = append(dst, capturePath...)
		case partCmd:
			dst = append(dst, cmd...)
		case partID:
			dst = append(dst, id...)
		case partCap:
			dst = strconv.AppendInt(dst, int64(capBytes), 10)
		case partHalf:
			dst = strconv.AppendInt(dst, half, 10)
		}
	}
	return dst
}

// appendWrappedJSON appends the whole PreToolUse response — envelope, wrapper
// and all — to dst, JSON-escaped. It is wrapCommand plus the encoder, fused:
// the scaffolding is escaped once at startup, the path once per HOME, and only
// the model's command is escaped per call.
//
// escapedPath must be the JSON-escaped form of capturePath. The id is hex and
// the two numbers are decimal, so neither can contain a byte JSON escapes.
//
// Unlike wrapCommand this does not re-validate the path's shell safety: the
// only untrusted part of it is $HOME, which captureDirReady checks once per
// process. Callers that build a path some other way must call shellSafeArg
// themselves.
func appendWrappedJSON(dst []byte, cmd, escapedPath, id string, capBytes int) []byte {
	half := int64(capBytes / 2)
	dst = append(dst, protocol.PreInputHead...)
	for i, lit := range wrapLitsJSON {
		dst = append(dst, lit...)
		if i >= len(wrapSlots) {
			break
		}
		switch wrapSlots[i] {
		case partPath:
			dst = append(dst, escapedPath...)
		case partCmd:
			dst = protocol.AppendJSONString(dst, cmd)
		case partID:
			dst = append(dst, id...)
		case partCap:
			dst = strconv.AppendInt(dst, int64(capBytes), 10)
		case partHalf:
			dst = strconv.AppendInt(dst, half, 10)
		}
	}
	return append(dst, protocol.PreInputTail...)
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
//
// The common command carries no JSON escape, and then the returned string
// aliases raw instead of copying it: the caller only appends it into a buffer
// before returning, so nothing outlives the request. An escaped command is
// unescaped into scratch, which the caller owns for the same window.
func decodeCappableCommand(raw json.RawMessage, scratch []byte) (string, []byte, bool) {
	v, typ, _, err := jsonparser.Get(raw, "command")
	if err != nil || typ != jsonparser.String {
		return "", scratch, false
	}
	if bytes.IndexByte(v, '\\') >= 0 {
		u, uerr := jsonparser.Unescape(v, scratch[:0])
		if uerr != nil {
			return "", scratch, false
		}
		if cap(u) > cap(scratch) {
			scratch = u // Unescape grew the buffer: keep the larger one pooled
		}
		cmd := bytesconv.B2S(u)
		return cmd, scratch, cappable(cmd)
	}
	cmd := bytesconv.B2S(v)
	return cmd, scratch, cappable(cmd)
}

// captureDir caches the capture directory per HOME. os.MkdirAll is two
// syscalls and, measured, was 92% of this hook's CPU when it ran per call;
// the daemon handles thousands of calls under one HOME, so it runs once. HOME
// is re-read each time — it is a map lookup, not a syscall — because the test
// suite moves it between cases and a frozen directory would silently write a
// later case's captures into an earlier case's temp dir.
type captureDirState struct {
	home       string
	dir        string
	escapedDir string // dir with JSON escaping already applied
	ok         bool
}

var captureDirCache atomic.Pointer[captureDirState]

// captureDirReady reports whether the capture directory exists and is
// writable, creating it if needed, and returns its path. A failure here just
// means no capture is available for this command, not an error worth
// propagating, so this reports a bool rather than the underlying error.
func captureDirReady() *captureDirState {
	home := os.Getenv("HOME")
	if st := captureDirCache.Load(); st != nil && st.home == home {
		return st
	}
	dir := cache.CaptureDir()
	st := &captureDirState{
		home:       home,
		dir:        dir,
		escapedDir: string(protocol.AppendJSONString(nil, dir)),
		// $HOME is not ours: a quote or a control byte in it would break out of
		// the wrapper's shell quoting. Checked here, once, so the hot path does
		// not rescan a constant — the rest of the capture path is hex plus a
		// fixed suffix and cannot contain either.
		ok: shellSafeArg(dir) && os.MkdirAll(dir, 0o700) == nil,
	}
	captureDirCache.Store(st)
	return st
}

// capScratch is one call's working memory: the bytes hashed into the capture
// id, the capture path built from that id, and the rewritten command. Pooling
// the three together keeps the whole rewrite allocation-free after warmup.
type capScratch struct {
	hash  []byte
	path  []byte
	epath []byte // the same path, JSON-escaped
	out   []byte
	cmd   []byte
}

var capPool = sync.Pool{New: func() any {
	return &capScratch{
		hash:  make([]byte, 0, 256),
		path:  make([]byte, 0, 128),
		epath: make([]byte, 0, 128),
		out:   make([]byte, 0, 1024),
		cmd:   make([]byte, 0, 256),
	}
}}

// captureID is sha256(session+cmd+seq)[:16] in hex — the same content address
// cache.RefHashOf computes, written straight into a caller-owned array so the
// id costs no allocation of its own.
func captureID(dst *[32]byte, scratch []byte, session, cmd string, seq uint64) []byte {
	scratch = append(scratch[:0], session...)
	scratch = append(scratch, cmd...)
	scratch = strconv.AppendUint(scratch, seq, 10)
	sum := sha256.Sum256(scratch)
	hex.Encode(dst[:], sum[:16])
	return scratch
}

// handleBashPreToolUse rewrites a Bash command so the shell bounds its own
// output. Commands the scanner rejects are passed through in silence: writing
// anything at all — "{}" included — would make Claude Code record a
// hook_success attachment that then rides the prefix for the rest of the session.
func handleBashPreToolUse(inp *protocol.HookInput, w io.Writer) error {
	sc, _ := capPool.Get().(*capScratch)
	if sc == nil {
		sc = &capScratch{}
	}
	// No defer: this is the pre-execution hot path for every Bash call.
	cmd, cmdBuf, ok := decodeCappableCommand(inp.ToolInput, sc.cmd)
	sc.cmd = cmdBuf
	if !ok {
		capPool.Put(sc)
		return nil
	}
	cd := captureDirReady()
	if !cd.ok {
		capPool.Put(sc) // capture unavailable: run the command untouched
		return nil
	}

	var idBuf [32]byte
	sc.hash = captureID(&idBuf, sc.hash, inp.SessionID, cmd, captureSeq.Add(1))
	id := bytesconv.B2S(idBuf[:])

	sc.path = append(sc.path[:0], cd.dir...)
	sc.path = append(sc.path, '/')
	sc.path = append(sc.path, idBuf[:]...)
	sc.path = append(sc.path, ".out"...)
	path := bytesconv.B2S(sc.path)

	// The hex id and ".out" need no escaping, so the escaped path is the
	// escaped directory with the same tail — no rescan of $HOME per call.
	sc.epath = append(sc.epath[:0], cd.escapedDir...)
	sc.epath = append(sc.epath, '/')
	sc.epath = append(sc.epath, idBuf[:]...)
	sc.epath = append(sc.epath, ".out"...)

	need := 2*len(cmd) + 8*len(path) + wrapOverhead + len(protocol.PreInputHead)
	if cap(sc.out) < need {
		sc.out = make([]byte, 0, need)
	}
	// The path is single-quoted into a shell command. The id is hex and the
	// directory is fixed, but $HOME is not ours — a quote in it would break out
	// of the quoting. wrapCommand refuses such a path and returns nil, and then
	// the command runs untouched.
	sc.out = appendWrappedJSON(sc.out[:0], cmd, bytesconv.B2S(sc.epath), id, capBytes)
	b := sc.out
	_, err := w.Write(b)
	capPool.Put(sc)
	return err
}
