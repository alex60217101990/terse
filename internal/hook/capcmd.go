package hook

import "github.com/alex60217101990/terse/internal/bytesconv"

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
