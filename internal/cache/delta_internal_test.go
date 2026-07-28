package cache

import (
	"strconv"
	"strings"
	"testing"
)

// TestPutDiffScratch_ClearsPooledLineTails is a pinning regression test.
//
// putDiffScratch used to reset aLines/bLines via `s.aLines[:0]` only,
// leaving the string headers in [len:cap] untouched. Each of those headers
// is a zero-copy unsafe.String view over the *entire* input byte slice
// passed to the diff that produced it (see splitLinesInto), so a stale tail
// keeps that whole input buffer reachable from the free list until a later,
// at-least-as-large diff overwrites every slot — pinning megabytes of dead
// content behind a warm scratch buffer.
//
// This test drains the free list, runs one big diff (so aLines/bLines grow
// to a sizeable cap while aliasing a large input buffer) followed by one
// tiny diff (which only overwrites the first slot), then pulls the scratch
// back off the free list directly and asserts every element of
// aLines[:cap] / bLines[:cap] is the zero string. It is deterministic: no
// GC or object-identity games required, since a stale header presents
// simply as a non-empty string in a slot beyond the live length.
func TestPutDiffScratch_ClearsPooledLineTails(t *testing.T) {
	drainDiffScratchFreeList(t)

	const numLines = 200
	var oldBuf, newBuf strings.Builder
	for i := range numLines {
		oldBuf.WriteString("line-")
		oldBuf.WriteString(strconv.Itoa(i))
		oldBuf.WriteString("-old-content-padding-xxxxxxxxxxxxxxxxxxxx\n")
		if i == numLines/2 {
			newBuf.WriteString("line-")
			newBuf.WriteString(strconv.Itoa(i))
			newBuf.WriteString("-CHANGED\n")
		} else {
			newBuf.WriteString("line-")
			newBuf.WriteString(strconv.Itoa(i))
			newBuf.WriteString("-old-content-padding-xxxxxxxxxxxxxxxxxxxx\n")
		}
	}
	oldBig := []byte(oldBuf.String())
	newBig := []byte(newBuf.String())
	if d := UnifiedDiff(oldBig, newBig, 1); d == "" {
		t.Fatal("expected non-empty diff for big input")
	}

	// Small diff reuses the now-warm scratch: it only overwrites the first
	// slot of aLines/bLines, leaving the rest to reveal any stale tail from
	// the big diff above.
	oldSmall := []byte("a\n")
	newSmall := []byte("b\n")
	if d := UnifiedDiff(oldSmall, newSmall, 1); d == "" {
		t.Fatal("expected non-empty diff for small input")
	}

	select {
	case s := <-diffScratchFreeList:
		for i, str := range s.aLines[:cap(s.aLines)] {
			if str != "" {
				t.Fatalf("aLines tail not cleared at index %d: %q (len=%d cap=%d) - stale string pins a prior diff's input buffer", i, str, len(s.aLines), cap(s.aLines))
			}
		}
		for i, str := range s.bLines[:cap(s.bLines)] {
			if str != "" {
				t.Fatalf("bLines tail not cleared at index %d: %q (len=%d cap=%d) - stale string pins a prior diff's input buffer", i, str, len(s.bLines), cap(s.bLines))
			}
		}
	default:
		t.Fatal("expected a warm scratch buffer on the free list after two UnifiedDiff calls")
	}
}

// drainDiffScratchFreeList empties the package-level free list so the test
// starts from a known-empty state regardless of what ran before it.
func drainDiffScratchFreeList(t *testing.T) {
	t.Helper()
	for {
		select {
		case <-diffScratchFreeList:
		default:
			return
		}
	}
}
