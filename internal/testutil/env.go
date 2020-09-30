package testutil

import (
	"os"
	"testing"
)

// Based on https://github.com/frankban/quicktest/blob/577841610793d24f99e31cc2c0ef3a541fefd7c7/patch.go#L34-L64
// Licensed under the MIT license
// Copyright (c) 2017 Canonical Ltd.

// Setenv sets an environment variable to a temporary value for the
// duration of the test.
//
// At the end of the test (see "Deferred execution" in the package docs), the
// environment variable is returned to its original value.
func Setenv(t *testing.T, name, val string) {
	setenv(t, name, val, true)
}

// Unsetenv unsets an environment variable for the duration of a test.
func Unsetenv(t *testing.T, name string) {
	setenv(t, name, "", false)
}

// setenv sets or unsets an environment variable to a temporary value for the
// duration of the test
func setenv(t *testing.T, name, val string, valOK bool) {
	oldVal, oldOK := os.LookupEnv(name)
	if valOK {
		os.Setenv(name, val)
	} else {
		os.Unsetenv(name)
	}
	t.Cleanup(func() {
		if oldOK {
			os.Setenv(name, oldVal)
		} else {
			os.Unsetenv(name)
		}
	})
}
