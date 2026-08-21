package tokens_test

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/tokens"
)

// The property under test: importing this package costs nothing until Count is
// first called. The o200k vocabulary is ~200k entries; decompressing and
// indexing it costs several megabytes of heap and tens of milliseconds. If any
// of that happened at init, every process that merely links the package would
// pay it.
//
// That is not hypothetical. The CLI fallback path spawns qdf-hook once per tool
// invocation, and the overwhelming majority of those invocations are
// passthroughs that return before a gate ever asks for a token count. Paying
// the vocabulary on a passthrough would be a per-invocation tax for nothing.
//
// The proof is the shape of the allocation curve, not its absolute size: the
// first Count allocates megabytes and the second allocates almost nothing. Work
// done at init would show up as a flat curve — both calls cheap — because the
// cost would already have been paid before the measurement started.
//
// It runs in a child process because the assertion is about the FIRST call in a
// process, and the other tests in this package call Count too.
const lazyChildEnv = "QDF_TOKENS_LAZY_CHILD"

func TestVocabularyLoadsLazily(t *testing.T) {
	if os.Getenv(lazyChildEnv) == "1" {
		lazyChild(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestVocabularyLoadsLazily$", "-test.v")
	cmd.Env = append(os.Environ(), lazyChildEnv+"=1")
	out, err := cmd.CombinedOutput()
	t.Logf("child output:\n%s", out)
	if err != nil {
		t.Fatalf("child process failed: %v", err)
	}
	// A -test.run typo would exit 0 having run nothing, which would make this
	// test vacuous. Require evidence the assertions actually ran.
	if !strings.Contains(string(out), "first Count allocated") {
		t.Fatal("child did not run the laziness assertions")
	}
}

func lazyChild(t *testing.T) {
	t.Helper()
	const (
		// The vocabulary is 200k entries; a map that size cannot fit in less.
		// Set well below the observed ~30 MB so the bound survives a Go map or
		// tiktoken-go implementation change, while still being far above what
		// any pure-encoding call could allocate.
		minFirstCall = 4 << 20
		// A second encode of the same short string touches the vocabulary but
		// does not rebuild it. Generous, because it bounds encoder scratch
		// rather than the vocabulary.
		maxSecondCall = 256 << 10
	)

	var before, afterFirst, afterSecond runtime.MemStats

	// TotalAlloc is cumulative and monotonic, so it measures work done rather
	// than what happens to be live when a GC runs.
	runtime.ReadMemStats(&before)
	if n := tokens.Count("lazily loaded vocabulary"); n == 0 {
		t.Fatal("Count returned 0 for a non-empty string")
	}
	runtime.ReadMemStats(&afterFirst)
	if n := tokens.Count("lazily loaded vocabulary"); n == 0 {
		t.Fatal("Count returned 0 for a non-empty string")
	}
	runtime.ReadMemStats(&afterSecond)

	first := afterFirst.TotalAlloc - before.TotalAlloc
	second := afterSecond.TotalAlloc - afterFirst.TotalAlloc
	t.Logf("first Count allocated %d bytes, second allocated %d bytes", first, second)

	if first < minFirstCall {
		t.Errorf("first Count allocated only %d bytes (want >= %d): the vocabulary "+
			"was already resident, so something loads it before Count is called",
			first, minFirstCall)
	}
	if second > maxSecondCall {
		t.Errorf("second Count allocated %d bytes (want <= %d): the vocabulary is "+
			"being rebuilt per call, not cached", second, maxSecondCall)
	}
}

// Counting an empty string must not load the vocabulary either — it is the
// cheapest possible caller and gets the short-circuit.
func TestCountEmptyDoesNotLoad(t *testing.T) {
	if got := tokens.Count(""); got != 0 {
		t.Errorf("Count(\"\") = %d, want 0", got)
	}
	if n := testing.AllocsPerRun(100, func() { _ = tokens.Count("") }); n != 0 {
		t.Errorf("Count(\"\") allocated %v times per run, want 0", n)
	}
}
