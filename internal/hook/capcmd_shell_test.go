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
	for line := range strings.SplitSeq(out, "\n") {
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

// TestWrapped_TrailingCommentDoesNotBreakSyntax protects against a "#" in the
// agent's own command swallowing the closing brace: `echo hi  # list files` is
// ordinary output, not shell syntax the wrapper is allowed to choke on.
func TestWrapped_TrailingCommentDoesNotBreakSyntax(t *testing.T) {
	out, code := runWrapped(t, "echo hi  # list files", 1600)
	if out != "hi\n" {
		t.Errorf("output = %q, want %q", out, "hi\n")
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

// TestWrapped_UnwritableCapturePathFallsBackUnwrapped protects against a
// missing capture directory silently eating the agent's command: without the
// writability probe, the failing redirect stops the command before it runs at
// all and reports a false failure for a command that was never broken.
func TestWrapped_UnwritableCapturePathFallsBackUnwrapped(t *testing.T) {
	dir := t.TempDir()
	cap := filepath.Join(dir, "nonexistent-subdir", "x.out")
	wrapped := string(wrapCommand(nil, "echo SHOULD-RUN; exit 3", cap, "x", 1600))
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
	if string(out) != "SHOULD-RUN\n" {
		t.Errorf("output = %q, want %q — the command must still run and print", out, "SHOULD-RUN\n")
	}
	if code != 3 {
		t.Errorf("exit code = %d, want 3 — the command's own exit code, not a redirect failure", code)
	}
}

// TestWrapped_NoLingeringBookkeepingVars protects the feature's promise that
// nothing changes except where output goes: __qrc and __qn are bookkeeping
// this wrapper introduces, and the agent's shell persists across calls, so
// they must not still be set once the wrapped command has finished.
func TestWrapped_NoLingeringBookkeepingVars(t *testing.T) {
	dir := t.TempDir()
	cap := filepath.Join(dir, "x.out")
	wrapped := string(wrapCommand(nil, "echo hi", cap, "x", 1600))
	c := exec.Command("/bin/sh", "-c", wrapped+`; printf 'qrc=%s qn=%s\n' "${__qrc:-GONE}" "${__qn:-GONE}"`)
	c.Dir = dir
	out, err := c.Output()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(string(out), "qrc=GONE qn=GONE") {
		t.Errorf("bookkeeping vars leaked into the caller's shell: %q", out)
	}
}
