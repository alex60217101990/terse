// Package tokens estimates how many tokens a string costs.
//
// qdf-hook exists to reduce the tokens a tool result consumes, so every
// never-worse gate has to be expressed in tokens. Bytes are a poor proxy: hex
// costs about 1.1 bytes per token while prose costs about 4.8, so two outputs
// of equal size can differ fourfold in what is actually paid. Byte-equal
// changes are not token-equal either, which is how a marker change worth 18% of
// the tokens can look like a 5% change in bytes and never get made.
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

// Count returns an estimated o200k token count for s.
//
// The model walks maximal runs of a single character class and charges each run
// by class. One space before a word is free: BPE folds it into the following
// token (" world" is a single token), so only whitespace runs of two or more
// are charged.
func Count(s string) int {
	if len(s) == 0 {
		return 0
	}

	milli := 0
	for i := 0; i < len(s); {
		c := class(s[i])
		j := i + 1
		for j < len(s) && class(s[j]) == c {
			j++
		}
		runLen := j - i

		switch c {
		case classNewline:
			milli += runLen * mtPerNewline
		case classPunct:
			milli += runLen * mtPerPunct
		case classSpace:
			// A single space attaches to the following word for free.
			if runLen > 1 {
				milli += ceilDiv(runLen*1000, mbPerTokSpace) * 1000
			}
		case classLower:
			milli += ceilDiv(runLen*1000, mbPerTokLower) * 1000
		case classUpper:
			milli += ceilDiv(runLen*1000, mbPerTokUpper) * 1000
		case classDigit:
			milli += ceilDiv(runLen*1000, mbPerTokDigit) * 1000
		default: // classHigh
			milli += ceilDiv(runLen*1000, mbPerTokHigh) * 1000
		}
		i = j
	}

	n := (milli + 999) / 1000
	if n < 1 {
		n = 1 // a non-empty string always costs something
	}
	return n
}

func ceilDiv(a, b int) int { return (a + b - 1) / b }

const (
	classLower = iota
	classUpper
	classDigit
	classSpace
	classNewline
	classPunct
	classHigh
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
	case b >= 0x80:
		return classHigh
	default:
		return classPunct
	}
}
