package detect

import (
	"regexp"
	"strconv"
	"strings"
)

// ansiRE matches ANSI/VT escape sequences (colors, cursor moves, progress
// redraws) — pure presentation bytes that carry no information for the model.
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;?]*[\x20-\x2f]*[\x40-\x7e]`)

// SqueezeOutput compresses unstructured terminal output losslessly-in-meaning:
// it strips ANSI escape sequences and collapses runs of identical consecutive
// lines into a single "line  ⨯N" marker. It returns the input unchanged when
// neither transform shrinks it, so callers can gate on len().
//
// This is display compression for output no structural detector matched; the
// "⨯N" run-length markers are self-describing, so no expansion step is needed.
func SqueezeOutput(content string) string {
	stripped := content
	if strings.IndexByte(content, 0x1b) >= 0 {
		stripped = ansiRE.ReplaceAllString(content, "")
	}

	var b strings.Builder
	b.Grow(len(stripped))

	// Single forward pass, no []string materialized: collapse runs of identical
	// consecutive lines. Groups are joined by '\n' (a newline before every group
	// but the first), matching the previous Split-based output exactly.
	groupIdx := 0
	writeGroup := func(line string, run int) {
		if groupIdx > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)
		if run > 1 {
			b.WriteString("  ⨯")
			b.WriteString(strconv.Itoa(run))
		}
		groupIdx++
	}
	var prev string
	run := 0
	have := false
	for line := range strings.SplitSeq(stripped, "\n") {
		if have && line == prev {
			run++
			continue
		}
		if have {
			writeGroup(prev, run)
		}
		prev, run, have = line, 1, true
	}
	if have {
		writeGroup(prev, run)
	}

	// Never-worse: return the original unless the result is strictly smaller.
	// (A short repeated line's "  ⨯N" marker can exceed the bytes it saves, so
	// guard on total length, not just "did anything collapse".)
	out := b.String()
	if len(out) >= len(content) {
		return content
	}
	return out
}
