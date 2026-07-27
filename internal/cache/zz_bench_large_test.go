package cache_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/cache"
)

func BenchmarkUnifiedDiff_LargeFileSmallChange(b *testing.B) {
	var old, newer strings.Builder
	for i := range 4000 {
		fmt.Fprintf(&old, "line-%d\n", i)
		if i == 2000 {
			newer.WriteString("line-CHANGED\n")
		} else {
			fmt.Fprintf(&newer, "line-%d\n", i)
		}
	}
	oldB := []byte(old.String())
	newB := []byte(newer.String())
	for b.Loop() {
		_ = cache.UnifiedDiff(oldB, newB, 3)
	}
}
