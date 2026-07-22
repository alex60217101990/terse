package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// readOnlyPrefixes are command prefixes considered safe to cache.
// These commands have no side effects and produce deterministic output
// for a given working directory state.
var readOnlyPrefixes = []string{
	"git status", "git log", "git diff", "git branch", "git show", "git stash list",
	"ls", "find", "cat", "pwd", "which", "echo",
	"go env", "go version", "go list",
	"printenv", "env",
}

// IsReadOnlyCommand reports whether a command is safe to cache.
func IsReadOnlyCommand(command string) bool {
	trimmed := strings.TrimSpace(command)
	for _, prefix := range readOnlyPrefixes {
		if trimmed == prefix || strings.HasPrefix(trimmed, prefix+" ") {
			return true
		}
	}
	return false
}

// bashCacheDir returns the bash cache directory, creating it if needed.
func bashCacheDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".qdf-hook", "bash-cache")
	_ = os.MkdirAll(dir, 0o700)
	return dir
}

// bashCacheKey computes the cache file path for a command+cwd pair.
// Key: sha256(command + "\x00" + cwd)[:8] as hex.
func bashCacheKey(command, cwd string) string {
	h := sha256.Sum256([]byte(command + "\x00" + cwd))
	return filepath.Join(bashCacheDir(), fmt.Sprintf("%x.entry", h[:8]))
}

// bashCacheEntry is the on-disk JSON format.
type bashCacheEntry struct {
	OutputHash string `json:"hash"`
	Output     string `json:"output"`
	TS         int64  `json:"ts"` // unix seconds
}

// bashCacheTTL returns the cache TTL in seconds (default 30, override via QDF_BASH_CACHE_TTL_SEC).
func bashCacheTTL() int64 {
	if v := os.Getenv("QDF_BASH_CACHE_TTL_SEC"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return 30
}

// BashCacheGet returns cached output for a command+cwd pair, if valid and unexpired.
func BashCacheGet(command, cwd string) (string, bool) {
	path := bashCacheKey(command, cwd)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var entry bashCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return "", false
	}
	if time.Now().Unix()-entry.TS >= bashCacheTTL() {
		return "", false
	}
	return entry.Output, true
}

// BashCacheSet stores output for a command+cwd pair.
func BashCacheSet(command, cwd, output string) {
	h := sha256.Sum256([]byte(output))
	entry := bashCacheEntry{
		OutputHash: fmt.Sprintf("%x", h[:8]),
		Output:     output,
		TS:         time.Now().Unix(),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	path := bashCacheKey(command, cwd)
	_ = os.WriteFile(path, data, 0o600)
}

// BashOutputHash returns the SHA-256 hash prefix of output (for compact summary tokens).
func BashOutputHash(output string) string {
	h := sha256.Sum256([]byte(output))
	return fmt.Sprintf("%x", h[:8])
}
