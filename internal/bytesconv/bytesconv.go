// Package bytesconv provides zero-copy conversions between string and []byte.
//
// These avoid the full copy that the built-in string([]byte) / []byte(string)
// conversions make. They are only safe under a strict contract:
//
//   - The result MUST be treated as read-only. Writing through a []byte returned
//     by S2B corrupts the source string (undefined behavior) — strings are
//     immutable in Go and may be interned or shared.
//   - The result MUST NOT outlive the input. It aliases the input's backing
//     memory; once the input is gone the alias dangles.
//
// Use them for hand-off to functions that only READ the bytes (hashing,
// parsing, diffing). When in doubt, use the built-in conversion.
package bytesconv

import "unsafe"

// S2B returns a read-only []byte view of s without copying.
func S2B(s string) []byte {
	if s == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// B2S returns a string view of b without copying. b must not be mutated for the
// lifetime of the returned string.
func B2S(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}
