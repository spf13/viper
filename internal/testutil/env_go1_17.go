//go:build go1.17
// +build go1.17

package testutil

import (
	"testing"
)

// Setenv sets an environment variable to a temporary value for the
// duration of the test.
//
// This shim can be removed once support for Go <1.17 is dropped.
func Setenv(t *testing.T, name, val string) {
	t.Helper()

	t.Setenv(name, val)
}
