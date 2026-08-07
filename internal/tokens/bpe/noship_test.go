package bpe_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The shipped binaries must not carry the multi-megabyte vocabulary. Anything
// reachable from cmd/qdf-hook is shipped; this package must not be.
func TestNotReachableFromShippedBinaries(t *testing.T) {
	const self = "github.com/alex60217101990/terse/internal/tokens/bpe"
	goBin := goTool(t)

	// Import paths, not "./cmd/...": the test runs with its own package
	// directory as the working directory, where a relative path to cmd/ does
	// not resolve.
	for _, target := range []string{"github.com/alex60217101990/terse/cmd/qdf-hook"} {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(goBin, "list", "-deps", target)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("go list -deps %s: %v\n%s", target, err, stderr.String())
		}
		for _, dep := range strings.Fields(stdout.String()) {
			if dep == self {
				t.Fatalf("%s links %s; the BPE vocabulary must never ship", target, self)
			}
		}
	}
}

// goTool locates the go command. It is normally on PATH, but a toolchain
// installed outside PATH (gvm, a CI cache) still exposes GOROOT, and this
// guard is too important to skip just because the binary moved.
func goTool(t *testing.T) string {
	t.Helper()
	if p, err := exec.LookPath("go"); err == nil {
		return p
	}
	if root := os.Getenv("GOROOT"); root != "" {
		p := filepath.Join(root, "bin", "go")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	t.Fatal("cannot locate the go command: not on PATH and not under $GOROOT/bin")
	return ""
}
