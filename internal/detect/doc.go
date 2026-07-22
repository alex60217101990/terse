// Package detect analyzes raw tool output and turns recognizable shapes into
// compact, information-preserving summaries.
//
// All functions are pure and perform no I/O: each takes the raw output as a
// string and returns either a summary or a signal (empty string / false) that
// the input did not match that shape. This keeps the package table-testable and
// lets the caller in internal/hook decide, per detector, whether a summary is
// worth emitting — the rule everywhere is "never worse": emit the compact form
// only when it is strictly smaller than the original.
//
// The main entry points are:
//
//   - [IsJSONArray] / [AnalyzeJSONArray] — columnar schema + per-column stats
//     for a JSON array of objects, parsed zero-copy.
//   - [IsGoTestOutput] / [SummarizeGoTest] — pass/fail counts plus only the
//     failing cases from `go test -v`.
//   - [IsGitLogOutput] / [SummarizeGitLog] — a compact commit table.
//   - [IsGoBenchOutput] / [SummarizeGoBench] — an aligned benchmark table.
//   - [SqueezeOutput] — ANSI stripping and run-length collapse of repeated
//     lines for output no structural detector matched.
package detect
