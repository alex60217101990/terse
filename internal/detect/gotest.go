package detect

import (
	"strings"
)

// IsGoTestOutput returns true if s looks like `go test -v` output.
// Cheap heuristic: looks for "=== RUN" or standard PASS/FAIL markers.
func IsGoTestOutput(s string) bool {
	return strings.Contains(s, "=== RUN") ||
		(strings.Contains(s, "--- PASS") && strings.Contains(s, "--- FAIL")) ||
		(strings.Contains(s, "--- PASS:") && strings.Contains(s, "\nok "))
}

// SummarizeGoTest condenses go test -v output into a compact summary.
// For passing runs: one-line summary. For failing runs: full failure details preserved.
func SummarizeGoTest(s string) string {
	var (
		passCount, failCount, skipCount int
		duration                        string
		pkg                             string
		pkgCount                        int // distinct package-result lines seen
		failLines                       []string
		currentIndented                 []string // indented lines for the current test
	)

	for line := range strings.SplitSeq(s, "\n") {
		switch {
		case strings.HasPrefix(line, "=== RUN"):
			currentIndented = nil
		case strings.HasPrefix(line, "--- PASS:"):
			passCount++
			currentIndented = nil
		case strings.HasPrefix(line, "--- FAIL:"):
			failCount++
			failLines = append(failLines, line)
			failLines = append(failLines, currentIndented...)
			currentIndented = nil
		case strings.HasPrefix(line, "--- SKIP:"):
			skipCount++
			currentIndented = nil
		case strings.HasPrefix(line, "ok  \t") || strings.HasPrefix(line, "ok\t"):
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				pkg = parts[1]
				duration = parts[2]
				pkgCount++
			}
		case strings.HasPrefix(line, "FAIL\t"):
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pkg = parts[1]
				pkgCount++
			}
		case strings.HasPrefix(line, "    "):
			currentIndented = append(currentIndented, line)
		}
	}

	// When the run spanned multiple packages, the single tracked pkg/duration
	// would misattribute the summary to just the last one — label it by count
	// instead, and drop the (per-package) duration.
	if pkgCount > 1 {
		pkg = itoa(pkgCount) + " packages"
		duration = ""
	}

	var sb strings.Builder
	if failCount == 0 {
		if pkg != "" {
			sb.WriteString("[go test PASS] ")
			sb.WriteString(pkg)
			if duration != "" {
				sb.WriteString(" ")
				sb.WriteString(duration)
			}
			sb.WriteString("\n")
		}
		sb.WriteString("[")
		if passCount > 0 {
			sb.WriteString(itoa(passCount))
			sb.WriteString(" PASS")
		}
		if skipCount > 0 {
			if passCount > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(itoa(skipCount))
			sb.WriteString(" SKIP")
		}
		sb.WriteString("]\n")
	} else {
		sb.WriteString("[go test FAIL] ")
		sb.WriteString(pkg)
		sb.WriteString("\n")
		sb.WriteString("[")
		sb.WriteString(itoa(passCount))
		sb.WriteString(" PASS, ")
		sb.WriteString(itoa(failCount))
		sb.WriteString(" FAIL")
		if skipCount > 0 {
			sb.WriteString(", ")
			sb.WriteString(itoa(skipCount))
			sb.WriteString(" SKIP")
		}
		sb.WriteString("]\n")
		sb.WriteString("FAILURES:\n")
		for _, l := range failLines {
			sb.WriteString(l)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := 20
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
