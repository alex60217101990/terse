package hook

import (
	"strconv"

	"github.com/alex60217101990/terse/internal/bytesconv"
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

// wrapCommand appends to dst a rewrite of cmd that captures the command's full
// output and prints a bounded view of it.
//
// Under capBytes the capture is printed verbatim, so the transcript is
// byte-identical to an unwrapped run. Over it, both ends are kept — a file's
// header and a build's first error live at the top, summaries and failure counts
// at the bottom — with one elision line between them carrying the recovery
// handle.
//
// dst is appended to, never reallocated by the caller's convention: pass a
// pooled buffer with enough capacity and this does no allocation at all.
func wrapCommand(dst []byte, cmd, capturePath, id string, capBytes int) []byte {
	half := strconv.Itoa(capBytes / 2)
	cb := strconv.Itoa(capBytes)

	dst = append(dst, "{ "...)
	dst = append(dst, cmd...)
	dst = append(dst, " ; } > '"...)
	dst = append(dst, capturePath...)
	dst = append(dst, "' 2>&1; __qrc=$?; __qn=$(wc -c < '"...)
	dst = append(dst, capturePath...)
	dst = append(dst, "'); if [ \"$__qn\" -le "...)
	dst = append(dst, cb...)
	dst = append(dst, " ]; then cat '"...)
	dst = append(dst, capturePath...)
	dst = append(dst, "'; else head -c "...)
	dst = append(dst, half...)
	dst = append(dst, " '"...)
	dst = append(dst, capturePath...)
	dst = append(dst, "'; printf '\\n... %s bytes elided, full output: qdf-hook expand "...)
	dst = append(dst, id...)
	dst = append(dst, "\\n' \"$__qn\"; tail -c "...)
	dst = append(dst, half...)
	dst = append(dst, " '"...)
	dst = append(dst, capturePath...)
	// The exit status must become the original command's, not cat/head/tail's.
	// Running it inside a subshell sets $? without an unqualified "exit", which
	// would terminate the caller's shell process before anything after this
	// wrapper on the same line gets to run.
	dst = append(dst, "'; fi; (exit \"$__qrc\")"...)
	return dst
}

// wrapOverhead is the fixed byte cost of wrapCommand's scaffolding, used to size
// a buffer exactly once. Deliberately generous: a few spare bytes cost nothing,
// a reallocation costs an allocation on every Bash call.
const wrapOverhead = 320
