// +build !toml

// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package viper

import (
	"testing"
	"time"
)

func initTOML(reset bool) {

	if reset {
		Reset()
	}

	v.config["title"] = "TOML Example"

	dob, _ := time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	v.config["owner"] = map[string]interface{}{
		"organization": "MongoDB",
		"Bio":          "MongoDB Chief Developer Advocate & Hacker at Large",
		"dob":          dob,
	}
}

func assertConfigValue(t *testing.T, v *Viper) {
}
