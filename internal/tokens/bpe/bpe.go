// Package bpe wraps an exact o200k byte-pair encoder for use as ground truth
// when validating the approximate counter in the parent package.
//
// o200k is NOT Claude's tokenizer. It is a public encoding with similar
// behaviour, used here as a stable, reproducible proxy. Numbers derived from
// it are for RELATIVE comparison between two candidate outputs — they are not
// a statement about production token cost.
//
// This package must never be linked into a shipped binary: it carries a
// multi-megabyte vocabulary and is orders of magnitude slower than
// tokens.Count. noship_test.go enforces that.
package bpe

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

// vocabPath is the vendored copy of the o200k_base rank table. It lives under
// testdata so it is obvious the file is fixture data rather than something the
// tool needs at runtime; go:embed reads it regardless of the directory name.
const vocabPath = "testdata/o200k_base.tiktoken.gz"

//go:embed testdata/o200k_base.tiktoken.gz
var vocabFS embed.FS

var (
	once sync.Once
	enc  *tiktoken.Tiktoken
)

// Count returns the exact number of o200k tokens in s.
//
// It panics if the vendored vocabulary cannot be loaded. This is test-only
// code, and a silently wrong ground truth is far worse than a crash.
func Count(s string) int {
	once.Do(load)
	// nil/nil: no token is treated as special, so a payload that happens to
	// contain "<|endoftext|>" is counted as the ordinary text it is rather
	// than panicking on a disallowed special token.
	return len(enc.Encode(s, nil, nil))
}

func load() {
	// tiktoken-go resolves vocabularies through a package-global loader that
	// fetches over HTTP by default. Replacing it with the offline one keeps
	// every call in this process network-free, which is the point: CI runs
	// this with no egress.
	tiktoken.SetBpeLoader(offlineLoader{})

	var err error
	enc, err = tiktoken.GetEncoding(tiktoken.MODEL_O200K_BASE)
	if err != nil {
		panic("bpe: build encoding: " + err.Error())
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
		return nil, fmt.Errorf("bpe: only o200k_base is vendored, cannot serve %q", tiktokenBpeFile)
	}

	f, err := vocabFS.Open(vocabPath)
	if err != nil {
		return nil, fmt.Errorf("bpe: open vendored vocabulary: %w", err)
	}
	defer f.Close()

	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("bpe: gunzip vocabulary: %w", err)
	}
	defer zr.Close()

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
			return nil, fmt.Errorf("bpe: malformed vocabulary line %q", line)
		}
		raw, err := base64.StdEncoding.DecodeString(tok)
		if err != nil {
			return nil, fmt.Errorf("bpe: decode token %q: %w", tok, err)
		}
		rank, err := strconv.Atoi(rankStr)
		if err != nil {
			return nil, fmt.Errorf("bpe: parse rank in %q: %w", line, err)
		}
		ranks[string(raw)] = rank
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("bpe: read vocabulary: %w", err)
	}
	if len(ranks) == 0 {
		return nil, fmt.Errorf("bpe: vendored vocabulary %s is empty", vocabPath)
	}
	return ranks, nil
}
