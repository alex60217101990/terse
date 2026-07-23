package main

import "testing"

func TestIsQdfHookCommand(t *testing.T) {
	cases := []struct {
		cmd, sub string
		want     bool
	}{
		{"/Users/x/.local/bin/qdf-hook precompact", "precompact", true},
		{"qdf-hook read", "read", true},
		{"/tmp/qdf-hook.test precompact", "precompact", true},         // go test binary
		{"/opt/homebrew/bin/sqz hook precompact", "precompact", false}, // must NOT match
		{"atuin hook claude-code", "read", false},
		{"qdf-hook read", "bash", false},
		{"qdf-hook", "read", false}, // no subcommand arg
	}
	for _, c := range cases {
		if got := isQdfHookCommand(c.cmd, c.sub); got != c.want {
			t.Errorf("isQdfHookCommand(%q, %q) = %v, want %v", c.cmd, c.sub, got, c.want)
		}
	}
}

func TestMergeHooks_DoesNotFalseMatchOtherTools(t *testing.T) {
	// A PreCompact hook from another tool ending in "precompact" must not block
	// qdf-hook's own precompact from being installed.
	hb := hooksBlock{
		"PreCompact": []hookEntry{
			{Hooks: []hookCmd{{Type: "command", Command: "/opt/homebrew/bin/sqz hook precompact"}}},
		},
	}
	added := mergeHooks(hb, "qdf-hook")
	if added != len(qdfHooks) {
		t.Errorf("expected all %d hooks added, got %d", len(qdfHooks), added)
	}

	var sawSqz, sawQdf bool
	for _, e := range hb["PreCompact"] {
		for _, h := range e.Hooks {
			if h.Command == "/opt/homebrew/bin/sqz hook precompact" {
				sawSqz = true
			}
			if isQdfHookCommand(h.Command, "precompact") {
				sawQdf = true
			}
		}
	}
	if !sawSqz {
		t.Error("sqz precompact hook was dropped")
	}
	if !sawQdf {
		t.Error("qdf-hook precompact was not installed")
	}

	// Idempotency: a second merge adds nothing.
	if again := mergeHooks(hb, "qdf-hook"); again != 0 {
		t.Errorf("second merge should add 0, added %d", again)
	}
}
