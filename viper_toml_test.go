// +build toml

// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package viper

import (
	"bytes"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var tomlExample = []byte(`
title = "TOML Example"

[owner]
organization = "MongoDB"
Bio = "MongoDB Chief Developer Advocate & Hacker at Large"
dob = 1979-05-27T07:32:00Z # First class dates? Why not?`)

func initTOML(reset bool) {
	if reset {
		Reset()
	}
	SetConfigType("toml")
	r := bytes.NewReader(tomlExample)
	unmarshalReader(r, v.config)
}

func assertConfigValue(t *testing.T, v *Viper) {
	assert.Equal(t, `value is `+path.Base(v.configPaths[0]), v.GetString(`key`))
}
