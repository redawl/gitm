//go:build !release

package util

func Assert(f func() bool) {
	if !f() {
		panic("Assertion failed")
	}
}
