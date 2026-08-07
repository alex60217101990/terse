// Package tokens estimates how many tokens a string costs.
//
// qdf-hook exists to reduce the tokens a tool result consumes, so every
// never-worse gate has to be expressed in tokens. Bytes are a poor proxy: hex
// costs about 1.1 bytes per token while prose costs about 4.8, so two outputs
// of equal size can differ fourfold in what is actually paid. Byte-equal
// changes are not token-equal either, which is how a marker change worth 18% of
// the tokens can look like 5% in bytes and never get made.
//
// Count is a fast approximation, not an encoder. It sits on the every-payload
// hot path, so it is single-pass and allocation-free. The exact encoder lives
// in the bpe subpackage and is used only by tests and the replay harness; the
// agreement tests pin how far the two are allowed to drift.
//
// The reference encoding is o200k, which is a public proxy for Claude's
// tokenizer rather than the real thing. Counts are for comparing two candidate
// outputs, not for estimating production cost.
package tokens

// Weights parameterises Count. It exists so `qdf-replay calibrate` can search
// over the REAL counting function instead of a linear stand-in: Count rounds up
// once per run, which a closed-form least-squares fit cannot express, so fitting
// a linear proxy would optimise the wrong thing.
//
// All fields are bytes per token in milli-bytes (4400 == 4.4 bytes per token),
// except MtPerNewline which is a flat cost in milli-tokens.
//
// Non-ASCII is split by UTF-8 sequence length rather than lumped together. A
// single weight cannot fit Cyrillic, CJK and emoji at once — they differ by
// several times in bytes per token, and forcing one number to cover all three
// drags the ASCII weights off to compensate.
type Weights struct {
	MbPerTokLower int
	MbPerTokUpper int
	MbPerTokDigit int
	MbPerTokSpace int
	MbPerTokPunct int // runs like "===" or "://" merge, so charge per run
	MbPerTokHigh2 int // 2-byte UTF-8: Latin-1 supplement, Cyrillic, Greek
	MbPerTokHigh3 int // 3-byte UTF-8: CJK, most symbols, box drawing
	MbPerTokHigh4 int // 4-byte UTF-8: emoji and astral planes
	MtPerNewline  int
}

// Default returns the fitted weights Count uses.
func Default() Weights {
	return Weights{
		MbPerTokLower: mbPerTokLower,
		MbPerTokUpper: mbPerTokUpper,
		MbPerTokDigit: mbPerTokDigit,
		MbPerTokSpace: mbPerTokSpace,
		MbPerTokPunct: mbPerTokPunct,
		MbPerTokHigh2: mbPerTokHigh2,
		MbPerTokHigh3: mbPerTokHigh3,
		MbPerTokHigh4: mbPerTokHigh4,
		MtPerNewline:  mtPerNewline,
	}
}

var defaultWeights = Default()

// Count returns an estimated o200k token count for s.
//
// The model walks maximal runs of a single character class and charges each run
// by class. One space before a word is free: BPE folds it into the following
// token (" world" is a single token), so only whitespace runs of two or more
// are charged.
func Count(s string) int { return CountWith(s, defaultWeights) }

// CountWith is Count with explicit weights. Calibration is its only caller;
// everything else should use Count so there is exactly one fitted set in play.
func CountWith(s string, w Weights) int {
	if len(s) == 0 {
		return 0
	}

	milli := 0
	for i := 0; i < len(s); {
		c := class(s[i])
		j := i + 1
		// A multi-byte rune's continuation bytes belong to the run its lead byte
		// opened; classifying them separately would split every non-ASCII
		// character into its own run.
		for j < len(s) && (class(s[j]) == c || (c >= classHigh2 && isCont(s[j]))) {
			j++
		}
		runLen := j - i

		switch c {
		case classNewline:
			milli += runLen * w.MtPerNewline
		case classSpace:
			// A single space attaches to the following word for free.
			if runLen > 1 {
				milli += perRun(runLen, w.MbPerTokSpace)
			}
		case classLower:
			milli += perRun(runLen, w.MbPerTokLower)
		case classUpper:
			milli += perRun(runLen, w.MbPerTokUpper)
		case classDigit:
			milli += perRun(runLen, w.MbPerTokDigit)
		case classPunct:
			milli += perRun(runLen, w.MbPerTokPunct)
		case classHigh2:
			milli += perRun(runLen, w.MbPerTokHigh2)
		case classHigh3:
			milli += perRun(runLen, w.MbPerTokHigh3)
		default: // classHigh4
			milli += perRun(runLen, w.MbPerTokHigh4)
		}
		i = j
	}

	n := (milli + 999) / 1000
	if n < 1 {
		n = 1 // a non-empty string always costs something
	}
	return n
}

// perRun charges a run of runLen bytes at mbPerTok milli-bytes per token,
// rounding up to a whole token and returning milli-tokens.
func perRun(runLen, mbPerTok int) int {
	return ceilDiv(runLen*1000, mbPerTok) * 1000
}

func ceilDiv(a, b int) int { return (a + b - 1) / b }

func isCont(b byte) bool { return b&0xC0 == 0x80 }

const (
	classLower = iota
	classUpper
	classDigit
	classSpace
	classNewline
	classPunct
	classHigh2
	classHigh3
	classHigh4
)

func class(b byte) int {
	switch {
	case b >= 'a' && b <= 'z':
		return classLower
	case b >= 'A' && b <= 'Z':
		return classUpper
	case b >= '0' && b <= '9':
		return classDigit
	case b == ' ' || b == '\t':
		return classSpace
	case b == '\n' || b == '\r':
		return classNewline
	case b < 0x80:
		return classPunct
	case b < 0xE0:
		// Includes continuation bytes (0x80-0xBF), which only reach here when a
		// string starts mid-rune; charging them as 2-byte is the closest guess.
		return classHigh2
	case b < 0xF0:
		return classHigh3
	default:
		return classHigh4
	}
}
