package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

// readOnlyPrefixes are command prefixes considered safe to cache.
// These commands have no side effects and produce deterministic output
// for a given working directory state.
// The cache only ever replaces output when a fresh run is byte-identical to a
// recent one (see HandleBash), and it runs in PostToolUse *after* the command
// has already executed. So a prefix here can never suppress a side effect —
// the worst case for a mis-classified command is a wasted cache write. We
// still restrict the list to genuinely read-only verbs whose output is stable
// enough to repeat within the TTL; non-deterministic commands (date, ps, top)
// simply never hit and are left out to avoid pointless writes.
var readOnlyPrefixes = []string{
	// git — read-only porcelain/plumbing.
	"git status", "git log", "git diff", "git branch", "git show", "git stash list",
	// gh (GitHub CLI) — read-only subcommands only. Mutating verbs (create,
	// merge, close, edit) and `gh api` (which can carry -X POST/PUT/DELETE)
	// are deliberately excluded.
	"gh pr view", "gh pr list", "gh pr diff", "gh pr checks", "gh pr status",
	"gh issue view", "gh issue list", "gh issue status",
	"gh repo view", "gh run list", "gh run view",
	"gh release view", "gh release list", "gh workflow list", "gh workflow view",
	"gh label list", "gh search", "gh status",
	// Filesystem inspection.
	"ls", "find", "cat", "pwd", "which", "echo", "head", "wc", "tree",
	"stat", "file", "du", "df", "realpath", "basename", "dirname",
	// Search — large, frequently-repeated output.
	"grep", "egrep", "fgrep", "rg", "ag",
	// Content hashing — deterministic by definition.
	"sha256sum", "shasum", "md5sum",
	// Go toolchain — read-only queries. `go build`/`go test` are excluded
	// (they produce artifacts / run code and are non-deterministic).
	"go env", "go version", "go list", "go doc", "go vet",
	"go mod why", "go mod graph", "go mod verify",
	// jq — deterministic transform of its input.
	"jq",
	// Container / cluster inspection — read-only queries.
	"docker ps", "docker images", "docker inspect", "docker version",
	"kubectl get", "kubectl describe", "kubectl version",
	// Node package managers — dependency listing.
	"npm ls", "npm list", "npm view", "pnpm list", "yarn list",
	// Environment.
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
		OutputHash: ShortHex(h[:8]),
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
	return ShortHex(h[:8])
}
