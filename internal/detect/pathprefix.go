package detect

import "strings"

// Path-dense output (grep/glob summaries, PostCompact file manifests) often
// repeats one long shared directory across every line. FoldPathPrefix folds
// that shared directory into a single declaration line plus a 3-byte §P§
// token per line — lossless and self-describing (no external state or
// expansion step is needed to read it back).
const (
	pathPrefixMinLines = 5  // fewer path-bearing lines: folding isn't worth it
	pathPrefixMinLen   = 16 // shorter shared directory: folding isn't worth it
)

// FoldPathPrefix losslessly folds the longest shared directory prefix across
// the path-bearing lines of content into one "§P=<prefix>§" declaration
// line, substituting the 3-byte token "§P§" for that prefix everywhere it
// recurs. Non-path lines are copied verbatim. It is never-worse: content is
// returned unchanged whenever no shared prefix clears the size bar, or the
// folded form wouldn't come out strictly smaller.
//
// Pass 1 (scanPathPrefix) never allocates: it walks content with
// strings.IndexByte only, tracking the running common prefix and the count
// of path-bearing lines purely as sub-slices of content. Pass 2 (the only
// allocating step) runs only once folding is already known to help.
func FoldPathPrefix(content string) string {
	// Bail out up front if content already contains our own token prefix
	// ("§P§" or the "§P=...§" declaration). Folding on top of an existing
	// occurrence would make original text indistinguishable from our own
	// substitution, corrupting lossless reconstruction. strings.Contains is
	// a single cheap scan (no allocation), so this preserves the zero-alloc
	// non-match invariant and costs the common case exactly one extra pass.
	if strings.Contains(content, "§P") {
		return content
	}
	prefix, count := scanPathPrefix(content)
	if count < pathPrefixMinLines || len(prefix) < pathPrefixMinLen {
		return content
	}
	// Each folded line trades len(prefix) bytes for the short §P§ token
	// (treated as 3 bytes here — an estimate, not exact UTF-8 accounting);
	// the declaration line itself costs len(prefix) plus a small fixed
	// wrapper overhead. This is only a cheap gate to decide whether folding
	// is worth attempting — the final never-worse check below catches any
	// case where the estimate undershoots reality.
	saving := count*(len(prefix)-3) - len(prefix) - 8
	if saving <= 0 {
		return content
	}
	return buildFoldedPathPrefix(content, prefix, saving)
}

// scanPathPrefix finds the longest common directory prefix shared by every
// path-bearing line in content, and how many lines carry a path. It is
// allocation-free: prefix (when non-empty) is always a sub-slice of content.
func scanPathPrefix(content string) (prefix string, count int) {
	start := 0
	for start <= len(content) {
		nl := strings.IndexByte(content[start:], '\n')
		var line string
		if nl < 0 {
			line = content[start:]
		} else {
			line = content[start : start+nl]
		}
		if _, seg, ok := pathSegment(line); ok {
			if count == 0 {
				prefix = seg
			} else {
				prefix = commonPrefix(prefix, seg)
			}
			count++
		}
		if nl < 0 {
			break
		}
		start += nl + 1
	}
	if prefix == "" {
		return "", count
	}
	return dirBoundary(prefix), count
}

// pathSegment finds the path-like token in line: from the first '/' to the
// first of ':', a space, a tab, or end of line. It reports the token's
// starting byte offset within line, the token itself, and whether one was
// found. Zero-alloc: both results are sub-slices of line.
func pathSegment(line string) (start int, seg string, ok bool) {
	i := strings.IndexByte(line, '/')
	if i < 0 {
		return 0, "", false
	}
	j := i
	for j < len(line) && line[j] != ':' && line[j] != ' ' && line[j] != '\t' {
		j++
	}
	return i, line[i:j], true
}

// commonPrefix returns the longest common byte prefix of a and b, as a
// sub-slice of a (zero-alloc).
func commonPrefix(a, b string) string {
	n := min(len(b), len(a))
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return a[:i]
}

// dirBoundary cuts s back to the last '/' it contains, so a raw byte-common
// prefix (which may end mid-filename, e.g. ".../pkg/file") becomes a full
// directory path (".../pkg"). Returns "" if s has no '/' at all.
func dirBoundary(s string) string {
	i := strings.LastIndexByte(s, '/')
	if i < 0 {
		return ""
	}
	return s[:i]
}

// buildFoldedPathPrefix renders the folded form: a "§P=<prefix>§" declaration
// line followed by every line of content, with prefix replaced by "§P§"
// wherever a line's path segment starts with it. estSaving sizes the
// Builder; it's only a hint.
func buildFoldedPathPrefix(content, prefix string, estSaving int) string {
	var b strings.Builder
	if grow := len(content) - estSaving; grow > 0 {
		b.Grow(grow)
	}
	b.WriteString("§P=")
	b.WriteString(prefix)
	b.WriteString("§\n")

	start := 0
	for start <= len(content) {
		nl := strings.IndexByte(content[start:], '\n')
		var line string
		hasNL := nl >= 0
		if hasNL {
			line = content[start : start+nl]
		} else {
			line = content[start:]
		}
		if i, seg, ok := pathSegment(line); ok && strings.HasPrefix(seg, prefix) {
			b.WriteString(line[:i])
			b.WriteString("§P§")
			b.WriteString(seg[len(prefix):])
			b.WriteString(line[i+len(seg):])
		} else {
			b.WriteString(line)
		}
		if !hasNL {
			break
		}
		b.WriteByte('\n')
		start += nl + 1
	}

	out := b.String()
	if len(out) >= len(content) {
		return content
	}
	return out
}
