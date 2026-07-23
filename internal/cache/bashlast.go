package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	qdf "github.com/alex60217101990/qdf"
)

// bashLastEntry is the last output of a command+cwd, for re-run delta encoding.
// Serialized with qdf OptBalanced (fast decode of repetitive text).
type bashLastEntry struct {
	Output string
	TS     int64
}

// BashLastDir returns (and creates) the per-command last-output directory.
func BashLastDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".qdf-hook", "bashlast")
	_ = os.MkdirAll(dir, 0o700)
	return dir
}

// bashLastPath keys the store by sha256(command \x00 cwd)[:16]. The command and
// cwd are hashed via zero-copy views — no concatenation buffer is retained.
func bashLastPath(command, cwd string) string {
	h := sha256.New()
	h.Write(s2b(command))
	h.Write([]byte{0})
	h.Write(s2b(cwd))
	var sum [32]byte
	h.Sum(sum[:0])
	return filepath.Join(BashLastDir(), hex.EncodeToString(sum[:16])+".blob")
}

// s2b returns a zero-copy read-only []byte view of s (never mutated here).
func s2b(s string) []byte {
	if s == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BashLastGet returns the previous output for command+cwd, if any. Decoded
// zero-copy (the returned string aliases the freshly read buffer).
func BashLastGet(command, cwd string) (string, bool) {
	data, err := os.ReadFile(bashLastPath(command, cwd))
	if err != nil {
		return "", false
	}
	var e bashLastEntry
	if err := qdf.Unmarshal(data, &e, qdf.WithNoCopy()); err != nil {
		return "", false
	}
	return e.Output, true
}

// BashLastPut stores the current output for command+cwd (plain write).
func BashLastPut(command, cwd, output string) {
	e := bashLastEntry{Output: output, TS: time.Now().Unix()}
	data, err := qdf.Marshal(&e, qdf.OptBalanced)
	if err != nil {
		return
	}
	_ = os.WriteFile(bashLastPath(command, cwd), data, 0o600)
}
