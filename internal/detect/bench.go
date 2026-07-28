package detect

import (
	"strings"
	"unicode/utf8"
)

// IsGoBenchOutput returns true if s contains Go benchmark output.
func IsGoBenchOutput(s string) bool {
	return strings.Contains(s, "ns/op") && strings.Contains(s, "Benchmark")
}

// SummarizeGoBench formats benchmark output as a compact aligned table.
func SummarizeGoBench(s string) string {
	// nsop/bop/allocsop hold just the numeric token (aliasing s, no allocation);
	// the unit suffix is written at output time. Empty means the metric was
	// absent on that line (e.g. a run without -benchmem).
	type row struct {
		name     string
		nsop     string
		bop      string
		allocsop string
	}
	var rows []row

	for line := range strings.SplitSeq(s, "\n") {
		if !strings.HasPrefix(line, "Benchmark") {
			continue
		}
		// Single pass over whitespace-delimited fields (no Fields slice). The
		// metric label at field i marks field i-1 as its value — exactly the
		// pairing the old loop did over fields[i], fields[i+1] for i >= 1, so
		// field 1 (the iteration count) is never treated as a value.
		var name, nsop, bop, allocsop, prev string
		count := 0
		for i := 0; i < len(line); {
			for i < len(line) && isASCIISpace(line[i]) {
				i++
			}
			if i >= len(line) {
				break
			}
			start := i
			for i < len(line) && !isASCIISpace(line[i]) {
				i++
			}
			tok := line[start:i]
			count++
			switch {
			case count == 1:
				name = tok
			case count >= 3:
				switch tok {
				case "ns/op":
					nsop = prev
				case "B/op":
					bop = prev
				case "allocs/op":
					allocsop = prev
				}
			}
			prev = tok
		}
		if count < 4 {
			continue
		}
		rows = append(rows, row{name: name, nsop: nsop, bop: bop, allocsop: allocsop})
	}

	if len(rows) == 0 {
		return s
	}

	var sb strings.Builder
	sb.WriteString("[go bench — ")
	sb.WriteString(itoa(len(rows)))
	sb.WriteString(" benchmarks]\n")
	for _, r := range rows {
		// Name column, left-justified to 42 runes (fmt %-42s measures width in
		// runes). Truncate on a rune boundary so a multi-byte name is never
		// split into invalid UTF-8 — matches SummarizeGitLog's truncation.
		if off, trunc := runeTrunc(r.name, 40, 37); trunc {
			sb.WriteString(r.name[:off])
			sb.WriteString("...")
			writePadTo(&sb, 40, 42) // 37 kept + 3 dots = 40 runes written
		} else {
			sb.WriteString(r.name)
			writePadTo(&sb, utf8.RuneCountInString(r.name), 42)
		}
		sb.WriteByte(' ')
		writeMetric(&sb, r.nsop, " ns/op")
		sb.WriteString("  ")
		writeMetric(&sb, r.bop, " B/op")
		sb.WriteString("  ")
		writeMetric(&sb, r.allocsop, " allocs/op")
		sb.WriteByte('\n')
	}
	return sb.String()
}

// writePadTo writes (width-have) spaces, or none when have >= width.
func writePadTo(sb *strings.Builder, have, width int) {
	for range width - have {
		sb.WriteByte(' ')
	}
}

// writeMetric writes val+unit when val is non-empty, otherwise nothing — so a
// missing metric prints as the empty string the old fmt "%s" of "" produced.
func writeMetric(sb *strings.Builder, val, unit string) {
	if val != "" {
		sb.WriteString(val)
		sb.WriteString(unit)
	}
}
