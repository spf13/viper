// +build !go1.7

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

func TestWriteConfigHCLError(t *testing.T) {
	v := New()
	fs := afero.NewMemMapFs()
	v.SetFs(fs)
	v.SetConfigName("c")
	v.SetConfigType("hcl")
	err := v.ReadConfig(bytes.NewBuffer(hclExample))
	if err != nil {
		t.Fatal(err)
	}
	errExpected := "While marshaling config: HCL output unsupported before Go 1.7"
	err = v.WriteConfigAs("c.hcl")
	assert.Equal(t, errExpected, err.Error())
}
