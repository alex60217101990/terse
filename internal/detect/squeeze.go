package detect

import (
	"regexp"
	"strconv"
	"strings"
)

// ansiRE matches ANSI/VT escape sequences (colors, cursor moves, progress
// redraws) — pure presentation bytes that carry no information for the model.
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

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

	lines := strings.Split(stripped, "\n")
	var b strings.Builder
	b.Grow(len(stripped))

	i := 0
	collapsed := false
	for i < len(lines) {
		j := i + 1
		for j < len(lines) && lines[j] == lines[i] {
			j++
		}
		run := j - i
		b.WriteString(lines[i])
		if run > 1 {
			b.WriteString("  ⨯")
			b.WriteString(strconv.Itoa(run))
			collapsed = true
		}
		if j < len(lines) {
			b.WriteByte('\n')
		}
		i = j
	}

	out := b.String()
	if !collapsed && len(out) == len(content) {
		return content // nothing changed
	}
	return out
}
