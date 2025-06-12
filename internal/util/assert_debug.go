//go:build !release

package util

// Assert panics if b is false.
// If compile with -tags release, Assert will be compiled out
func Assert(b bool) {
	if !b {
		panic("Assertion failed")
	}
}
