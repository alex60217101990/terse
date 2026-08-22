package main

import (
	"os"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/cache"
)

// TestRunExpand_ResolvesCapture is the other half of output capping: the
// wrapper prints "qdf-hook expand <id>" next to the elided bytes, so that
// command has to find the capture. It used to look in the ref store only,
// which meant every capped command advertised a handle that did not resolve.
func TestRunExpand_ResolvesCapture(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	const id = "0123456789abcdef0123456789abcdef"
	const body = "the full output that was elided\nsecond line\n"
	if err := os.MkdirAll(cache.CaptureDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache.CapturePath(id), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(t, func() {
		if err := runExpand(id); err != nil {
			t.Errorf("runExpand: %v", err)
		}
	})
	if got != body {
		t.Errorf("expand printed %q, want %q", got, body)
	}
}

// TestRunExpand_UnknownID keeps the failure legible: neither store has it.
func TestRunExpand_UnknownID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	err := runExpand("deadbeefdeadbeefdeadbeefdeadbeef")
	if err == nil {
		t.Fatal("expected an error for an unknown id")
	}
	if !strings.Contains(err.Error(), "no capture") {
		t.Errorf("error should name both stores, got: %v", err)
	}
}

// TestRunExpand_StillResolvesRefs keeps the original path working.
func TestRunExpand_StillResolvesRefs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	const body = "content behind a ref"
	hash := cache.RefHashOf(body)
	cache.RefPut(hash, body)
	got := captureStdout(t, func() {
		if err := runExpand("§ref:" + hash + "§"); err != nil {
			t.Errorf("runExpand: %v", err)
		}
	})
	if got != body {
		t.Errorf("expand printed %q, want %q", got, body)
	}
}

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what
// it wrote. runExpand prints straight to stdout by design (it is a CLI).
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		var sb strings.Builder
		buf := make([]byte, 4096)
		for {
			n, rerr := r.Read(buf)
			sb.Write(buf[:n])
			if rerr != nil {
				break
			}
		}
		done <- sb.String()
	}()
	fn()
	os.Stdout = orig
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	out := <-done
	if err := r.Close(); err != nil {
		t.Fatalf("close pipe reader: %v", err)
	}
	return out
}
