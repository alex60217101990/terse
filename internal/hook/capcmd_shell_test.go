package hook

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runWrapped executes a wrapped command through /bin/sh and returns its stdout
// and exit code — the two things the model and the agent actually observe.
func runWrapped(t *testing.T, cmd string, capBytes int) (string, int) {
	t.Helper()
	dir := t.TempDir()
	cap := filepath.Join(dir, "x.out")
	wrapped := string(wrapCommand(nil, cmd, cap, "x", capBytes))
	c := exec.Command("/bin/sh", "-c", wrapped)
	c.Dir = dir
	out, err := c.Output()
	code := 0
	var ee *exec.ExitError
	if err != nil {
		if !errors.As(err, &ee) {
			t.Fatalf("run %q: %v", wrapped, err)
		}
		code = ee.ExitCode()
	}
	return string(out), code
}

func TestWrapped_UnderCapIsByteIdentical(t *testing.T) {
	out, code := runWrapped(t, "printf 'hello\\n'", 1600)
	if out != "hello\n" {
		t.Errorf("output = %q, want %q — under the cap nothing may change", out, "hello\n")
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

func TestWrapped_PropagatesExitCode(t *testing.T) {
	if _, code := runWrapped(t, "exit 7", 1600); code != 7 {
		t.Errorf("exit code = %d, want 7 — agents branch on success", code)
	}
}

func TestWrapped_KeepsBothEndsAndElides(t *testing.T) {
	out, _ := runWrapped(t, "seq 1 500", 200)
	if !strings.HasPrefix(out, "1\n2\n") {
		t.Errorf("head of the output was lost: %.40q", out)
	}
	if !strings.HasSuffix(out, "500\n") {
		t.Errorf("tail of the output was lost: %.40q", out[max(0, len(out)-40):])
	}
	if !strings.Contains(out, "qdf-hook expand x") {
		t.Error("elision line must carry the recovery handle")
	}
	if len(out) > 200+200 {
		t.Errorf("capped output is %d bytes, far over the %d cap", len(out), 200)
	}
}

func TestWrapped_CdSurvives(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	cap := filepath.Join(dir, "x.out")
	wrapped := string(wrapCommand(nil, "cd sub", cap, "x", 1600))
	// The agent's shell is persistent: a cd inside the wrapper must still apply
	// to the next command, which brace grouping guarantees and a subshell breaks.
	c := exec.Command("/bin/sh", "-c", wrapped+"; pwd")
	c.Dir = dir
	out, err := c.Output()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(string(out), "sub") {
		t.Errorf("cd did not survive the wrapper: %q", out)
	}
}

// TestWrapped_NoticeFitsBudget holds the elision line to its budget: it is the
// only text this feature adds to the model's context, and every token of it is
// re-billed on each later turn of the session.
func TestWrapped_NoticeFitsBudget(t *testing.T) {
	out, _ := runWrapped(t, "seq 1 500", 200)
	var notice string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "qdf-hook expand") {
			notice = line
			break
		}
	}
	if notice == "" {
		t.Fatal("no elision line found")
	}
	// 4 bytes per token is the conservative factor used throughout this feature.
	if tokens := len(notice) / 4; tokens > 20 {
		t.Errorf("notice is ~%d tokens, budget is 20: %q", tokens, notice)
	}
}
