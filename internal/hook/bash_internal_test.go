package hook

import (
	"strings"
	"testing"
)

// worth gates a summary by ratio: the loose gate keeps a 25–49% win that the
// strict gate would discard (the old single 0.5 gate silently dropped those).
func TestWorth_Gates(t *testing.T) {
	content := strings.Repeat("x", 1000)
	sixtyPct := strings.Repeat("x", 600) // 60% of original -> 40% saved

	if worth(sixtyPct, content, minSummaryRatio) {
		t.Error("strict 0.5 gate must reject a 60% summary")
	}
	if !worth(sixtyPct, content, minSummaryRatioLoose) {
		t.Error("loose 0.75 gate must accept a 60% summary")
	}
	if worth("", content, minSummaryRatioLoose) {
		t.Error("empty summary must never be worth it")
	}
}
