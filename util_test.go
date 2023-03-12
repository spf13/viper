// Copyright © 2016 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Viper is a application configuration system.
// It believes that applications can be configured a variety of ways
// via flags, ENVIRONMENT variables, configuration files retrieved
// from the file system, or a remote key/value store.

package viper

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper/internal/testutil"
)

func TestCopyAndInsensitiviseMap(t *testing.T) {
	var (
		given = map[string]interface{}{
			"Foo": 32,
			"Bar": map[interface{}]interface{}{
				"ABc": "A",
				"cDE": "B",
			},
		}
		expected = map[string]interface{}{
			"foo": 32,
			"bar": map[string]interface{}{
				"abc": "A",
				"cde": "B",
			},
		}
	)

	got := copyAndInsensitiviseMap(given)

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Got %q\nexpected\n%q", got, expected)
	}

	if _, ok := given["foo"]; ok {
		t.Fatal("Input map changed")
	}

	if _, ok := given["bar"]; ok {
		t.Fatal("Input map changed")
	}

	m := given["Bar"].(map[interface{}]interface{})
	if _, ok := m["ABc"]; !ok {
		t.Fatal("Input map changed")
	}
}

func TestAbsPathify(t *testing.T) {
	skipWindows(t)

	home := userHomeDir()
	homer := filepath.Join(home, "homer")
	wd, _ := os.Getwd()

	testutil.Setenv(t, "HOMER_ABSOLUTE_PATH", homer)
	testutil.Setenv(t, "VAR_WITH_RELATIVE_PATH", "relative")

	tests := []struct {
		input  string
		output string
	}{
		{"", wd},
		{"sub", filepath.Join(wd, "sub")},
		{"./", wd},
		{"./sub", filepath.Join(wd, "sub")},
		{"$HOME", home},
		{"$HOME/", home},
		{"$HOME/sub", filepath.Join(home, "sub")},
		{"$HOMER_ABSOLUTE_PATH", homer},
		{"$HOMER_ABSOLUTE_PATH/", homer},
		{"$HOMER_ABSOLUTE_PATH/sub", filepath.Join(homer, "sub")},
		{"$VAR_WITH_RELATIVE_PATH", filepath.Join(wd, "relative")},
		{"$VAR_WITH_RELATIVE_PATH/", filepath.Join(wd, "relative")},
		{"$VAR_WITH_RELATIVE_PATH/sub", filepath.Join(wd, "relative", "sub")},
	}

	for _, test := range tests {
		got := absPathify(jwwLogger{}, test.input)
		if got != test.output {
			t.Errorf("Got %v\nexpected\n%q", got, test.output)
		}
	}
}
