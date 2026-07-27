package detect

import "strings"

// minFoldBlock is the smallest block (in bytes) worth folding. Below this a
// back-reference marker would cost more than the duplicate it removes.
const minFoldBlock = 64

// maxMarkerFirstLine caps how much of a block's first line the marker echoes,
// in runes, so a very long header line can't bloat the marker.
const maxMarkerFirstLine = 56

// FoldRepeatedBlocks collapses NON-adjacent duplicate blocks within a single
// payload — the case [SqueezeOutput]'s consecutive run-length pass cannot see.
// Tool output that repeats whole sections (e.g. an MCP batch that re-dumps the
// same grep result under several query headers) shrinks a lot here.
//
// A "block" is a maximal run of lines containing no blank line; blocks are
// separated by runs of two-or-more newlines. Separators and first occurrences
// are copied byte-for-byte, so non-duplicated content is never altered. The
// first occurrence of a block is kept verbatim; each later identical occurrence
// (at least minFoldBlock bytes, and only when the marker is strictly shorter)
// is replaced by a self-describing one-line back-reference naming the block by
// its first line. No expansion step is needed — the referent is above in the
// same text, exactly like the "⨯N" run-length markers.
//
// It returns the input unchanged when nothing folds, so callers can gate on
// len().
func FoldRepeatedBlocks(content string) string {
	if len(content) < minFoldBlock*2 {
		return content
	}

	var b strings.Builder
	b.Grow(len(content))
	seen := make(map[string]struct{})
	folded := false

	n := len(content)
	segStart := 0
	i := 0
	emit := func(block string) {
		// Key on the block without trailing newlines so the final block of the
		// payload (which carries the file's terminating '\n') still matches an
		// identical mid-payload occurrence that a separator trimmed. The tail is
		// preserved verbatim on both the kept and the folded paths.
		key := strings.TrimRight(block, "\n")
		tail := block[len(key):]
		if len(key) >= minFoldBlock {
			if _, dup := seen[key]; dup {
				if m := blockMarker(key); len(m) < len(key) {
					b.WriteString(m)
					b.WriteString(tail)
					folded = true
					return
				}
			} else {
				seen[key] = struct{}{}
			}
		}
		b.WriteString(block)
	}

	for i < n {
		if content[i] == '\n' {
			j := i + 1
			for j < n && content[j] == '\n' {
				j++
			}
			if j-i >= 2 { // blank line -> block boundary
				emit(content[segStart:i])
				b.WriteString(content[i:j]) // separator, verbatim
				segStart = j
				i = j
				continue
			}
		}
		i++
	}
	emit(content[segStart:n]) // trailing block

	if !folded {
		return content
	}
	out := b.String()
	if len(out) >= len(content) {
		return content
	}
	return out
}

// blockMarker builds a single-line back-reference that names a folded block by
// its (truncated) first line, so the model can tell which earlier block it is.
func blockMarker(block string) string {
	first, _, _ := strings.Cut(block, "\n")
	first = strings.TrimSpace(first)
	trunc := false
	if r := []rune(first); len(r) > maxMarkerFirstLine {
		first = string(r[:maxMarkerFirstLine])
		trunc = true
	}
	var m strings.Builder
	m.WriteString("⟦↑ repeat: \"")
	m.WriteString(first)
	if trunc {
		m.WriteString("…")
	}
	m.WriteString("\"⟧")
	return m.String()
}
