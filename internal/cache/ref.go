package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	qdf "github.com/alex60217101990/qdf"
)

// ShortHex returns the lowercase hex encoding of b using a stack-resident
// scratch buffer. Unlike fmt.Sprintf("%x", b) it boxes no interface argument
// and lets the caller's backing array (e.g. a [32]byte hash) stay on the stack;
// only the returned string is heap-allocated. b must be at most 32 bytes.
func ShortHex(b []byte) string {
	var buf [64]byte // 32 bytes -> 64 hex chars, the max we ever encode
	n := hex.Encode(buf[:], b)
	return string(buf[:n])
}

// refEntry is the on-disk payload of a content-addressed blob, serialized with
// qdf OptBalanced (the repetitive-payload default: ~38x faster decode than JSON
// at equal wire, and the payload is exactly command/log text).
type refEntry struct {
	Content string
	TS      int64 // unix seconds, for gc age
}

// RefsDir returns (and creates) the content-addressed blob directory.
func RefsDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".qdf-hook", "refs")
	_ = os.MkdirAll(dir, 0o700)
	return dir
}

// refHash returns sha256(content)[:16] as 32 hex chars. The input is hashed via
// a zero-copy view (no string->[]byte allocation); sha256 only reads it.
func refHash(content string) string {
	b := unsafe.Slice(unsafe.StringData(content), len(content))
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:16])
}

// RefPath is the blob file path for a hash.
func RefPath(hash string) string { return filepath.Join(RefsDir(), hash+".blob") }

// RefSeen reports whether a blob for hash already exists (one stat).
func RefSeen(hash string) bool {
	_, err := os.Stat(RefPath(hash))
	return err == nil
}

// RefPut writes content under hash. Plain write (no tmp+rename): the blob store
// is a rebuildable cache, so a torn write just fails to decode on read.
func RefPut(hash, content string) {
	e := refEntry{Content: content, TS: time.Now().Unix()}
	data, err := qdf.Marshal(&e, qdf.OptBalanced)
	if err != nil {
		return
	}
	_ = os.WriteFile(RefPath(hash), data, 0o600)
}

// RefGet returns the content stored under hash. Decodes zero-copy (the returned
// string aliases the freshly read buffer, kept alive by the string header).
func RefGet(hash string) (string, bool) {
	data, err := os.ReadFile(RefPath(hash))
	if err != nil {
		return "", false
	}
	var e refEntry
	if err := qdf.Unmarshal(data, &e, qdf.WithNoCopy()); err != nil {
		return "", false
	}
	return e.Content, true
}

// Dedup replaces content that was already emitted (this or an earlier session,
// byte-identical) with a compact §ref token; otherwise it registers the content
// and returns ("", false) so the caller emits it in full this first time.
// minSize gates tiny outputs where a ~60-byte token would not pay off.
func Dedup(content string, minSize int) (token string, deduped bool) {
	if len(content) < minSize {
		return "", false
	}
	hash := refHash(content)
	if RefSeen(hash) {
		return fmt.Sprintf("§ref:%s§ (%d bytes, identical to earlier output — qdf-hook expand %s)",
			hash, len(content), hash), true
	}
	RefPut(hash, content)
	return "", false
}
