package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"

	qdf "github.com/alex60217101990/qdf"
	"github.com/alex60217101990/terse/internal/bytesconv"
)

// lastEntry is the previous output of a tool call, for re-run delta encoding.
// Serialized with qdf OptBalanced (fast decode of repetitive text).
type lastEntry struct {
	Output string
	TS     int64
}

// LastOutDir returns the per-tool-call last-output directory (created lazily on write).
func LastOutDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".qdf-hook", "lastout")
}

// LastOutputKey builds a tool-agnostic store key: sha256(tool \x00 input)[:16].
// A re-run of the same tool with identical arguments maps to the same key.
// tool and input are hashed via zero-copy views (never mutated).
func LastOutputKey(tool string, input []byte) string {
	h := sha256.New()
	h.Write(bytesconv.S2B(tool))
	h.Write([]byte{0})
	h.Write(input)
	var sum [32]byte
	h.Sum(sum[:0])
	return hex.EncodeToString(sum[:16])
}

func lastPath(key string) string { return filepath.Join(LastOutDir(), key+".blob") }

// LastOutputGet returns the previous output for a key, if any. Decoded zero-copy
// (the returned string aliases the freshly read buffer).
func LastOutputGet(key string) (string, bool) {
	data, err := os.ReadFile(lastPath(key))
	if err != nil {
		return "", false
	}
	var e lastEntry
	if err := qdf.Unmarshal(data, &e, qdf.WithNoCopy()); err != nil {
		return "", false
	}
	return e.Output, true
}

// LastOutputPut stores the current output for a key (plain write).
func LastOutputPut(key, output string) {
	e := lastEntry{Output: output, TS: time.Now().Unix()}
	_ = marshalBlobPooled(lastPath(key), &e)
}
