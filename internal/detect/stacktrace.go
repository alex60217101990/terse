package detect

import "strings"

// SummarizeStackTrace keeps the exception/panic message and the top/bottom
// frames of a long stack trace, eliding the middle. "" means either the
// content isn't a recognized trace format, or there aren't enough frames for
// folding to pay off — the caller then leaves the output untouched.
//
// traceHeadFrames/traceTailFrames frames are kept verbatim from the front and
// back of the trace; everything in between collapses into a single
// "⋯ <k> frames" marker line. Folding only happens once the trace holds more
// than traceHeadFrames+traceTailFrames+traceMinFold frames, so short traces
// (the common case for a real, actionable error) pass through unmodified.
const (
	traceHeadFrames = 5
	traceTailFrames = 2
	traceMinFold    = 4 // only fold when frames > headFrames+tailFrames+minFold
	traceMaxLines   = 100000
)

// traceFrame is a half-open [start, end) byte range into the original content
// covering exactly one stack frame (including its trailing newline, if any).
type traceFrame struct{ start, end int }

// SummarizeStackTrace detects python/go/java/node/rust stack traces via a
// cheap 256-byte prefix guard, then folds the middle frames when there are
// enough of them to be worth it.
func SummarizeStackTrace(content string) string {
	head := content
	if len(head) > 256 {
		head = head[:256]
	}
	switch {
	case strings.HasPrefix(head, "Traceback (most recent call last)"):
		return summarizePyTrace(content)
	case strings.HasPrefix(head, "panic: ") || strings.Contains(head, "\ngoroutine "):
		return summarizeGoPanic(content)
	case strings.Contains(head, "\n\tat ") || strings.HasPrefix(head, "\tat ") ||
		strings.Contains(head, "\n    at ") || strings.HasPrefix(head, "    at "):
		return summarizeAtTrace(content)
	case strings.Contains(head, "stack backtrace:"):
		return summarizeRustTrace(content)
	default:
		return ""
	}
}

// lineEnd returns the offset just past the line that starts at pos, including
// its trailing '\n' — or len(content) if that line runs to EOF unterminated.
func lineEnd(content string, pos int) int {
	nl := strings.IndexByte(content[pos:], '\n')
	if nl < 0 {
		return len(content)
	}
	return pos + nl + 1
}

// buildTraceSummary renders the common "header + head frames + fold marker +
// tail frames + epilogue" shape shared by every flavor. Returns "" when there
// aren't enough frames to fold, or when the result isn't strictly smaller
// than content (never-worse gate).
func buildTraceSummary(content string, headerEnd int, frames []traceFrame, epilogueStart int) string {
	total := len(frames)
	if total <= traceHeadFrames+traceTailFrames+traceMinFold {
		return ""
	}

	var b strings.Builder
	b.Grow(headerEnd + (len(content) - epilogueStart) + 64)
	b.WriteString(content[:headerEnd])
	for i := range traceHeadFrames {
		b.WriteString(content[frames[i].start:frames[i].end])
	}
	b.WriteString("⋯ ")
	b.WriteString(itoa(total - traceHeadFrames - traceTailFrames))
	b.WriteString(" frames\n")
	for i := total - traceTailFrames; i < total; i++ {
		b.WriteString(content[frames[i].start:frames[i].end])
	}
	b.WriteString(content[epilogueStart:])

	out := b.String()
	if len(out) >= len(content) {
		return ""
	}
	return out
}

// --- python ---
//
// Traceback (most recent call last):
//   File "...", line N, in name
//     source line
//   ... (repeated)
// ValueError: ...  <- epilogue, kept verbatim

func summarizePyTrace(content string) string {
	if tooManyLines(content) {
		return ""
	}
	headerEnd := lineEnd(content, 0)

	var frames []traceFrame
	pos := headerEnd
	for pos < len(content) {
		if !strings.HasPrefix(content[pos:], `  File "`) {
			break
		}
		fileEnd := lineEnd(content, pos)
		if fileEnd >= len(content) {
			// No following source line — count it as a (short) frame anyway.
			frames = append(frames, traceFrame{pos, fileEnd})
			pos = fileEnd
			break
		}
		frameEnd := lineEnd(content, fileEnd)
		frames = append(frames, traceFrame{pos, frameEnd})
		pos = frameEnd
	}

	return buildTraceSummary(content, headerEnd, frames, pos)
}

// --- go ---
//
// panic: ...
//
// goroutine 1 [running]:
// example.com/pkg.func(0x0)
// 	/src/pkg/file.go:10 +0x0
// ... (repeated)

func summarizeGoPanic(content string) string {
	if tooManyLines(content) {
		return ""
	}
	idx := strings.Index(content, "\ngoroutine ")
	var headerStart int
	switch {
	case idx >= 0:
		headerStart = idx + 1
	case strings.HasPrefix(content, "goroutine "):
		headerStart = 0
	default:
		return "" // panic: prefix only, no goroutine header — nothing to fold
	}
	headerEnd := lineEnd(content, headerStart)

	var frames []traceFrame
	pos := headerEnd
	for pos < len(content) {
		funcEnd := lineEnd(content, pos)
		if funcEnd >= len(content) || !strings.HasPrefix(content[funcEnd:], "\t") {
			break
		}
		locEnd := lineEnd(content, funcEnd)
		frames = append(frames, traceFrame{pos, locEnd})
		pos = locEnd
	}

	return buildTraceSummary(content, headerEnd, frames, pos)
}

// --- java / node ---
//
// Error: message (possibly multiple header/"Caused by:" lines)
// 	at foo (file.js:1:1)          <- or "    at foo(...)"
// ... (repeated)
// (anything after the last "at" line, e.g. a nested "Caused by:", is epilogue)

func isAtFrameLine(line string) bool {
	return strings.HasPrefix(line, "\tat ") || strings.HasPrefix(line, "    at ")
}

func summarizeAtTrace(content string) string {
	if tooManyLines(content) {
		return ""
	}
	// Header is everything up to the first frame line.
	headerEnd := 0
	for headerEnd < len(content) {
		next := lineEnd(content, headerEnd)
		if isAtFrameLine(content[headerEnd:next]) {
			break
		}
		headerEnd = next
	}
	if headerEnd >= len(content) {
		return "" // guard matched somewhere, but no frame line actually found
	}

	var frames []traceFrame
	pos := headerEnd
	for pos < len(content) {
		next := lineEnd(content, pos)
		if !isAtFrameLine(content[pos:next]) {
			break
		}
		frames = append(frames, traceFrame{pos, next})
		pos = next
	}

	return buildTraceSummary(content, headerEnd, frames, pos)
}

// --- rust ---
//
// stack backtrace:
//    0: rust_out::main
//              at ./src/main.rs:2:5
// ... (repeated)
// note: ...  <- epilogue

func isRustNumberedLine(line string) bool {
	i := 0
	for i < len(line) && line[i] == ' ' {
		i++
	}
	j := i
	for j < len(line) && line[j] >= '0' && line[j] <= '9' {
		j++
	}
	return j > i && j < len(line) && line[j] == ':'
}

func summarizeRustTrace(content string) string {
	if tooManyLines(content) {
		return ""
	}
	idx := strings.Index(content, "stack backtrace:")
	if idx < 0 {
		return ""
	}
	headerEnd := lineEnd(content, idx)

	var frames []traceFrame
	pos := headerEnd
	for pos < len(content) {
		numEnd := lineEnd(content, pos)
		numLine := strings.TrimSuffix(content[pos:numEnd], "\n")
		if !isRustNumberedLine(numLine) {
			break
		}
		frameEnd := numEnd
		if numEnd < len(content) {
			locEnd := lineEnd(content, numEnd)
			locLine := strings.TrimLeft(strings.TrimSuffix(content[numEnd:locEnd], "\n"), " ")
			if strings.HasPrefix(locLine, "at ") {
				frameEnd = locEnd
			}
		}
		frames = append(frames, traceFrame{pos, frameEnd})
		pos = frameEnd
	}

	return buildTraceSummary(content, headerEnd, frames, pos)
}

// tooManyLines bounds adversarial input: a trace with more than traceMaxLines
// newlines skips folding rather than risking pathological scan time.
func tooManyLines(content string) bool {
	return strings.Count(content, "\n") > traceMaxLines
}
