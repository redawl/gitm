//go:build !release

package util

var IsRelease = false

// Assert panics if b is false.
// If compile with -tags release, Assert will be compiled out
func Assert(b bool) {
	if !b {
		panic("Assertion failed")
	}
}
