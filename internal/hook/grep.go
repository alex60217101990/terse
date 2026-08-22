package hook

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/alex60217101990/terse/internal/hookcore"
)

// grepFileCap is the max matching lines shown per file before eliding.
const grepFileCap = 8

// grepGroupsPool recycles the file→matches map buildGrepSummary builds on every
// call. Its []grepMatch values alias per-call substrings of content, so the map
// is clear()ed before going back (no stale content pinned) and dropped when it
// grew past maxPooledGrepGroups — clear() empties entries but does not shrink
// the bucket array, so a one-off huge-file grep must not pin an oversized map
// in the pool. Mirrors detect's seenPool/shouldPoolDedupMap pattern.
var grepGroupsPool = sync.Pool{New: func() any { return make(map[string][]grepMatch) }}

const maxPooledGrepGroups = 4096

// shouldPoolGrepGroups reports whether a groups map with n entries (measured
// BEFORE clear()) is small enough to return to grepGroupsPool. Isolated as a
// pure function so the size-guard boundary can be tested deterministically,
// mirroring detect's shouldPoolDedupMap.
func shouldPoolGrepGroups(n int) bool { return n <= maxPooledGrepGroups }

// HandleGrep is retained for backward compatibility; Grep is handled by the
// generic pipeline via Dispatch (buildGrepSummary is its tool-specific step).
func HandleGrep(r io.Reader, w io.Writer) error { return Dispatch(hookcore.NewDiskStore(), r, w) }

type grepMatch struct {
	line string // line number as text (kept as-is)
	text string
}

// parseGrepLine splits a ripgrep/grep content line "file:linenum:text".
// It reports ok=false unless the segment between the first two colons is all
// digits (the line-number) AND the leading segment is a plausible file path
// (see plausibleGrepFile). The digit rule alone misreads colon-delimited text
// — ISO timestamps ("2024-01-01T00:00:00 msg"), clock times ("12:30:45 msg"),
// and config dumps ("service:8000:desc") — as "file:line:text", fabricating
// grep structure over plain logs. The path check closes that hole.
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
	if !plausibleGrepFile(s[:i]) {
		return "", "", "", false
	}
	return s[:i], num, rest[j+1:], true
}

// plausibleGrepFile reports whether seg is believable as a grep path segment.
// A real grep/ripgrep path virtually always contains a '/' or a '.' and never
// a space; a colon-delimited log/config field ("service", "12", "2024-01-01T00")
// has none of the former. It also rejects a date-shaped lead (4 digits then a
// '-', e.g. "2024-01-15.log") so an ISO date that happens to carry a '.' can't
// sneak through the path-char test. Zero-alloc: byte scans only, no allocation.
//
// Accepted residual risk: dotted non-path left segments ("db.host:5432:x",
// "10.0.0.1:8080:x") still pass this test and classify as grep. This is
// deliberate — the resulting summary is lossy but fully recoverable via the
// withRecovery footer in dispatch.go, so nothing is unrecoverably dropped.
func plausibleGrepFile(seg string) bool {
	if seg == "" {
		return false
	}
	// Date-shaped lead: 4 leading digits followed by '-' (e.g. "2024-01-01").
	if len(seg) >= 5 && seg[4] == '-' &&
		seg[0] >= '0' && seg[0] <= '9' &&
		seg[1] >= '0' && seg[1] <= '9' &&
		seg[2] >= '0' && seg[2] <= '9' &&
		seg[3] >= '0' && seg[3] <= '9' {
		return false
	}
	hasPathChar := false
	for k := range len(seg) {
		switch seg[k] {
		case ' ', '\t':
			return false
		case '/', '.':
			hasPathChar = true
		}
	}
	return hasPathChar
}

// buildGrepSummary returns (summary, action). action is "grouped" for content
// mode, "tree" when delegated to the file-tree compressor, or "" (empty
// summary) when the input doesn't look like grep output.
func buildGrepSummary(content string) (string, string) {
	groups, ok := grepGroupsPool.Get().(map[string][]grepMatch)
	if !ok {
		panic("grepGroupsPool: unexpected type")
	}
	// Cleared and returned to the pool on every exit path (len captured before
	// clear() so the size guard sees the true entry count). The result string
	// has already copied out any needed substrings via the Builder below.
	defer func() {
		n := len(groups) // captured before clear() so the guard sees the true count
		clear(groups)
		if shouldPoolGrepGroups(n) {
			grepGroupsPool.Put(groups)
		}
	}()
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
