package hook

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/alex60217101990/terse/internal/hookcore"
)

// grepFileCap is the max matching lines shown per file before eliding.
const grepFileCap = 8

// HandleGrep is retained for backward compatibility; Grep is handled by the
// generic pipeline via Dispatch (buildGrepSummary is its tool-specific step).
func HandleGrep(r io.Reader, w io.Writer) error { return Dispatch(hookcore.NewDiskStore(), r, w) }

type grepMatch struct {
	line string // line number as text (kept as-is)
	text string
}

// parseGrepLine splits a ripgrep/grep content line "file:linenum:text".
// It reports ok=false unless the segment between the first two colons is all
// digits (the line-number), which distinguishes content mode from bare paths.
func parseGrepLine(s string) (file, line, text string, ok bool) {
	i := strings.IndexByte(s, ':')
	if i <= 0 || i == len(s)-1 {
		return "", "", "", false
	}
	rest := s[i+1:]
	j := strings.IndexByte(rest, ':')
	if j <= 0 {
		return "", "", "", false
	}
	num := rest[:j]
	for _, c := range num {
		if c < '0' || c > '9' {
			return "", "", "", false
		}
	}
	return s[:i], num, rest[j+1:], true
}

// buildGrepSummary returns (summary, action). action is "grouped" for content
// mode, "tree" when delegated to the file-tree compressor, or "" (empty
// summary) when the input doesn't look like grep output.
func buildGrepSummary(content string) (string, string) {
	groups := make(map[string][]grepMatch)
	// SplitSeq: single forward pass, no []string materialized. lineCount
	// replaces len(lines) for the bare-list ratio below.
	parsed, lineCount := 0, 0
	for ln := range strings.SplitSeq(strings.TrimSpace(content), "\n") {
		lineCount++
		file, num, text, ok := parseGrepLine(ln)
		if !ok {
			continue
		}
		parsed++
		groups[file] = append(groups[file], grepMatch{line: num, text: text})
	}

	// If almost nothing parsed as content matches, treat the output as a bare
	// path list (files_with_matches) and reuse the Glob tree compressor. The
	// `parsed == 0` short-circuit handles a single bare path (lineCount 1,
	// parsed 0) — where `parsed < lineCount/2` is `0 < 0` (false) and would
	// otherwise emit a bogus "0 matches in 0 files".
	if parsed == 0 || parsed < lineCount/2 {
		if tree := buildGlobTree(content); tree != "" {
			return tree, "tree"
		}
		return "", ""
	}

	// Alphabetical file order — same as the previous first-seen + sort.Strings.
	order := slices.Sorted(maps.Keys(groups))
	var sb strings.Builder
	total := 0
	for _, file := range order {
		ms := groups[file]
		total += len(ms)
		plural := "es"
		if len(ms) == 1 {
			plural = ""
		}
		fmt.Fprintf(&sb, "%s (%d match%s)\n", file, len(ms), plural)
		shown := ms
		if len(shown) > grepFileCap {
			shown = shown[:grepFileCap]
		}
		for _, m := range shown {
			sb.WriteString("  ")
			sb.WriteString(m.line)
			sb.WriteString(": ")
			sb.WriteString(strings.TrimSpace(m.text))
			sb.WriteByte('\n')
		}
		if len(ms) > grepFileCap {
			fmt.Fprintf(&sb, "  ... +%d more\n", len(ms)-grepFileCap)
		}
	}
	fmt.Fprintf(&sb, "[grep: %d matches in %d files]\n", total, len(order))
	return sb.String(), "grouped"
}
