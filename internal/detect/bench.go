package detect

import (
	"fmt"
	"strings"
)

// IsGoBenchOutput returns true if s contains Go benchmark output.
func IsGoBenchOutput(s string) bool {
	return strings.Contains(s, "ns/op") && strings.Contains(s, "Benchmark")
}

// SummarizeGoBench formats benchmark output as a compact aligned table.
func SummarizeGoBench(s string) string {
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
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		r := row{name: fields[0]}
		for i := 1; i+1 < len(fields); i++ {
			switch fields[i+1] {
			case "ns/op":
				r.nsop = fields[i] + " ns/op"
			case "B/op":
				r.bop = fields[i] + " B/op"
			case "allocs/op":
				r.allocsop = fields[i] + " allocs/op"
			}
		}
		rows = append(rows, r)
	}

	if len(rows) == 0 {
		return s
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "[go bench — %d benchmarks]\n", len(rows))
	for _, r := range rows {
		name := r.name
		// Truncate on a rune boundary so a multi-byte name (e.g. a subtest
		// with a non-ASCII rune) is never split into invalid UTF-8 — matches
		// SummarizeGitLog's []rune truncation.
		if rs := []rune(name); len(rs) > 40 {
			name = string(rs[:37]) + "..."
		}
		fmt.Fprintf(&sb, "%-42s %s  %s  %s\n", name, r.nsop, r.bop, r.allocsop)
	}
	return sb.String()
}
