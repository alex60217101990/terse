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
		pkgFailed                       bool     // saw a package-level FAIL line
		crash                           bool     // panic / timeout / data race
	)

	for line := range strings.SplitSeq(s, "\n") {
		if !crash && (strings.HasPrefix(line, "panic:") ||
			strings.HasPrefix(line, "fatal error:") ||
			strings.Contains(line, "test timed out after") ||
			strings.Contains(line, "DATA RACE")) {
			crash = true
		}
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
			var f [3]string
			if fieldsInto(line, f[:]) >= 3 {
				pkg = f[1]
				duration = f[2]
				pkgCount++
			}
		case strings.HasPrefix(line, "FAIL\t"):
			pkgFailed = true
			var f [2]string
			if fieldsInto(line, f[:]) >= 2 {
				pkg = f[1]
				pkgCount++
			}
		case line == "FAIL":
			pkgFailed = true
		case strings.HasPrefix(line, "    "):
			// Indented subtest results (`    --- FAIL: T/sub`) were folded into
			// the parent's detail but never tallied — count them here so the
			// headline [N PASS, M FAIL] includes subtests. The line still goes
			// into currentIndented so the FAILURES block is unchanged.
			switch t := strings.TrimLeft(line, " "); {
			case strings.HasPrefix(t, "--- PASS:"):
				passCount++
			case strings.HasPrefix(t, "--- FAIL:"):
				failCount++
			case strings.HasPrefix(t, "--- SKIP:"):
				skipCount++
			}
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

	// A crash/timeout/data-race, or a package-level FAIL with no per-test
	// --- FAIL: line, carries its diagnostic OUTSIDE the structured markers this
	// summarizer captures. Worse, a crash/timeout has failCount==0, so it would
	// be reported as PASS. Return "" so the pipeline passes the full, unmodified
	// output through — the failure cause must reach the model intact.
	if crash || (pkgFailed && failCount == 0) {
		return ""
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

// fieldsInto fills dst with up to len(dst) whitespace-delimited fields of s
// (each aliasing s, no allocation) and returns how many it wrote — a bounded,
// zero-alloc strings.Fields for callers that need only the first few fields of
// a result line and can supply a stack array.
func fieldsInto(s string, dst []string) int {
	n := 0
	for i := 0; i < len(s) && n < len(dst); {
		for i < len(s) && isASCIISpace(s[i]) {
			i++
		}
		if i >= len(s) {
			break
		}
		start := i
		for i < len(s) && !isASCIISpace(s[i]) {
			i++
		}
		dst[n] = s[start:i]
		n++
	}
	return n
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
