package detect

import (
	"strings"
)

// isASCIISpace reports whether c is an ASCII whitespace byte — the ASCII subset
// of unicode.IsSpace, which is all that separates fields in git-log output.
func isASCIISpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\f' || c == '\r'
}

// firstFieldEnd returns the byte index of the first ASCII-whitespace byte in s
// (the end of the first whitespace-delimited field), or -1 if s has none.
func firstFieldEnd(s string) int {
	for i := range len(s) {
		if isASCIISpace(s[i]) {
			return i
		}
	}
	return -1
}

// isHexASCII reports whether every byte of s is a hex digit. s is expected to
// be ASCII; a multi-byte rune's bytes are all >= 0x80 and fail the range test,
// matching the previous rune-iterating check.
func isHexASCII(s string) bool {
	for i := range len(s) {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

// runeTrunc reports whether s exceeds limit runes and, if so, the byte offset
// after its first keep runes — so a caller emits s[:off] + "...". It replaces a
// []rune conversion (which allocated a rune slice the length of the string) with
// a single UTF-8 scan and no allocation. keep must be <= limit.
func runeTrunc(s string, limit, keep int) (off int, trunc bool) {
	runes := 0
	keepOff := -1
	for i := range s { // ranging a string yields the byte index of each rune start
		if runes == keep {
			keepOff = i
		}
		runes++
	}
	if runes <= limit {
		return 0, false
	}
	return keepOff, true
}

// IsGitLogOutput returns true if s looks like `git log --oneline` output.
// Heuristic: at least 2 of the first ~6 non-empty lines start with a 7–40 char
// hex hash followed by a second field. Scanned line-by-line and field-by-field
// via IndexByte — no SplitN slice, no per-line Fields slice.
func IsGitLogOutput(s string) bool {
	s = strings.TrimSpace(s)
	matched, scanned := 0, 0
	for len(s) > 0 && scanned < 6 {
		line := s
		if nl := strings.IndexByte(s, '\n'); nl >= 0 {
			line, s = s[:nl], s[nl+1:]
		} else {
			s = ""
		}
		scanned++
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		he := firstFieldEnd(line)
		if he < 0 {
			continue // single field: the old code's len(fields) < 2
		}
		if strings.TrimLeft(line[he:], " \t\v\f\r") == "" {
			continue // hash only, no second field
		}
		hash := line[:he]
		if len(hash) < 7 || len(hash) > 40 {
			return false
		}
		if isHexASCII(hash) {
			matched++
		}
	}
	return matched >= 2
}

// SummarizeGitLog returns a compact table of commits.
// Keeps hash (short), and truncates long commit messages to 72 characters.
func SummarizeGitLog(s string) string {
	var body strings.Builder
	commits := 0
	rest := strings.TrimSpace(s)
	for len(rest) > 0 {
		line := rest
		if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
			line, rest = rest[:nl], rest[nl+1:]
		} else {
			rest = ""
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		he := firstFieldEnd(line)
		if he < 0 {
			continue // single field: the old code's len(fields) < 2
		}
		msg := strings.TrimLeft(line[he:], " \t\v\f\r")
		if msg == "" {
			continue // hash only, no message field
		}
		hash := line[:he]
		if len(hash) > 8 {
			hash = hash[:8]
		}
		body.WriteString(hash)
		body.WriteString(" ")
		// Truncate long messages to keep output compact. Cut on a rune
		// boundary so multi-byte characters are never split.
		if off, trunc := runeTrunc(msg, 72, 69); trunc {
			body.WriteString(msg[:off])
			body.WriteString("...")
		} else {
			body.WriteString(msg)
		}
		body.WriteString("\n")
		commits++
	}

	var sb strings.Builder
	sb.WriteString("[git log — ")
	sb.WriteString(itoa(commits))
	sb.WriteString(" commits]\n")
	sb.WriteString(body.String())
	return sb.String()
}
