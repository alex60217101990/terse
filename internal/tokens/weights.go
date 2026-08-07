package tokens

// Class weights used by Count.
//
// These are STARTING values for the search, not a fit. `qdf-replay calibrate`
// replaces this whole block with the result of a coordinate-descent fit against
// the exact BPE over the committed synthetic corpus.
//
// Do not hand-tune. Refit and paste, so every number stays traceable to a
// corpus a contributor can regenerate.
const (
	// Bytes per token within a run of the given class, in milli-bytes
	// (4400 == 4.4 bytes per token).
	mbPerTokLower = 4400 // "handler", " return" — the common case
	mbPerTokUpper = 2000 // "HTTPServer" fragments into more pieces
	mbPerTokDigit = 3000 // BPE groups digits in runs of up to three
	mbPerTokSpace = 6000 // long indentation runs merge aggressively
	mbPerTokPunct = 1100 // mostly one token each; runs like "===" merge
	mbPerTokHigh2 = 2500 // 2-byte UTF-8: Cyrillic, Greek, Latin-1 supplement
	mbPerTokHigh3 = 3000 // 3-byte UTF-8: CJK, box drawing, most symbols
	mbPerTokHigh4 = 2000 // 4-byte UTF-8: emoji and astral planes

	// Flat per-occurrence cost, in milli-tokens.
	mtPerNewline = 1000
)
