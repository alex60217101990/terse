package detect

import "strings"

// minFoldBlock is the smallest block (in bytes, trailing newlines excluded)
// worth folding. Below this a back-reference marker would cost more than the
// duplicate it removes.
const minFoldBlock = 64

// maxMarkerFirstLine caps how much of a block's first line the marker echoes,
// in runes, so a very long header line can't bloat the marker.
const maxMarkerFirstLine = 56

// Marker pieces. Byte lengths are compile-time constants, so a marker's size is
// computed without building the string — the fold decision needs the length
// before committing, and on a fold the pieces are written straight into the
// output builder (no intermediate marker allocation).
const (
	markerPre = "⟦↑ repeat: \""
	markerSuf = "\"⟧"
	markerEll = "…"
)

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
// is replaced by a self-describing one-line back-reference naming it by its
// first line. No expansion step is needed — the referent is above in the same
// text, exactly like the "⨯N" run-length markers.
//
// It returns the input unchanged when nothing folds, so callers can gate on
// len(). The output builder is allocated lazily — only once a fold actually
// happens — so the common no-duplicate payload pays for nothing but the dedup
// map and a single scan.
func FoldRepeatedBlocks(content string) string {
	// A non-adjacent duplicate needs at least two blocks, i.e. a blank-line
	// boundary. No "\n\n" (or too small) ⇒ nothing to do, allocate nothing.
	if len(content) < minFoldBlock*2 || !strings.Contains(content, "\n\n") {
		return content
	}

	var b strings.Builder
	active := false // builder in use (set on the first fold)
	written := 0    // bytes of content already committed to b
	seen := make(map[string]struct{})

	// fold checks one block [bs,be); on a duplicate that shrinks, it lazily
	// starts the builder, flushes the verbatim gap since the last write, and
	// emits the marker plus the block's preserved trailing newlines.
	fold := func(bs, be int) {
		block := content[bs:be]
		key := strings.TrimRight(block, "\n")
		if len(key) < minFoldBlock {
			return
		}
		if _, dup := seen[key]; !dup {
			seen[key] = struct{}{}
			return
		}
		first, _, _ := strings.Cut(key, "\n")
		first = strings.TrimSpace(first)
		trunc := false
		if r := []rune(first); len(r) > maxMarkerFirstLine {
			first = string(r[:maxMarkerFirstLine])
			trunc = true
		}
		mlen := len(markerPre) + len(first) + len(markerSuf)
		if trunc {
			mlen += len(markerEll)
		}
		if mlen >= len(key) { // marker wouldn't be smaller — keep verbatim
			return
		}
		if !active {
			b.Grow(len(content))
			active = true
		}
		b.WriteString(content[written:bs]) // verbatim gap (kept blocks + separators)
		b.WriteString(markerPre)
		b.WriteString(first)
		if trunc {
			b.WriteString(markerEll)
		}
		b.WriteString(markerSuf)
		b.WriteString(block[len(key):]) // preserved trailing newlines
		written = be
	}

	n := len(content)
	segStart := 0
	i := 0
	for i < n {
		if content[i] == '\n' {
			j := i + 1
			for j < n && content[j] == '\n' {
				j++
			}
			if j-i >= 2 { // blank line -> block boundary; separator [i,j) kept verbatim
				fold(segStart, i)
				segStart = j
				i = j
				continue
			}
		}
		i++
	}
	fold(segStart, n) // trailing block

	if !active {
		return content
	}
	b.WriteString(content[written:]) // verbatim tail after the last fold
	return b.String()
}
