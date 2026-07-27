package detect_test

import (
	"os"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
)

// Measured savings on a real ctx_batch_execute payload (an MCP batch that
// re-dumps the same sections under a second query). Asserts a meaningful
// reduction and logs the ratio so regressions in the fold heuristic surface.
func TestFoldRepeatedBlocks_RealBatchPayload(t *testing.T) {
	raw, err := os.ReadFile("testdata/ctx_batch_execute.txt")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	in := string(raw)
	out := detect.FoldRepeatedBlocks(in)

	saved := len(in) - len(out)
	pct := float64(saved) * 100 / float64(len(in))
	t.Logf("ctx_batch_execute: %d -> %d bytes (-%.1f%%)", len(in), len(out), pct)

	if pct < 25 {
		t.Fatalf("expected >=25%% reduction on duplicate-heavy batch, got %.1f%%", pct)
	}
}
