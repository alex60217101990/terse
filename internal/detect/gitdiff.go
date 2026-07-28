package detect

import "strings"

// gitdiff folding: keep every header/hunk/+/- line verbatim; collapse runs of
// more than gitDiffKeepCtx unchanged context lines into "⋯ <k> unchanged".
const (
	gitDiffKeepCtx  = 6
	gitDiffMaxLines = 200000 // bound adversarial input
)

// IsGitDiffOutput reports whether content looks like unified git-diff output.
// Guard is byte-cheap: looks only at the first 512 bytes.
func IsGitDiffOutput(content string) bool {
	head := content
	if len(head) > 512 {
		head = head[:512]
	}
	if strings.HasPrefix(head, "diff --git ") || strings.Contains(head, "\ndiff --git ") {
		return true
	}
	// Bare `--- a/... / +++ b/... / @@` form (git show, some tools).
	return strings.HasPrefix(head, "--- ") && strings.Contains(head, "\n+++ ") && strings.Contains(head, "\n@@ ")
}

// SummarizeGitDiff collapses long unchanged-context runs. Returns "" when the
// input is not a diff or nothing collapses (caller treats "" as no-win).
func SummarizeGitDiff(content string) string {
	if !IsGitDiffOutput(content) {
		return ""
	}
	var b strings.Builder
	// Output only drops or copies bytes from content, never adds net length, so
	// len(content) is a tight upper bound — one Grow avoids the Builder's
	// realloc chain on large diffs.
	b.Grow(len(content))
	collapsed := false
	lines := 0
	ctxStart := -1 // byte offset where the current context run began
	ctxCount := 0
	pos := 0
	flush := func(end int) {
		if ctxStart < 0 {
			return
		}
		if ctxCount > gitDiffKeepCtx {
			// keep first 2 + last 2 context lines, fold the middle
			kept := 0
			p := ctxStart
			for kept < 2 && p < end {
				nl := strings.IndexByte(content[p:end], '\n')
				if nl < 0 {
					break
				}
				p += nl + 1
				kept++
			}
			// find start of last 2 lines by scanning back from end
			q := end
			for range 2 {
				if q <= ctxStart {
					q = ctxStart
					break
				}
				r := strings.LastIndexByte(content[ctxStart:q-1], '\n')
				if r < 0 {
					q = ctxStart
					break
				}
				q = ctxStart + r + 1
			}
			if q > p {
				b.WriteString(content[ctxStart:p])
				b.WriteString("⋯ ")
				b.WriteString(itoa(ctxCount - 4))
				b.WriteString(" unchanged\n")
				b.WriteString(content[q:end])
				collapsed = true
			} else {
				b.WriteString(content[ctxStart:end])
			}
		} else {
			b.WriteString(content[ctxStart:end])
		}
		ctxStart, ctxCount = -1, 0
	}
	for pos < len(content) && lines < gitDiffMaxLines {
		nl := strings.IndexByte(content[pos:], '\n')
		end := len(content)
		if nl >= 0 {
			end = pos + nl + 1
		}
		line := content[pos:end]
		isCtx := len(line) > 0 && line[0] == ' '
		if isCtx {
			if ctxStart < 0 {
				ctxStart = pos
			}
			ctxCount++
		} else {
			flush(pos)
			b.WriteString(line)
		}
		pos = end
		lines++
	}
	flush(pos)
	if !collapsed {
		return ""
	}
	out := b.String()
	if len(out) >= len(content) {
		return ""
	}
	return out
}
