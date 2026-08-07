package detect

import (
	"strconv"
	"strings"
)

// Claude Code renders a Read as cat -n output: every line carries a
// "<number><tab>" gutter. For a contiguous range those numbers are perfectly
// redundant — line k is the start line plus k — but they are not free: measured
// with o200k over 1249 real Read payloads, the gutter is 15.9% of all Read
// tokens (263k tokens in one project's archive).
//
// It is not redundant to the READER, though. The model cites file:line, and a
// model that has to count from the top of a 400-line file will miscount. So the
// gutter is thinned rather than dropped: every gutterKeepEvery-th line keeps its
// number as an anchor, bounding any counting error to half that interval, and
// the rest lose it. Anchors cost 1.7 points of the 15.9 — dropping the gutter
// entirely saves 15.9%, keeping every 10th saves 14.2%.
const gutterKeepEvery = 10

// ThinLineNumbers removes the per-line number gutter from cat -n style content,
// keeping it on the first line, the last line, and every gutterKeepEvery-th
// line number. It returns "" when content is not gutter-bearing, so callers can
// skip it without a second scan.
//
// The rule for recognising a gutter is deliberately strict: EVERY non-empty
// line must carry one, and the numbers must run consecutively. Anything looser
// risks mangling a file whose own content happens to start with digits and a
// tab (a TSV, a numbered list), and content is what this tool must never
// damage. A payload that fails the check is left alone.
func ThinLineNumbers(content string) string {
	first, last, ok := scanGutter(content)
	if !ok {
		return ""
	}

	var b strings.Builder
	b.Grow(len(content))
	n := first
	for rest := content; ; n++ {
		line, tail, more := cutLine(rest)
		body, hasGutter := splitGutter(line)
		switch {
		case !hasGutter || n == first || n == last || n%gutterKeepEvery == 0:
			b.WriteString(line)
		default:
			// The tab goes with the number. Keeping it as an empty gutter field
			// would cost a token per line for nothing, and dropping it puts the
			// file's own indentation flush against the left margin, which is
			// what the content actually looks like.
			b.WriteString(body)
		}
		if !more {
			break
		}
		b.WriteByte('\n')
		rest = tail
	}
	return b.String()
}

// scanGutter validates that content is consecutive cat -n output and reports
// the first and last line numbers. It allocates nothing.
func scanGutter(content string) (first, last int, ok bool) {
	if content == "" {
		return 0, 0, false
	}
	n, lines := 0, 0
	for rest := content; ; {
		line, tail, more := cutLine(rest)
		// A trailing empty line is the newline that ends the last real line, not
		// a line of its own.
		if line == "" && !more {
			break
		}
		num, _, hasGutter := parseGutter(line)
		if !hasGutter {
			return 0, 0, false
		}
		switch {
		case lines == 0:
			first, n = num, num
		case num != n+1:
			return 0, 0, false
		default:
			n = num
		}
		lines++
		if !more {
			break
		}
		rest = tail
	}
	// Two lines cannot pay for the header a caller would add, and a single line
	// has nothing to thin.
	if lines < gutterKeepEvery {
		return 0, 0, false
	}
	return first, n, true
}

// cutLine splits off the first line of s, reporting whether more followed.
func cutLine(s string) (line, rest string, more bool) {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i], s[i+1:], true
	}
	return s, "", false
}

// parseGutter splits a "<number><tab>" prefix off line. Leading spaces are
// tolerated: some renderers right-align the gutter.
func parseGutter(line string) (num int, body string, ok bool) {
	i := 0
	for i < len(line) && line[i] == ' ' {
		i++
	}
	d := i
	for d < len(line) && line[d] >= '0' && line[d] <= '9' {
		d++
	}
	if d == i || d >= len(line) || line[d] != '\t' {
		return 0, line, false
	}
	num, err := strconv.Atoi(line[i:d])
	if err != nil {
		return 0, line, false
	}
	return num, line[d+1:], true
}

// splitGutter is parseGutter without the number.
func splitGutter(line string) (body string, ok bool) {
	_, body, ok = parseGutter(line)
	return body, ok
}
