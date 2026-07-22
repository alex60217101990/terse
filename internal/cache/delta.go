package cache

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"
)

// IsBinaryContent returns true if data is not safe to display as text.
// Checks for null bytes (binary marker) and invalid UTF-8.
func IsBinaryContent(data []byte) bool {
	if bytes.IndexByte(data, 0) >= 0 {
		return true
	}
	return !utf8.Valid(data)
}

var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

// UnifiedDiff computes a unified diff between old and new.
// contextLines is the number of unchanged lines to show around each hunk.
// Returns "" if old == new.
func UnifiedDiff(old, newer []byte, contextLines int) string {
	if bytes.Equal(old, newer) {
		return ""
	}
	oldLines := splitLines(old)
	newLines := splitLines(newer)

	// Compute LCS edit script using Myers' algorithm (O((N+M)D)).
	edits := myersDiff(oldLines, newLines)
	if len(edits) == 0 {
		return ""
	}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	writeHunks(buf, oldLines, newLines, edits, contextLines)

	return buf.String()
}

// splitLines splits b into lines preserving content.
// The last line may or may not end with \n.
func splitLines(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	lines := strings.Split(string(b), "\n")
	// strings.Split adds an empty string after a trailing newline; remove it.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// edit represents a single edit operation in the diff.
type edit struct {
	oldIdx int  // line index in old (-1 if insertion)
	newIdx int  // line index in new (-1 if deletion)
	kind   byte // '=' equal, '-' delete, '+' insert
}

// myersDiff returns the minimal edit script between a and b.
// Uses O((N+M)D) Myers algorithm with full trace backtracking.
func myersDiff(a, b []string) []edit {
	n, m := len(a), len(b)
	if n == 0 && m == 0 {
		return nil
	}

	// v holds the furthest-reaching x for diagonal k=x-y,
	// stored at index k+offset to avoid negative indexing.
	offset := n + m + 1
	v := make([]int, 2*offset+1)

	// trace stores the v array at each step d for backtracking.
	trace := make([][]int, 0, n+m)

	for d := 0; d <= n+m; d++ {
		snap := make([]int, 2*offset+1)
		copy(snap, v)
		trace = append(trace, snap)

		for k := -d; k <= d; k += 2 {
			var x int
			kIdx := k + offset
			if k == -d || (k != d && v[kIdx-1] < v[kIdx+1]) {
				x = v[kIdx+1] // move down (insert from b)
			} else {
				x = v[kIdx-1] + 1 // move right (delete from a)
			}
			y := x - k
			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}
			v[kIdx] = x
			if x >= n && y >= m {
				return backtrack(a, b, trace, d, offset)
			}
		}
	}
	return nil
}

func backtrack(a, b []string, trace [][]int, d, offset int) []edit {
	x, y := len(a), len(b)
	var edits []edit

	for i := d; i > 0; i-- {
		v := trace[i]
		k := x - y
		kIdx := k + offset

		var prevK int
		if k == -i || (k != i && v[kIdx-1] < v[kIdx+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}
		prevX := v[prevK+offset]
		prevY := prevX - prevK

		// Add equal edits for the snake portion of this step.
		for x > prevX && y > prevY {
			x--
			y--
			edits = append(edits, edit{oldIdx: x, newIdx: y, kind: '='})
		}
		// Add the actual edit (delete or insert).
		if x > prevX {
			x--
			edits = append(edits, edit{oldIdx: x, newIdx: -1, kind: '-'})
		} else {
			y--
			edits = append(edits, edit{oldIdx: -1, newIdx: y, kind: '+'})
		}
	}
	// Remaining equal prefix.
	for x > 0 && y > 0 {
		x--
		y--
		edits = append(edits, edit{oldIdx: x, newIdx: y, kind: '='})
	}
	// Reverse (we built it backwards).
	for i, j := 0, len(edits)-1; i < j; i, j = i+1, j-1 {
		edits[i], edits[j] = edits[j], edits[i]
	}
	return edits
}

// writeHunks writes unified diff hunks to buf.
func writeHunks(buf *bytes.Buffer, a, b []string, edits []edit, ctx int) {
	n := len(edits)

	// Find hunk ranges (indices into edits slice).
	type hunkRange struct{ lo, hi int }
	var hunks []hunkRange

	i := 0
	for i < n {
		if edits[i].kind == '=' {
			i++
			continue
		}
		lo := i - ctx
		if lo < 0 {
			lo = 0
		}
		j := i
		for j < n {
			if edits[j].kind != '=' {
				j++
				continue
			}
			k := j
			for k < n && edits[k].kind == '=' {
				k++
			}
			if k-j <= ctx && k < n {
				j = k
			} else {
				break
			}
		}
		hi := j + ctx
		if hi > n {
			hi = n
		}
		hunks = append(hunks, hunkRange{lo, hi})
		i = hi
	}

	// Walk the full edit list with running counters to compute hunk positions.
	oldLine := 0 // 0-based position in old file
	newLine := 0 // 0-based position in new file
	editIdx := 0

	for _, h := range hunks {
		// Advance counters to hunk start.
		for editIdx < h.lo {
			e := edits[editIdx]
			switch e.kind {
			case '=':
				oldLine++
				newLine++
			case '-':
				oldLine++
			case '+':
				newLine++
			}
			editIdx++
		}

		// Count lines in this hunk.
		oldCount, newCount := 0, 0
		for _, e := range edits[h.lo:h.hi] {
			switch e.kind {
			case '=':
				oldCount++
				newCount++
			case '-':
				oldCount++
			case '+':
				newCount++
			}
		}

		// Write @@ header using 1-based line numbers.
		// For zero-count sides: position is "after line N" (use N, not N+1).
		oldStart := oldLine
		newStart := newLine
		if oldCount > 0 {
			oldStart = oldLine + 1
		}
		if newCount > 0 {
			newStart = newLine + 1
		}
		fmt.Fprintf(buf, "@@ -%d,%d +%d,%d @@\n", oldStart, oldCount, newStart, newCount)

		// Write hunk lines and advance counters.
		for _, e := range edits[h.lo:h.hi] {
			switch e.kind {
			case '=':
				fmt.Fprintf(buf, " %s\n", a[e.oldIdx])
				oldLine++
				newLine++
			case '-':
				fmt.Fprintf(buf, "-%s\n", a[e.oldIdx])
				oldLine++
			case '+':
				fmt.Fprintf(buf, "+%s\n", b[e.newIdx])
				newLine++
			}
		}
		editIdx = h.hi
	}
}
