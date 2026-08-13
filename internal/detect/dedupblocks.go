package detect

import (
	"strings"
	"sync"
)

// seenPool / seenNormPool recycle the two dedup maps across FoldRepeatedBlocks
// calls (which may run concurrently in the daemon — sync.Pool is safe). Each
// map is clear()ed before it goes back so no entry (whose keys alias the prior
// call's content) survives into the next call.
var (
	seenPool     = sync.Pool{New: func() any { return make(map[string]struct{}) }}
	seenNormPool = sync.Pool{New: func() any { return make(map[string]string) }}
)

// maxPooledDedupMapLen caps how large a map FoldRepeatedBlocks will return to
// seenPool/seenNormPool. clear() empties a map's entries but Go's runtime
// never shrinks its bucket array back down, so a map that grew huge for one
// exceptional (e.g. deliberately adversarial or pathologically repetitive)
// payload would otherwise pin that bucket memory in the pool for every
// smaller call that follows. Past this many entries the map is dropped
// instead of pooled, and the pool's New allocates a fresh small one on the
// next Get.
const maxPooledDedupMapLen = 4096

// shouldPoolDedupMap reports whether a dedup map with n entries (measured
// BEFORE clear()) is small enough to return to its sync.Pool. The boundary is
// the whole of maxPooledDedupMapLen's rationale, isolated into a pure
// function so it can be tested without relying on sync.Pool's unspecified
// same-P Get/Put timing.
func shouldPoolDedupMap(n int) bool { return n <= maxPooledDedupMapLen }

// putSeen returns seen to seenPool unless it grew past maxPooledDedupMapLen
// entries, in which case it is dropped instead (see that constant's doc
// comment). n must be the map's length captured BEFORE this call clears it —
// clear() empties the map in place, so measuring len(seen) after clearing
// would always read 0 and defeat the cap check.
func putSeen(seen map[string]struct{}, n int) {
	clear(seen)
	if shouldPoolDedupMap(n) {
		seenPool.Put(seen)
	}
}

// putSeenNorm is putSeen's counterpart for seenNormPool.
func putSeenNorm(seenNorm map[string]string, n int) {
	clear(seenNorm)
	if shouldPoolDedupMap(n) {
		seenNormPool.Put(seenNorm)
	}
}

// minFoldBlock is the smallest block (in bytes, trailing newlines excluded)
// worth folding. Below this a back-reference marker would cost more than the
// duplicate it removes.
const minFoldBlock = 64

// maxMarkerFirstLine caps how much of a block's first line the marker echoes,
// in runes, so a very long header line can't bloat the marker.
const maxMarkerFirstLine = 56

// maxFuzzyDiffs bounds how many volatile tokens may differ between two
// otherwise-identical blocks before a near-duplicate fold is refused. Every
// differing token is spelled out in the marker, so this also caps marker bloat.
const maxFuzzyDiffs = 4

// Marker pieces. Byte lengths are compile-time constants, so a marker's size is
// computed without building the string — the fold decision needs the length
// before committing, and on a fold the pieces are written straight into the
// output builder (no intermediate marker allocation).
//
// ASCII brackets rather than the ⟦ ⟧ this used to emit. Measured with o200k,
// ⟦ and ⟧ cost three tokens EACH — they are outside the vocabulary and fall
// back to per-byte pieces — so the pair alone was six tokens of pure framing
// on every folded block. "[" and "]" cost one each, and the marker prefix goes
// from seven tokens to four. Nothing parses these markers back, so the change
// is one-way and safe; only IsHookOutput in cmd/qdf-replay has to keep
// recognising the old form, because an archive spans both.
const (
	markerPre = "[repeat: \""
	markerSuf = "\"]"
	markerEll = "…"
)

// Fuzzy (near-duplicate) marker pieces. A fuzzy marker names the repeated block
// by its first line, exactly like the exact marker, then lists EVERY volatile
// token that differs as "old→new" pairs — so no differing byte is ever hidden.
const (
	fuzzyPre    = "[repeat of \""
	fuzzyExcept = "\" except "
	fuzzyArrow  = "->"
	fuzzySep    = ", "
	fuzzySuf    = "]"
)

// blockClass classifies a byte for FoldRepeatedBlocks' single content scan:
// bit blockNewline marks '\n' (a block boundary candidate) and bit blockDigit
// marks an ASCII digit (a near-duplicate signal). Folding both into one table
// load keeps the hot loop at one load + one zero-test per byte — the same shape
// as a bare '\n' compare — so digit detection costs the no-duplicate path
// essentially nothing.
const (
	blockNewline = 1 << 0
	blockDigit   = 1 << 1
)

var blockClass = func() (t [256]byte) {
	t['\n'] = blockNewline
	for c := byte('0'); c <= '9'; c++ {
		t[c] = blockDigit
	}
	return t
}()

// tokenPair is one differing volatile token between a base block (old) and a
// later near-duplicate (new). Both fields come verbatim from the source, so the
// marker reproduces them byte-for-byte.
type tokenPair struct {
	old string
	new string
}

// truncFirstLine returns the first line of s (its content up to the first
// '\n', with surrounding space trimmed), capped at maxMarkerFirstLine runes.
// trunc reports whether the cap clipped the line. It slices s in place via a
// forward rune walk — no []rune materialization and no new backing array — so
// a marker's first-line echo allocates nothing. Both the exact and fuzzy
// marker paths share this single truncation source of truth.
func truncFirstLine(s string) (first string, trunc bool) {
	first, _, _ = strings.Cut(s, "\n")
	first = strings.TrimSpace(first)
	count := 0
	for idx := range first { // ranges rune-by-rune; idx is each rune's start byte
		if count == maxMarkerFirstLine {
			return first[:idx], true
		}
		count++
	}
	return first, false
}

// isDigitByte reports whether c is an ASCII decimal digit.
func isDigitByte(c byte) bool { return c >= '0' && c <= '9' }

// isHexByte reports whether c is a lowercase-hex digit [0-9a-f].
func isHexByte(c byte) bool { return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') }

// isTSByte reports whether c may appear inside an RFC3339-ish timestamp run
// once one has started ([0-9] plus the punctuation glue).
func isTSByte(c byte) bool {
	switch c {
	case ':', 'T', 'Z', '.', '+', '-':
		return true
	}
	return c >= '0' && c <= '9'
}

// volatileRun returns the end index of a maximal "volatile" run beginning at i,
// or i itself if block[i] does not start one (a literal byte). It is the single
// source of truth for normalizer boundaries: both normalizeBlock and diffTokens
// walk with it, so a run is normalized to '#' exactly when it is treated as one
// differing token — the two views can never disagree.
//
// Classification, in priority order, at a byte that is a digit or [a-f]:
//  1. an RFC3339-ish timestamp: 4 digits + '-' prefix, then a maximal run of
//     [0-9] and the glue bytes -:TZ.+ ;
//  2. a maximal run of lowercase-hex [0-9a-f] of length >= 8 ;
//  3. a maximal run of ASCII digits (length >= 1).
//
// Anything else (including a hex-letter that neither starts a >=8 hex run nor a
// digit run) is literal.
func volatileRun(block string, i int) int {
	c := block[i]
	isDigit := isDigitByte(c)
	isHexLetter := c >= 'a' && c <= 'f'
	if !isDigit && !isHexLetter {
		return i // literal
	}
	n := len(block)

	// (1) timestamp: dddd- prefix.
	if isDigit && i+4 < n && block[i+4] == '-' &&
		isDigitByte(block[i+1]) && isDigitByte(block[i+2]) && isDigitByte(block[i+3]) {
		j := i
		for j < n && isTSByte(block[j]) {
			j++
		}
		return j
	}

	// (2) hex run of length >= 8.
	hj := i
	for hj < n && isHexByte(block[hj]) {
		hj++
	}
	if hj-i >= 8 {
		return hj
	}

	// (3) digit run of length >= 1.
	if isDigit {
		dj := i
		for dj < n && isDigitByte(block[dj]) {
			dj++
		}
		return dj
	}

	return i // hex letter, but neither a >=8 hex run nor a digit run — literal
}

// normalizeBlock writes into buf a copy of block with every volatile run
// replaced by a single '#', using volatileRun for boundaries, and returns the
// filled slice. buf is reused across calls (pass buf[:0]); it grows once and is
// never aliased into the result maps (callers copy via string(...)).
func normalizeBlock(block string, buf []byte) []byte {
	buf = buf[:0]
	for i := 0; i < len(block); {
		if e := volatileRun(block, i); e > i {
			buf = append(buf, '#')
			i = e
		} else {
			buf = append(buf, block[i])
			i++
		}
	}
	return buf
}

// diffTokens walks two blocks in lockstep over the SAME volatile boundaries and
// collects the ordered differing token pairs. ok is false unless the blocks are
// byte-identical outside their volatile tokens, tokenize to the same structure,
// differ in at least one but no more than maxFuzzyDiffs tokens, AND every pair's
// old value is unique among the collected pairs. That uniqueness requirement is
// load-bearing for zero-loss: the fuzzy marker lists old→new pairs with no
// positional information, so if two pairs shared the same old value a reader
// could not tell which occurrence of old maps to which new — reconstruction
// would be ambiguous. When ok, the returned pairs cover EVERY difference AND
// unambiguously identify which new value replaces which old value, so a marker
// listing them loses nothing.
func diffTokens(a, b string) (pairs []tokenPair, ok bool) {
	ia, ib := 0, 0
	for ia < len(a) && ib < len(b) {
		ea := volatileRun(a, ia)
		eb := volatileRun(b, ib)
		aTok := ea > ia
		bTok := eb > ib
		if aTok != bTok {
			return nil, false // token structure diverges
		}
		if !aTok {
			if a[ia] != b[ib] {
				return nil, false // literal bytes differ
			}
			ia++
			ib++
			continue
		}
		ta := a[ia:ea]
		tb := b[ib:eb]
		if ta != tb {
			pairs = append(pairs, tokenPair{old: ta, new: tb})
			if len(pairs) > maxFuzzyDiffs {
				return nil, false
			}
		}
		ia = ea
		ib = eb
	}
	if ia != len(a) || ib != len(b) {
		return nil, false // one block has trailing content the other lacks
	}
	if len(pairs) == 0 {
		return nil, false // identical — the exact-dup path owns this case
	}
	// Refuse when two pairs share the same old value: the marker has no
	// positional information, so which occurrence of old maps to which new
	// could not be reconstructed — ambiguous, so keep the block verbatim
	// instead of losing information. pairs is capped at maxFuzzyDiffs (4), so
	// this O(n²) scan is free.
	for i := range pairs {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[i].old == pairs[j].old {
				return nil, false
			}
		}
	}
	return pairs, true
}

// FoldRepeatedBlocks collapses NON-adjacent duplicate blocks within a single
// payload — the case [SqueezeOutput]'s consecutive run-length pass cannot see.
// Tool output that repeats whole sections (e.g. an MCP batch that re-dumps the
// same grep result under several query headers) shrinks a lot here.
//
// A "block" is a maximal run of lines containing no blank line; blocks are
// separated by runs of two-or-more newlines. Separators and first occurrences
// are copied byte-for-byte, so non-duplicated content is never altered. The
// first occurrence of a block is kept verbatim; each later occurrence (at least
// minFoldBlock bytes, and only when the marker is strictly shorter) is replaced
// by a self-describing one-line back-reference naming it by its first line:
//
//   - exact duplicate → [repeat: "<first line>"]
//   - near-duplicate (identical apart from volatile tokens — ids, hex digests,
//     timestamps) → [repeat of "<first line>" except <old>-><new>, …], which
//     lists every differing token so the reader can reconstruct the block with
//     ZERO information loss. This fold is refused (block kept verbatim) if any
//     two differing tokens share the same old value: with no positional
//     information in the marker, which occurrence maps to which new value
//     would be ambiguous (see diffTokens).
//
// No expansion step is needed — the referent is above in the same text, exactly
// like the "⨯N" run-length markers.
//
// It returns the input unchanged when nothing folds, so callers can gate on
// len(). The output builder is allocated lazily — only once a fold actually
// happens — so the common no-duplicate payload pays for nothing but the dedup
// map(s) and a single scan. The near-duplicate machinery (normalize + second
// map) is touched only for blocks that miss the exact map AND actually contain
// a volatile token, so token-free payloads keep the exact-dup fast path.
func FoldRepeatedBlocks(content string) string {
	// A non-adjacent duplicate needs at least two blocks, i.e. a blank-line
	// boundary. No "\n\n" (or too small) ⇒ nothing to do, allocate nothing.
	if len(content) < minFoldBlock*2 || !strings.Contains(content, "\n\n") {
		return content
	}

	var b strings.Builder
	active := false // builder in use (set on the first fold)
	written := 0    // bytes of content already committed to b
	seen := seenPool.Get().(map[string]struct{})
	defer func() { putSeen(seen, len(seen)) }() // len() captured before putSeen's clear()
	var seenNorm map[string]string              // normalized key → first (base) block; lazy, pooled
	defer func() {
		if seenNorm != nil {
			putSeenNorm(seenNorm, len(seenNorm)) // len() captured before putSeenNorm's clear()
		}
	}()
	var scratch []byte // reused normalize buffer; lazy

	// emitFuzzy writes a near-duplicate marker for block [bs,be) (verbatim key,
	// with block==content[bs:be]) referencing base, listing pairs. It reports
	// whether it folded: false (block kept verbatim) unless the marker is
	// strictly shorter than the block.
	emitFuzzy := func(bs, be int, key, block, base string, pairs []tokenPair) bool {
		first, trunc := truncFirstLine(base)
		mlen := len(fuzzyPre) + len(first)
		if trunc {
			mlen += len(markerEll)
		}
		mlen += len(fuzzyExcept)
		for i, p := range pairs {
			if i > 0 {
				mlen += len(fuzzySep)
			}
			mlen += len(p.old) + len(fuzzyArrow) + len(p.new)
		}
		mlen += len(fuzzySuf)
		if mlen >= len(key) { // marker wouldn't be smaller — keep verbatim
			return false
		}
		if !active {
			b.Grow(len(content))
			active = true
		}
		b.WriteString(content[written:bs]) // verbatim gap (kept blocks + separators)
		b.WriteString(fuzzyPre)
		b.WriteString(first)
		if trunc {
			b.WriteString(markerEll)
		}
		b.WriteString(fuzzyExcept)
		for i, p := range pairs {
			if i > 0 {
				b.WriteString(fuzzySep)
			}
			b.WriteString(p.old)
			b.WriteString(fuzzyArrow)
			b.WriteString(p.new)
		}
		b.WriteString(fuzzySuf)
		b.WriteString(block[len(key):]) // preserved trailing newlines
		written = be
		return true
	}

	// fold checks one block [bs,be). On an exact duplicate that shrinks it emits
	// the exact marker; otherwise, on a near-duplicate that shrinks, the fuzzy
	// marker. In both cases it lazily starts the builder, flushes the verbatim
	// gap since the last write, and preserves the block's trailing newlines. vol
	// reports whether the block carries a volatile run (a digit, or a hex run) —
	// tracked by the caller's single scan so the near-dup path costs nothing on
	// token-free payloads.
	fold := func(bs, be int, vol bool) {
		block := content[bs:be]
		key := strings.TrimRight(block, "\n")
		if len(key) < minFoldBlock {
			return
		}
		if _, dup := seen[key]; dup {
			// Exact duplicate: unchanged fast path.
			first, trunc := truncFirstLine(key)
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
			return
		}

		// Near-duplicate path: only for blocks that actually carry a volatile
		// token. Token-free blocks normalize to themselves, so a matching
		// normalized key would be a byte-identical block already caught above.
		if !vol {
			seen[key] = struct{}{}
			return
		}
		scratch = normalizeBlock(key, scratch)
		if base, ok := seenNorm[string(scratch)]; ok { // no-alloc string(bytes) lookup
			if pairs, ok := diffTokens(base, key); ok {
				if emitFuzzy(bs, be, key, block, base, pairs) {
					// Folded: the block is now a marker, not verbatim. Do NOT
					// record it in seen — a later exact copy folds against the
					// same verbatim base instead of chaining off this marker.
					return
				}
			}
			// Kept verbatim (too different, or marker not shorter): a future
			// exact copy may back-reference this occurrence.
			seen[key] = struct{}{}
			return
		}
		// First occurrence of this normalized shape: keep verbatim, register it
		// as the base for future near-duplicates and for exact-copy folding.
		if seenNorm == nil {
			seenNorm = seenNormPool.Get().(map[string]string)
		}
		seenNorm[string(scratch)] = key
		seen[key] = struct{}{}
	}

	// One pass over content locates block boundaries AND flags whether the
	// current segment contains an ASCII digit — so the near-dup normalize step
	// never triggers a second scan. vol is a cheap, conservative gate, NOT the
	// authoritative classifier: normalizeBlock/diffTokens (via volatileRun) stay
	// exact. Every real volatile token this format targets — request ids, byte
	// counts, RFC3339 timestamps, hex digests — carries at least one digit, so a
	// digit is a reliable "might be a near-dup base" signal. A digit-free block
	// is treated as token-free: at worst a purely-alphabetic hex run (e.g.
	// "deadbeef") is not folded, which only forgoes a fold (never-worse) and
	// never alters output. The per-byte cost is one blockClass load plus a
	// zero-test — the same shape as a bare '\n' compare — so the no-duplicate
	// hot path pays almost nothing for the added digit detection.
	n := len(content)
	segStart := 0
	vol := false
	i := 0
	for i < n {
		class := blockClass[content[i]]
		if class == 0 {
			i++
			continue
		}
		if class == blockNewline {
			j := i + 1
			for j < n && content[j] == '\n' {
				j++
			}
			if j-i >= 2 { // blank line -> block boundary; separator [i,j) kept verbatim
				fold(segStart, i, vol)
				segStart = j
				i = j
				vol = false
				continue
			}
			i++
			continue
		}
		vol = true // blockDigit: segment carries a digit — a possible near-dup token
		i++
	}
	fold(segStart, n, vol) // trailing block

	if !active {
		return content
	}
	b.WriteString(content[written:]) // verbatim tail after the last fold
	return b.String()
}
