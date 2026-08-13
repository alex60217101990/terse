package tokens_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// corpusRoot holds committed, invented samples that stand in for real tool
// output. Nothing here is generated: the numbers a benchmark or a gate produces
// are only comparable across commits if the input is byte-identical.
const corpusRoot = "testdata/corpus"

// corpusFingerprint is SHA-256 over the sorted (filename, content) pairs of the
// corpus, recorded here so any change to the samples has to be deliberate.
//
// The trap this closes: `make fmt` runs `gofumpt -w .` over the whole tree, and
// two samples used to be named *.go. gofumpt rewrote them — it had already done
// so in a working tree before this test existed — which silently changes every
// token count derived from the corpus while leaving the tests green. The
// samples are now *.go.txt so no Go tool claims them, and this fingerprint
// catches the next mechanism (an editor, a line-ending normalisation, a
// well-meaning cleanup) rather than trusting that renaming was enough.
//
// If you intend to change the corpus: run this test, take the "got" value, and
// paste it here in the same commit as the sample change.
const corpusFingerprint = "d59582e222a9a867cd4dd0ca5e705e884adee8e154c75354cfdb0e2bf9b5c088"

func TestCorpusFingerprint(t *testing.T) {
	names, contents := loadCorpus(t)

	h := sha256.New()
	for i, name := range names {
		// Length-prefix-free but unambiguous: names cannot contain a NUL, so a
		// rename can never be disguised as a content change or vice versa.
		h.Write([]byte(name))
		h.Write([]byte{0})
		h.Write(contents[i])
		h.Write([]byte{0})
	}
	got := hex.EncodeToString(h.Sum(nil))
	if got != corpusFingerprint {
		t.Errorf("corpus fingerprint changed:\n got  %s\n want %s\n"+
			"A sample was added, removed, or rewritten. If that was deliberate, "+
			"update corpusFingerprint in the same commit; if not, something "+
			"reformatted the corpus behind your back.", got, corpusFingerprint)
	}
}

// No sample may carry a Go file extension: `gofumpt -w .` would rewrite it.
func TestCorpusHasNoGoFiles(t *testing.T) {
	names, _ := loadCorpus(t)
	for _, name := range names {
		if filepath.Ext(name) == ".go" {
			t.Errorf("%s ends in .go; Go tooling will rewrite it — use .go.txt", name)
		}
	}
}

// loadCorpus returns the sample names and contents, sorted by name so both the
// fingerprint and any benchmark built on it are order-independent.
func loadCorpus(tb testing.TB) (names []string, contents [][]byte) {
	tb.Helper()
	entries, err := os.ReadDir(corpusRoot)
	if err != nil {
		tb.Fatalf("read corpus dir: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		tb.Fatalf("corpus dir %s is empty", corpusRoot)
	}
	sort.Strings(names)
	for _, name := range names {
		b, err := os.ReadFile(filepath.Join(corpusRoot, name))
		if err != nil {
			tb.Fatalf("read %s: %v", name, err)
		}
		contents = append(contents, b)
	}
	return names, contents
}
