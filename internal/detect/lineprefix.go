package detect

import "strings"

// Command output is line-oriented and its lines are usually near-neighbors of
// each other: paths under one directory, stack frames in one package, log lines
// with one timestamp shape, indented code at one depth. That shared head is
// paid for once per line.
//
// FoldLinePrefixes declares it once per RUN instead. Consecutive lines that
// share a boundary-aligned prefix get it hoisted into a "[^=<prefix>]" line and
// keep only their suffix, marked with "^".
//
// Measured with o200k over 1853 real Bash payloads of 256 tokens or more —
// everything the pipeline currently hands back untouched — this is worth 11.5%.
// The alternatives measured on the same corpus: the existing whole-payload
// squeeze 0.5%, block folding 0.0%, single-global-prefix path folding 2.0%,
// common-indent stripping 2.4%, and deduplicating identical lines 2.7%. Command
// output is not repetitive; it is incrementally similar, which is a different
// shape and needs a different transform.
//
// Classic front-coding — "<n>^suffix", where n is the count of bytes shared
// with the PREVIOUS line — measures better still, 16.2%. It is not built,
// deliberately: reading it back requires counting characters into the line
// above, and that is exactly the operation a language model is worst at. Those
// 4.7 points buy a format whose reader can silently reconstruct the wrong path.
// Every line here carries its own suffix verbatim under a prefix named in full.
const (
	// linePrefixMinRun is how many consecutive lines must share a prefix. Two
	// already pays: the declaration costs one line, each member saves one
	// prefix.
	linePrefixMinRun = 2
	// linePrefixMinLen is the shortest prefix worth hoisting, in bytes. The
	// sweep from 6 to 12 spans 11.3% to 11.5%, so this is a plateau rather than
	// a tuned edge; the high end is taken because longer prefixes mean fewer and
	// more meaningful groups.
	linePrefixMinLen = 12
	// linePrefixSeps are the characters a hoisted prefix may end on, so it never
	// cuts mid-token — a prefix ending inside a filename reads as a typo.
	linePrefixSeps = "/ :\t.-_"
)

// FoldLinePrefixes folds shared prefixes of consecutive lines. It returns ""
// when content is not foldable or the folded form would not be smaller, so
// callers can skip it without a second scan.
//
// Callers must still apply their own never-worse check in tokens. Bytes are
// only a proxy here, and the cheap one: this function refuses obvious losses
// without tokenising, and the caller pays for the tokenizer once.
func FoldLinePrefixes(content string) string {
	// Guard, exactly as FoldPathPrefix does: if the payload already contains
	// something that reads as our own output, folding on top of it would make
	// original text indistinguishable from a substitution, and reconstruction
	// would corrupt a line nobody touched.
	if strings.Contains(content, linePrefixDecl) || strings.HasPrefix(content, "^") ||
		strings.Contains(content, "\n^") {
		return ""
	}

	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) < linePrefixMinRun+1 {
		return ""
	}
	trailingNL := strings.HasSuffix(content, "\n")

	var b strings.Builder
	b.Grow(len(content))
	folded := false
	for i := 0; i < len(lines); {
		pre, j := runPrefix(lines, i)
		if j-i < linePrefixMinRun {
			b.WriteString(lines[i])
			b.WriteByte('\n')
			i++
			continue
		}
		folded = true
		b.WriteString(linePrefixDecl)
		b.WriteString(pre)
		b.WriteString("]\n")
		for _, l := range lines[i:j] {
			b.WriteByte('^')
			b.WriteString(l[len(pre):])
			b.WriteByte('\n')
		}
		i = j
	}
	if !folded {
		return ""
	}
	out := b.String()
	if !trailingNL {
		out = strings.TrimSuffix(out, "\n")
	}
	if len(out) >= len(content) {
		return ""
	}
	return out
}

const linePrefixDecl = "[^="

// runPrefix extends a run from lines[i] for as long as every member shares a
// boundary-aligned prefix of at least linePrefixMinLen bytes, and reports that
// prefix with the index one past the run.
func runPrefix(lines []string, i int) (prefix string, end int) {
	prefix, end = lines[i], i+1
	for end < len(lines) {
		next := cutAtBoundary(commonPrefix(prefix, lines[end]))
		if len(next) < linePrefixMinLen {
			break
		}
		prefix, end = next, end+1
	}
	// A run of one never folds, and its "prefix" is the whole line — which would
	// hoist a line into a declaration and leave an empty suffix.
	if end-i < linePrefixMinRun {
		return "", i + 1
	}
	return prefix, end
}

// cutAtBoundary trims s back to the last separator it contains, so a hoisted
// prefix always ends on a token boundary. Returns "" when there is none.
func cutAtBoundary(s string) string {
	if i := strings.LastIndexAny(s, linePrefixSeps); i >= 0 {
		return s[:i+1]
	}
	return ""
}
