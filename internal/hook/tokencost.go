package hook

// statsTokens resolves the (in, out) token pair for one analytics row from counts
// that were already paid for, and never tokenises anything itself.
//
// Counting is exact, which means a full BPE pass over the payload: measured at
// roughly 15 MB/s, allocating per regex match — about 330 allocations per
// kilobyte. A never-worse GATE is worth that, because it decides whether the
// model reads the compressed form at all. A statistics row is not: this hook runs
// synchronously in the user's tool-call path, and nobody trades interactive
// latency for a prettier number in `qdf stats`. Doing it anyway is what turned
// one Bash payload into 261k allocations and 66ms.
//
// So the rows that carry exact token counts are exactly the rows where a gate
// needed them, which are the rows where compression actually happened — the ones
// whose saving is non-zero and worth reporting precisely. Everything else records
// zeroes, and analytics.eventTokens reads a zero PAIR as "no token data here",
// falling back to the byte estimate with EstimatedTokens set. Estimated and
// flagged, never estimated and presented as exact.
//
// Both sides are zeroed together for that reason: zeroing one alone would make
// in-out come out negative and read as the hook inflating its own output.
//
// inKnown/outKnown are the counts a gate already computed, or -1 when unknown.
func statsTokens(in, out string, inKnown, outKnown int) (int, int) {
	if inKnown < 0 || outKnown < 0 {
		// One known side is not enough to report a saving, and counting the other
		// is the cost this function exists to avoid.
		if inKnown >= 0 && out == in {
			return inKnown, inKnown
		}
		return 0, 0
	}
	return inKnown, outKnown
}
