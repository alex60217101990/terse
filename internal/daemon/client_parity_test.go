package daemon

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// lookForCompiler returns the argv prefix for the first available C
// compiler among "zig cc" and "cc" (in that order), skipping the test if
// neither is on PATH.
func lookForCompiler(t *testing.T) []string {
	t.Helper()
	if _, err := exec.LookPath("zig"); err == nil {
		return []string{"zig", "cc"}
	}
	if _, err := exec.LookPath("cc"); err == nil {
		return []string{"cc"}
	}
	t.Skip("neither zig nor cc found on PATH; skipping native client parity test")
	return nil
}

// buildQC compiles ../../client/qc.c into a temp binary using cc (the argv
// prefix returned by lookForCompiler) and returns the binary's path.
func buildQC(t *testing.T, cc []string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "qdf-hookc")
	args := append(append([]string{}, cc[1:]...), "-O2", "-o", bin, "../../client/qc.c")
	cmd := exec.Command(cc[0], args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compile client/qc.c with %v: %v\n%s", cc, err, out)
	}
	return bin
}

// runProcess runs bin with args, feeding stdin, and returns stdout. It fails
// the test on a non-zero exit.
func runProcess(t *testing.T, bin string, args []string, stdin []byte) []byte {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Stdin = bytes.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("run %s %v: %v\nstderr: %s", bin, args, err, stderr.String())
	}
	return stdout.Bytes()
}

// TestNativeClient_ParityWithGoClient proves qdf-hookc's reply is
// byte-identical to DialAndProxy's for the same payload against the daemon.
//
// It uses two independently started daemons (each with its own fresh store),
// for the same reason TestDialAndProxy_MatchesDaemonReply does: the hook
// pipeline is stateful (Read's mtime cache, generic dedup) keyed on content
// across calls to the same store, so sending the identical payload twice to
// one daemon would legitimately get a different second reply (a
// dedup/"unchanged" marker) — that's correct dispatch behavior, not a client
// bug, and would make this test flaky for the wrong reason. Two fresh stores
// each see the payload as a first-time request, so any byte difference
// between the two replies can only come from the client path itself.
func TestNativeClient_ParityWithGoClient(t *testing.T) {
	cc := lookForCompiler(t)
	bin := buildQC(t, cc)

	sockNative := startTestDaemon(t)
	sockGo := startTestDaemon(t)
	payload := readFixture(t, "post")

	native := runProcess(t, bin, []string{sockNative}, payload)

	var goReply bytes.Buffer
	if err := DialAndProxy(sockGo, bytes.NewReader(payload), &goReply, 2*time.Second); err != nil {
		t.Fatalf("DialAndProxy: %v", err)
	}

	if !bytes.Equal(native, goReply.Bytes()) {
		t.Errorf("native client reply != Go client reply\nnative: %q\ngo: %q", native, goReply.Bytes())
	}
}
