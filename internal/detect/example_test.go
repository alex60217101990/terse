package detect_test

import (
	"fmt"

	"github.com/alex60217101990/terse/internal/detect"
)

func ExampleIsJSONArray() {
	fmt.Println(detect.IsJSONArray(`[{"id":1,"name":"a"}]`))
	fmt.Println(detect.IsJSONArray("just some log output"))
	// Output:
	// true
	// false
}

func ExampleSqueezeOutput() {
	// Runs of identical consecutive lines collapse to a self-describing marker.
	out := detect.SqueezeOutput("connecting...\nconnecting...\nconnecting...\nconnected\n")
	fmt.Print(out)
	// Output:
	// connecting...  ⨯3
	// connected
}

func ExampleAnalyzeJSONArray() {
	data := []byte(`[{"status":"ok","ms":12},{"status":"ok","ms":18},{"status":"err","ms":9}]`)
	stats, _ := detect.AnalyzeJSONArray(data, 1000)
	fmt.Println("rows:", stats.RowCount)
	fmt.Println("columns:", len(stats.Columns))
	// Output:
	// rows: 3
	// columns: 2
}
