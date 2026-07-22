package hook

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

// grepFileCap is the max matching lines shown per file before eliding.
const grepFileCap = 8

// HandleGrep compresses Grep tool output. Content-mode output (file:line:text)
// is regrouped so each file path is printed once, with its matching lines
// capped per file. files_with_matches output (bare paths) is delegated to the
// same directory-tree compressor Glob uses.
func HandleGrep(r io.Reader, w io.Writer) error {
	start := time.Now()
	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}
	if inp.ToolResponse == nil {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	content := inp.ToolResponse.Content
	if len(content) <= 256 {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	summary, action := buildGrepSummary(content)
	// Never-worse: only replace when the summary is actually shorter.
	if summary == "" || len(summary) >= len(content) {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	_ = analytics.Record(analytics.Event{
		TS:       time.Now().UnixNano(),
		SID:      inp.SessionID,
		Hook:     "grep",
		Action:   action,
		BytesIn:  len(content),
		BytesOut: len(summary),
		DurNS:    time.Since(start).Nanoseconds(),
	})
	return protocol.EncodeOutput(w, protocol.Replace(summary))
}

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
	for k := 0; k < len(num); k++ {
		if num[k] < '0' || num[k] > '9' {
			return "", "", "", false
		}
	}
	return s[:i], num, rest[j+1:], true
}

// buildGrepSummary returns (summary, action). action is "grouped" for content
// mode, "tree" when delegated to the file-tree compressor, or "" (empty
// summary) when the input doesn't look like grep output.
func buildGrepSummary(content string) (string, string) {
	lines := strings.Split(strings.TrimSpace(content), "\n")

	groups := make(map[string][]grepMatch)
	var order []string
	parsed := 0
	for _, ln := range lines {
		file, num, text, ok := parseGrepLine(ln)
		if !ok {
			continue
		}
		parsed++
		if _, seen := groups[file]; !seen {
			order = append(order, file)
		}
		groups[file] = append(groups[file], grepMatch{line: num, text: text})
	}

	// If almost nothing parsed as content matches, treat the output as a bare
	// path list (files_with_matches) and reuse the Glob tree compressor.
	if parsed < len(lines)/2 {
		if tree := buildGlobTree(content); tree != "" {
			return tree, "tree"
		}
		return "", ""
	}

	sort.Strings(order)
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
