package detect

import (
	"fmt"
	"strings"
)

const (
	tableMinRows  = 6 // minimum data rows before summarizing pays
	tableHeadKeep = 5 // first rows kept verbatim
	tableTailKeep = 2 // last rows kept verbatim
	tableMaxLines = 100000
)

// SummarizeTable compresses tabular output (fixed-width `docker ps`-style or
// single-char-delimited CSV/TSV): header + kept head/tail rows + a row count.
// Returns "" unless the input parses as a table AND the summary is strictly
// smaller. Lossy (middle rows elided) — the caller adds a recovery footer.
func SummarizeTable(content string) string {
	// cheap guard: need at least tableMinRows+1 newlines and either a 2+space
	// run or a delimiter in the first line.
	first, rest, ok := strings.Cut(content, "\n")
	if !ok || len(first) == 0 || len(first) > 4096 {
		return ""
	}
	comma := strings.Count(first, ",")
	tab := strings.Count(first, "\t")
	gap := strings.Contains(first, "  ")
	if comma < 2 && tab < 2 && !gap {
		return ""
	}
	// count lines without allocating a slice
	n := strings.Count(rest, "\n")
	if len(rest) > 0 && !strings.HasSuffix(rest, "\n") {
		n++
	}
	if n < tableMinRows {
		return ""
	}
	switch {
	case comma >= 2:
		return summarizeDelimited(content, first, ',', comma+1, n)
	case tab >= 2:
		return summarizeDelimited(content, first, '\t', tab+1, n)
	default:
		return summarizeFixedWidth(content, first, n)
	}
}

// countByte returns the number of occurrences of c in s without the small
// allocation strings.Count(s, string(c)) can incur.
func countByte(s string, c byte) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			n++
		}
	}
	return n
}

// buildTableSummary renders the common "header + head rows + elision + tail
// rows" shape shared by the delimited and fixed-width paths.
func buildTableSummary(header string, rows, cols int, head, tail []string) string {
	var b strings.Builder
	b.Grow(len(header) + 32 + tableHeadKeep*32 + tableTailKeep*32)
	fmt.Fprintf(&b, "[TABLE %d rows × %d cols]\n", rows, cols)
	b.WriteString(header)
	b.WriteByte('\n')
	for _, l := range head {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "… +%d rows\n", rows-(tableHeadKeep+tableTailKeep))
	for _, l := range tail {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	return b.String()
}

// summarizeDelimited handles single-char-delimited tables (CSV/TSV): every
// data line must have exactly cols-1 occurrences of sep. Any inconsistent
// line (a real one, not the empty artifact of a trailing newline) aborts the
// match — a malformed row means this probably isn't a clean table.
func summarizeDelimited(content, header string, sep byte, cols, rows int) string {
	if rows > tableMaxLines || rows <= tableHeadKeep+tableTailKeep {
		return ""
	}
	_, rest, _ := strings.Cut(content, "\n")

	var head [tableHeadKeep]string
	var tail [tableTailKeep]string
	tailStart := rows - tableTailKeep

	idx := 0
	for {
		nl := strings.IndexByte(rest, '\n')
		var line string
		if nl < 0 {
			if rest == "" {
				break
			}
			line = rest
			rest = ""
		} else {
			line = rest[:nl]
			rest = rest[nl+1:]
		}

		if countByte(line, sep) != cols-1 {
			return ""
		}

		switch {
		case idx < tableHeadKeep:
			head[idx] = line
		case idx >= tailStart:
			tail[idx-tailStart] = line
		}
		idx++
		if nl < 0 {
			break
		}
	}
	if idx != rows {
		return ""
	}

	out := buildTableSummary(header, rows, cols, head[:], tail[:])
	if len(out) >= len(content) {
		return ""
	}
	return out
}

// headerBoundaries returns, for a fixed-width table header, the byte offsets
// where a run of 2+ spaces ends and a column label begins. len(bounds)+1 is
// the column count; each bounds[i]-1 is where a well-aligned data line must
// also hold a space.
func headerBoundaries(header string) []int {
	var bounds []int
	i := 0
	for i < len(header) {
		if header[i] != ' ' {
			i++
			continue
		}
		start := i
		for i < len(header) && header[i] == ' ' {
			i++
		}
		if i-start >= 2 && i < len(header) {
			bounds = append(bounds, i)
		}
	}
	return bounds
}

// summarizeFixedWidth handles space-aligned tables (`docker ps`, `kubectl
// get`, `ls -l`, ...). Column boundaries are derived from the header's 2+
// space runs; a data line "fits" if it holds a space one byte before each
// boundary. Individual misaligned lines are tolerated (short values, ragged
// trailing columns) as long as >=80% of rows fit — this is not stripped
// down to per-line strictness like the delimited path because whitespace
// alignment is inherently looser than an explicit separator.
func summarizeFixedWidth(content, header string, rows int) string {
	if rows > tableMaxLines || rows <= tableHeadKeep+tableTailKeep {
		return ""
	}
	bounds := headerBoundaries(header)
	if len(bounds) < 2 {
		return ""
	}
	cols := len(bounds) + 1

	_, rest, _ := strings.Cut(content, "\n")

	var head [tableHeadKeep]string
	var tail [tableTailKeep]string
	tailStart := rows - tableTailKeep

	hits, total, idx := 0, 0, 0
	for {
		nl := strings.IndexByte(rest, '\n')
		var line string
		if nl < 0 {
			if rest == "" {
				break
			}
			line = rest
			rest = ""
		} else {
			line = rest[:nl]
			rest = rest[nl+1:]
		}

		total++
		aligned := true
		for _, b := range bounds {
			if b-1 >= len(line) || line[b-1] != ' ' {
				aligned = false
				break
			}
		}
		if aligned {
			hits++
		}

		switch {
		case idx < tableHeadKeep:
			head[idx] = line
		case idx >= tailStart:
			tail[idx-tailStart] = line
		}
		idx++
		if nl < 0 {
			break
		}
	}
	if idx != rows || total == 0 || hits*5 < total*4 { // hits/total < 0.8
		return ""
	}

	out := buildTableSummary(header, rows, cols, head[:], tail[:])
	if len(out) >= len(content) {
		return ""
	}
	return out
}
