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

// ThinLineNumberRuns is [ThinLineNumbers] for output that is only PARTLY a
// numbered listing: it thins every maximal run of consecutively numbered lines
// and leaves everything else byte-identical. It returns "" when nothing changed.
//
// The shape it is for is a file dumped through Bash rather than Read —
// `cat -n f`, `nl f`, `wc -l f && cat -n f`, or a loop echoing a header before
// each file. The strict whole-payload check rejects all but the barest of those
// (one stray header line is enough), yet a numbered listing costs the same
// tokens whoever printed it: over one project's archive, strict thinning of
// passthrough output is worth 1.35 points of the corpus and this is worth 2.35.
//
// A run must reach gutterKeepEvery lines before it is touched. That is what
// keeps the relaxed rule safe: a TSV or a numbered list would have to hold ten
// perfectly consecutive integers before this could reach it, and even then it
// only thins numbers a reader can re-derive by counting from the anchor above.
func ThinLineNumberRuns(content string) string {
	runs := scanGutterRuns(content)
	if len(runs) == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(len(content))
	run, pos, changed := 0, 0, false
	for rest := content; ; {
		line, tail, more := cutLine(rest)
		num, body, hasGutter := parseGutter(line)
		switch {
		case run == len(runs) || pos < runs[run].start || !hasGutter:
			b.WriteString(line)
		default:
			last := pos == runs[run].end
			if pos == runs[run].start || last || num%gutterKeepEvery == 0 {
				b.WriteString(line)
			} else {
				b.WriteString(body)
				changed = true
			}
			if last {
				run++
			}
		}
		if !more {
			break
		}
		b.WriteByte('\n')
		rest, pos = tail, pos+1
	}
	if !changed {
		return ""
	}
	return b.String()
}

// gutterRun is one maximal stretch of consecutively numbered lines, as line
// indices into the payload.
type gutterRun struct{ start, end int }

// scanGutterRuns reports the runs long enough to be worth thinning. It walks
// the content once and allocates only the (short) run list.
func scanGutterRuns(content string) []gutterRun {
	if content == "" {
		return nil
	}
	var runs []gutterRun
	start, prev, length, pos := 0, 0, 0, 0
	flush := func(end int) {
		if length >= gutterKeepEvery {
			runs = append(runs, gutterRun{start, end})
		}
		length = 0
	}
	for rest := content; ; {
		line, tail, more := cutLine(rest)
		num, _, hasGutter := parseGutter(line)
		switch {
		case !hasGutter:
			flush(pos - 1)
		case length > 0 && num == prev+1:
			length++
			prev = num
		default:
			flush(pos - 1)
			start, prev, length = pos, num, 1
		}
		if !more {
			flush(pos)
			break
		}
		rest, pos = tail, pos+1
	}
	return runs
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
