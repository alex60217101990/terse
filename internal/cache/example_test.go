package cache_test

import (
	"fmt"

	"github.com/alex60217101990/terse/internal/cache"
)

func ExampleShortHex() {
	fmt.Println(cache.ShortHex([]byte{0xde, 0xad, 0xbe, 0xef}))
	// Output: deadbeef
}

func ExampleUnifiedDiff() {
	old := []byte("line one\nline two\nline three\n")
	newer := []byte("line one\nline TWO\nline three\n")
	fmt.Print(cache.UnifiedDiff(old, newer, 1))
	// Output:
	// @@ -1,3 +1,3 @@
	//  line one
	// -line two
	// +line TWO
	//  line three
}
