package tokens

// Class weights used by Count.
//
// These are the INITIAL hand-set values. `qdf-replay calibrate` replaces this
// whole block with a least-squares fit against the exact BPE over the committed
// synthetic corpus. Do not tune them by hand afterwards — refit and paste, so
// every number stays traceable to a corpus a contributor can regenerate.
const (
	// Bytes per token within a run of the given class, in milli-bytes
	// (4400 == 4.4 bytes per token).
	mbPerTokLower = 4400 // "handler", " return" — the common case
	mbPerTokUpper = 2000 // "HTTPServer" fragments into more pieces
	mbPerTokDigit = 3000 // BPE groups digits in runs of up to three
	mbPerTokHigh  = 1500 // non-ASCII: multi-byte and rarely merged
	mbPerTokSpace = 8000 // long indentation runs merge aggressively

	// Flat per-occurrence costs, in milli-tokens.
	mtPerNewline = 1000
	mtPerPunct   = 900
)
