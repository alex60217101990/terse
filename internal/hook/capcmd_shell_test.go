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
	capPath := filepath.Join(dir, "x.out")
	wrapped := string(wrapCommand(nil, cmd, capPath, "x", capBytes))
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
	capPath := filepath.Join(dir, "x.out")
	wrapped := string(wrapCommand(nil, "cd sub", capPath, "x", 1600))
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
	capPath := filepath.Join(dir, "nonexistent-subdir", "x.out")
	wrapped := string(wrapCommand(nil, "echo SHOULD-RUN; exit 3", capPath, "x", 1600))
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
	capPath := filepath.Join(dir, "x.out")
	wrapped := string(wrapCommand(nil, "echo hi", capPath, "x", 1600))
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

// TestWrapped_MatchesPlainRun is the differential the widened skip list rests
// on: for every shape the scanner now accepts, running the command wrapped
// must produce the same bytes, the same exit code and the same effect on the
// caller's shell as running it plainly.
//
// Both sides run in one /bin/sh -c, with the probe appended after the command,
// so a lost cd or a lost variable shows up as a difference in the probe's
// output rather than as a silent pass.
func TestWrapped_MatchesPlainRun(t *testing.T) {
	cases := []struct {
		name  string
		cmd   string
		probe string // runs after cmd, in the same shell
	}{
		{"plain", "echo hello", ""},
		{"pipeline", "seq 1 20 | grep 7", ""},
		{"pipeline exit code", "true | false", ""},
		{"early exit stage", "yes | head -3", ""},
		{"merge redirect", "sh -c 'echo out; echo err 1>&2' 2>&1", ""},
		{"own redirect", "printf 'x\\n' > f; cat f", ""},
		{"append redirect", "echo one >> f; echo two >> f; cat f", ""},
		{"discard stderr", "ls nonexistent 2>/dev/null; echo rc=$?", ""},
		{"stdin redirect", "printf 'b\\na\\n' > in; sort < in", ""},
		{"heredoc", "cat <<'EOF'\nalpha\nbeta\nEOF", ""},
		{"heredoc into pipe", "cat <<'EOF' | sort -r\na\nb\nEOF", ""},
		{"substitution", "echo v=$(printf 42)", ""},
		{"backtick", "echo v=`printf 43`", ""},
		{"assignment survives", "V=$(printf 7)", "echo after=$V"},
		{"export survives", "export QQ=9", "echo qq=$QQ"},
		{"cd survives", "cd /", "pwd"},
		{"trailing comment", "echo hi # a note", ""},
		{"failing command", "sh -c 'exit 3'", "echo rc=$?"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if !cappable(c.cmd) {
				t.Fatalf("the scanner refuses %q, so this case is not exercising the wrapper", c.cmd)
			}
			plain, plainCode := runInShell(t, c.cmd, c.probe, "")
			dir := t.TempDir()
			capPath := filepath.Join(dir, "x.out")
			wrapped := string(wrapCommand(nil, c.cmd, capPath, "x", 1600))
			got, gotCode := runInShell(t, wrapped, c.probe, dir)
			if got != plain {
				t.Errorf("output differs:\n plain %q\n wrapped %q", plain, got)
			}
			if gotCode != plainCode {
				t.Errorf("exit code differs: plain %d, wrapped %d", plainCode, gotCode)
			}
		})
	}
}

// runInShell runs body (plus an optional probe) in one shell and returns its
// merged output and exit code. dir is the working directory; empty means a
// fresh temp dir, so the two sides of a comparison never see each other's files.
func runInShell(t *testing.T, body, probe, dir string) (string, int) {
	t.Helper()
	if dir == "" {
		dir = t.TempDir()
	}
	script := body
	if probe != "" {
		script += "\n" + probe
	}
	c := exec.Command("/bin/sh", "-c", script)
	c.Dir = dir
	out, err := c.CombinedOutput()
	code := 0
	var ee *exec.ExitError
	if err != nil {
		if !errors.As(err, &ee) {
			t.Fatalf("run %q: %v", script, err)
		}
		code = ee.ExitCode()
	}
	return string(out), code
}
