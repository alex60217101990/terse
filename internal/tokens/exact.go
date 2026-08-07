// Package tokens counts how many tokens a string costs.
//
// qdf-hook exists to reduce the tokens a tool result consumes, so every
// never-worse gate has to be expressed in tokens. Bytes are a poor proxy: hex
// costs about 1.1 bytes per token while prose costs about 4.8, so two outputs
// of equal size can differ fourfold in what is actually paid. Byte-equal
// changes are not token-equal either, which is how a marker change worth 18% of
// the tokens can look like 5% in bytes and never get made.
//
// Count is exact, not an estimate. A character-class approximator was built and
// measured first; against the exact encoder its decision agreement was 0.92 at
// a 20% token margin, which is precisely the band the worth() gates operate in
// (ratios of 0.5 and 0.75). It was deleted rather than kept as a second source
// of truth: there is one counter, and it is the real one.
//
// The vocabulary is embedded gzipped and decompressed on FIRST USE only.
// Importing this package costs nothing; lazy_test.go pins that. It matters
// because the CLI fallback path spawns a process per invocation, and every
// passthrough returns before a gate ever asks for a count.
//
// The reference encoding is o200k, which is a public proxy for Claude's
// tokenizer rather than the real thing. Counts are for comparing two candidate
// outputs, not for estimating production cost.
package tokens

import (
	"bufio"
	"compress/gzip"
	"embed"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/pkoukk/tiktoken-go"
)

// vocabPath is the vendored copy of the o200k_base rank table. It is production
// data, not a fixture, so it lives in its own directory rather than testdata.
const vocabPath = "vocab/o200k_base.tiktoken.gz"

//go:embed vocab/o200k_base.tiktoken.gz
var vocabFS embed.FS

var (
	once sync.Once
	enc  *tiktoken.Tiktoken
)

// Count returns the exact number of o200k tokens in s.
//
// The first call decompresses and indexes the embedded vocabulary, which costs
// a few megabytes and a few tens of milliseconds; every later call is encoding
// only. Callers on a hot path should short-circuit obvious cases (an empty or
// tiny payload, a passthrough) before reaching here.
//
// It panics if the vendored vocabulary cannot be loaded. The vocabulary is
// compiled into the binary, so a failure means the binary itself is corrupt,
// and a silently wrong token count is far worse than a crash: it would flip
// the gates that decide whether compression is applied at all.
func Count(s string) int {
	// An empty payload is answered without touching the vocabulary, so a caller
	// that only ever counts nothing never pays for the load.
	if s == "" {
		return 0
	}
	once.Do(load)
	// nil/nil: no token is treated as special, so a payload that happens to
	// contain "<|endoftext|>" is counted as the ordinary text it is rather
	// than panicking on a disallowed special token.
	return len(enc.Encode(s, nil, nil))
}

func load() {
	// tiktoken-go resolves vocabularies through a package-global loader that
	// fetches over HTTP by default. Replacing it with the offline one keeps
	// every call in this process network-free, which is the point: neither CI
	// nor a user's machine may depend on egress to count a token.
	tiktoken.SetBpeLoader(offlineLoader{})

	var err error
	enc, err = tiktoken.GetEncoding(tiktoken.MODEL_O200K_BASE)
	if err != nil {
		panic("tokens: build encoding: " + err.Error())
	}
}

// offlineLoader implements tiktoken.BpeLoader against the embedded vocabulary.
//
// tiktoken-go hands the loader the upstream blob URL it would otherwise fetch.
// We ignore it as a source but check it, so that asking for an encoding this
// package did not vendor fails loudly instead of being served o200k ranks.
type offlineLoader struct{}

func (offlineLoader) LoadTiktokenBpe(tiktokenBpeFile string) (map[string]int, error) {
	if !strings.Contains(tiktokenBpeFile, "o200k_base") {
		return nil, fmt.Errorf("tokens: only o200k_base is vendored, cannot serve %q", tiktokenBpeFile)
	}

	f, err := vocabFS.Open(vocabPath)
	if err != nil {
		return nil, fmt.Errorf("tokens: open vendored vocabulary: %w", err)
	}
	defer func() { _ = f.Close() }()

	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("tokens: gunzip vocabulary: %w", err)
	}
	defer func() { _ = zr.Close() }()

	// o200k_base is one "<base64-token> <rank>" pair per line.
	ranks := make(map[string]int, 200000)
	sc := bufio.NewScanner(zr)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		tok, rankStr, ok := strings.Cut(line, " ")
		if !ok {
			return nil, fmt.Errorf("tokens: malformed vocabulary line %q", line)
		}
		raw, err := base64.StdEncoding.DecodeString(tok)
		if err != nil {
			return nil, fmt.Errorf("tokens: decode token %q: %w", tok, err)
		}
		rank, err := strconv.Atoi(rankStr)
		if err != nil {
			return nil, fmt.Errorf("tokens: parse rank in %q: %w", line, err)
		}
		ranks[string(raw)] = rank
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("tokens: read vocabulary: %w", err)
	}
	if len(ranks) == 0 {
		return nil, fmt.Errorf("tokens: vendored vocabulary %s is empty", vocabPath)
	}
	return ranks, nil
}
