// +build go1.7

// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package viper

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var hclWriteExpected = []byte(`"foos" = {
  "foo" = {
    "key" = 1
  }

  "foo" = {
    "key" = 2
  }

  "foo" = {
    "key" = 3
  }

  "foo" = {
    "key" = 4
  }
}

"id" = "0001"

"name" = "Cake"

"ppu" = 0.55

"type" = "donut"`)

func TestWriteConfigHCL(t *testing.T) {
	v := New()
	fs := afero.NewMemMapFs()
	v.SetFs(fs)
	v.SetConfigName("c")
	v.SetConfigType("hcl")
	err := v.ReadConfig(bytes.NewBuffer(hclExample))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.WriteConfigAs("c.hcl"); err != nil {
		t.Fatal(err)
	}
	read, err := afero.ReadFile(fs, "c.hcl")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, hclWriteExpected, read)
}
